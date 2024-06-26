# Setlist to Playlist

Cool CLI that creates a Spotify playlist based on a Setlist.fm entry, written in Go.

[![Continuous Integration](https://github.com/mathcale/setlist-to-playlist/actions/workflows/ci.yaml/badge.svg)](https://github.com/mathcale/setlist-to-playlist/actions/workflows/ci.yaml)

## Usage

Just run the following command, replacing the URL with the one you want to create a playlist from:

```sh
setlist-to-playlist --url https://www.setlist.fm/setlist/blink182/2024/autodromo-de-interlagos-sao-paulo-brazil-53aa1325.html
```

## Installation

TBA

## Architecture

TBA

## Development

### Prerequisites

- Go 1.22.3 (or later)
- GNU Make

### Setup

1. Clone the repository
2. Run `make tidy` to install dependencies

### Running

1. Just execute `go run ./cmd/cli/main.go --url SOME_URL_HERE` to run the CLI

### Testing

```sh
make test
```

### Building

```sh
make build
```

### Reset config and authentication state

```sh
make clean-all
```

## License

[GNU GPLv3](LICENSE)
