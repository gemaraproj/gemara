all: tidy test testcov cuefmtcheck lintcue cuegen dirtycheck lintinsights

tidy:
	@echo "  >  Tidying go.mod ..."
	@go mod tidy
	@echo "  >  Tidying cue.mod ..."
	@cd schemas && cue mod tidy

test:
	@echo "  >  Running tests ..."
	@go vet ./...
	@go test ./...

testcov:
	@echo "Running tests and generating coverage output ..."
	@go test ./... -coverprofile coverage.out -covermode count
	@sleep 2 # Sleeping to allow for coverage.out file to get generated
	@echo "Current test coverage : $(shell go tool cover -func=coverage.out | grep total | grep -Eo '[0-9]+\.[0-9]+') %"

# Verify CUE formatting in ./schemas
cuefmtcheck:
	@echo "  >  Verifying CUE formatting in ./schemas ..."
	@cue fmt --check --files ./schemas

lint:
	@echo "  >  Linting Go files ..."
	@golangci-lint run

lintcue:
	@echo "  >  Linting CUE files (with module support) ..."
	@cd schemas && cue eval . --all-errors --verbose

cuegen:
	@echo "  >  Generating types from cue schema ..."
	@cd schemas && cue exp gengotypes .
	@mv schemas/cue_types_gen.go generated_types.go
	@go build -o cmd/types_tagger/types_tagger cmd/types_tagger/main.go
	@cmd/types_tagger/types_tagger generated_types.go
	@rm cmd/types_tagger/types_tagger

dirtycheck:
	@echo "  >  Checking for uncommitted changes ..."
	@if [ -n "$$(git status --porcelain)" ]; then \
		echo "  >  Uncommitted changes to generated files found!"; \
		echo "  >  Run make cuegen and commit the results."; \
		exit 1; \
	else \
		echo "  >  No uncommitted changes to generated files found."; \
	fi

oscalgenerate:
	@echo "  >  Generating OSCAL testdata from Gemara artifacts..."
	@mkdir -p artifacts
	@go run ./cmd/oscal_export catalog ./test-data/good-osps.yml --output ./artifacts/catalog.json
	@go run ./cmd/oscal_export guidance ./test-data/good-guidance.yaml --catalog-output ./artifacts/guidance.json --profile-output ./artifacts/profile.json

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

.PHONY: tidy test testcov lintcue cuegen dirtycheck lintinsights serve build clean stop restart check-jekyll
