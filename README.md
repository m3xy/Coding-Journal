# project-code

# Build
## With full Dockerisation
Run `./dockerBuild` in the root directory. This script will build all necessary containers.

## Without full Dockerisation
Run:
+ `npm i` from the =frontend/= directory,
+ `go build .` from the =backend/= directory,
+ `sh ./setupDB.sh` from the =db/= directory. This script sets up a container for the database.
Docker is still necessary for this step.

# Run
## With full Dockerisation
Run `./run` in the root directory. 

Alternatively, run:
+ `docker start mariadb` to start the database,
+ `docker run --name backend -net host -d ci-backend` to start the backend.
+ `docker run --name frontent -net host -d ci-frontend` to start the frontend.

Then, after initial setup, because the containers, have already been created, run:
+ `docker start mariadb`,
+ `docker start backend`,
+ `docker start frontend`.

## Without full Dockerisation
Run:
+ `docker start mariadb`
+ `backend/backend` to run the backend.
+ `cd frontend && npm run start` to run the frontend.
