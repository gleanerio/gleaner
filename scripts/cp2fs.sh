#!/bin/bash
# script to copy output to the file system
# usage:  ./cp2fs.sh local/test

mc_cmd() {
        mc ls $1 | awk '{print $5}'
}

for i in $(mc_cmd $1); do
    mc cp $1/$i $i.nq
done

