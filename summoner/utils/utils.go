package utils

import (
	"crypto/sha1"
	"net/url"
	"strings"
)

// Config struct
type Config struct {
	Minio struct {
		Endpoint        string `json:"endpoint"`
		AccessKeyID     string `json:"accessKeyID"`
		SecretAccessKey string `json:"secretAccessKey"`
	} `json:"minio"`
	Source string `json:"source"`
}

// DomainNameShort takes a URL, pulls them domain and removes the dots
func DomainNameShort(dn string) (string, string, error) {
	u, err := url.Parse(dn)
	if err != nil {
		panic(err)
	}

	// do we need to deal with port numbers from the host?
	// host, port, err := net.SplitHostPort(u.Host)
	// if err != nil {
	// 	log.Printf("Error parsing the domain name %v", err)
	// }

	// return host, port, u.Scheme, err

	// rewrite the host

	return strings.Replace(u.Host, ".", "", -1), u.Scheme, err
}

// GetSHA1 returns the sha1 string for the given byte array
func GetSHA1(b []byte) string {
	h := sha1.New()
	h.Write(b)
	bs := h.Sum(nil)
	return string(bs)
}
