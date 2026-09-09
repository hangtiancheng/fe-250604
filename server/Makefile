.PHONY: dev swifty build

dev: ## Hot-reload the chat server with air
	air -c .air.toml

swifty: ## Run the swifty TUI
	go run ./cmd/swifty

build: ## Build both binaries into ./tmp
	go build -o ./tmp/main ./cmd
	go build -o ./tmp/swifty ./cmd/swifty

sync:
	rm -rf ~/.swifty
	cp -r ~/.swifty ~/.swifty

remote:
	node ./internal/swifty/remote/fe/index.js
