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
2. Build app `./scripts/build-mac.sh` or `./scripts/build-win.bat`. Output will be in `build` folder

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
