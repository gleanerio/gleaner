#!/bin/bash

# bind vs volume


docker volume create my-vol

docker run --rm  --mount type=bind,source=$(pwd),target=/gleaner/config \
	--mount type=volume,source=my-vol,target=/TMPDIR \
        --net=host  --entrypoint=/gleaner/gleaner nsfearthcube/gleaner:latest -cfg=/gleaner/config/config $1

docker volume rm my-vol
