# SPDX-License-Identifier: Apache-2.0

all: tidy cuefmtcheck lintcue lintinsights test

#
# SCHEMA VALIDATION TESTS
#

# -count=1: the tests read the *.cue schemas at runtime, which Go's test cache
# cannot track, so cached passes can go stale against edited schemas.
test:
	@echo "  >  Running schema validation tests ..."
	@cd test && go test -v -count=1 ./...
	@cd cmd && go test -v -count=1 ./...
	@echo "  >  Schema validation tests complete."


#
# CUE DEVELOPMENT TOOLS
#

tidy:
	@echo "  >  Tidying cue.mod ..."
	@cue mod tidy

tidycheck: tidy
	@echo "  >  Checking CUE module tidiness ..."
	@if [ -n "$$(git status --porcelain cue.mod 2>/dev/null)" ]; then \
		echo "Error: cue.mod is not tidy. Please run 'make tidy' and commit the changes."; \
		git diff cue.mod; \
		exit 1; \
	fi
	@echo "  >  CUE module is tidy."

cuefmtcheck:
	@echo "  >  Verifying CUE formatting ..."
	@cue fmt --check --files .

lintcue:
	@echo "  >  Linting CUE files (with module support) ..."
	@cue eval . --all-errors --verbose

#
# SECURITY INSIGHTS VALIDATION
#
lintinsights:
	@echo "  >  Linting security-insights.yml ..."
	@curl -O --silent https://raw.githubusercontent.com/ossf/security-insights-spec/refs/tags/v2.1.0/schema.cue
	@cue vet -d '#SecurityInsights' security-insights.yml schema.cue
	@rm schema.cue
	@echo "  >  Linting security-insights.yml complete."

#
# GENERATE OPENAPI FROM SCHEMAS
#
# The website (github.com/gemaraproj/website) generates its reference pages
# by running the cmd/ CLI against a checkout of this repo; downstream type
# generation (e.g. gemara-react) consumes generated/openapi.yaml.
#
# Embedded CUE definitions are emitted as `allOf` against the embedded
# schema, so a consumer enumerating a schema's fields must resolve `allOf`.
# Field descriptions on `$ref` properties use the `{description, allOf:
# [{$ref}]}` form, which is how OpenAPI 3.0 permits both.
#

GENERATED_DIR := generated
OPENAPI_YAML := $(GENERATED_DIR)/openapi.yaml
MANIFEST_JSON := $(GENERATED_DIR)/schema-manifest.json

genopenapi:
	@echo "  >  Converting CUE schema to OpenAPI ..."
	@mkdir -p $(GENERATED_DIR)
	@cd cmd && go run . cue2openapi --schema .. --output ../$(OPENAPI_YAML) --manifest ../$(MANIFEST_JSON)
	@echo "  >  OpenAPI schema generation complete!"

#
# TEST BACKWARD COMPATIBILITY
#
# Repo whose releases hold the v1 baseline OpenAPI asset.
REPO ?= gemaraproj/gemara

breaking-check:
	@echo "  >  Checking for breaking schema changes ..."
	@filter='.[] | select(.prerelease == false) | select(.tag_name | test("^v1\\.")) | select(any(.assets[]?; .name == "openapi.yaml")) | .tag_name'; \
	tags=$$(gh api "repos/$(REPO)/releases" --paginate --jq "$$filter") \
		|| { echo "  >  ERROR: failed to query releases from $(REPO); aborting breaking-change check" >&2; exit 1; }; \
	tag=$$(printf '%s\n' "$$tags" | sort -V | tail -n1); \
	if [ -z "$$tag" ]; then \
		echo "  >  No v1 baseline release with an openapi.yaml asset found; skipping breaking-change check"; \
		exit 0; \
	fi; \
	echo "  >  Comparing against v1 baseline $$tag"; \
	tmp=$$(mktemp -d); \
	trap 'rm -rf "$$tmp"' EXIT; \
	gh release download "$$tag" --repo "$(REPO)" --pattern openapi.yaml --dir "$$tmp" \
		|| { echo "  >  ERROR: failed to download baseline asset for $$tag" >&2; exit 1; }; \
	allow=""; \
	if [ -f .oasdiff-allow ]; then allow="--allow ../.oasdiff-allow"; fi; \
	( cd cmd && go run . breaking-check --schema .. --base "$$tmp/openapi.yaml" $$allow )
	@echo "  >  Backward compatibility check complete."

#
# REMOVE GENERATED FILES
#

cleanup:
	@echo "  >  Removing generated files..."
	@rm -rf $(GENERATED_DIR)
	@echo "  >  Cleanup complete!"

.PHONY: tidy tidycheck cuefmtcheck lintcue lintinsights test breaking-check cleanup genopenapi
