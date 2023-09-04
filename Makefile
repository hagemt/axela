GO ?= $(shell command -v go)

app: sane
	$(GO) install ./... && ./scripts/env.sh "$(shell $(GO) env GOPATH)/bin/axela"
.PHONY: app

check: lint
	make test
.PHONY: check

lint: sane
	$(GO) fmt ./...
	golangci-lint run ./...
	gosec -tests ./{brains,cmd,say}/...
	for SCRIPT in ./scripts/* ; do shellcheck $$SCRIPT ; done
.PHONY: lint

sane: whisper.cpp
	@[ -x '$(GO)' ] || echo '--- missing Go: brew install golang # macOS'
	@echo '--- using GO=$(GO) and loading Textsynth or OpenAI (GPT-4)'
.PHONY: sane

test: sane
	$(GO) test -v ./...
.PHONY: test

tidy: sane
	$(GO) mod tidy
.PHONY: tidy

whisper.cpp: .gitmodules
	git submodule update --init
	./scripts/download_models.sh
	./scripts/env_auth_models.sh
