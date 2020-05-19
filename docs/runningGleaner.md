# Notes for the screencast

## About

These are the notes I used while recording the screencast 
at: https://www.youtube.com/watch?v=12figImXgDk

## Steps


Get the files we need from the GitHub repo releases section
https://github.com/earthcubearchitecture-project418/gleaner/releases

Make and exporta directory for the data volume for the Docker containers.

```
mkdir /home/tmp/dv
export DATAVOL=/home/tmp/dv
```

We need to grab any context files we use and note their URL and file location
in the config file.   We are just using schema.org for now.   You can use other
contexts in the JSON-LD files and Gleaner will attempt to fetch them over the net.  This
is a slow process though and can result in many thousands of network calls.  By
placing the context here we can cache them locally for use.  As such, this 
is a highly recommended practice.  

You can fetch the current schema.org context with:

```bash
curl -L -H "Accept: application/ld+json" -H "Content-Type: application/ld+json" https://schema.org > jsonldcontext.jsonld
```

Minio client (or use your web browser)
Ref: https://docs.min.io/docs/minio-client-complete-guide.html

You can get the minio client for you OS at:
```
wget https://dl.min.io/client/mc/release/linux-amd64/mc
chmod 755 mc
./mc config host add minio http://0.0.0.0:9000 gleaneraccess gleanersecret --api S3v4
```

After running gleaner you can look for the output graphs and load the data into Jena

```bash
./mc cat local/gleaner/results/runid/samplesearth_graph.nq | curl -X POST --header "Content-Type:application/n-quads" -d @- http://localhost:3030/demo/data
```

