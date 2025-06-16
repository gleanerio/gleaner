#!/bin/bash

POSITIONAL_ARGS=()

while [[ $# -gt 0 ]]; do
    case $1 in
        -b|--bucket)
            BUCKET="$2"
            shift # past argument
            shift # past value
        ;;
        -s|--sparqlurl)
            SPARQL="$2"
            shift # past argument
            shift # past value
        ;;
        --default)
            DEFAULT=YES
            shift # past argument
        ;;
        -*|--*)
            echo "Unknown option $1"
            exit 1
        ;;
        *)
            POSITIONAL_ARGS+=("$1") # save positional arg
            shift # past argument
        ;;
    esac
done

set -- "${POSITIONAL_ARGS[@]}" # restore positional parameters

echo "S3 BUCKET  = ${BUCKET}"
echo "SPARQL URL = ${SPARQL}"
echo "DEFAULT    = ${DEFAULT}"

mc_cmd() {
   mc ls ${BUCKET} | awk '{print $6}'
}

# If you use this for ntriples, be sure to compute and/or add in a graph in the URL target
for i in $(mc_cmd ${BUCKET}); do
    echo "-------------start-------------"
    echo Next: $i
    # mc cat $1/$i | jsonld format -q | curl -X POST -H 'Content-Type:text/x-nquads' --data-binary  @- $2
    #       mc cat $1/$i | curl -X POST -H 'Content-Type:text/x-nquads' --data-binary  @- $2   #  For nquads source
    echo "-------------done--------------"
done

