#!/bin/bash
# A wrapper script for loading RDF into Blazegraph from Minio
# usage:  load2blaze.sh souceBucket

mc_cmd() {
        mc ls $1 | awk '{print $6}'
}

# If you use this for ntriples, be sure to add in a graph in the URL target
for i in $(mc_cmd $1); do
      echo "-------------start-------------"
      echo Next: $i
      mc cat $1/$i | jsonld format -q | curl -X POST -H 'Content-Type:text/x-nquads' --data-binary  @- https://ts.collaborium.io/blazegraph/namespace/queue/sparql
      echo "-------------done--------------"
done

