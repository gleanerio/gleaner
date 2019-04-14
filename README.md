# Gleaner

## Quick run

* Download the compose file from Github
* docker-compose -f gleanerServices.yml up -d
* docker pull nsfearthcube/gleaner:latest
* docker run --rm -ti nsfearthcube/gleaner:latest
* docker run --rm -ti -e MINIO_SECRET_KEY -e MINIO_ACCESS_KEY  nsfearthcube/gleaner:2.0.2
* docker run --rm -ti --env-file ./secret/kv.env nsfearthcube/gleaner:2.0.2

## About

> Based on RDA P13 interest and EarthCube follow on work 
> I am currently working on the UI and run time patterns
> for Gleaner.  (April 2019)


Gleaner is the index builder for Project 418.  It is composed of two main 
elements.  

The Summoner, which uses sitemap files to access and parse facility 
resources pages.  Summoner places the results of these calls into a S3 API 
compliant storage.  

The Miller, which takes the JSON-LD documents pulled and stored by 
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


## Running
This section is in development.  We are working to make Gleaner a tool
that is easy to deploy and use.  

```
docker run -it earthcube/gleaner:2.0.1 gleaner
```


* make output (code will check) and datavol to pass as arg to docker-compose
* export DATAVOL=full path to dv to use
* docker-compose -f gleaner-compose.yml up -d
* get mc for use with minio
* edit config 
* mc copy to minio
* run

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

## Running notes

docker-compose -f gleaner-compose.yml up -d
mc cp config.json local/gleaner
