#!/bin/bash

parent_dir=$( cd "$(dirname "${BASH_SOURCE[0]}")"; cd ..; pwd -P )
image_id=$(docker buildx build -q --load $parent_dir)

docker run --rm -i $image_id
exit_code=$(echo $?)

docker rmi $image_id

exit $exit_code

