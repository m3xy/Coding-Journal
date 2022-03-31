# Introduction

Project-code is a journal that peer-review the code behind some research papers.
It's main aim is to run code and algorithms which produce reproducible results based
on the input that is given to it.

# Getting Started

Before getting started, you will need to make a `backend/secrets.env` file, containing:
+ A key for `BACKEND_TOKEN`,
+ A key for `JWT_SECRET`.

Make sure those values are secure.

## Running

### With Docker

You will need:
+ `docker`
+ `docker-compose`

If your computer has docker, you can simply start the project by running `docker-compose up -d`.

If you are getting an issue with the backend, run `docker-compose restart backend`.

### With Podman
 
This configuration can be run in the school machines.

You will need:
+ `podman` - installed on school machines
+ `podman-compose` - Run `pip3 install --user podman-compose` and make sure `$HOME/.local/bin` is in your `PATH`.
+ `go version 1.17-1.18` - cf. Installing dependencies

To run follow these steps:
1. On the root folder, run `podman-compose up --build -d`
2. Go to the `backend` directory
3. On the `backend` directory, run `source .env`. This will set up some required variables for the backend.
4. Run `go run .` to run the backend. Make sure the database (running under `podman-compose`) is running
beforehands.
5. To shut everything down, exit the backend process, and in the root folder, do `podman-compose down`.

## Installing Dependencies
Many dependencies in the project require versions that are not installed by default on the school machines.

#### Go
The school machines support upto version 1.16.12. But version 1.17 is required.

To get go version 1.17/1.18, you will need a go version manager (e.g. [g](https://github.com/stefanmaric/g))

To do so:
+ Run `curl -sSL https://git.io/g-install | sh -s -- -y` to install g. This will install g and the latest version of go.
+ Restart your terminal. Then, if you do `go version`, you should see `version 1.18` appearing on the terminal. 

#### Node
The school machines support node up to version 1.10. This version is still maintained by RedHat as an LTS,
but it does not the support many modern node features. Therefore, node version 1.17 is required at least.
It can be installed using `volta`.

To do so:
+ Run `curl https://get.volta.sh | bash`, then `~/.volta/bin/volta install node@latest` to install the latest
version of Node.
+ Restart the terminal. If you do `node --version`, you should at the very least see `version 1.17` or newer
appearing on screen.


## Troubleshooting
Any time a problem occurs, the best solution is to check at logs.
For the frontend or the database (db), one can do `podman-compose logs <frontend/db>`.
For the backend, the logs are written to `logs/cs3099-backend.log`.
