package cli

import (
	"fmt"
	configTypes "github.com/gleanerio/gleaner/internal/config"
	nConfig "github.com/gleanerio/nabu/pkg/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"path"
	"strings"
	"time"
)

type Source = configTypes.Sources
type SourceConfig = configTypes.SourcesConfig
type MinoConfigType = configTypes.Minio

var Prov, Org bool

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "generate gleaner.io config files from a directory that has been intialized",
	Long: `Generate creates config files for the gleaner.io tools (gleaner and nabu). Before running command 
run 
# gleaner config init --confName {default: local}

Usually you will need to edit servers.yaml and sources.csv.
A copy of the files (one per DAY) is saved.
`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("generate called")
		//generateCfg(cfgPath, cfgName, sourcesVal) // sources moved into localConfig
		generateCfg(cfgPath, cfgName)
	},
}

func init() {
	configCmd.AddCommand(generateCmd)
	// Here you will define your flags and configuration settings
	generateCmd.Flags().BoolVar(&Prov, "prov", true, "include prov/source buckets in nabu conf")
	generateCmd.Flags().BoolVar(&Org, "org", true, "include orgs bucket in nabu conf")
	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// generateCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// generateCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// need to have more options.
func generateCfg(cfgPath string, cfgName string) error {
	var err error
	var minioCfg MinoConfigType
	var gleaner, nabu, servers *viper.Viper
	var sources []Source

	var configDir = path.Join(cfgPath, cfgName)

	servers, err = configTypes.ReadServersConfig(configBaseFiles["servers"], configDir)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	gleaner, err = configTypes.ReadGleanerConfig(configBaseFiles["gleaner"], configDir)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	nabu, err = nConfig.ReadNabuConfig(configBaseFiles["nabu"], configDir)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	sourcesVal := servers.GetString("sourcesSource.location")
	sources, err = configTypes.ReadSourcesCSV(sourcesVal, configDir)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	//var mi interface{}
	var date string
	currentTime := time.Now()
	date = currentTime.Format("20060102")
	// sources
	// need a check to see if it is an absolute path, so read not needed, and
	if !(strings.HasPrefix(sourcesVal, "https://") || strings.HasPrefix(sourcesVal, "http://")) {
		fmt.Println("make copy of sources")
		var original = path.Join(configDir, sourcesVal)
		if strings.HasPrefix(sourcesVal, "/") { // how do we test filesystem agnostic
			original = sourcesVal
		}
		var config = path.Join(configDir, date+sourcesVal)
		_, err = copy(original, config)
		if err != nil {
			panic(fmt.Errorf("error when copying sources: %v", err))
		}
	}

	fmt.Println("make copy of servers.yaml")
	original := path.Join(configDir, configBaseFiles["servers"])
	config := path.Join(configDir, date+configBaseFiles["servers"])
	_, err = copy(original, config)
	if err != nil {
		panic(fmt.Errorf("error when copying servers.yaml: %v", err))
	}
	//****** READ SERVERS CONFIG FILE ***
	// load minio
	//mi =  servers.Get("minio")
	//no idea why the unmarshall is not working
	//basically means env substitution needs to be handled by us
	// frig frig... do not use lowercase... those are private variables
	var ms = servers.Sub("minio")
	//s.BindEnv("address", "MINIO_ADDRESS")
	//s.BindEnv("p
	//ort", "MINIO_PORT")
	//s.BindEnv("ssl", "MINIO_USE_SSL")
	//s.BindEnv("accesskey", "MINIO_ACCESS_KEY")
	//s.BindEnv("secretkey", "MINIO_SECRET_KEY")
	//s.BindEnv("bucket", "MINIO_BUCKET")
	//s.AutomaticEnv()
	//err = s.Unmarshal( &minioCfg)
	minioCfg, err = configTypes.ReadMinioConfig(ms)
	if err != nil {
		panic(fmt.Errorf("error when writing config: %v", err))
	}
	sparqlSub := servers.Sub("sparql")
	sparqlCfg, err := configTypes.ReadSparqlConfig(sparqlSub)

	s3Sub := servers.Sub("s3")
	s3Cfg, err := configTypes.ReadS3Config(s3Sub)
	//s3Cfg.Bucket =  servers.GetString("minio.bucket")
	//s3Cfg := servers.Sub("s3")
	//servers.Set("s3.bucket", servers.GetString("minio.bucket"))

	// since not fully defined in mapping. things are missing
	//hdlsCfg  :=  servers.Get("headless")

	fmt.Println("Regnerate gleaner")
	gleaner.SetConfigType("yaml")
	var fn = path.Join(configDir, date+gleanerFileNameBase) // copy previous
	err = gleaner.WriteConfigAs(fn)
	if err != nil {
		panic(fmt.Errorf("error when writing config: %v", err))
	}

	gleaner.Set("minio", minioCfg)
	gleaner.Set("sources", sources)

	//gleaner.Set("summoner.headless", hdlsCfg) // since not fully defined in mapping. things are missing

	// hack to get rid of the sourcetype
	//err =  gleaner.UnmarshalKey("sitemaps", &sm)
	//gleaner.Set("sitemaps", sm)
	fn = path.Join(configDir, gleanerFileNameBase)
	err = gleaner.WriteConfigAs(fn)
	if err != nil {
		panic(fmt.Errorf("error when writing config: %v", err))
	}

	fmt.Println("Regnerate nabu")
	nabu.SetConfigType("yaml")
	fn = path.Join(configDir, date+nabuFilenameBase) // copy previous
	err = nabu.WriteConfigAs(fn)
	if err != nil {
		panic(fmt.Errorf("error when writing config: %v", err))
	}
	nabu.Set("minio", minioCfg)
	nabu.Set("sparql", sparqlCfg)
	//nabu.Set("objects", s3Cfg)
	////nabu.Set("objects", servers.Get("s3"))
	//var prefix []string
	//for _, s := range sources {
	//	if s.Active {
	//		prefix = append(prefix, s.Name)
	//	}
	//}
	//nabu.Set("objects.prefix", prefix)
	//var prefixOff []string
	//for _, s := range sources {
	//	if !s.Active {
	//		prefixOff = append(prefixOff, s.Name)
	//	}
	//}
	//nabu.Set("objects.prefixOff", prefixOff)

	//nabu.Set("objects", servers.Get("s3"))
	//var prefix []string
	//for _, s := range sources {
	//	if s.Active {
	//		prefix = append(prefix, s.Name)
	//	}
	//}
	//s3Cfg.Prefix= prefix
	var prefixSources []Source

	for _, s := range sources {
		if s.Active {
			prefixSources = append(prefixSources, s)
		}
	}
	s3Cfg.Prefix = configTypes.SourceToNabuPrefix(prefixSources, Prov)
	if Org {
		s3Cfg.Prefix = append(s3Cfg.Prefix, "org") // TODO: add flags for prov and or
	}

	//var prefixOff []string
	//for _, s := range sources {
	//	if !s.Active {
	//		prefixOff = append(prefixOff, s.Name)
	//	}
	//}
	//s3Cfg.PrefixOff =prefixOff
	var prefixOffSources []Source
	for _, s := range sources {
		if !s.Active {
			prefixOffSources = append(prefixOffSources, s)
		}
	}
	s3Cfg.PrefixOff = configTypes.SourceToNabuPrefix(prefixOffSources, Prov)
	nabu.Set("objects", s3Cfg)
	//nabu.Set("sitemaps", sources)
	//// hack to get rid of the sourcetype
	//err =  nabu.UnmarshalKey("sitemaps", &sm)
	//nabu.Set("sitemaps", sm)
	fn = path.Join(configDir, nabuFilenameBase)
	err = nabu.WriteConfigAs(fn)
	if err != nil {
		panic(fmt.Errorf("error when writing config: %v", err))
	}
	return err
}
