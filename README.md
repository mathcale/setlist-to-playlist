# Setlist to Playlist

Cool CLI that creates a Spotify or Apple Music playlist based on a Setlist.fm entry, written in Go.

[![Continuous Integration](https://github.com/mathcale/setlist-to-playlist/actions/workflows/ci.yaml/badge.svg)](https://github.com/mathcale/setlist-to-playlist/actions/workflows/ci.yaml)

## Usage

Just run the following command, replacing the URL with the one you want to create a playlist from:

```sh
setlist-to-playlist --url https://www.setlist.fm/setlist/blink182/2024/autodromo-de-interlagos-sao-paulo-brazil-53aa1325.html
```

By default the playlist is created on Spotify. Use `--provider` (or `-p`) to pick the service, either `spotify` (the default) or `apple-music`:

```sh
setlist-to-playlist --provider apple-music --url https://www.setlist.fm/setlist/blink182/2024/autodromo-de-interlagos-sao-paulo-brazil-53aa1325.html
```

If you omit `--provider` and run the CLI in an interactive terminal, a menu asks you to choose between Spotify and Apple Music. When there is no terminal (CI, pipes), it uses Spotify without prompting. An explicit `--provider` never shows the menu.

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

### Apple Music setup (optional)

Only needed for `--provider apple-music`. Apple does not offer a free public API for this, so the CLI uses the same tokens as the [music.apple.com](https://music.apple.com) web player. Sign in there, then:

1. **Bearer token:** open your browser devtools, go to the Network tab, reload the page and click any request to `amp-api.music.apple.com`. Copy the value of the `Authorization` request header **without** the `Bearer ` prefix
2. **Media user token:** in devtools, open Application (Chrome) or Storage (Firefox/Safari) > Cookies > `https://music.apple.com` and copy the value of the `media-user-token` cookie

The CLI asks for both on the first Apple Music run and stores them next to the other credentials.

A few things to keep in mind:

- This approach is unofficial and can break whenever Apple changes the web player
- The tokens expire periodically; when that happens the CLI will ask for new ones
- The playlist link printed at the end is a library link, so it only works for the account that owns the playlist
- Library playlists may take a moment to show up in the Music app

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
