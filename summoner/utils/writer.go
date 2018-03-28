package utils

import (
	"log"
)

// MinioWrite writes a byte array to the system using sha1 file name
// todo, add in bucket based on domain name?
func MinioWrite(file []byte, name string) error {

	// TODO..  set up some bolt based KV systems to hold some metadata
	//  sha1 -> file name (URL)
	//  sha1 -> metadata  (or prov)

	log.Printf("Write file %s length %d to Minio S3", name, len(file))
	return nil

}
