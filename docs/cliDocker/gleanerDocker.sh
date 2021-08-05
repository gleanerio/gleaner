#!/bin/bash

# gleaner-cli 
# A wrapper script for invoking gleaner-cli with docker
# Put this script in $PATH as `rgleaner-cli`
# -getsdo pulls down latest schema.org context
# -update pulls down the latest version of the gleaner-cli 
# -cfgtemplate pulls down the latest config template 
# -getcompose pulls down the latest basic compose file
# -help displays help message

PROGNAME="$(basename $0)"
VERSION="v0.0.1"

if [[ $1 == "-init" ]];
then 
    curl -O https://schema.org/version/latest/schemaorg-current-https.jsonld
    curl -O https://raw.githubusercontent.com/earthcubearchitecture-project418/gleaner/master/configs/template_v2.0.yaml
    curl -O https://raw.githubusercontent.com/earthcubearchitecture-project418/gleaner/master/deployment/setenvIS.sh
    curl -O https://raw.githubusercontent.com/earthcubearchitecture-project418/gleaner/master/deployment/gleaner-IS.yml
    docker pull fils/gleaner:latest
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

## put in call to pull down schema.org context. 

# Guards (checks for dependencies)
check_cmd_in_path docker
check_cmd_in_path curl
# check_cmd_in_path docker-machine
# docker-machine active > /dev/null 2>&1 || error 2 "No active docker-machine VM found."

# Set up mounted volumes, environment, and run our containerized command
# podman needs --privileged to mount /dev/shm
#exec podman run \
  #--privileged \
  #--interactive --tty --rm \
  #--volume "$PWD":/wd \
  #--workdir /wd \
  #"localhost/nsfearthcube/gleaner:latest" "$@"

exec docker run \
  --interactive --tty --rm \
  --volume "$PWD":/wd \
  --workdir /wd \
  "fils/gleaner:latest" "$@"

