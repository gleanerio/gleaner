package check

import (
	"fmt"
	"log"
	"net/http"

	"github.com/minio/minio-go"
)

// GleanerSetup set up  elements of Gleaner are running

// GleanerCheck checks the setup
func GleanerCheck(mc *minio.Client) (bool, error) {
	syscheck(mc)

	// look for the buckets
	// todo make a loop on all the needed buckets
	found, err := mc.BucketExists("gleaner")
	if err != nil {
		log.Fatalln(err)
	}

	if found {
		log.Println("Gleaner Bucket found.")
	} else {
		log.Println("Gleaner Bucket not found.")
	}

	// look for the docker containers like tika, shacl

	fmt.Println(urlCheck())

	return true, err
}

// need to check, tika, shacl, headless
func urlCheck() bool {
	s := false

	resp, err := http.Get("http://localhost:7001") // 9998  7000 32772
	if err != nil {
		return s
	}

	if resp.Status == "200 OK" {
		s = true
	}
	return s
}

// syscheck is a place holder function for work to be done....
func syscheck(mc *minio.Client) {
	fmt.Println("System setup check placeholder")
	s := "valid"
	fmt.Printf("System check results: %s\n", s)
}
