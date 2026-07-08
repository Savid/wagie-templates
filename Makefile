.DEFAULT_GOAL := help
.PHONY: test validate check-runbooks tidy clean help

## test: validate templates against wagie core
test:
	go test -shuffle=on ./...

## validate: report per-file validation (optional FILTER=<path substring>)
validate:
	go run ./cmd/validate $(FILTER)

## check-runbooks: verify template retrieval phrases rank the intended panda runbook first
check-runbooks:
	./scripts/check-runbooks.sh

## tidy: tidy go modules
tidy:
	go mod tidy

## clean: remove Go test cache for this module
clean:
	go clean -testcache

## help: show this help
help:
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'
