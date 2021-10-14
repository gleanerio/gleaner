package acquire

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"github.com/gleanerio/gleaner/internal/common"
	minio "github.com/minio/minio-go/v7"
	"github.com/spf13/viper"
)

func addToJsonListIfValid(v1 *viper.Viper, jsonlds []string, new_json string) ([]string, error) {
    valid, err := isValid(v1, new_json)
    if err != nil {
        return jsonlds, fmt.Errorf("error checking for valid json: %s", err)
    }
    if !valid {
        return jsonlds, fmt.Errorf("invalid json; continuing")
    }
    return append(jsonlds, new_json), nil
}


// Do we want to also validate JSON we get from headless browsers?
func isValid(v1 *viper.Viper, jsonld string) (bool, error) {
    proc, options := common.JLDProc(v1)

    var myInterface map[string]interface{}

    err := json.Unmarshal([]byte(jsonld), &myInterface)
    if err != nil {
        return false, fmt.Errorf("Error in unmarshaling json: %s", err)
    }

    _, err = proc.ToRDF(myInterface, options) // returns triples but toss them, just validating
    if err != nil {                           // it's wasted cycles.. but if just doing a summon, needs to be done here
        return false, fmt.Errorf("Error in JSON-LD to RDF call: %s", err)
    }

    return true, nil
}

func Upload(v1 *viper.Viper, mc *minio.Client, logger *log.Logger, bucketName string, site string,  urlloc string, jsonld string) {
	sha, err := common.GetNormSHA(jsonld, v1) // Moved to the normalized sha value
	if err != nil {
		logger.Printf("ERROR: URL: %s Action: Getting normalized sha  Error: %s\n", urlloc, err)
	}
	objectName := fmt.Sprintf("summoned/%s/%s.jsonld", site, sha)
	contentType := "application/ld+json"
	b := bytes.NewBufferString(jsonld)
	size := int64(b.Len()) // gets set to 0 after upload for some reason

	usermeta := make(map[string]string) // what do I want to know?
	usermeta["url"] = urlloc
	usermeta["sha1"] = sha

	// TODO
	// Make prov based on object name (org and object SHA)
	// DO this by writing a nanopub object to Minio..   then collect them up into a graph later...
	// I need:  re3 of source, url of json-ld, sha of jsonld, date
	// RESID  string SHA256 string RE3    string SOURCE string DATE   string
	// sha points to object
	// source to where I got it from
	// also need what I searc for and display as the URL to link to
	// need to revidw the subject and diurl I am using in the UI

	err = StoreProv(v1, mc, site, sha, urlloc)
	if err != nil {
		log.Println(err)
	}

	// Upload the file with FPutObject
	_, err = mc.PutObject(context.Background(), bucketName, objectName, b, int64(b.Len()), minio.PutObjectOptions{ContentType: contentType, UserMetadata: usermeta})
	if err != nil {
		logger.Printf("%s", objectName)
		logger.Fatalln(err) // Fatal?   seriously?    I guess this is the object write, so the run is likely a bust at this point, but this seems a bit much still.
	}
	logger.Printf("Uploaded Bucket:%s File:%s Size %d\n", bucketName, objectName, size)
}
