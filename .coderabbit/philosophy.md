<!-- SPDX-License-Identifier: Apache-2.0 -->
# Gemara design principles for reviewers

Distilled from the Gemara whitepaper
(https://openssf.org/wp-content/uploads/2026/03/04_OpenSSF_Gemara-Project_Whitepaper.pdf)
and the maintainers' schema conventions. Use these as review criteria for changes to the
`.cue` schema.

## What the schema is

Gemara is a seven-layer model for automated risk assessment. Each layer builds on the
one below it, so a layer may reference lower layers but never the reverse.

| Layer | Subject | Output |
| --- | --- | --- |
| 1 | Vectors and Guidance | document |
| 2 | Threats and Controls | document |
| 3 | Risks and Policy | document |
| 4 | Sensitive Activities (the pivot) | none; implementation, not schema |
| 5 | Intent and Behavior Evaluation | timestamped log |
| 6 | Preventive and Remediative Enforcement | timestamped log |
| 7 | Audit and Continuous Monitoring | timestamped log |

- Definition layers (1-3) produce documents. Measurement layers (5-7) produce
  timestamped logs. A change that blurs this split needs a reason.
- Layer 4 has no schema. Do not add one.
- Use the whitepaper's terms with their defined meanings: Control, Threat, Vector,
  Guidance, Policy, Assessment Requirement, Evaluation, Enforcement, Audit. Flag a
  field or doc comment that uses one loosely (for example a Policy that is really
  tool configuration).

## Principles to enforce

1. **Intent, not derivable data.** The schema records intent. If a value can be
   computed from other fields, from the document's position in the model, or from the
   enclosing document, do not add a field for it. Leaner is better.
2. **Derive identity, never mint it.** No UUID or GUID fields. Identity comes from
   existing fields, such as an id scoped by its enclosing document.
3. **No property bags.** Reject free-form maps, `{...}` open structs, `[string]: _`
   and `_` fields used as extension points. The whitepaper names customizable fields
   as a failure of earlier GRC schemas, because every tool then needs foreknowledge
   of the custom fields.
4. **Close enums.** Prefer a closed set of values. If a value set must stay open,
   the PR must say why a closed set would not work.
5. **Opinionated over flexible.** The schema should be strict enough for an
   automation ecosystem to form around it. When in doubt, do not add flexibility.
6. **A new field needs a use case.** Ask who reads it and what they do with it. Fields
   that only communicate (draft, version, type, state) are fine; rules such as
   boundaries and pointers need a consumer.
7. **Keep CUE for what CUE can express.** Cross-document and procedural rules belong
   in the tool-developer spec on the website, not in the schema.

## Compatibility

- Renaming a field, removing a field, or adding a required field is a breaking
  change.
- Tightening a constraint is allowed within v1. Adding an optional field is not
  breaking either. A stable replacement is added alongside the old field, which is
  deprecated.
- Experimental definitions can change in a minor release.
- If the existing doc comment already states the rule a change enforces, the change
  is a patch, not a break.
