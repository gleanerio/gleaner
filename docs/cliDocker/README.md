# Gleaner Docker instance

## About

This is a new approach for quick starts with Gleaner.  It is a script that exposes
a containerized version of Gleaner as a CLI interface.

You can use the -init flag to pull down all the support files you need including
the Docker Compose file for setting up the object store, a triplestore and 
the support for headless indexing.  

## Prerequisites

You need Docker installed

## Steps

Download the script gleanerDocker.sh from https://github.com/earthcubearchitecture-project418/gleaner/tree/master/docs/cliDocker You may need to make it run-able with 

```bash
curl -O https://raw.githubusercontent.com/earthcubearchitecture-project418/gleaner/master/docs/cliDocker/gleanerDocker.sh

chmod 755 gleanerDocker.sh
```

Next you can run the script with the -init flag to pull down all the support files you need.

```bash
./gleanerDocker.sh -init
```

This will also download the needed docker image and the support files. 
Your directory should look like this now:

```bash
fils@ubuntu:~/clidocker# ls -lt
total 1356
-rw-r--r-- 1 fils fils    1281 Aug 15 14:07 gleaner-IS.yml
-rw-r--r-- 1 fils fils     290 Aug 15 14:07 setenvIS.sh
-rw-r--r-- 1 fils fils    1266 Aug 15 14:07 template_v2.0.yaml
-rw-r--r-- 1 fils fils 1371350 Aug 15 14:07 schemaorg-current-https.jsonld
-rwxr-xr-x 1 fils fils    1852 Aug 15 14:06 gleanerDocker.sh
```

Let's see if we can setup our support infrastructure for Gleaner.  The 
file gleaner-IS.yml is a docker compose file that will set up the object store,
and a triplestore.

To do this we need to set up a few environment variables.  To do this we can 
leverage the setenvIS.sh script.  This script will set up the environment we need.
Note you can also use a .env file or other approaches.  You can references 
the [Environment variables in Compose](https://docs.docker.com/compose/environment-variables/) documentation.  

```bash
root@ubuntu:~/clidocker# source setenvIS.sh 
root@ubuntu:~/clidocker# docker-compose -f gleaner-IS.yml up -d
Creating network "clidocker_traefik_default" with the default driver
Creating clidocker_triplestore_1 ... done
Creating clidocker_s3system_1    ... done
Creating clidocker_headless_1    ... done
```

Note:  In a fresh run all the images will be pulled down.  This may take a while.

In the end, you should be able to see these images running:

```bash
root@ubuntu:~/clidocker# docker ps
CONTAINER ID        IMAGE                            COMMAND                  CREATED              STATUS              PORTS                                              NAMES
a26f7c945479        nawer/blazegraph                 "docker-entrypoint.s…"   About a minute ago   Up About a minute   0.0.0.0:9999->9999/tcp                             clidocker_triplestore_1
f3a4197c42be        minio/minio:latest               "/usr/bin/docker-ent…"   About a minute ago   Up About a minute   0.0.0.0:9000->9000/tcp, 0.0.0.0:54321->54321/tcp   clidocker_s3system_1
062f029462b1        chromedp/headless-shell:latest   "/headless-shell/hea…"   About a minute ago   Up About a minute   0.0.0.0:9222->9222/tcp  
```

At this point we should be able to do a run.  During the init process a 
working config file was downloaded.   

> Note:  This config file will change...  it's pointing to an OIH partner 
> and I will not do that for the release.  I have a demo site I will use.  

Next we need to setup our object for Gleaner.  Gleaner itself can do this 
task so we will use 

```bash
root@ubuntu:~/clidocker# ./gleanerDocker.sh -setup -cfg template_v2.0
main.go:35: EarthCube Gleaner
main.go:110: Setting up buckets
check.go:58: Gleaner Bucket gleaner not found, generating
main.go:117: Buckets generated.  Object store should be ready for runs
```

> Note:  Here is where we go off the rails.  The config file uses 0.0.0.0 as the 
> location and this is not working.   You need to edit the config file with the 
> "real" IP of the host machine.  In may case is this 192.168.122.77.  This is 
> obviously still a local network IP but it does work.  I am still investigating 
> this issue.

We can now do a run with the example template file.  

> Note:  Best to delete the "sitegraph" node, I will do that soon.  It should 
> work, but is currently slow and gives little feedback

If everything goes well, you should see something like the following:

```bash
root@ubuntu:~/clidocker# ./gleanerDocker.sh -cfg template_v2.0
main.go:35: EarthCube Gleaner
main.go:122: Validating access to object store
check.go:39: Validated access to object store: gleaner.
org.go:156: Building organization graph (nq)
org.go:163: {samplesearth  https://samples.earth/sitemap.xml false https://www.re3data.org/repository/samplesearth Samples Earth (DEMO Site) https://samples.earth}
main.go:154: Sitegraph(s) processed
summoner.go:16: Summoner start time: 2021-08-15 14:34:08.907152656 +0000 UTC m=+0.067250623 
resources.go:74: samplesearth : 202
 100% |██████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████| (202/202, 20 it/s)        
summoner.go:34: Summoner end time: 2021-08-15 14:34:20.36804137 +0000 UTC m=+11.528139340 
summoner.go:35: Summoner run time: 0.191015 
webfeed.go:37: 1758
millers.go:26: Miller start time: 2021-08-15 14:34:20.368063453 +0000 UTC m=+11.528161421 
millers.go:40: Adding bucket to milling list: summoned/samplesearth
millers.go:51: Adding bucket to prov building list: prov/samplesearth
 100% |█████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████| (202/202, 236 it/s)        
graphng.go:77: Assembling result graph for prefix: summoned/samplesearth to: milled/samplesearth
graphng.go:78: Result graph will be at: results/runX/samplesearth_graph.nq
pipecopy.go:16: Start pipe reader / writer sequence
graphng.go:84: Pipe copy for graph done
millers.go:80: Miller end time: 2021-08-15 14:34:21.84702814 +0000 UTC m=+13.007126109 
millers.go:81: Miller run time: 0.024649 

```

At this point you have downloaded the JSON-LD documents if all has gone well.

I need to document loading these into the triplestore.  You can use something like
the scripts like minio2blaze.sh at https://github.com/earthcubearchitecture-project418/gleaner/tree/master/scripts.

I need to work up documentation for that though.  Also, those scripts require
that you have mc installed.  The Minio Client, which can be installed
following their [Minio Client Quickstate Guide](https://docs.min.io/docs/minio-client-quickstart-guide.html).

Let's do a quick stab at it... 

Download the minio2blaze.sh script.




## Notes

```basg
docker inspect -f '{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}' clidocker_s3system_1
```
