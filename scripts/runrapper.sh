#!/bin/bash
mc_cmd() {
        mc ls local/gleaner-milled | awk '{print $5}'
}


for i in $(mc_cmd); do
    echo "$i"
    mc cat local/gleaner-milled/$i | rapper - -c -i ntriples http://example.org/base/
done

