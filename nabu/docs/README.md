#  Loading Gleaner Output into Triplestores

## About

Nabu primarily is a tool for reading from an S3 object store and writing 
to a triplestore.  The object stores can be any S3 compliant object stores
so AWS S3,  Google Cloud Storage, Wasabi, or others.   For most cases I am using
Minio, an open source S3 object store. 

Similarly, the triplestore can be any standards compliant triplestore.  Here
the primary standards we need implemented include

* SPARQL 1.1 with Update support
* SPARQL 1.1 over HTTP

As noted in the documentation late, if using the _bulk_ upload feature you 
need set the METHOD and ContentType your triplestore will expect. 

## Basic operations

Nabu needs a configuration file.  A template for this can be seen 
in [example.yaml](../config/example.yaml).  

Commands are like the following:
### Help
nabu --help

### Prefix: load all objects in the bucket/prefixes

The mode "prefix" in Nabu is used for loading a S3 object prefix 
path into a triplestore

```bash
nabu --cfg example.yaml  prune

nabu --cfg example.yaml  prune --prefix summoned/amgeo
```

```bash
nabu --cfg .example.yaml prefix

nabu --cfg example.yaml prefix -prefix summoned/amgeo

```

```bash
nabu prefix --cfg ../gleaner/configs/nabu
```
and gleaner generated configuration: 
```
nabu prefix --cfgPath directory --cfgName name 
```
eg use generated
```
nabu prefix --cfgPath ../gleaner/configs --cfgName local 
```
### Prune: load all objects in the bucket/prefixes

The mode "prune" in Nabu is to sync a prefix to the graph (remove graphs no longer in use, add new ones)

Note that updated graphs become new objects, since the object name is the SHA256 of the object


```bash
nabu prune --cfg file 
```
and gleaner generated configuration:
```bash
nabu prefix --cfgPath directory --cfgName name 
```
eg use generated
```bash
nabu prefix --cfgPath ../gleaner/configs --cfgName local 
```


### Bulk

This commands loads all the triples into the triplestore using the bulk load 
approach.  This is a SPARQL UPDATE call, vs the classic SPARQL command.  Nabu
will generate all the triples into a temporary file and then use that load into the 
triplestore.  This file will be removed after it is used. 

Required configuration entry

```yaml
sparql:
  endpoint: http://localhost/blazegraph/namespace/earthcube/sparql
  endpointBulk: http://coreos.lan:3030/testing/data
  endpointMethod: PUT
  contentType: application/n-quads
  authenticate: false
  username: ""
  password: ""
```
The bulk loading endpoint for many triplestores is different from the default
SPARQL endpoint.  Also, different vendors will likely require different methods
and content type.  These are only needed in the case where you are using the
_bulk_ command in Nabu.  For example:

GraphDB example ([reference](https://graphdb.ontotext.com/documentation/10.2/))
```yaml
endpointBulk: http://example.org:7200/repositories/testing/statements
endpointMethod: PUT
contentType: application/n-quads
```

Jena example ([reference](https://jena.apache.org/tutorials/index.html))
```yaml
endpointBulk: http://example.org:3030/testing/data
endpointMethod: PUT
contentType: application/n-quads
```

Blazegraph example ([reference](https://github.com/blazegraph/database/wiki/REST_API))
```yaml
endpointBulk: http://example.org9090/blazegraph/namespace/kb/sparql
endpointMethod: POST
contentType: text/x-nquads
```



Bulk load  the  specified source in the objects-prefix node
of the configuration file, use the _--prefix_ flag to specify the source.

```bash
nabu bulk --cfg ./example.yaml  --prefix summoned/providera
```

Bulk load all the sources defined in the objects-prefix node
of the configuration file.

```bash
nabu bulk --cfg ./example.yaml   
```


### Release

The _release_ command is used to build out release graphs.  These are the entire
set of objects associated with a provider, rolled up in one file.  These are done 
as nquads with the named graph following the pattern as defined in the ADR
[0001-URN-decision](https://github.com/gleanerio/nabu/blob/dev/decisions/0001-URN-decision.md).


To build a release graphs for a specified source in the objects-prefix node
of the configuration file, use the _--prefix_ flag to specify the source.  

```bash
nabu release --cfg ./example.yaml  --prefix summoned/providera
```

To build all the release graphs for the sources defined in the objects-prefix node
of the configuration file.

```bash
nabu release --cfg ./example.yaml   
```

### Object: load one object in the bucket/prefixes

The mode "object" in Nabu is used for loading a S3 object
path into a triplestore
```
nabu object --cfg file objectId
```
eg
```
nabu object --cfg ../gleaner/configs/nabu milled/opentopography/ffa0df033bb3a8fc9f600c80df3501fe1a2dbe93.rdf
```

and gleaner generated configuration:
```
nabu object --cfgPath directory --cfgName name  objectId
```
eg use generated
```
nabu object --cfgPath ../gleaner/configs --cfgName local milled/opentopography/ffa0df033bb3a8fc9f600c80df3501fe1a2dbe93.rdf
```

### Using URL based configuration

Nabu can also read the configuration file from over the network

```
go run ../../cmd/nabu/main.go release --cfgURL https://provisium.io/data/nabuconfig.yaml --prefix summoned/dataverse --endpoint localoxi
```