# Buckets

## NOTE

code isn't wired in or working for any purpose yet.  it's a placeholder for a plan

8-06-22
Perhaps create an archives prefix along with summoner and milled.
Then move

.../summoned/orgx
to
.../archived/orgx/[DATE]

This would be done for each sitegraph run and also for any full sitemap run, 
but not for diff sitemap runs.  



## About

Code here needs to manage the buckets that a crawl goes into. 

Buckets can be moved for archive reasons or simply purged.  

The sitemap.xml + prov graph does not tell us much really.  We don't
know if a DO has been updated without a hash.  We can not rely on 
the sitemap update date.  

On each index we can "honor" the sitemap and not index a 
resource in prov (from s3select calls) or "ignore" the 
sitemap and do a file index.  

We can "honor" for a time too.   N days for example.  

Config file section

update mode: honor   One of honor, ignore, age

The process is easy

ignore

- remove everything and index

Do we remove all objects?   or move to X.1  then run.  

honor 
Get the URLs from the sitemap, get the URLs form 
the s3select call on the prov bucket 

- URL in prov but not in sitemap?  remove it
- URL not in prov, but in sitemap?   get it (queue it)
- URL in prov and sitemap  ignore it

age
Like honor but ensure prov age > sitemap age before doing anything


