#!/bin/bash
# A wrapper script for loading RDF into Jena 
# usage:  load2Blaze.sh directory 
pushd $1

files=$( ls -1  *.nq | grep -v shacl | grep -v prov   )
counter=0
for i in $files ; do
      echo "-------------start-------------"
      echo Next: $i
      # rapper -e -c -i nquads $i 
      curl -X POST -H 'Content-Type:text/x-nquads' --data-binary @$i http://graph.openknowledge.network/blazegraph/namespace/kb/sparql
      echo "-------------done--------------"
done

popd

