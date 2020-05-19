#!/bin/bash
mc_cmd() {
        mc ls $1 | awk '{print $5}'
}

for i in $(mc_cmd $1); do
    echo "--------start----------------"
    echo $i
    mc cat $1/$i | rapper --count --input ntriples --input-uri "http://gleaner.earthcube.org/" - 
    echo "--------end------------------"
done
