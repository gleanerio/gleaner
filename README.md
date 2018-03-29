## Gleaner

This is the new crawler for P418.  The old crawler is still around to address
some unique use cases still not addressed.  

The following two items were cultivated from the garden.  They are;


### Summoner

Fill in description and futher information.

### Millers

Fill in description and futher information.

### S3 commands

We use the minio s3 system.  To address listing and removing entries from the s3 system
use commands like

```
mc rm -r local/earthreforg/*
mc rm -r --force local/earthreforg


ls       List files and folders.
mb       Make a bucket or a folder.
cat      Display file and object contents.
pipe     Redirect STDIN to an object or file or STDOUT.
share    Generate URL for sharing.
cp       Copy files and objects.
mirror   Mirror buckets and folders.
find     Finds files which match the given set of parameters.
stat     Stat contents of objects and folders.
diff     List objects with size difference or missing between two folders or buckets.
rm       Remove files and objects.
events   Manage object notifications.
watch    Watch for file and object events.
policy   Manage anonymous access to objects.
session  Manage saved sessions for cp command.
config   Manage mc configuration file.
update   Check for a new software update.
version  Print version info.

```
