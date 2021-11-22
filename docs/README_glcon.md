# Gleaner Console  (glcon)

## About

Console about merging tasks for gleaner and nabu into a single command. 
It creates configuration files, can setup the triplestore, run gleaner and run nabu

`glcon --help`


## Workflow
1. User initializes configuration directory: `glcon config init --cfgName name`
2. user edits files in directory 
3. user generates configuration: `glcon config generate --cfgName name`
4. User runs gleaner setup: `glcon gleaner setup --cfgName name`
5. user runs gleaner: `glcon gleaner batch --cfgName name`
6. user runs nabu prefix: `glcon nabu prefix --cfgName name`
7. user runs nabu prune: `glcon nabu prune --cfgName name`
8. User edits configuration, and regenerates configurations, runs gleaner and nabu

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

### Executing Nabu
`glcon nabu prefix --cfgName name`
Load graphs from prefix to triplestore

Environmental variable substition will occur:
(need list of env avaiables)

`glcon nabu  prune --cfgName name`
Prune graphs in triplestore not in objectVal store

Reads the gleaner configuration file, and executes gleaner.
Environmental variable substition will occur:
(need list of env avaiables)
##### Environment variables
	("minio.address", "MINIO_ADDRESS")
	("minio.port", "MINIO_PORT")
	("minio.ssl", "MINIO_USE_SSL")
	("minio.accesskey", "MINIO_ACCESS_KEY")
	("minio.secretkey", "MINIO_SECRET_KEY")
	("minio.bucket", "MINIO_BUCKET")
	("minio.domain", "S3_DOMAIN")
	("sparql.endpoint", "SPARQL_ENDPOINT")
	("sparql.authenticate", "SPARQL_AUTHENTICATE")
	("sparql.username", "SPARQL_USERNAME")
	("sparql.password", "SPARQL_PASSWORD")
	("gleaner.headless", "GLEANER_HEADLESS_ENDPOINT")
	("gleaner.threads", "GLEANER_THREADS")
	("gleaner.mode", "GLEANER_MODE")

## Notes

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

`glcon config init --cfgName name`
will create a directory in configs, with base configs to modify

`glcon config generate --cfgName name`
will generate gleaner and nabu configurations, and make copies 

## Configuration
* glcon config init --cfgName X 
* glcon config generate --cfgName X
* glcon config validate --cfgName X

## Gleaner
* glcon gleaner setup --cfgName X  
* glcon gleaner batch  --cfgName X  

## Nabu
* glcon nabu prefix --cfgName X
* glcon nabu prune --cfgName X
* glcon nabu object --cfgName X

## Refs

* https://farazdagi.com/2014/rest-and-long-running-jobs/ 
* Cobra Golang commander
