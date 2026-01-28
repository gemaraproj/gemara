all: tidy cuefmtcheck lintcue lintinsights gendocs test-links cleanup


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
# WEBSITE DEVELOPMENT TOOLS
#
check-jekyll:
	@if ! command -v jekyll >/dev/null 2>&1; then \
		echo "ERROR: Jekyll not found."; \
		echo "  >  Install Jekyll: gem install jekyll bundler && cd docs && bundle install"; \
		exit 1; \
	fi

serve: check-jekyll gendocs
	@echo "  >  Starting Jekyll documentation site..."
	@echo "  >  Site will be available at: http://localhost:4000/gemara"
	@echo ""
	@cd docs && bundle exec jekyll serve --host 0.0.0.0 --livereload

build: check-jekyll gendocs
	@echo "  >  Building Jekyll documentation site..."
	@cd docs && bundle exec jekyll build

stop:
	@echo "  >  Use Ctrl+C to stop the Jekyll server if it's running."

restart: stop serve

#
# GENERATE WEBSITE DOCUMENTATION FROM SCHEMAS & LEXICON
#

GENERATED_DIR := generated
OPENAPI_YAML := $(GENERATED_DIR)/openapi.yaml
MANIFEST_JSON := $(GENERATED_DIR)/schema-manifest.json
SPEC_DIR := $(GENERATED_DIR)/spec
DOCS_SCHEMA_DIR := docs/schema
SCHEMA_NAV := docs/schema-nav.yml

genopenapi:
	@echo "  >  Converting CUE schema to OpenAPI ..."
	@mkdir -p $(GENERATED_DIR)
	@cd cmd/cue2openapi && go run . -schema ../.. -output ../../$(OPENAPI_YAML) -manifest ../../$(MANIFEST_JSON)
	@echo "  >  OpenAPI schema generation complete!"

genmd: genopenapi
	@echo "  >  Generating markdown from OpenAPI ..."
	@mkdir -p $(SPEC_DIR)
	@cd cmd/openapi2md && go run . -input ../../$(OPENAPI_YAML) -output ../../$(SPEC_DIR) -nav ../../$(SCHEMA_NAV)
	@echo "  >  Markdown generation complete!"

gendocs: genmd
	@echo "  >  Copying schema pages to $(DOCS_SCHEMA_DIR)/ for website ..."
	@mkdir -p $(DOCS_SCHEMA_DIR)
	@sh "$(CURDIR)/cmd/scripts/parse-nav.sh" "$(SCHEMA_NAV)" list-pages | while IFS='|' read -r filename title; do \
		if [ -f "$(SPEC_DIR)/$$filename.md" ]; then \
			{ \
				echo "---"; \
				echo "layout: page"; \
				echo "title: $$title"; \
				echo "---"; \
				echo ""; \
				cat "$(SPEC_DIR)/$$filename.md"; \
			} > "$(DOCS_SCHEMA_DIR)/$$filename.md"; \
		fi; \
	done
	@echo "  >  Updating schema list in $(DOCS_SCHEMA_DIR)/index.md ..."
	@if [ -f "$(DOCS_SCHEMA_DIR)/index.md" ]; then \
		schema_list_file="$(DOCS_SCHEMA_DIR)/index.md.schema_list.tmp"; \
		sh "$(CURDIR)/cmd/scripts/parse-nav.sh" "$(SCHEMA_NAV)" list-pages | while IFS='|' read -r filename title; do \
			[ -f "$(DOCS_SCHEMA_DIR)/$$filename.md" ] && echo "- [$$title]($$filename.html)"; \
		done > "$$schema_list_file"; \
		awk -v list_file="$$schema_list_file" ' \
			BEGIN { \
				while ((getline line < list_file) > 0) { \
					schema_list = schema_list line "\n"; \
				} \
				close(list_file); \
			} \
			/<!-- SCHEMA_LIST_START -->/ { \
				print; \
				print ""; \
				printf "%s", schema_list; \
				print ""; \
				skip=1; \
				next \
			} \
			/<!-- SCHEMA_LIST_END -->/ { print; skip=0; next } \
			skip==0 { print } \
		' "$(DOCS_SCHEMA_DIR)/index.md" > "$(DOCS_SCHEMA_DIR)/index.md.tmp" && \
		rm -f "$$schema_list_file" && \
		mv "$(DOCS_SCHEMA_DIR)/index.md.tmp" "$(DOCS_SCHEMA_DIR)/index.md"; \
	fi
	@echo "  >  Generating definitions table from lexicon ..."
	@if [ -f "docs/model/02-definitions.md.template" ]; then \
		cp "docs/model/02-definitions.md.template" "docs/model/02-definitions.md"; \
	fi
	@cd cmd/lexicon2md && go run . -lexicon ../../docs/lexicon.yaml -output ../../docs/model/02-definitions.md
	@echo "  >  Linking defined terms across documentation ..."
	@cd cmd/termlinker && go run . -lexicon ../../docs/lexicon.yaml -docs ../../docs
	@echo "  >  Documentation generation complete!"

#
# TEST GENERATED DOCUMENTATION
#

test-links:
	@echo "  >  Validating all site pages and links with html-proofer..."
	@cd docs && bundle exec htmlproofer _site \
		--allow-hash-href \
		--disable-external \
		--ignore-empty-alt \
		--only-4xx \
		--ignore-files '/model\/02-definitions\.html/' \
		--root-dir "$$(pwd)/_site" \

#
# REMOVE GENERATED DOCUMENTATION
#

clean-jekyll:
	@echo "  >  Cleaning jekyll build artifacts..."
	@rm -rf generated docs/_site docs/.jekyll-cache docs/.jekyll-metadata
	@echo "  >  Clean jekyll build artifacts complete!"

cleanup-links:
	@echo "  >  Removing termlinker-generated links from documentation ..."
	@cd cmd/termlinker && go run . -lexicon ../../docs/lexicon.yaml -docs ../../docs -cleanup
	@echo "  >  Link cleanup complete!"

cleanup: clean-jekyll cleanup-links
	@echo "  >  Removing generated documentation files and links..."
	@sh "$(CURDIR)/cmd/scripts/parse-nav.sh" "$(SCHEMA_NAV)" list-pages | while IFS='|' read -r filename title; do \
		rm -f "$(DOCS_SCHEMA_DIR)/$$filename.md"; \
	done
	@rm -f docs/model/02-definitions.md
	@git checkout -- docs/schema/index.md 2>/dev/null || true
	@rm -rf docs/_site docs/.jekyll-cache docs/.jekyll-metadata
	@echo "  >  Cleanup complete!"

.PHONY: tidy tidycheck cuefmtcheck lintcue lintinsights serve build test-links html-proofer clean cleanup cleanup-links stop restart check-jekyll genopenapi genmd gendocs
