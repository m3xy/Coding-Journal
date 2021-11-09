#!/bin/sh
#The container containing the database
container="mariadb"


cd $(dirname "$(readlink -f "$0")")

#Stop the container
docker stop $container

#Remove all unused containers, networks, images (both dangling and unreferenced) and volumes.
docker system prune --volumes -a -f

# Use this if you don't want to delete the database volume
# docker system prune
