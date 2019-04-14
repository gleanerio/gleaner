#!/bin/bash
#a wrapper script for invoking 
mc_cmd() {
        mc ls local/gleaner-milled | awk '{print $5}'
}


for i in $(mc_cmd); do
    echo "$i"
    mc cat local/gleaner-milled/$i | jenaload
done

