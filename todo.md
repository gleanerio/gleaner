# TODO

## About

This is a file with some next step items 

* Fix the graph builder from tika and add in upper level URI to point to
the file being indexed

* wrap the CLI into docker to make it easier to distribute  (ref:
https://spin.atomicobject.com/2015/11/30/command-line-tools-docker/)
Do the docker push to a container for just me...

* next, build out the web UI to do the same thing via the web intereface

Thoughts:

Three buckets...  
	* Queue  (load a config file)
	* Running   (one config file that is current)
	* Completed (collection of output somehow)  maybe a buket or buket "directory"

Use the webhook listening code demonstrated in /home/fils/src/go/src/whoi.edu/switchBoard

remove rundir
remove (unreference writerdf) like in prov


MINIO_ACCESS_KEY=AKIAIOSFODNN7EXAMPLE
MINIO_SECRET_KEY=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY

https://docs.minio.io/docs/minio-bucket-notification-guide.html#webhooks


