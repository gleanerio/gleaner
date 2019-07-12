# TODO for Release One

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

The 1m samples.earth summon took about 3 1/2 hours

