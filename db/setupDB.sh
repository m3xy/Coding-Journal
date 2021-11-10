#!/bin/sh

#Shell variables to store Docker DB:
# -> Image tag
image="maria-db"
# -> Container name
container="mariadb"
# -> Volume name
volume="my-db"

cd $(dirname "$(readlink -f "$0")")

#If the container already exists
if [ "$(docker ps -q -f name=$container)" ];then
    docker exec -it $container mysql -u root -p # -> Execute the running instance
else
    docker build -t $image . # -> Otherwise, build the image

	#   -> Check if the volume already exists (if not, create one - to store db persistently)
    if [ ! "$(docker volume ls -q -f name=$volume)" ];then
        docker volume create $volume
    fi
    docker run --name $container -dp 3307:3306 -v $volume:/var/lib/mysql $image #   -> Run the image as a container
fi
