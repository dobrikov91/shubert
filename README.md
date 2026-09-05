# SHubert

[![Docker Image CI](https://github.com/dobrikov91/shubert/actions/workflows/docker-image.yml/badge.svg)](https://github.com/dobrikov91/shubert/actions/workflows/docker-image.yml)

## Description
The tool allows to use midi controller for sending shell commands. Could be useful for pet projects with api. App can sits in background and it has a web ui for any controller configuration.

## Usage
[ENG](docs/help-en.md)
[RUS](docs/help-rus.md)

## Build
0. Install go lang https://go.dev/doc/install
1. Clone the repo `git clone https://github.com/dobrikov91/shubert.git`
2. Build app `./scripts/build-mac.sh`, `./scripts/build-win.bat`, or `./scripts/build-linux.sh`. Output will be in `build` folder

The web UI (HTML/CSS/JS/images) is embedded into the binary at build time, so the `templates` folder is **not** required for distribution — only the single executable (plus the `data` folder, which holds your config) needs to be shipped/copied.

## Release
Version is tracked in `version.txt` (embedded into the binary at build time — no manual `-ldflags` needed).

To cut a new release:
1. Bump the version number in `version.txt` and commit it (`git commit -am "Bump version to X.Y.Z"`)
2. Push a matching git tag: `git tag vX.Y.Z && git push --tags`
3. The [`release` workflow](.github/workflows/release.yml) picks up the tag, builds native binaries for Linux/macOS/Windows, and attaches them to a new [GitHub Release](https://github.com/dobrikov91/shubert/releases) automatically

The git tag only triggers the workflow — it does not itself set the app's version, so step 1 must happen first (and the tag should match `version.txt`).

Pushing the tag also triggers the [Docker workflow](.github/workflows/docker-image.yml), which publishes a matching versioned image (`dobrikov91/shubert:X.Y.Z`) to Docker Hub alongside `:latest`.

## Docker
Note: app inside the docker will execute commands inside the container. I found it useful to call web API of another service.

0. Install docker
```
curl -fsSL https://get.docker.com -o get-docker.sh
sh get-docker.sh
```
1. Navigate to the project folder
2. Change variables in `compose.yaml` if required
3. Run `sudo docker compose up -d`

## License
This project is licensed under the [MIT License](LICENSE).
