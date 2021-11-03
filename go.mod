module github.com/gleanerio/gleaner

go 1.14

require (
	github.com/PuerkitoBio/goquery v1.8.0
	github.com/apache/thrift v0.14.1 // indirect
	github.com/araddon/dateparse v0.0.0-20210429162001-6b43995a97de
	github.com/aws/aws-sdk-go v1.41.12
	github.com/chromedp/chromedp v0.6.5
	github.com/gocarina/gocsv v0.0.0-20211020200912-82fc2684cc48
	github.com/gorilla/mux v1.8.0
	github.com/gosuri/uilive v0.0.4 // indirect
	github.com/gosuri/uiprogress v0.0.1
	github.com/knakk/rdf v0.0.0-20190304171630-8521bf4c5042
	github.com/mafredri/cdp v0.32.0
	github.com/minio/minio-go/v7 v7.0.15
	github.com/piprate/json-gold v0.4.0
	github.com/rs/xid v1.2.1
	github.com/schollz/progressbar/v3 v3.8.3
	github.com/spf13/cobra v1.2.1
	github.com/spf13/viper v1.9.0
	github.com/stretchr/testify v1.7.0
	github.com/xitongsys/parquet-go v1.6.0
	github.com/xitongsys/parquet-go-source v0.0.0-20211010230925-397910c5e371
	go.etcd.io/bbolt v1.3.2
	golang.org/x/crypto v0.0.0-20210817164053-32db794688a5
)

//replace (
//    github.com/gleanerio/gleaner/internal/config => "./internal/config"
//    github.com/gleanerio/gleaner/internal/objects => "./internal/objects"
//)

require (
	github.com/boltdb/bolt v1.3.1
	github.com/oxffaa/gopher-parse-sitemap v0.0.0-20191021113419-005d2eb1def4
)
