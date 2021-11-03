# ECAM Gleaner Demo

## Prerequsits

The only real requirement is a running Docker installation.  You can learn more about
docker and how to install it for your system at https://www.docker.com/

## Downloads

* Pull down the docker config from 
https://github.com/gleanerio/gleaner/tree/master/deployments   We will use the 
gleaner-base.yml
* Pull down the "starterpack.zip" from https://github.com/gleanerio/gleaner/releases
* Pull down the copy of gleaner for your system from the distribution page
* The minio S3 client mc is also needed.  https://min.io/download
	* You will likely need to "chmod +x mc" that client
	* You can also see usage with: ./mc --help
	* The command "./mc config" is used to generate your files.  Locate them and edit them
	with the access and secret keys you used in the prod.env file

## Set up your environment

* Make a directory to act at the data volume.  This is where files and data are persisted between container
runs.   For this demo we will use a directory called DV located at /root/DV
* You will need to build an environment file.  There is one called demo.env you can pull out of
the starterpack.zip
* You can also copy this file to .env as well so that it is default read by docker compose
* These environment files should be in the same directory you will place your docker compose file.
* You will need a docker configuration file.  There are a few (the ones ending .yml) in the starterpack.zip 
file.  You can also find them at the GitHub repo for Gleaner in the deployments directory.
* Be sure to export these variables to the shell you are running gleaner from.

## Setup runs

* You will need to run docker-compose -f config.yml up -d (or equivelant) to setup the 
docker containers needed by Gleaner.  The staterpack.zip contains a basic compose file you can 
start with and modify to your needs. 
* You need to run Gleaner with the flag to check and create the needed buckets 
	* `./gleaner -setup`
	* `./gleaner -address=localhost -port=9000 -setup`
* You need to copy the shape graphs into the gleaner-shacl bucket to support the
optional SHACL validation miller.
`mc cp *.ttl local/gleaner-shacl`

## First run

With all the containers running and the propper buckets in place in minio we are ready 
to run Gleaner.

We need to ensure our configuration file is ready to go.  There are a couple included in the starterpack.zip
you can use for some testing runs.

```gleaner -configfile ssdb.yml```

## Docker notes

You will not really need to use these commands.  However, if you want to clean up some 
of the docker containers, the following commands are useful.  Careful, these commands
are rather sweeping.

* kill all running containers with docker kill $(docker ps -q)
* delete all stopped containers with docker rm $(docker ps -a -q)
* delete all images with docker rmi $(docker images -q)

