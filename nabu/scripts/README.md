# Scripts

## About

These are some scripts that can be used to load graphs into
various graph databases


## Archives


Blaze (POST)

```
mc cat $1/$i | curl -X POST -H 'Content-Type:text/x-nquads' --data-binary  @- $2
```

Jena (PUT)

```
mc cat $1/$i | curl -X PUT --header "Content-Type:application/n-quads" -d @- http://localhost:3030/$2/data
```

Meili

```
mc cat $1/$i | curl  -H 'Content-Type: application/json' -X POST -d @- http://127.0.0.1:7700/indexes/movies/documents
```

rapper

```
mc cat $1/$i | rapper --count --input ntriples --input-uri "http://example.org/" -
```

Zinc

```
mc cat $1/$i | curl -u admin:"Complexpass#123" -X PUT -d @- http://localhost:4080/api/myshinynewindex/document
```
