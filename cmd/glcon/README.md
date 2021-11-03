# Gleaner Console  (glcon)

## About

Console is the start of exploring building an interface to submit jobs into gleaner.

`glcon --help`


## Workflow
1. User initializes configuration directory: `glcon config init --cfg name`
2. user edits files in directory
3. user generates configuration: `glcon config generate --cfg name`
4. User runs gleaner setup: `glcon gleaner setup -- cfg name`
5. user runs gleaner: `glcon gleaner batch -cfg name`
6. User edits configuration, and regenerates configurations, and runs gleaner

### configuration building:

`glcon config init --cfgName name`
will create a directory in configs, with base configs to modify
configurtion is generated in configs/{name}
User will need to edit the servers.yaml, and sources.csv. They may also need to edit gleaner_base.yaml, and nabu.base.yaml

`glcon config generate --cfgName name`
will generate gleaner and nabu configurations, and make copies (one per day, for now)

The routine will merge information from servers.yaml with gleaner_base.yaml to create gleaner
The routine will merge information from servers.yaml with nabu_base.yaml to create nBU
The routine will use sources.csv to create a list of sources to process in the output of gleaner and nabu
Environmental variable substition will occur:
(need list of env avaiables)

### Executing Gleaner
`glcon gleaner setup --cfgName name`
Reads the gleaner configuration file, and checks if the s3 minio service is available,
and creates buckets if needed
Environmental variable substition will occur:
(need list of env avaiables)

`glcon gleaner batch --cfgName name`
Reads the gleaner configuration file, and executes gleaner.
Environmental variable substition will occur:
(need list of env avaiables)

##### env variables
	minioSubtress.BindEnv("address", "MINIO_ADDRESS")
	minioSubtress.BindEnv("port", "MINIO_PORT")
	minioSubtress.BindEnv("ssl", "MINIO_USE_SSL")
	minioSubtress.BindEnv("accesskey", "MINIO_ACCESS_KEY")
	minioSubtress.BindEnv("secretkey", "MINIO_SECRET_KEY")
	minioSubtress.BindEnv("secretkey", "MINIO_SECRET_KEY")
	minioSubtress.BindEnv("bucket", "MINIO_BUCKET")
	viperSubtree.BindEnv("domain", "S3_DOMAIN")
	viperSubtree.BindEnv("endpoint", "SPARQL_ENDPOINT")
	viperSubtree.BindEnv("authenticate", "SPARQL_AUTHENTICATE")
	viperSubtree.BindEnv("username", "SPARQL_USERNAME")
	viperSubtree.BindEnv("password", "SPARQL_PASSWORD")
	viperSubtree.BindEnv("headless", "GLEANER_HEADLESS_ENDPOINT")
	viperSubtree.BindEnv("threads", "GLEANER_THREADS")
	viperSubtree.BindEnv("mode", "GLEANER_MODE")
## Notes

console needs to:

Create configuration files for Gleaner and Nabu:
* init a config files for gleaner and nabu
  * generate templates to a config directory 
    * example servers and service, (tikka? use flag?).
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

* glcon config init --cfgName X 
* glcon config generate --cfgName X
* glcon config validate --cfgName X

TBD
* glcon gleaner check ---cfgName X  
* glcon gleaner run   (fire off a webhook call to the webhook listener in gleaner)
* glcon gleaner doctor
*  glcon gleaner newcfg --cfgName Y


## Refs

* https://farazdagi.com/2014/rest-and-long-running-jobs/ 
* Cobra Golang commander
