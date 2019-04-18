# Gleaner

## About

> Based on RDA P13 interest and EarthCube follow on work 
> I am currently working on the UI and run time patterns
> for Gleaner.  (April 2019)


Gleaner is the index builder for Project 418.  It is composed of two main 
elements.  

*The Summoner*, which uses site map files to access and parse facility 
resources pages.  Summoner places the results of these calls into a S3 API 
compliant storage. 

*The Miller*, which takes the JSON-LD documents pulled and stored by 
summoner and runs them through various millers.  These millers can do 
various things.  The current millers are:

* text:  build a text index in raw bleve
* spatial: parse and build a spatial index using a geohash server
* graph: convert the JSON-LD to RDF for SPARQL queries

A set of other millers exist that are more experimental

* tika: access the actual data files referneced by the JSON-LD and process
    through Apache Tika.  The extracted text is then indexed in text system allowing 
    full text search on the document contents.
* blast: like text, but using the blast package built on bleve
* fdptika: like tika, but using Frictionless Data Packages
* ftpgraph: like graph, but pulling JSON-LD files from Frictionless Data Packages
* prov: build a basic prov graph from the overall gleaner process
* shacl: validate the facility resoruces against defined SHACL shape graphs 

## How to run (or at least try..., this is still a work in progress)

1) Get the compose file via curl, wget, httpie or your favorite method.
curl https://raw.githubusercontent.com/earthcubearchitecture-project418/gleaner/master/deployments/gleanerServices.yml -o gleanerService.yml

2) Set up your environment variables, (I have no clue how this is done in Windows...   ).  One path is to make env file the following.  

```bash
# Set environments
export MINIO_ACCESS_KEY="KEYHERE"
export MINIO_SECRET_KEY="SECRETHERE"
export DATAVOL="/home/nemo/dataVolumes/gleaner"
```

Source this or set your environment variables in the manner you wish.

3) Pull the containers (or just let that happen when you invoke them)

4) If you use docker-compose then try

`docker-compose -f gleanerService.yml up -d`

or for swarm

`docker stack deploy --compose-file gleanerService.yml gleaner`

5) Grab the binary from the release page or pull the gleaner container.  At this 
point best to grab the release binary unless you are comfortable running command line apps from containers.  I'll document the later but likely always provide both.

- Docker hub: https://cloud.docker.com/u/nsfearthcube/repository/docker/nsfearthcube/gleaner with
   ```docker pull nsfearthcube/gleaner:latest```
- Gleaner release page at <https://github.com/earthcubearchitecture-project418/gleaner/releases>

5.5) Run the gleaner -checksetup command to validate connections and make the requiredminio buckets if they are missing  (this is NOT working yet)

6) Make a config file.  Pull the example one and edit it.   This is the worst part 
mostly likely and I am going to switch from JSON to YAML for configs.   (I should never have used JSON for config, sorry about that).  Reference: <https://github.com/earthcubearchitecture-project418/gleaner/blob/master/configs/basic_config.json> 

7) Copy your config file into the minio object store in the correct bucket with 
the correct object name.  The easiest way if you don't have a local s3 API 
compatible client is to pull the minio mc client from https://hub.docker.com/r/minio/mc/

8) Copy the config file in...   make sure to do this each time you edit it.
```mc cp my_config.json local/gleaner-config/config.json```

9) Ok, finally ready at step 8 to see if this even works.  
```go run cmd/gleaner/main.go```

## Next Steps

Update:

* fix handler for multi-sitemap sites (like BCO-DMO)
* have a validator check for sitemaps for when the web ui allows them to be submitted
* handle URL submissions that are sitemaps or resource URLs like for type: Organization
* review existing spatial indexer for issues around resources with multiple types (like bbox and points)

Add a new web ui to the system that:

* allows editing the config JSON
* allows index runs to be started and indexes to be built
* front the config JSON (or the go struct) to a UI for CRUD operations...  it's possible
    that we want to use JSON scheme here and some of the various Javascript libs for
    JSON schema to forms


