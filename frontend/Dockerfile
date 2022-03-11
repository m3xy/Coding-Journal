# -------
# Dockerfile: Docker image to run the NodeJS backend.
# Author(s): 190014935
# ------

# ----
# Node docker image
# ----

FROM node:17-alpine

WORKDIR /usr/src/app
COPY package*.json ./
RUN npm i

COPY . ./
EXPOSE 3000
CMD ["npm", "run", "start:dev"]

#CMD [ -d "node_modules" ] && npm run start:dev || npm ci && npm run start:dev