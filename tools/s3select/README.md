# S3Select

## Notes

[minio select api quickstart guide](https://docs.min.io/docs/minio-select-api-quickstart-guide.html)

```
$ curl "https://population.un.org/wpp/Download/Files/1_Indicators%20(Standard)/CSV_FILES/WPP2019_TotalPopulationBySex.csv" > TotalPopulation.csv
$ mc mb myminio/mycsvbucket
$ gzip TotalPopulation.csv
$ mc cp TotalPopulation.csv.gz myminio/mycsvbucket/sampledata/
```