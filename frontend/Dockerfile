# -------
# Dockerfile: Docker image to run the NodeJS backend.
# Author(s): , 190014935
# ------

# ----
# Node docker image
# ----
FROM node:16 AS cs3099node
WORKDIR /usr/src/app
COPY . ./
RUN npm i
EXPOSE 23409
CMD ["npm", "run", "start"]
