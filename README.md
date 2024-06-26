# Setlist to Playlist

Cool CLI that creates a Spotify playlist based on a Setlist.fm entry, written in Go.

[![Continuous Integration](https://github.com/mathcale/setlist-to-playlist/actions/workflows/ci.yaml/badge.svg)](https://github.com/mathcale/setlist-to-playlist/actions/workflows/ci.yaml)

## Usage

Just run the following command, replacing the URL with the one you want to create a playlist from:

```sh
setlist-to-playlist --url https://www.setlist.fm/setlist/blink182/2024/autodromo-de-interlagos-sao-paulo-brazil-53aa1325.html
```

## Installation

### Step 1: downloading the binary

1. Download the latest release for your OS from the [releases page](https://github.com/mathcale/setlist-to-playlist/releases/latest)
2. Extract the tarball to a folder in your PATH (like `$HOME/.local/bin` or `/usr/local/bin`)
3. Run `setlist-to-playlist --help` to check if the installation was successful

### Step 2: Generating the Spotify API credentials

1. Go to the [Spotify Developer Dashboard](https://developer.spotify.com/dashboard/applications)
2. Click on "Create app" and fill in the required fields, but pay attention to the following:
   - The `Redirect URIs` field **MUST** be `http://localhost:8080/callback`
   - The `Web API` checkbox **MUST** be checked
3. Copy the `Client ID` and `Client Secret` to a safe place

### Step 3: Generating the Setlist.fm API key

1. Go to the [Setlist.fm API page](https://www.setlist.fm/settings/apps) and fill in the required fields
2. Copy the generated `API Key` to a safe place

### Step 4: Configuring the CLI

On the first run, the CLI will ask for the credentials you generated in the previous steps. Just copy and paste them when prompted and the CLI will store them in a TOML configuration file in your home directory (`$HOME/.config/setlist-to-playlist` on Linux and `$HOME/Library/Application Support/setlist-to-playlist` on macOS).

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
