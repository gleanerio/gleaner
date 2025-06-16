# Simple Test Script

## About

This is a simple test script to just do a few functional tests with the nabu command.

You will need to set up the variables at the top to point to:

* NABU = The location of nabu binary or the go command to run the code base
* CFG = The location of a valid config file
* BLAZE = For me this is just a SPARQL endpoint I can run a simple count SPARQL on and parse with jq
* GRAPHDB = The other triplestore I am testing

These last two are too specific, I can make them generic, but this is a rather bespoke script
anyway.  I doubt many would run this but you might inspect it.  

This script test two new flags

```
--dangerous
```

which is needed with the ```clear``` command to make sure you know this is a VERY DANGEROUS command.

```
--endpoint
```

which is used with the [new config file](../../config/example.yaml) node ```endpoints``` to provide a 
means to supply multiple SPARQL endpoints and to define SPARQL, SPARQL UPDATE and any 
possible BULK loaders an application might have via SPARQL over HTTP approaches.  

Example commands then look like, in this case for the ```prune``` command.  

```bash
nabu prune --cfg ./myconfig.yaml  --prefix summoned/test --endpoint graphdb
```
