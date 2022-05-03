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
	github.com/stretchr/testify v1.7.1
	github.com/xitongsys/parquet-go v1.6.0
	github.com/xitongsys/parquet-go-source v0.0.0-20211010230925-397910c5e371
	go.etcd.io/bbolt v1.3.6
	golang.org/x/crypto v0.0.0-20210921155107-089bfa567519
)

// just using bolt github.com/boltdb/bolt would be ok... but it complains if we mix.
//replace  github.com/boltdb/bolt v1.3.1 => "go.etcd.io/bbolt" v1.3.6

require (
	github.com/BurntSushi/toml v1.1.0 // indirect
	github.com/anmitsu/go-shlex v0.0.0-20200514113438-38f4b401e2be // indirect
	github.com/apache/thrift v0.16.0 // indirect
	github.com/gleanerio/nabu v0.0.0-20211214151422-eda9e525f196
	github.com/google/uuid v1.3.0 // indirect
	github.com/jessevdk/go-flags v1.5.0 // indirect
	github.com/kisielk/errcheck v1.6.0 // indirect
	github.com/m3db/prometheus_client_golang v0.8.1 // indirect
	github.com/m3db/prometheus_client_model v0.1.0 // indirect
	github.com/m3db/prometheus_common v0.1.0 // indirect
	github.com/m3db/prometheus_procfs v0.8.1 // indirect
	github.com/oxffaa/gopher-parse-sitemap v0.0.0-20191021113419-005d2eb1def4
	github.com/pborman/uuid v1.2.1 // indirect
	github.com/prometheus/common v0.33.0 // indirect
	github.com/samclarke/robotstxt v0.0.0-20171127213916-2817654b7988
	github.com/samuel/go-thrift v0.0.0-20210915161234-7b67f98e972f // indirect
	github.com/sirupsen/logrus v1.8.1
	github.com/tidwall/gjson v1.14.0
	github.com/tidwall/sjson v1.2.3
	github.com/twmb/murmur3 v1.1.6 // indirect
	github.com/uber-common/cadence-samples v0.0.0-20220107175756-7a5db1b8efd7 // indirect
	github.com/uber-go/tally v3.4.3+incompatible // indirect
	github.com/uber/tchannel-go v1.22.3 // indirect
	github.com/utahta/go-openuri v0.1.0
	go.uber.org/atomic v1.9.0 // indirect
	go.uber.org/cadence v0.19.0 // indirect
	go.uber.org/multierr v1.8.0 // indirect
	go.uber.org/net/metrics v1.3.1 // indirect
	go.uber.org/thriftrw v1.29.2 // indirect
	go.uber.org/yarpc v1.60.0 // indirect
	go.uber.org/zap v1.21.0 // indirect
	golang.org/x/exp/typeparams v0.0.0-20220414153411-bcd21879b8fd // indirect
	golang.org/x/net v0.0.0-20220412020605-290c469a71a5 // indirect
	golang.org/x/oauth2 v0.0.0-20220223155221-ee480838109b
	golang.org/x/sys v0.0.0-20220412211240-33da011f77ad // indirect
	golang.org/x/time v0.0.0-20220411224347-583f2d630306 // indirect
	golang.org/x/xerrors v0.0.0-20220411194840-2f41105eb62f // indirect
	google.golang.org/api v0.60.0
	google.golang.org/genproto v0.0.0-20220414192740-2d67ff6cf2b4 // indirect
	honnef.co/go/tools v0.3.0 // indirect
)

// local replace. gleaner and nabu at same level
//replace  github.com/gleanerio/nabu v0.0.0-20211107193830-958398c3aaef  => "../nabu"
