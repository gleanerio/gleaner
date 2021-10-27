package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/earthcubearchitecture-project418/gleaner/internal/common"
	"github.com/earthcubearchitecture-project418/gleaner/internal/millers"
	"github.com/earthcubearchitecture-project418/gleaner/internal/summoner"
	"github.com/gorilla/mux"
	"github.com/minio/minio-go/v7"
	"github.com/rs/xid"
	"github.com/spf13/viper"
)

var viperVal string

// var setupVal bool  // not implemented in web version yet

// Profile is a struct for holding results tot be sent to the
// user about the index request
type Profile struct {
	ID   string
	QLen int32
	Qpos int32
}

// MyServer struct for mux router
type MyServer struct {
	r *mux.Router
}

func init() {
	log.SetFlags(log.Lshortfile)
	// log.SetOutput(ioutil.Discard) // turn off all logging

	// flag.BoolVar(&setupVal, "setup", false, "Run Gleaner configuration check and exit")
	flag.StringVar(&viperVal, "cfg", "config", "Configuration file")
}

func main() {
	log.Println("EarthCube Gleaner (Web)")
	flag.Parse() // parse any command line flags...

	htmlRouter := mux.NewRouter()
	htmlRouter.PathPrefix("/").Handler(http.StripPrefix("/", http.FileServer(http.Dir("./static"))))
	http.Handle("/", &MyServer{htmlRouter})

	jp := mux.NewRouter()
	jp.HandleFunc("/job", MakeJob).Methods("POST")
	http.Handle("/job", &MyServer{jp})

	jg := mux.NewRouter()
	jg.HandleFunc("/job/{ID}", CheckJob).Methods("GET")
	http.Handle("/job/", &MyServer{jg})

	log.Printf("Listening on 9900. Go to http://127.0.0.1:9900/")
	err := http.ListenAndServe(":9900", nil)
	// http 2.0 http.ListenAndServeTLS(":443", "server.crt", "server.key", nil)
	if err != nil {
		log.Fatal(err)
	}
}

func readConfig(filename string, defaults map[string]interface{}) (*viper.Viper, error) {
	v := viper.New()
	for key, value := range defaults {
		v.SetDefault(key, value)
	}
	v.SetConfigName(filename)
	v.AddConfigPath(".")
	v.AutomaticEnv()
	err := v.ReadInConfig()
	return v, err
}

// MakeJob generates a new go fun c call for the indexing
func MakeJob(w http.ResponseWriter, r *http.Request) {
	var v1 *viper.Viper
	var err error
	guid := xid.New()

	// Viper config
	v1, err = readConfig("./config", map[string]interface{}{
		"sources": map[string]string{
			"name":     fmt.Sprintf("%s_samplesearth", guid.String()), // error..   will overwrite others..   guid+name
			"url":      "https://samples.earth/sitemap.xml",
			"headless": "false",
		},
		"gleaner": map[string]string{
			"runid":  fmt.Sprintf("%s_web", guid.String()),
			"summon": "true",
			"mill":   "true",
		},
	})
	if err != nil {
		panic(fmt.Errorf("error when reading config: %v", err))
	}

	log.Println(v1)

	//vars := mux.Vars(r)
	b, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	log.Println(string(b))

	// make the return struct
	profile := Profile{guid.String(), 1, 1}

	js, err := json.Marshal(profile)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	mc := common.MinioConnection(v1)
	go cli(mc, v1) // this is a hack.. need a semaphore system here...

	w.WriteHeader(http.StatusAccepted)
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

// func cli(mc *minio.Client, cs utils.Config) {
func cli(mc *minio.Client, v1 *viper.Viper) {
	mcfg := v1.GetStringMapString("gleaner")

	if mcfg["summon"] == "true" {
		summoner.Summoner(mc, v1)
	}

	if mcfg["mill"] == "true" {
		millers.Millers(mc, v1) // need to remove rundir and then fix the compile
	}
}

// CheckJob will return the status of a job
func CheckJob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["ID"]

	fmt.Fprintf(w, "Job check: %s\n", id)
}

func (s *MyServer) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	rw.Header().Set("Access-Control-Allow-Origin", "*")
	rw.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	rw.Header().Set("Access-Control-Allow-Headers",
		"Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

	// Let the Gorilla work
	s.r.ServeHTTP(rw, req)
}

func addDefaultHeaders(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		fn(w, r)
	}
}
