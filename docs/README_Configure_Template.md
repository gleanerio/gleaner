# Configure Using glcon and Templates

You do not need to have a container stack running to run    `glcon config`
But run `glcon gleaner` and `glcon nabu`, you will need to. 

## OVERVIEW glcon Configuration generation
`glcon config` is used to create configuration files for gleaner and nabu
The pattern is to intiialize a configuration directory, edit files, and generate new 
configuration files for gleaner and nabu. Inside a configuration, you will need to edit a localConfiguration file
Edit/add sources in a csv listing, and generate the configurations.

### initialize a configuration directory
use  glcon command can intialize a configuration directory, and allow for the generation of gleaner and nabu configurations

```
glcon config init -cfgName test
```

initializes a configuration in configs with name of 'test'
Inside you will find
```
test % ls
gleaner_base.yaml	readme.txt		sources.csv
nabu_base.yaml		localConfig.yaml 
README_Configure_Template.md
```

## EDIT the files

Usually, you will only need to edit the [localConfig.yaml](./localConfig.yaml) and sources.csv
The localConfig.yaml

#### localConfig.yaml
```yaml
---
minio:
  address: 0.0.0.0 # can be overridden with MINIO_ADDRESS
  port: 9000 # can be overridden with MINIO_PORT
  accessKey: worldsbestaccesskey # can be overridden with MINIO_ACCESS_KEY
  secretKey: worldsbestsecretkey # can be overridden with MINIO_SECRET_KEY
  ssl: false # can be overridden with MINIO_SSL
  bucket: gleaner # can be overridden with MINIO_BUCKET
sparql:
  endpoint: http://localhost/blazegraph/namespace/earthcube/sparql
s3:
  bucket: gleaner # sync with above... can be overridden with MINIO_BUCKET... get's zapped if it's not here.
  domain: us-east-1

#headless field in gleaner.summoner
headless: http://127.0.0.1:9222
sourcesSource:
   type: csv
   location: sources.csv
```

First, in the "mino:" section make sure the accessKey and secretKey here match the access keys for your minio.
These can be overridden with the environent variables:
* "MINIO_ACCESS_KEY"
* "MINIO_SECRET_KEY"

#### sourcesSource
[sources.csv](./sources_custom.csv) is utilized to generate the information needed to retrieve (summon), process (mill), upload (prefix), and cull old records (prune)
This is done in order to facilitate the managing a list of sources, in a spreadsheet, rather than a yamil file.
a csv file with the fields below


This is designed to be edited in a spreadsheet, or exported by url as csv from a google spreadsheet

```csv
hack,SourceType,Active,Name,ProperName,URL,Headless,Domain,PID,Logo,CredentialsFile
1,sitegraph,FALSE,aquadocs,AquaDocs,https://oih.aquadocs.org/aquadocs.json ,FALSE,https://aquadocs.org,http://hdl.handle.net/1834/41372,,
3,sitemap,TRUE,opentopography,OpenTopography,https://opentopography.org/sitemap.xml,FALSE,http://www.opentopography.org/,https://www.re3data.org/repository/r3d100010655,https://opentopography.org/sites/opentopography.org/files/ot_transp_logo_2.png,
,sitemap,TRUE,iris,IRIS,http://ds.iris.edu/files/sitemap.xml,FALSE,http://iris.edu,https://www.re3data.org/repository/r3d100010268,http://ds.iris.edu/static/img/layout/logos/iris_logo_shadow.png,
```

Fields: 
1. hack:a hack to make the fields are properly read.
2. SourceType : [sitemap, sitegraph, googledrive] type of source
3. Active: [TRUE,FALSE] is source active. 
4. Name: short name of source. It should be one word (no space) and be lower case.
5. ProperName: Long name of source that will be added to organization record for provenance
6. URL: URL of sitemap or sitegraph.
7. Headless: [FALSE,TRUE] should be set to false unless you know this site uses JavaScript to place the JSON-LD into the page.  This is true of some sites and it is supported but not currently auto-detected.  So you will need to know this and set it.  For most place, this will be false.
   if the json-ld is generated in a page dynamically, then use , TRUE
8. Domain: 
9. PID: a unique identifier for the source. Perfered that is is a research id.
10. Logo: while no longer used, logo of the source
11. CredentialsFile: (ONLY NEEDED FOR type:googledrive) environment variable pointing to a google api key.
12. any additional feilds you wish. This might be used to generate information about sources for a website.

#### Configuration of your source
You configure the source in the localConfig.yaml (or override via the command line)
```yaml
# looks for Mysources.csv in configuration directory
sourcesSource:
   type: csv
   location: Mysources.csv
```
This can also be a remote url starting with http:// or https://
```yaml
# pulls from a google sheet
sourcesSource:
   type: csv
   location: https://docs.google.com/spreadsheets/d/{key}/gviz/tq?tqx=out:csv&sheet={sheet_name}
```

If you start the name with a '/' a full path to the file is assumed.
```yaml
# file outside of local configuration directory
sourcesSource:
   type: csv
   location: /home/user/ourSources.csv
```

## GENERATE the configuration files
```
glcon generate -cfgName test
```
This will generate files  ['gleaner'](./gleaner) and  ['nabu'](./nabu) and make copies of the existing configuration files




## Some Gleaner Configuration details
This is a summary of a few portions of the configuration files generated.
More details are at: [GleanerConfiguration.md](../docs/GleanerConfiguration.md) 

Open the file ['gleaner'](./gleaner), and you will see  is actually quite a bit information this file, but for this starting demo only a few things we need to worry about.  The default file will look like:

```yaml
---
minio:
  address: 0.0.0.0
  port: 9000
  accessKey: worldsbestaccesskey
  secretKey: worldsbestsecretkey
  ssl: false
  bucket: gleaner
gleaner:
  runid: runX # this will be the bucket the output is placed in...
  summon: true # do we want to visit the web sites and pull down the files
  mill: true
context:
  cache: true
contextmaps:
  - prefix: "https://schema.org/"
    file: "./configs/schemaorg-current-https.jsonld"
  - prefix: "http://schema.org/"
    file: "./configs/schemaorg-current-https.jsonld"
summoner:
  after: ""      # "21 May 20 10:00 UTC"   
  mode: full  # full || diff:  If diff compare what we have currently in gleaner to sitemap, get only new, delete missing
  threads: 5
  delay:  # milliseconds (1000 = 1 second) to delay between calls (will FORCE threads to 1) 
  headless: http://127.0.0.1:9222  # URL for headless see docs/headless
millers:
  graph: true
# will be built from sources.csv
sources:
  - sourcetype: sitegraph
    name: aquadocs
    logo: ""
    url: https://oih.aquadocs.org/aquadocs.json
    headless: false
    pid: http://hdl.handle.net/1834/41372
    propername: AquaDocs
    domain: https://aquadocs.org
    active: false
  - sourcetype: sitemap
    name: opentopography
    logo: https://opentopography.org/sites/opentopography.org/files/ot_transp_logo_2.png
    url: https://opentopography.org/sitemap.xml
    headless: false
    pid: https://www.re3data.org/repository/r3d100010655
    propername: OpenTopography
    domain: http://www.opentopography.org/
    active: false
```

A few things we need to look at.

First, in the "mino:" section make sure the accessKey and secretKey here are the ones you utlize.
Note: blank these out, and used environment variables (TODO:Need to describe them)

Next, lets look at the "gleaner:" section.  We can set the runid to something.  This is the ID for a run and it allows you to later make different runs and keep the resulting graphs organized.  It can be set to any lower case string with no spaces. 

The miller and summon sections are true and we will leave them that way.  It means we want Gleaner to both fetch the resources and process (mill) them.  

Now look at the "miller:"  section when lets of pick what milling to do.   Currently it is set with only graph set to true.  Let's leave it that way for now.  This means Gleaner will only attempt to make graph and not also run validation or generate prov reports for the process.  

The final section we need to look at is the "sources:" section.   
Here is where the fun is.  While there are two types, sitegraph and sitemaps we will normally use sitemap type. 
There is a third type that involves configuring and pulling from a

A standard [sitemap](./SourceSitemap.md) is below:
```yaml
sources:
  - sourcetype: sitemap
      name: opentopography
      logo: https://opentopography.org/sites/opentopography.org/files/ot_transp_logo_2.png
      url: https://opentopography.org/sitemap.xml
      headless: false
      pid: https://www.re3data.org/repository/r3d100010655
      propername: OpenTopography
      domain: http://www.opentopography.org/
      active: true
```

A sitegraph 
```yaml
sources:
  - sourcetype: sitegraph
    name: aquadocs
    logo: ""
    url: https://oih.aquadocs.org/aquadocs.json
    headless: false
    pid: http://hdl.handle.net/1834/41372
    propername: AquaDocs
    domain: https://aquadocs.org
    active: false
```

A [google Drive](./SourceGoogleDrive.md)
```yaml
sources:
- sourcetype: googledrive
  name: ecrr_submitted
  logo: https://www.earthcube.org/sites/default/files/doc-repository/logo_earthcube_full_horizontal.png
  url: https://drive.google.com/drive/u/0/folders/1TacUQqjpBbGsPQ8JPps47lBXMQsNBRnd
  headless: false
  pid: ""
  propername: Earthcube Resource Registry
  domain: http://www.earthcube.org/resourceregistry/
  active: true
  credentialsfile: configs/credentials/gleaner-331805-030e15e1d9c4.json
  other: {}
```
These are the sources we wish to pull and process. 
Each source has a type, and 8 entries though at this time we no longer use the "logo" value. 
It was used in the past to provide a page showing all the sources and 
a logo for them.  However, that's really just out of scope for what we want to do. 
You can leave it blank or set it to any value, it wont make a difference.  

The name is what you want to call this source.  It should be one word (no space) and be lower case. 

The url value needs to point to the URL for the site map XML file.  This will be created and served by the data provider. 

The headless value should be set to false unless you know this site uses JavaScript to place the JSON-LD into the page.  This is true of some sites and it is supported but not currently auto-detected.  So you will need to know this and set it.  For most place, this will be false. 

You can have as many sources as you wish.  For an example look the configure file for the CDF Semantic Network at: https://github.com/gleanerio/CDFSemanticNetwork/blob/master/configs/cdf.yaml




