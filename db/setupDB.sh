#!/bin/sh

#Shell variables to store Docker DB:
# -> Image tag
image="maria-db"
# -> Container name
container="mariadb"
# -> Volume name
volume="my-db"

if [ "$(docker ps -q -f name=$container)" ]                                     #If the container already exists
then
    docker exec -it $container mysql -u root -p                                 # -> Execute the running instance
else
    docker build -t $image . -q                                                 # -> Otherwise, build the image
    if [ ! "$(docker volume ls -q -f name=$volume)" ]                           #   -> Check if the volume already exists (if not, create one - to store db persistently)
    then
        docker volume create $volume
    fi
    docker run --name $container -dp 3307:3306 -v $volume:/var/lib/mysql $image #   -> Run the image as a container
fi