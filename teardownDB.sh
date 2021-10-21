#!/bin/sh
container="mariadb"
docker stop $container

#Note: this deletes all unused images also
docker system prune --volumes -a -f

# Use this if you don't want to delete the database volume
# docker system prune