.PHONY: build test tidy clean-tokens clean-config clean-all install

build:
	@go build -ldflags="-w -s -buildid=" -trimpath -o ./bin/setlist-to-playlist ./cmd/cli/main.go

test:
	@./scripts/test.sh

tidy:
	@go mod tidy

clean-tokens:
	rm -rf ~/.config/setlist-to-playlist/spotify_auth.json

clean-config:
	rm -rf ~/.config/setlist-to-playlist/config.toml

clean-all: clean-tokens clean-config

install: build
	@cp ./bin/setlist-to-playlist ~/.local/bin
