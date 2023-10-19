#!/bin/bash

parent_dir=$( cd "$(dirname "${BASH_SOURCE[0]}")"; cd ..; pwd -P )
image_id=$(docker buildx build -q --load $parent_dir)

container_id=$(docker run --rm -d --privileged=true $image_id)

docker exec -i $container_id go test ./...
exit_code=$(echo $?)

docker stop $container_id
docker rmi -f $image_id

exit $exit_code

