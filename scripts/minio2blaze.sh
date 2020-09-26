#!/bin/bash
# A wrapper script for loading RDF into Jena from Minio
# usage:  load2jena.sh souceBucket targetDataBase targetGraph
# example:  load2jena.sh local/gleaner-milled/run1 index common 
# example: ./load2jena.sh local/gleaner-milled/runtwo earthcube runid
# todo replace the following sections with $1 $2 $3 from above command invoking
mc_cmd() {
        mc ls $1 | awk '{print $5}'
}

# If you use this for ntriples, be sure to add in a graph in the URL target
for i in $(mc_cmd $1); do
      echo "-------------start-------------"
      echo Next: $i
      mc cat $1/$i | curl -X POST -H 'Content-Type:text/x-nquads' -d @- https://graph.geodex.org/blazegraph/namespace/cdf/sparql
      echo "-------------done--------------"
done

