#!/bin/bash
# A wrapper script for loading RDF into Jena from Minio
# usage:  load2jena.sh souceBucket targetDataBase targetGraph
# example:  load2jena.sh local/gleaner-milled/run1 index common 
# example: ./load2jena.sh local/gleaner-milled/runtwo earthcube runid
# todo replace the following sections with $1 $2 $3 from above command invoking
mc_cmd() {
        mc ls $1 | awk '{print $5}'
}

for i in $(mc_cmd $1); do
    echo "--------start----------------"
    echo $i
    mc cat $1/$i | rapper --count --input ntriples --input-uri "http://gleaner.earthcube.org/" - 
    echo "--------end------------------"
done

