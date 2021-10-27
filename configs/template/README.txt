
Edit:
* sources
* mino

Edit the servers.yaml
add the cofiguraiton parameters for servers

edit sources.csv
* use a spreadsheet to edit


Possible environment Variables

	minioSubtress.BindEnv("address", "MINIO_ADDRESS")
	minioSubtress.BindEnv("port", "MINIO_PORT")
	minioSubtress.BindEnv("ssl", "MINIO_USE_SSL")
	minioSubtress.BindEnv("accesskey", "MINIO_ACCESS_KEY")
	minioSubtress.BindEnv("secretkey", "MINIO_SECRET_KEY")
	viperSubtree.BindEnv("endpoint", "SPARQL_ENDPOINT")
	viperSubtree.BindEnv("bucket", "S3_BUCKET")
	viperSubtree.BindEnv("domain", "S3_DOMAIN")