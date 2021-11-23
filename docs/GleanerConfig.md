# Gleaner Configuration file

This assumes that you have a container stack running

```
s3 store
triple store
headless
```
## Gleaner Configuration generation
Files can be generated using glcon. Described in [README_CONFIGURE_Template](./README_Configure_Template.md)

## Gleaner Configuration

When generated, the Gleaner configuration file named `gleaner` in the configs/{config} directory, but any name with a .yaml ending is acceptable.
There is actually quite a bit in this file, but for this starting demo only a few things we need to worry about.  The default file will look like:

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

First, in the "mino:" section make sure the accessKey and secretKey here match the ones you have and set via your demo.env file. 

Next, lets look at the "gleaner:" section.  We can set the runid to something.  This is the ID for a run and it allows you to later make different runs and keep the resulting graphs organized.  It can be set to any lower case string with no spaces. 

The miller and summon sections are true and we will leave them that way.  It means we want Gleaner to both fetch the resources and process (mill) them.  

Now look at the "miller:"  section when lets of pick what milling to do.   Currently it is set with only graph set to true.  Let's leave it that way for now.  This means Gleaner will only attempt to make graph and not also run validation or generate prov reports for the process.  

The final section we need to look at is the "sources:" section.   
Here is where the fun is.  While there are two types, sitegraph and sitemaps we will normally use sitemap type. 


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

A [Google Drive](./SourceGoogleDrive.md)
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




