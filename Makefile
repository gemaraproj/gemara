all: tidy cuefmtcheck lintcue lintinsights

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


# Verify CUE formatting
cuefmtcheck:
	@echo "  >  Verifying CUE formatting ..."
	@cue fmt --check --files .

lintcue:
	@echo "  >  Linting CUE files (with module support) ..."
	@cue eval . --all-errors --verbose

lintinsights:
	@echo "  >  Linting security-insights.yml ..."
	@curl -O --silent https://raw.githubusercontent.com/ossf/security-insights-spec/refs/tags/v2.1.0/schema.cue
	@cue vet -d '#SecurityInsights' security-insights.yml schema.cue
	@rm schema.cue
	@echo "  >  Linting security-insights.yml complete."

# Documentation site targets
check-jekyll:
	@if ! command -v jekyll >/dev/null 2>&1; then \
		echo "ERROR: Jekyll not found."; \
		echo "  >  Install Jekyll: gem install jekyll bundler && cd docs && bundle install"; \
		exit 1; \
	fi

serve: check-jekyll
	@echo "  >  Starting Jekyll documentation site..."
	@echo "  >  Site will be available at: http://localhost:4000/gemara"
	@echo ""
	@cd docs && bundle exec jekyll serve --host 0.0.0.0 --livereload

build: check-jekyll
	@echo "  >  Building Jekyll documentation site..."
	@cd docs && bundle exec jekyll build

test-links:
	@echo "  >  Validating links with html-proofer..."
	@cd docs && bundle exec htmlproofer _site \
		--allow-hash-href \
		--disable-external \
		--ignore-empty-alt \
		--only-4xx \
		--root-dir "$$(pwd)/_site" \
		--ignore-urls "/localhost/,/0.0.0.0/,/127.0.0.1/,/\/pages\//"

clean:
	@echo "  >  Cleaning generated files..."
	@rm -rf generated docs/_site docs/.jekyll-cache docs/.jekyll-metadata
	@echo "  >  Clean complete!"

stop:
	@echo "  >  Use Ctrl+C to stop the Jekyll server if it's running."

restart: stop serve

GENERATED_DIR := generated
OPENAPI_YAML := $(GENERATED_DIR)/openapi.yaml
MANIFEST_JSON := $(GENERATED_DIR)/schema-manifest.json
SPEC_DIR := $(GENERATED_DIR)/spec
DOCS_SCHEMA_DIR := docs/schema

genopenapi:
	@echo "  >  Converting CUE schema to OpenAPI ..."
	@mkdir -p $(GENERATED_DIR)
	@cd cmd/cue2openapi && go run . -schema ../.. -output ../../$(OPENAPI_YAML) -manifest ../../$(MANIFEST_JSON)
	@echo "  >  OpenAPI schema generation complete!"

genmd: genopenapi
	@echo "  >  Generating markdown from OpenAPI ..."
	@mkdir -p $(SPEC_DIR)
	@cd cmd/openapi2md && go run . -input ../../$(OPENAPI_YAML) -output ../../$(SPEC_DIR) -manifest ../../$(MANIFEST_JSON)
	@echo "  >  Markdown generation complete!"

gendocs: genmd
	@echo "  >  Copying schema pages to $(DOCS_SCHEMA_DIR)/ for website ..."
	@mkdir -p $(DOCS_SCHEMA_DIR)
	@for f in $(SPEC_DIR)/*.md; do \
		[ -f "$$f" ] || continue; \
		base=$$(basename "$$f" .md); \
		title=$$(sh "$(CURDIR)/cmd/scripts/schema-display-name.sh" "$$base"); \
		{ \
			echo "---"; \
			echo "layout: page"; \
			echo "title: $$title"; \
			echo "---"; \
			echo ""; \
			cat "$$f"; \
		} > "$(DOCS_SCHEMA_DIR)/$$base.md"; \
	done
	@echo "  >  Generating $(DOCS_SCHEMA_DIR)/index.md ..."
	@{ \
		echo "---"; \
		echo "layout: page"; \
		echo "title: Schema"; \
		echo "nav-title: Schema"; \
		echo "---"; \
		echo ""; \
		echo "Schema documentation generated from CUE. One page per schema file."; \
		echo ""; \
		for f in $(SPEC_DIR)/base.md $(SPEC_DIR)/metadata.md $(SPEC_DIR)/mapping.md $(SPEC_DIR)/layer-1.md $(SPEC_DIR)/layer-2.md $(SPEC_DIR)/layer-3.md $(SPEC_DIR)/layer-5.md; do \
			[ -f "$$f" ] || continue; \
			base=$$(basename "$$f" .md); \
			title=$$(sh "$(CURDIR)/cmd/scripts/schema-display-name.sh" "$$base"); \
			echo "- [$$title]($$base.html)"; \
		done; \
	} > "$(DOCS_SCHEMA_DIR)/index.md"
	@echo "  >  Documentation generation complete!"

.PHONY: tidy tidycheck cuefmtcheck lintcue lintinsights serve build test-links clean stop restart check-jekyll genopenapi genmd gendocs
