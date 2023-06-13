module github.com/gleanerio/gleaner

// to update the nabu dependency used by glcon
// go get github.com/gleanerio/nabu@dev
// go mod tidy

go 1.19

require (
	github.com/PuerkitoBio/goquery v1.8.0
	github.com/approvals/go-approval-tests v0.0.0-20220530063708-32d5677069bd
	github.com/aws/aws-sdk-go v1.41.12
	github.com/chromedp/chromedp v0.6.5
	github.com/gocarina/gocsv v0.0.0-20211020200912-82fc2684cc48
	github.com/gorilla/mux v1.8.0
	github.com/gosuri/uiprogress v0.0.1
	github.com/knakk/rdf v0.0.0-20190304171630-8521bf4c5042
	github.com/mafredri/cdp v0.32.0
	github.com/minio/minio-go/v7 v7.0.52
	github.com/ohler55/ojg v1.14.5
	github.com/piprate/json-gold v0.5.0
	github.com/rs/xid v1.4.0
	github.com/schollz/progressbar/v3 v3.8.3
	github.com/spf13/cobra v1.5.0
	github.com/spf13/viper v1.12.0
	github.com/stretchr/testify v1.8.1
	github.com/valyala/fasttemplate v1.2.2
	github.com/xitongsys/parquet-go v1.6.0
	github.com/xitongsys/parquet-go-source v0.0.0-20211010230925-397910c5e371
	golang.org/x/crypto v0.6.0
)

require (
	cloud.google.com/go/compute v1.6.1 // indirect
	github.com/andybalholm/brotli v1.0.4 // indirect
	github.com/andybalholm/cascadia v1.3.1 // indirect
	github.com/apache/thrift v0.14.1 // indirect
	github.com/chromedp/cdproto v0.0.0-20210122124816-7a656c010d57 // indirect
	github.com/chromedp/sysutil v1.0.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/fsnotify/fsnotify v1.5.4 // indirect
	github.com/gobwas/httphead v0.1.0 // indirect
	github.com/gobwas/pool v0.2.1 // indirect
	github.com/gobwas/ws v1.0.4 // indirect
	github.com/golang-jwt/jwt/v4 v4.4.2 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/golang/snappy v0.0.3 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/googleapis/gax-go/v2 v2.4.0 // indirect
	github.com/gorilla/websocket v1.4.2 // indirect
	github.com/gosuri/uilive v0.0.4 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/compress v1.16.0 // indirect
	github.com/klauspost/cpuid/v2 v2.2.4 // indirect
	github.com/magiconair/properties v1.8.6 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mattn/go-isatty v0.0.14 // indirect
	github.com/mattn/go-runewidth v0.0.13 // indirect
	github.com/meilisearch/meilisearch-go v0.21.1 // indirect
	github.com/minio/md5-simd v1.1.2 // indirect
	github.com/minio/sha256-simd v1.0.0 // indirect
	github.com/mitchellh/colorstring v0.0.0-20190213212951-d06e56a500db // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/pelletier/go-toml v1.9.5 // indirect
	github.com/pelletier/go-toml/v2 v2.0.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/pquerna/cachecontrol v0.0.0-20180517163645-1555304b9b35 // indirect
	github.com/rivo/uniseg v0.2.0 // indirect
	github.com/spf13/afero v1.8.2 // indirect
	github.com/spf13/cast v1.5.0 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/subosito/gotenv v1.3.0 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.0 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasthttp v1.37.1-0.20220607072126-8a320890c08d // indirect
	go.opencensus.io v0.23.0 // indirect
	golang.org/x/net v0.7.0 // indirect
	golang.org/x/sys v0.5.0 // indirect
	golang.org/x/term v0.5.0 // indirect
	golang.org/x/text v0.7.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto v0.0.0-20220519153652-3a47de7e79bd // indirect
	google.golang.org/grpc v1.46.2 // indirect
	google.golang.org/protobuf v1.28.0 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

require (
	github.com/gleanerio/nabu v0.0.0-20230526171436-5ace63727afa
	github.com/orandin/lumberjackrus v1.0.1
	github.com/oxffaa/gopher-parse-sitemap v0.0.0-20191021113419-005d2eb1def4
	github.com/sirupsen/logrus v1.9.0
	github.com/temoto/robotstxt v1.1.2
	github.com/tidwall/gjson v1.14.4
	github.com/tidwall/sjson v1.2.5
	github.com/utahta/go-openuri v0.1.0
	golang.org/x/oauth2 v0.0.0-20220411215720-9780585627b5
	google.golang.org/api v0.81.0
	gopkg.in/natefinch/lumberjack.v2 v2.0.0 // indirect
)

// to update the nabu dependency used by glcon
// go get github.com/gleanerio/nabu@dev

// local replace. gleaner and nabu at same level
//replace  github.com/gleanerio/nabu v0.0.0-20211107193830-958398c3aaef  => "../nabu"
