COMMIT := $(shell git rev-list --abbrev-commit -1 HEAD)

native:
	go build -ldflags "-s -w -X main.gitCommit=$(COMMIT)"

linux:
	GOOS=linux go build -ldflags "-s -w -X main.gitCommit=$(COMMIT)"

macosx:
	GOOS=darwin go build -ldflags "-s -w -X main.gitCommit=$(COMMIT)"

windows:
	GOOS=windows go build -ldflags "-s -w -X main.gitCommit=$(COMMIT)"

install:
	go install -ldflags "-s -w -X main.gitCommit=$(COMMIT)"

lint:
	golangci-lint run ./...

sec:
	grype . --add-cpes-if-none
	trivy fs .

.PHONY: native linux macosx windows install
