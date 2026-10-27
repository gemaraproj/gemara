// SPDX-License-Identifier: Apache-2.0

package schema_test

import (
	"maps"
	"os"
	"slices"
	"testing"

	"cuelang.org/go/cue"
	cueyaml "cuelang.org/go/encoding/yaml"
)

// executionEnvironment mirrors #ExecutionEnvironment for comparison in tests.
type executionEnvironment struct {
	Digests []string `json:"digests"`
}

// actorEnvironment holds the parts of an #Actor that identify where it ran.
type actorEnvironment struct {
	ID          string                `json:"id"`
	Environment *executionEnvironment `json:"execution-environment"`
}

// planParameter mirrors #Parameter. A parameter with accepted values is pinned by the
// policy; one without is left free, as a parameter that differs for each target must be.
type planParameter struct {
	ID             string   `json:"id"`
	AcceptedValues []string `json:"accepted-values"`
}

// acceptedMethod is what a Policy's assessment plan accepts for one evaluation method:
// the executor to run it and the parameters to run it with.
type acceptedMethod struct {
	Executor   actorEnvironment
	Parameters []planParameter
}

// recordedAssessment holds the parts of an #AssessmentLog that a policy's plan is compared with.
type recordedAssessment struct {
	Requirement struct {
		EntryID string `json:"entry-id"`
	} `json:"requirement"`
	Context struct {
		MethodID        string            `json:"method-id"`
		ParameterInputs map[string]string `json:"parameter-inputs"`
	} `json:"execution-context"`
	Executor *actorEnvironment `json:"executor"`
	Evidence []struct {
		ID                 string `json:"id"`
		ObservationVantage string `json:"observation-vantage"`
	} `json:"evidence"`
}

// loadValidated reads a YAML document, validates it against definition and
// returns it as a CUE value.
func loadValidated(t *testing.T, file, definition string) cue.Value {
	t.Helper()
	data, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("read %s: %v", file, err)
	}
	def := schemaValue.LookupPath(cue.ParsePath(definition))
	if err := cueyaml.Validate(data, def); err != nil {
		t.Fatalf("%s does not validate as %s: %v", file, definition, err)
	}
	expr, err := cueyaml.Extract(file, data)
	if err != nil {
		t.Fatalf("extract %s: %v", file, err)
	}
	return schemaCtx.BuildFile(expr)
}

// acceptedFromPlan returns what a Policy's assessment plan accepts for the evaluation
// method with methodID.
func acceptedFromPlan(t *testing.T, policy cue.Value, planID, methodID string) acceptedMethod {
	t.Helper()
	var plans []struct {
		ID      string `json:"id"`
		Methods []struct {
			ID       string           `json:"id"`
			Executor actorEnvironment `json:"executor"`
		} `json:"evaluation-methods"`
		Parameters []planParameter `json:"parameters"`
	}
	path := cue.MakePath(cue.Str("adherence"), cue.Str("assessment-plans"))
	if err := policy.LookupPath(path).Decode(&plans); err != nil {
		t.Fatalf("decode assessment-plans: %v", err)
	}
	for _, plan := range plans {
		for _, method := range plan.Methods {
			if plan.ID == planID && method.ID == methodID {
				return acceptedMethod{Executor: method.Executor, Parameters: plan.Parameters}
			}
		}
	}
	t.Fatalf("no method %s on plan %s", methodID, planID)
	return acceptedMethod{}
}

// logAssessments returns every assessment of a log, and the log's author, which is the
// producing actor of every assessment that names no executor of its own.
func logAssessments(t *testing.T, log cue.Value) ([]recordedAssessment, actorEnvironment) {
	t.Helper()
	var author actorEnvironment
	if err := log.LookupPath(cue.ParsePath("metadata.author")).Decode(&author); err != nil {
		t.Fatalf("decode metadata.author: %v", err)
	}
	var evaluations []struct {
		AssessmentLogs []recordedAssessment `json:"assessment-logs"`
	}
	if err := log.LookupPath(cue.ParsePath("evaluations")).Decode(&evaluations); err != nil {
		t.Fatalf("decode evaluations: %v", err)
	}
	var all []recordedAssessment
	for _, e := range evaluations {
		all = append(all, e.AssessmentLogs...)
	}
	return all, author
}

// environmentMismatches names each field of the recorded environment that differs from the
// accepted one. Because the executor and the producing actor are both #Actor, the comparison
// is field by field on one type.
func environmentMismatches(accepted, recorded actorEnvironment) []string {
	var fields []string
	if accepted.ID != recorded.ID {
		fields = append(fields, "id")
	}
	a, r := accepted.Environment, recorded.Environment
	if a == nil || r == nil {
		if a != r {
			fields = append(fields, "execution-environment")
		}
		return fields
	}
	if !slices.Equal(a.Digests, r.Digests) {
		fields = append(fields, "digests")
	}
	return fields
}

// parameterMismatches names each parameter whose recorded value the plan does not accept:
// a pinned parameter with another value or none, and an input the plan never declared.
// Parameters the plan leaves free (no accepted values) match any value, so the organisation
// or repository a scan was pointed at never counts against it.
func parameterMismatches(params []planParameter, inputs map[string]string) []string {
	declared := map[string]bool{}
	var ids []string
	for _, p := range params {
		declared[p.ID] = true
		if len(p.AcceptedValues) == 0 {
			continue
		}
		if v, ok := inputs[p.ID]; !ok || !slices.Contains(p.AcceptedValues, v) {
			ids = append(ids, p.ID)
		}
	}
	for id := range inputs {
		if !declared[id] {
			ids = append(ids, id)
		}
	}
	slices.Sort(ids)
	return ids
}

// rollUpVantage gives the vantage of an assessment from the evidence it cites: substrate only
// when every piece of evidence is, artifact when any rests on output the resource produced,
// and nothing when the evidence states no vantage.
func rollUpVantage(vantages []string) string {
	out := ""
	for _, v := range vantages {
		switch v {
		case "artifact":
			return "artifact"
		case "substrate":
			out = "substrate"
		}
	}
	return out
}

// checkResult is what comparing one log with its policy's plan reports.
type checkResult struct {
	environment []string
	parameters  []string
	vantage     map[string]string
}

// compareLog compares each assessment of a log with what the plan accepts for the method it
// ran, and reads each assessment's vantage from its own evidence.
func compareLog(t *testing.T, accepted acceptedMethod, log cue.Value) checkResult {
	t.Helper()
	assessments, author := logAssessments(t, log)
	res := checkResult{vantage: map[string]string{}}
	for _, a := range assessments {
		producer := author
		if a.Executor != nil {
			producer = *a.Executor
		}
		res.environment = append(res.environment, environmentMismatches(accepted.Executor, producer)...)
		res.parameters = append(res.parameters, parameterMismatches(accepted.Parameters, a.Context.ParameterInputs)...)
		var vantages []string
		for _, e := range a.Evidence {
			vantages = append(vantages, e.ObservationVantage)
		}
		res.vantage[a.Requirement.EntryID] = rollUpVantage(vantages)
	}
	res.environment = slices.Compact(res.environment)
	res.parameters = slices.Compact(res.parameters)
	return res
}

// TestLogAgainstAcceptedPlan shows the use the execution environment and the parameter inputs
// exist for: a Policy's plan accepts results of one evaluation method only from a pinned
// scanner image run with pinned parameters, and a log records what its assessments ran with.
// Every log validates either way; comparing it with the plan is what tells one produced as
// accepted from one produced with a wider token scope or another scanner image. The
// organisation and repository are free parameters, so a log of another target is not a
// mismatch, which a hash of the whole configuration could not allow.
func TestLogAgainstAcceptedPlan(t *testing.T) {
	policy := loadValidated(t, "./test-data/good-policy-executor-pvtr.yaml", "#Policy")
	accepted := acceptedFromPlan(t, policy, "AP-OSPS-BASELINE", "EV-PVTR-01")

	perCheck := map[string]string{"OSPS-AC-01.01": "substrate", "OSPS-QA-04.01": "artifact"}
	tests := []struct {
		name string
		log  string
		want checkResult
	}{
		{"log produced as the plan accepts", "accepted", checkResult{vantage: perCheck}},
		{"log of another repository under the same pins", "other-target", checkResult{vantage: perCheck}},
		{"log produced with a wider token scope", "parameter-drift", checkResult{parameters: []string{"token-scope"}, vantage: perCheck}},
		{"log produced by another scanner image", "digest-drift", checkResult{environment: []string{"digests"}, vantage: perCheck}},
	}
	inputs := map[string]map[string]string{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log := loadValidated(t, "./test-data/good-evaluation-log-author-environment-"+tt.log+".yaml", "#EvaluationLog")
			got := compareLog(t, accepted, log)
			if !slices.Equal(got.environment, tt.want.environment) {
				t.Errorf("environment mismatches = %v, want %v", got.environment, tt.want.environment)
			}
			if !slices.Equal(got.parameters, tt.want.parameters) {
				t.Errorf("parameter mismatches = %v, want %v", got.parameters, tt.want.parameters)
			}
			if !maps.Equal(got.vantage, tt.want.vantage) {
				t.Errorf("vantage per check = %v, want %v", got.vantage, tt.want.vantage)
			}
			assessments, _ := logAssessments(t, log)
			inputs[tt.log] = assessments[0].Context.ParameterInputs
		})
	}

	// The two accepted logs ran against different repositories, so a hash over their whole
	// configuration would differ although both ran as the plan accepts.
	if maps.Equal(inputs["accepted"], inputs["other-target"]) {
		t.Errorf("accepted and other-target record the same parameter inputs: %v", inputs["accepted"])
	}
}

// TestMixedEvidenceVantageIsArtifact pins the roll-up: an assessment that draws on evidence
// of both vantages rests on output the resource could influence.
func TestMixedEvidenceVantageIsArtifact(t *testing.T) {
	for _, tc := range []struct {
		name     string
		vantages []string
		want     string
	}{
		{"no evidence", nil, ""},
		{"evidence that states no vantage", []string{""}, ""},
		{"substrate only", []string{"substrate", "substrate"}, "substrate"},
		{"artifact only", []string{"artifact"}, "artifact"},
		{"both", []string{"substrate", "artifact"}, "artifact"},
		{"both, artifact first", []string{"artifact", "substrate"}, "artifact"},
	} {
		if got := rollUpVantage(tc.vantages); got != tc.want {
			t.Errorf("%s: rollUpVantage(%v) = %q, want %q", tc.name, tc.vantages, got, tc.want)
		}
	}
}
