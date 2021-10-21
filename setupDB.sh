#!/bin/sh
image="maria-db"
container="mariadb"
volume="my-db"

if [ "$(docker ps -q -f name=$container)" ]
then
    docker exec -it $container mysql -u root -p
else
    docker build -t $image . -q
    if [ ! "$(docker volume ls -q -f name=$volume)" ]
    then
        docker volume create $volume
    fi
    docker run --name $container -dp 3306:3306 -v $volume:/var/lib/mysql $image
fi