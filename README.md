# Gleaner

## About

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

