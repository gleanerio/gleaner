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

mc_dirlist() {
   mc ls ${BUCKET} | awk '{print $5}'
}

function mc_bucketlist {
   mc ls ${1} | awk '{print $6}'
}

# If you use this for ntriples, be sure to compute and/or add in a graph in the URL target
total=0
trpltotal=0
for i in $(mc_dirlist ${BUCKET}); do
    b=${BUCKET}/$i
    count=0
    triples=0
    for i in $(mc ls  ${b} | awk '{print $5}'); do  # 5 on iow, 6 here (mc version issue)
        #echo Next: $b$i
        let count++
        tcount=$(mc cat $b$i | jsonld format -q | wc -l)
        let triples=triples+tcount
        #       mc cat $1/$i | curl -X POST -H 'Content-Type:text/x-nquads' --data-binary  @- $2   #  For nquads source
    done
    string="${count} \t ${b}"
    echo -e "$string"
    echo ${triples} triple count
    let total=total+count
    let trpltotal=trpltotal+triples
done

echo -e "${total} \t total"
echo -e "${trpltotal} \t triple total"

