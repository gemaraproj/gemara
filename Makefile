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

clean:
	@echo "  >  Cleaning generated files..."
	@rm -rf docs/_site docs/.jekyll-cache docs/.jekyll-metadata
	@echo "  >  Clean complete!"

stop:
	@echo "  >  Use Ctrl+C to stop the Jekyll server if it's running."

restart: stop serve

.PHONY: tidy tidycheck cuefmtcheck lintcue lintinsights serve build clean stop restart check-jekyll
