# Console

## About

Console is the start of building an intereface to submit jobs into gleaner with.

## Notes

console needs to 

Create configuration files for Gleaner and Nabu:
* init a config files for gleaner and nabu
  * generate templates to a config directory 
    * example mino, (tikka? use flag?).
    * example sources.csv
    * example gleaner
    * example nabu
* generate/update a config file for gleaner/nabu
  * merge mino and sources configurations
* check setup
  * read a mino config
  * read a csv file with headers to manage 'sources'
    * validate csv format
* pull, push, check local config to minio config

configuration building:

`glcon -cfginit -cfg name`
will create a directory in configs, with base configs to modify

`glcon -cfggen -cfg name`
will generate gleaner and nabu configurations, and make copies 


FUTURE (use cobra)

* glcon config init --config=X 
* glcon config generate --config=X
* glcon config validate --config=X

TBD
* glcon gleaner check --config=X  
* glcon gleaner run   (fire off a webhook call to the webhook listener in gleaner)
* glcon gleaner doctor
*  glcon gleaner newcfg --config=Y


## Refs

* https://farazdagi.com/2014/rest-and-long-running-jobs/ 
* Cobra Golang commander
