#!/usr/bin/env sh

set -eux

# Launch a MySQL container and expose the port 3306 to the localhost.
# A `todolist` database is created in the MySQL container if it does not exist.

IMAGE_NAME=mysql
CONTAINER_NAME=todolist-mysql

if ! docker ps -a | grep $CONTAINER_NAME; then
	docker run -dp 3306:3306 --name "$CONTAINER_NAME" -e MYSQL_ROOT_PASSWORD=root "$IMAGE_NAME"
fi
docker exec -it "$CONTAINER_NAME" \
	mysql -uroot -proot -e "CREATE DATABASE todolist;"
