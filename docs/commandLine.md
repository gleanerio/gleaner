# Command line notes

## About

A few command line examples for parsing sitemaps and testing them

```
curl http://opencoredata.org/sitemap.xml | grep -i dataset | awk -F'[<>]' '/loc/ {print $3}'
```
httpclient http://tangram.geodex.org/ucheck?url=http://opencoredata.org/id/dataset/063382f3-4a92-4d34-92e8-fb87781c6471&format=human&shape=required

xargs -0 wc -l  < <(tr \\n \\0 <urls.txt)

while read in; do chmod 755 "$in"; done < file.txt

httpclient "http://tangram.geodex.org/ucheck?url=http://opencoredata.org/id/dataset/5f475e43-6e23-44f6-821b-795d5f1f82d2&format=human&shape=required"

while read in; do echo "http://tangram.geodex.org/ucheck?url=$in&format=human&shape=required"; done < 100urls.txt

awk '{ print "chmod 755 "$0"" | "/bin/sh" }' file.txt


