# Gleaner getting started

This guide is intended to walk through getting started with Gleaner in the easier manner.   It is not the only way to use this program and may not be the best way for your environment or use case.  However, it will give a general overview and many of the points here are common across all the ways to use this program.

First, what is Gleaner.  Gleaner is a tool for extracting structured data on the web from a set of define providers.  It is not a web crawler and will not follow links in the pages it access.  It use a sitemap file created by a provider which is a set of URLs to resources Gleaner will visit.  Gleaner then extracts the structured data from the web represented by JSON-LD.  Some readers will note there are others ways, like RDFA, to represent this data on a web site.  Gleaner only looks for JSON-LD at this time.

If you are interested in publishing this sort of data, take a look at the ESIP hosted Science on Scheme GitHub repository and also the Google Developer Guide on publishing this sort of data as a provider.  

## **Prerequisites** 

To start this guide you will need a few things first.  One is a computer with Docker installed.  Docker is a popular tool for creating and using containers.  Containers are packaged applications like databases, games or web servers.  The Docker runtime providers a cross platform environment to run this common container images.  Images are downloaded from the net and can be maintained and updated.  Containers can be run in large cloud based environments with sophisticated orchestration systems or on your local computer.  For this example we will be running on a rather simple Linux based server provided by NSF's XSEDE environment.  However, any personal computer will do just fine.   You can download Docekr your PC or Mac at https://www.docker.com/products/docker-desktop and Linux users can typically just use your distro's package management system.

Once you have Docker installed and verified its operation you will need to download the Gleaner "Stater Pack".  Reference the Gleaner releases at https://github.com/earthcubearchitecture-project418/gleaner/releases to download the needed files for this guide.   In particular the starterpack.zip located in the assets section of the releases.   

Gleaner is a command line application, so you will need a terminal program on your computer and be comfortable issuing basic commands.  

Testing on Windows has not taken place yet.  This documentation will be updated when that is done. These scripts are just simple Docker commands, so use them as a guide and we will work to generate the Windows scripts ASAP.

Let's make a directory and download the starter pack.  Visit the web site at  https://github.com/earthcubearchitecture-project418/gleaner/releases  and find the latest release.  In the assets drop down section you will find links to the various assets of that release.  We will need the starterpack.zip file.  

Here we will use wget to download the file, but you could use curl or just web browser to download this file. 

```bash
root@gleaner:~# mkdir gleanerRuns
root@gleaner:~# cd gleanerRuns/
root@gleaner:~/gleanerRuns# wget https://github.com/earthcubearchitecture-project418/gleaner/releases/download/2.0.11/starterpack.zip
[...]
Saving to: ‘starterpack.zip’
[...]
2019-11-05 04:53:10 (41.8 MB/s) - ‘starterpack.zip’ saved [1752/1752]

root@gleaner:~/gleanerRuns# unzip starterpack.zip
Archive:  starterpack.zip
  inflating: starterpack/demo.env
  inflating: starterpack/gleaner-base.yml
   creating: starterpack/shapefiles/
  inflating: starterpack/v2config.yaml
root@gleaner:~/gleanerRuns# cd starterpack/
root@gleaner:~/gleanerRuns/starterpack# ls -lt
total 164
-rw-r--r-- 1 root root    356 Oct 29 10:25 README.md
-rw-r--r-- 1 root root   1651 Oct 23 22:49 config.yaml
-rw-r--r-- 1 root root   1042 Oct 23 22:45 docker-compose.yml
-rwxr-xr-x 1 root root    320 Oct 23 22:42 runGleaner.sh
-rw-r--r-- 1 root root    137 Oct 23 21:40 demo.env
drwxr-xr-x 2 root root   4096 Oct 20 03:00 shapegraphs
-rw-rw-r-- 1 root root 141567 Oct  8 11:52 jsonldcontext.json
```

These files are the ones we will now start with. 

## **Setting the Environment Variables**

There are a few things we need to do first to set out the environment in which Gleaner will run.   One is actually set a few "environment variables".   If you look at the demo.env file you will see:

```bash
root@gleaner:~/gleanerRuns/starterpack# cat demo.env
# Set environments
export MINIO_ACCESS_KEY="MySecretAccessKey"
export MINIO_SECRET_KEY="MySecretSecretKeyforMinio"
export DATAVOL="./DV"
```

A note on these values.  The Minio entries are for the Minio object storage system that Gleaner uses and that we will install and run here soon via Docker.  You might wish to change these though note if you are just running Minio locally and for Gleaner only on a system not Internet accessible these are OK to start with.  Good practice would be to change them of course.    

#### Data Volume 

The DATAVOL variable will be used to define where Minio and other elements of the run store data.  You don't have to use a DATAVOL mount but it is a good idea.  First, it will persist your data should you shutdown and restart your Docker containers later, which you likely will.  Second, it will give a performance nudge over writing to the Docker file system, which is an abstraction over your native file system.   If you are more familiar with Docker, feel free to change this. 

#### Setting

We need to ensure these are set in our terminal (shell).   For most of you should be able to simply source this file and set the values.  In the following we will check for a variable, not find it, source the file and then confirm we see it. 

```bash
root@gleaner:~/gleanerRuns/starterpack# echo $DATAVOL

root@gleaner:~/gleanerRuns/starterpack# source demo.env 
root@gleaner:~/gleanerRuns/starterpack# echo $DATAVOL
./DV
```

Note that the file assumes BASH or a BASH compliant shell.  If you are running ZSH or another shell, you likely know how to set these.  Those of you more familiar with Docker might note you could copy this file to a new name.  Specifically a .env file located in the same directory as your docker-compose.yml file.  This should also work.  Also, the compose file actually reference this demo.env file as an "env_file" entry.   However, I have seen cases where some of these approaches do not always work, so manually setting them and confirming them is a good move unless you are more familiar with Docker and Docker annoyances.  

## Docker Compose command

We are not ready to set up the containers we need running to support Gleaner.  In the starter pack there is a file called docker-compose.yml.  It's a bit large for this document but you can view the version on GitHub at:  https://github.com/earthcubearchitecture-project418/gleaner/blob/master/docs/starterpack/docker-compose.yml

This file contains the instructions Docker will use to download and run the various container images we need.  If you look at the file you will see 6 images.  Their labels and a short description follow;

* mc
* glenaer
* minio
* tangram
* headless
* jena

### Getting the images

The first thing we can do is download the images we will need.  You do not need to do this separately, issueing the run command in Docker will check for and download any required images.  For this document, however, let's do it as a special command.   Note, you only need to download an image once, it will then been stored local to your system as an image and will run from there.  Later you can check for new versions or updates too.   If we look at our images we might not see anything if this is a new system with Docker.  

```bash
root@gleaner:~/gleanerRuns/staterpack# docker images
REPOSITORY          TAG                 IMAGE ID            CREATED             SIZE
```

Let's download our images.  Still in our starterpack directory we can use the command:

```bash
root@gleaner:~/gleanerRuns/starterpack# docker-compose -f docker-compose.yml pull
[... Docker will report out the progress here, I have removed it for brevity]
root@gleaner:~/gleanerRuns/starterpack# docker images
REPOSITORY                TAG                 IMAGE ID            CREATED             SIZE
chromedp/headless-shell   latest              2c051a7d9f70        10 hours ago        220MB
nsfearthcube/gleaner      latest              c5fb0023d473        2 weeks ago         104MB
minio/minio               latest              8869bca0366f        3 weeks ago         51MB
minio/mc                  latest              f4f9de663a7f        3 weeks ago         22.7MB
fils/p418tangram          0.1.15              833aa7811eb1        3 weeks ago         991MB
fcore/p418jena            0.1.11              879cafba0181        4 months ago        2.67GB
```

Depending on your network, this might take a minute or two.   After the pull command we can rerun the "docker images" command to now see the images we will be using for our run.   We host these at https://hub.docker.com/ and there are over a 100,000 containers there from the Docker community.  

### Run docker-compose

At this point we are ready to set up the environment.  As noted, do this without doing the above pull command is fine.  Docker will see what images it needs to satisfy a compose file and fetch them.  

We have:

* set up our environment variables
* downloaded our images

Next we need to issue the command to run these images.  We can see what containers we have already with the docker ps command as in:

```bash
root@gleaner:~/gleanerRuns/starterpack# docker ps
CONTAINER ID        IMAGE               COMMAND             CREATED             STATUS              PORTS               NAMES
```

Here we have no running containers.   Let's run some.

```bash
root@gleaner:~/gleanerRuns/starterpack# docker-compose -f docker-compose.yml up -d
WARNING: The Docker Engine you're using is running in swarm mode.

Compose does not use swarm mode to deploy services to multiple nodes in a swarm. All containers will be scheduled on the current node.

To deploy your application across the swarm, use `docker stack deploy`.

Creating network "starterpack_default" with the default driver
Creating network "starterpack_web" with driver "overlay"
Creating starterpack_jena_1 ... 
Creating starterpack_gleaner_1 ... 
Creating starterpack_minio_1 ... 
Creating starterpack_tangram_1 ... 
Creating starterpack_headless_1 ... 
Creating starterpack_mc_1 ... 
Creating starterpack_gleaner_1
Creating starterpack_jena_1
Creating starterpack_minio_1
Creating starterpack_mc_1
Creating starterpack_headless_1
Creating starterpack_jena_1 ... done
root@gleaner:~/gleanerRuns/starterpack# docker ps -a
CONTAINER ID        IMAGE                            COMMAND                  CREATED             STATUS                            PORTS                     NAMES
789f5343f06a        fils/p418tangram:0.1.15          "/bin/sh -c 'exec gu…"   12 seconds ago      Up 6 seconds                      0.0.0.0:8080->8080/tcp    starterpack_tangram_1
dc22eb223210        chromedp/headless-shell:latest   "/headless-shell/hea…"   12 seconds ago      Up 5 seconds                      0.0.0.0:32772->9222/tcp   starterpack_headless_1
7026b893f9b1        fcore/p418jena:0.1.11            "/usr/local/bin/entr…"   12 seconds ago      Up 4 seconds (health: starting)   0.0.0.0:3030->3030/tcp    starterpack_jena_1
730263a80255        minio/mc:latest                  "mc"                     12 seconds ago      Exited (0) 9 seconds ago                                    starterpack_mc_1
8beac8e063c2        minio/minio:latest               "/usr/bin/docker-ent…"   12 seconds ago      Up 6 seconds                      0.0.0.0:9000->9000/tcp    starterpack_minio_1
408610b309b9        nsfearthcube/gleaner:latest      "/gleaner/gleaner"       12 seconds ago      Exited (2) 8 seconds ago                                    starterpack_gleaner_1
root@gleaner:~/gleanerRuns/starterpack# 

```

A few things to note in the above:

#### swarm

The WARNING about swarm you likely will NOT see as you likely are NOT running in Swarm mode.  That is fine.  If you are running in swarm mode (you likely don't need this document) but you might be using the command

```bash
docker stack deploy --compose-file docker-compose.yml gleaner
```

as some swarm instances will not have the docker-compose command.

#### ports

You can see that our containers are using some ports to communicate on.  If you are experienced with Docker and running other containers you may have an issue where ports are already in use.  You will need to resolve that.  Note, Gleaner goes looking for services on these ports and I don't currently offer the ability to change that. So you will need to resolve it.   The required ports are: 8080, 32772, 3030 and 9000.   In future releases I will try and note that these should be changed to less popular ports.  8080 in particular may be problematic if you are running other containers as it's a popular port for local http services. 

## Gleaner Configuration

So now we are ready to review the Gleaner configuration file named config.yml.  There is actually quite a bit in this file, but for this starting demo only a few things we need to worry about.  The default file will look like:

```yaml
---
minio:
  address: localhost
  port: 9000
  accesskey: MySecretAccessKey
  secretkey: MySecretSecretKeyforMinio
  ssl: false
gleaner:
  runid: demo  # this will be the bucket the output is placed in...
  summon: true # do we want to visit the web sites and pull down the files
  mill: true
  tmpdir: ""
context:
  cache: true
contextmaps:
- prefix: "https://schema.org/"
  file: "/gleaner/config/jsonldcontext.json"
- prefix: "http://schema.org/"
  file: "/gleaner/config/jsonldcontext.json"
- prefix: "https://schema.org"
  file: "/gleaner/config/jsonldcontext.json"
- prefix: "http://schema.org"
  file: "/gleaner/config/jsonldcontext.json"
summoner:
  mode: diff  # [time, hash, none] diff: look for difference or full: delete old objects and replace
millers:
  graph: true
  shacl: false
  prov: false
shapefiles:
- ref: https://raw.githubusercontent.com/geoschemas-org/geoshapes/master/shapegraphs/googleRequired.ttl
- ref: https://raw.githubusercontent.com/geoschemas-org/geoshapes/master/shapegraphs/googleRecommendedCoverageCheck.ttl
sources:
- name: opencore
  logo: http://geodex.org/images/logos/EarthCubeLogo.png
  url: http://opencoredata.org/sitemap.xml
  headless: false
```

A few things we need to look at.

First, in the "mino:" section make sure the accessKey and secretKey here match the ones you have and set via your demo.env file. 

Next, lets look at the "gleaner:" section.  We can set the runid to something.  This is the ID for a run and it allows you to later make different runs and keep the resulting graphs organized.  It can be set to any lower case string with no spaces. 

The miller and summon sections are true and we will leave them that way.  It means we want Gleaner to both fetch the resources and process (mill) them.  

Now look at the "miller:"  section when lets of pick what milling to do.   Currently it is set with only graph set to true.  Let's leave it that way for now.  This means Gleaner will only attempt to make graph and not also run validation or generate prov reports for the process.  

The final section we need to look at is the "sources:" section.   Here is where the fun is.  

```yaml
sources:
- name: opencore
  logo: http://geodex.org/images/logos/EarthCubeLogo.png
  url: http://opencoredata.org/sitemap.xml
  headless: false
```

These are the sources we wish to pull and process.  Each source has 4 entries though at this time we no longer use the "logo" value.  It was used in the past to provide a page showing all the sources and a logo for them.  However, that's really just out of scope for what we want to do.  You can leave it blank or set it to any value, it wont make a difference.  

The name is what you want to call this source.  It should be one word (no space) and be lower case. 

The url value needs to point to the URL for the site map XML file.  This will be created and served by the data provider. 

The headless value should be set to false unless you know this site uses JavaScript to place the JSON-LD into the page.  This is true of some sites and it is supported but not currently auto-detected.  So you will need to know this and set it.  For most place, this will be false. 

You can have as many sources as you wish.  For an example look the configure file for the CDF Semantic Network at: https://github.com/earthcubearchitecture-project418/CDFSemanticNetwork/blob/master/configs/cdf.yaml

For this demo we will use the site map for Open Core data.  However, I would prefer to get a better and smaller example source that could highlight various capabilities and edge cases for this package.  

A more detailed review of the configuration file will be made and linked here.  

## **Run Gleaner via Docker**

With our configuration file ready we have arrived at the time when we can run Gleaner.  

```bash
 ./runGleaner.sh --setup
```

Note the option and the changes needs to run the binary here?

## **Reviewing the output**

The output from the Gleaner runs is located in the object store (default Minio). You can use the mc command in the related Docker container or any other s3 compatible viewer like CyberDuck. 



## **Load to a triplestore and query** 

A set of scripts are available in the repository to load the Gleaner output into a file system or triplestore. 