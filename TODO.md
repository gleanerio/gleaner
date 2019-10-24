# TODO for Release One


## TempFS issue

I need to pass a temp FS location in the config file which will be the 
local mount from the script.  So this is /gleaner/config

However, note in the config this should be blank if you are running the 
binary.  That will let the OS run the tmp file location which is better.  


## Config workflow

In the command line load the config object into the object store 
and read from there.   Load it the ID in gleaner-milled

Then read it from there.....
Use this approach to leverage from Docker but also for the web later...

## Near Term

[] Fix the process delete process in summoner with bucket prefix now in use
[] Buckets are really prefixes in gleaner in place..  but the var is bucket  (fix)
[] Log better...


## Items

* minio bucket validation (in gleaner)
* migrate to docker pull chromedp/headless-shell:latest and test
* migrate to pyshacl from topquadrant 
* migrate to yaml config from JSON
* load shacl from github?  or minio
* graphql UI based on various shapes 
* make the millers better functions in a package to improve reuse and further milling
* make a report of the results and more importantly, the errors
* WKT output from miller..   ref: https://github.com/twpayne/go-geom
* SHACL pulls from minio and initial test is for spatial element


New Summoner

ref: https://stackoverflow.com/questions/38654383/how-to-search-for-an-element-in-a-golang-slice 

* pull down the sitemap and place in object store first
* load the sitemap into a struct   (or should it be an in memory KV store)?
* compare the old sitemap with the new sitemap (if that is the case) 
* compare via structs or via KV pattern
* at some point we need to reconcile the objects existing to the sitemap.
* URL, LastModDate, sha256 of the data graph


New Object location.


## notes and such

Command line examples for gleaner in docker
```
docker run --rm -ti --network=host nsfearthcube/gleaner:2.0.6 -help 
docker run --rm -it --network=host nsfearthcube/gleaner:2.0.6 -setup -access=ACCESS -secret=SECRET
```


