# -------
# Dockerfile: Docker image to run the NodeJS backend.
# Author(s): 190019931, 190014935
# ------

# ----
# Node docker image
# ----

FROM node:17-alpine

WORKDIR /usr/src/app

# If node modules already exists locally, start frontend. - Otherwise creates them before starting (Takes a while for 1st time!)
CMD [ -d "node_modules" ] && npm run start:dev || npm ci && npm run start:dev
