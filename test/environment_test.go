// SPDX-License-Identifier: Apache-2.0

package schema_test

import (
	"os"
	"slices"
	"testing"

	"cuelang.org/go/cue"
	cueyaml "cuelang.org/go/encoding/yaml"
)

// executionEnvironment mirrors #ExecutionEnvironment for comparison in tests.
type executionEnvironment struct {
	Digests            []string `json:"digests"`
	ConfigDigest       string   `json:"config-digest"`
	ObservationVantage string   `json:"observation-vantage"`
}

// actorEnvironment holds the parts of an #Actor that identify where it ran.
type actorEnvironment struct {
	ID          string                `json:"id"`
	Environment *executionEnvironment `json:"execution-environment"`
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

// acceptedExecutor returns the executor a Policy names on the evaluation method with methodID.
func acceptedExecutor(t *testing.T, policy cue.Value, methodID string) actorEnvironment {
	t.Helper()
	methods := policy.LookupPath(cue.MakePath(cue.Str("adherence"), cue.Str("evaluation-methods")))
	iter, err := methods.List()
	if err != nil {
		t.Fatalf("evaluation-methods: %v", err)
	}
	for iter.Next() {
		var method struct {
			ID       string           `json:"id"`
			Executor actorEnvironment `json:"executor"`
		}
		if err := iter.Value().Decode(&method); err != nil {
			t.Fatalf("decode evaluation method: %v", err)
		}
		if method.ID == methodID {
			return method.Executor
		}
	}
	t.Fatalf("no evaluation method %s", methodID)
	return actorEnvironment{}
}

// logAuthor returns the author of a log.
func logAuthor(t *testing.T, log cue.Value) actorEnvironment {
	t.Helper()
	var author actorEnvironment
	if err := log.LookupPath(cue.ParsePath("metadata.author")).Decode(&author); err != nil {
		t.Fatalf("decode metadata.author: %v", err)
	}
	return author
}

// environmentMismatches names each field of the recorded environment that differs from the
// accepted one. Because the executor and the log author are both #Actor, the comparison is
// field by field on one type.
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
	if a.ConfigDigest != r.ConfigDigest {
		fields = append(fields, "config-digest")
	}
	if a.ObservationVantage != r.ObservationVantage {
		fields = append(fields, "observation-vantage")
	}
	return fields
}

// TestLogAuthorAgainstAcceptedExecutor shows the use the execution environment exists for: a
// Policy accepts results of one evaluation method only from an executor running a pinned
// scanner with a pinned configuration, and a log records the environment its author ran in.
// Both documents validate either way; comparing the two environments is what tells a log
// produced as accepted from one produced with a configuration the policy never accepted.
func TestLogAuthorAgainstAcceptedExecutor(t *testing.T) {
	policy := loadValidated(t, "./test-data/good-policy-executor-pvtr.yaml", "#Policy")
	accepted := acceptedExecutor(t, policy, "EV-PVTR-01")

	tests := []struct {
		name string
		log  string
		want []string
	}{
		{"log produced in the accepted environment", "./test-data/good-evaluation-log-author-environment-accepted.yaml", nil},
		{"log produced with a different scanner configuration", "./test-data/good-evaluation-log-author-environment-config-drift.yaml", []string{"config-digest"}},
		{"log whose author records no configuration", "./test-data/good-evaluation-log-author-environment.yaml", []string{"config-digest"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log := loadValidated(t, tt.log, "#EvaluationLog")
			got := environmentMismatches(accepted, logAuthor(t, log))
			if !slices.Equal(got, tt.want) {
				t.Errorf("mismatched fields = %v, want %v", got, tt.want)
			}
		})
	}
}
