# How to run

## Intro


## Steps


docker run -it  --mount type=bind,source="$(pwd)",target=/gleaner/config --net=host  --entrypoint=/gleaner/gleaner nsfearthcube/gleaner:2.0.7 -cfg=/gleaner/config/config



