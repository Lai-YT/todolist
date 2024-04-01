#!/usr/bin/env sh

set -eux

# Launch a MySQL container and expose the port 3306 to the localhost.
# A `todolist` database is created in the MySQL container if it does not exist.

IMAGE_NAME=mysql
CONTAINER_NAME=todolist-mysql
DATA_DIR=$PWD/.data
PORT=3306
MYSQL_ROOT_PASSWORD=root

# If the .data directory does not exist, create it.
if [ ! -d .data ]; then
	mkdir .data
fi
# If the container does not exist, create it.
if ! docker ps -a | grep $CONTAINER_NAME; then
	docker run \
		-dp $PORT:3306 \
		-v "$DATA_DIR":/var/lib/mysql \
		--name "$CONTAINER_NAME" \
		-e MYSQL_ROOT_PASSWORD="$MYSQL_ROOT_PASSWORD" \
		"$IMAGE_NAME"
fi
# If the `todolist` database does not exist, create it.
docker exec -it "$CONTAINER_NAME" \
	mysql -uroot -p$MYSQL_ROOT_PASSWORD \
	-e "CREATE DATABASE todolist;"
