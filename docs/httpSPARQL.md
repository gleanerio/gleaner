# SPARQL over HTTP

## About

looking at different patterns for loading large files over
http for SPARQL.

### GraphDB POST

```bash
curl -X POST -H 'Content-Type:application/n-quads'  --data-binary @May4Buildings.nq  http://192.168.86.45:32774/repositories/loadtest/rdf-graphs/services
```

### Blazegraph

```bash
curl -X POST -H 'Content-Type:text/x-nquads'  --data-binary @May4Buildings.nq  http://192.168.86.45:32772/blazegraph/namespace/loadtest/sparql
```

```bash
curl -H 'Accept: application/sparql-results+json' http://coreos.lan:9090/blazegraph/namespace/iow/sparql --data-urlencode 'query=select * where{ ?s ?p ?o } limit 10'
```

```bash
curl -H 'Accept: application/sparql-results+json' http://coreos.lan:9090/blazegraph/namespace/iow/sparql --data-urlencode 'query=SELECT (COUNT(DISTINCT ?graph) AS ?namedGraphsCount)(COUNT(*) AS ?triplesCount)WHERE {GRAPH ?graph {?subject ?predicate ?object}}'
```


### Oxigraph

```bash
curl -i -X PUT  -H 'Content-Type:text/x-nquads'    --data-binary @veupathdb_release.nq  http://localhost:7878/store
```

```bash
curl -i -X POST  -H 'Content-Type:text/x-nquads'    --data-binary @veupathdb_release.nq  http://localhost:7878/store
```