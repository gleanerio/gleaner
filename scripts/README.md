# Scripts

# Init Minio

After using docker-compose to spin up the various containers
this script will download mc, generate the initial required
bucket and then populate the first config file.


# Load to Jena

Look at the code in /home/fils/src/Projects/LDN/GoLDeN/internal/graph

but we should be able to do this with curl too...  

test with 
File Upload: /test/upload
Graph Store Protocol: /test/data 
Graph Store Protocol (Read): /test/get 
HTTP Quads: /test/
SPARQL Query: /test/query
SPARQL Query: /test/sparql
SPARQL Update: /test/update

something like

```
curl -X POST --header "Content-Type:application/n-triples"   -d  @test_prov.n3 http://localhost:3030/test/data?graph=test
```

