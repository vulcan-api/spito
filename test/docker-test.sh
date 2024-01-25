#!/bin/bash

parent_dir=$( cd "$(dirname "${BASH_SOURCE[0]}")"; cd ..; pwd -P )
if [[ $1 -ge "" ]]
then
  image_id=$(docker buildx build -q --load $parent_dir -f "test/arch.Dockerfile")
else
  image_id=$(docker buildx build -q --load $parent_dir -f "test/$1.Dockerfile")
fi

container_id=$(docker run --rm -d --privileged=true $image_id)

### INFO!
#I bet you are wondering what here is happening
#To be honest I am not sure but "sleep 2"
#Make it works
sleep 2

docker exec -i $container_id go test ./...
exit_code=$(echo $?)

docker stop $container_id
docker rmi -f $image_id

exit $exit_code
