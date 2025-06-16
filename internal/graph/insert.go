package graph

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"

	log "github.com/sirupsen/logrus"
)

func Insert(g, nt, spql, username, password string, auth bool) (string, error) {

	p := "INSERT DATA { "
	pab := []byte(p)
	gab := []byte(fmt.Sprintf(" graph <%s>  { ", g))
	u := " } }"
	uab := []byte(u)
	pab = append(pab, gab...)
	pab = append(pab, []byte(nt)...)
	pab = append(pab, uab...)

	req, err := http.NewRequest("POST", spql, bytes.NewBuffer(pab)) // PUT for any of the servers?
	if err != nil {
		log.Error(err)
	}

	req.Header.Set("Content-Type", "application/sparql-update") // graphdb  blaze and jena  alt might be application/sparql-results+xml
	req.Header.Set("Accept", "application/x-trig")              // graphdb

	if auth {
		req.SetBasicAuth(username, password)
	}

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		log.Error(err)
	}
	defer resp.Body.Close()

	log.Tracef("response Status: %s", resp.Status)
	log.Tracef("response Headers: %s", resp.Header)
	// TODO just string check for 200 or 204 rather than try to match
	if resp.Status != "200 OK" && resp.Status != "204 No Content" && resp.Status != "204 " {
		log.Infof("response Status: %s", resp.Status)
		log.Infof("response Headers: %s", resp.Header)
	}

	body, err := ioutil.ReadAll(resp.Body)
	// log.Println(string(body))
	if err != nil {
		log.Error("response Body:", string(body))
		log.Error("response Status:", resp.Status)
		log.Error("response Headers:", resp.Header)
	}

	return resp.Status, err
}
