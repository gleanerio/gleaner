# STAC Scanner

## TODO

* make the frame for the types to pull out the URLs I want
* need an in memory KV store or cache for the URLs to index.

##  About

This is just a test based on some discussion in DeCODER.  I will pull the top level DataCatalog.
It is only DataCatalogs and Items.

So, when we get this we should then be able to frame and send the results into a struct and then pull
the next URLs to index from the current data graph(s).   Process those and continue to add to the queue.

The JSON-LD can simply to sent to the object store like any other set of files.  