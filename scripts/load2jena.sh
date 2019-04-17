#!/bin/bash
#a wrapper script for invoking 
mc_cmd() {
        mc ls local/gleaner-milled | awk '{print $5}'
}

for i in $(mc_cmd); do
    echo "$i"
    mc cat local/gleaner-milled/$i | curl -X POST --header "Content-Type:application/n-triples" -d @- http://localhost:3030/test/data?graph=test
done

