#!/bin/bash
# A wrapper script for loading RDF into Blazegraph from Minio
# usage:  load2blaze.sh souceBucket sparqlEndpoint
# example: nas/gleaner-oih/summoned/cioos/  http://localhost:9999/blazegraph/namespace/kb/sparql

mc_cmd() {
        mc ls $1 | awk '{print $6}'
}

#if [ "$#" -ne 1  ]; then
        #echo "Illegal number of parameters"
#fi

# If you use this for ntriples, be sure to add in a graph in the URL target
for i in $(mc_cmd $1); do
      echo "-------------start-------------"
      echo Next: $i
      mc cat $1/$i | curl -X POST -H 'Content-Type:text/x-nquads' --data-binary  @- $2
      echo "-------------done--------------"
done

