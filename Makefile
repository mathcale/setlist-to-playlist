.PHONY: build test tidy

build:
	@go build -o ./bin/cli ./cmd/cli/main.go

test:
	@./scripts/test.sh

tidy:
	@go mod tidy

clean-tokens:
	rm -rf ~/.config/setlist-to-playlist/spotify_auth.json
