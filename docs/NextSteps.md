# Next steps

* Turn off writing graph to FS
* Bleve writes to memory file system and then copies to Minio

* Turn off (by flag?) all file based read writes...  (default off)

* Web UI for config file that allows editing..
* Saves config to minio..
* webhook in minio fires off to gleaner "dispatcher" that 
run one index at a time for as many "configs" in the configqueue bucket


So gleaner on "job" mode will simply fire up and look at configqueue bucket
and run the config files in there one at a time till done.  It doesn't
really need an IO..  it can write to stdout and it's shutdown could 
be triggered by ... special config?   seperate API endpoint?


