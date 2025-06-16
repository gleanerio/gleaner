# Notes on Loading to JanusGraph

These are a few notes and references on running and loading 
RDF into JanusGraph.  This has never been done in a 
production settings.  This was mostly something done to 
explore some alternative graph approaches.

## Refs

* https://github.com/costezki/rdf2gremlin
* https://docs.janusgraph.org/getting-started/installation/

## Notes
```
docker run -it -p 8182:8182 janusgraph/janusgraph
docker run --name janusgraph-default janusgraph/janusgraph:latest
```
