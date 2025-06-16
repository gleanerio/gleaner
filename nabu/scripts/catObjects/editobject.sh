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
      temp_file=$(mktemp)
      t=$(mc cat $1/$i)
      printf "$t . " > $temp_file
      mc cp $temp_file $1v2/$i
      rm $temp_file
done

