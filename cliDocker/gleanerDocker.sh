#!/bin/bash

# gleaner-cli : A wrapper script for invoking gleaner-cli with docker

# TODO:  add help with things like config for 
# cfg eco_local 
# cfg eco_local -source cfg_name


PROGNAME="$(basename $0)"
VERSION="v0.0.1"

# Pull down some of the needed docks if called with -init
if [[ $1 == "-init" ]];
then 
    curl -O https://schema.org/version/latest/schemaorg-current-https.jsonld
    curl -O https://raw.githubusercontent.com/earthcubearchitecture-project418/gleaner/master/configs/demo.yaml
    curl -O https://raw.githubusercontent.com/earthcubearchitecture-project418/gleaner/master/deployment/setenvIS.sh
    curl -O https://raw.githubusercontent.com/earthcubearchitecture-project418/gleaner/master/deployment/gleaner-IS.yml
    docker pull fils/gleaner:latest
    echo "\n See notes at: https://github.com/gleanerio/gleaner/blob/dev/docs/cliDocker/README.md \n"
    exit 0
fi

# Helper functions for guards
error(){
  error_code=$1
  echo "ERROR: $2" >&2
  echo "($PROGNAME wrapper version: $VERSION, error code: $error_code )" &>2
  exit $1
}
check_cmd_in_path(){
  cmd=$1
  which $cmd > /dev/null 2>&1 || error 1 "$cmd not found!"
}

# Guards (checks for dependencies)
check_cmd_in_path docker
check_cmd_in_path curl


# Podman:  podman needs --privileged to mount /dev/shm
#exec podman run \
  #--privileged \
  #--interactive --tty --rm \
  #--volume "$PWD":/wd \
  #--workdir /wd \
  #"localhost/nsfearthcube/gleaner:latest" "$@"

# Docker:  current docker command to do local volume mounts
exec docker run \
  --interactive --tty --rm \
  --volume "$PWD":/wd \
  --workdir /wd \
  "fils/gleaner:latest" "$@"

