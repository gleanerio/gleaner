
Edit:
* sources
* mino

Edit the servers.yaml
add the cofiguraiton parameters for servers

edit sources.csv
* use a spreadsheet to edit


Possible environment Variables

	gleaner("minio.address", "MINIO_ADDRESS")
	gleaner("minio.port", "MINIO_PORT")
	gleaner("minio.ssl", "MINIO_USE_SSL")
	gleaner("minio.accesskey", "MINIO_ACCESS_KEY")
	gleaner("minio.secretkey", "MINIO_SECRET_KEY")
	gleaner("sparql.endpoint", "SPARQL_ENDPOINT")
	gleaner("minio.bucket", "S3_BUCKET")
	nabu("minio.bucket", "S3_BUCKET")
	nabu("objects.bucket", "S3_BUCKET")
	nabu("objects.domain", "S3_DOMAIN")

Notes:

sources.csv does not include a sitegraph type (aquadocs is 100megs plus)
source_min.csv does include sitegraph