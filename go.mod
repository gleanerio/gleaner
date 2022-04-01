module github.com/gleanerio/gleaner

go 1.14

require (
	github.com/PuerkitoBio/goquery v1.8.0
	github.com/aws/aws-sdk-go v1.41.12
	github.com/chromedp/chromedp v0.6.5
	github.com/gocarina/gocsv v0.0.0-20211020200912-82fc2684cc48
	github.com/gorilla/mux v1.8.0
	github.com/gosuri/uiprogress v0.0.1
	github.com/knakk/rdf v0.0.0-20190304171630-8521bf4c5042
	github.com/mafredri/cdp v0.32.0
	github.com/minio/minio-go/v7 v7.0.15
	github.com/piprate/json-gold v0.4.1-0.20210813112359-33b90c4ca86c
	github.com/rs/xid v1.2.1
	github.com/schollz/progressbar/v3 v3.8.3
	github.com/spf13/cobra v1.2.1
	github.com/spf13/viper v1.9.0
	github.com/stretchr/testify v1.7.0
	github.com/xitongsys/parquet-go v1.6.0
	github.com/xitongsys/parquet-go-source v0.0.0-20211010230925-397910c5e371
	go.etcd.io/bbolt v1.3.6
	golang.org/x/crypto v0.0.0-20210817164053-32db794688a5
)

// just using bolt github.com/boltdb/bolt would be ok... but it complains if we mix.
//replace  github.com/boltdb/bolt v1.3.1 => "go.etcd.io/bbolt" v1.3.6

require (
	github.com/gleanerio/nabu v0.0.0-20211214151422-eda9e525f196
	github.com/oxffaa/gopher-parse-sitemap v0.0.0-20191021113419-005d2eb1def4
	github.com/samclarke/robotstxt v0.0.0-20171127213916-2817654b7988
	github.com/sirupsen/logrus v1.8.1
	github.com/tidwall/gjson v1.10.2
	github.com/tidwall/sjson v1.2.3
	github.com/utahta/go-openuri v0.1.0
	golang.org/x/oauth2 v0.0.0-20211005180243-6b3c2da341f1
	google.golang.org/api v0.60.0
)

// local replace. gleaner and nabu at same level
//replace  github.com/gleanerio/nabu v0.0.0-20211107193830-958398c3aaef  => "../nabu"
