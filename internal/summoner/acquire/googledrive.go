package acquire

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gleanerio/gleaner/internal/common"
	configTypes "github.com/gleanerio/gleaner/internal/config"
	"github.com/gleanerio/gleaner/internal/millers/graph"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"google.golang.org/api/googleapi"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"

	"github.com/minio/minio-go/v7"
)

const googleDriveType = "googledrive"

/*
Access files from a google drive.
Following keys in the source are needed:
  credentialsenv: GOOGLEAPIAUTH
  url: https://drive.google.com/drive/u/0/folders/1TacUQqjpBbGsPQ8JPps47lBXMQsNBRnd



credentialsenv:
This is a path and name to a file. There could be multiple services, each with different groups.
suggested path: configs/credentials/{file}.json

This should be a Service Credential JSON file, otherwise the original user needs to run glcon config gdrive.

Google Folder:
Files are associated with parent folder:
https://drive.google.com/drive/u/0/folders/1TacUQqjpBbGsPQ8JPps47lBXMQsNBRnd
googleparentfolderid is: 1TacUQqjpBbGsPQ8JPps47lBXMQsNBRnd
We put the full URL in the url field.

*/

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

//func getClientAPIKey(apiKey string) *http.Client {
//	// The file token.json stores the user's access and refresh tokens, and is
//	// created automatically when the authorization flow completes for the first
//	// time.
//	tokFile := "token.json"
//	tok, err := tokenFromFile(tokFile)
//	if err != nil {
//		tok = getTokenFromWeb(config)
//		saveToken(tokFile, tok)
//	}
//	return config.Client(context.Background(), tok)
//}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web %v", err)
	}
	return tok
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func GetDriveCredentials(authFilename string) (srv *drive.Service, err error) {
	ctx := context.Background()

	if authFilename == "" {
		b, err := ioutil.ReadFile("configs/credentials.json")
		//b, err := ioutil.ReadFile("configs/client_secret_255488082803-v2kja4qjaonb85gp8lv59333hpnt3n45.apps.googleusercontent.com.json")

		if err != nil {
			log.Printf("Unable to read client secret file: %v", err)
		}

		// If modifying these scopes, delete your previously saved token.json.
		config, err := google.ConfigFromJSON(b, drive.DriveMetadataReadonlyScope)
		if err != nil {
			log.Printf("Unable to parse client secret file to config: %v", err)
		}
		client := getClient(config)
		srv, err = drive.NewService(ctx, option.WithHTTPClient(client))

		srv.UserAgent = "EarthCube_DataBot/1.0"
		if err != nil {
			log.Printf("Unable to retrieve Drive client: %v", err)
		}
		return srv, err
	} else {
		// service credentials

		srv, err = drive.NewService(ctx, option.WithCredentialsFile(authFilename))
		if err != nil {
			log.Printf("Unable to retrieve Drive client: %v", err)
		}
		return srv, err
	}

	//r, err := srv.Files.List().PageSize(10).
	//	Fields("nextPageToken, files(id, name)").Do()
	//if err != nil {
	//	log.Fatalf("Unable to retrieve files: %v", err)
	//}
	//fmt.Println("Files:")
	//if len(r.Files) == 0 {
	//	fmt.Println("No files found.")
	//} else {
	//	for _, i := range r.Files {
	//		fmt.Printf("%s (%s)\n", i.Name, i.Id)
	//	}
	//}
}

func GetFileList(srv *drive.Service, parentFolder string, isSharedDrive bool, sharedDriveId string) ([]*drive.File,
	error) {
	var files []*drive.File
	var err error
	var parentQuery = fmt.Sprintf("'%s' in parents", parentFolder)
	pageToken := ""

	for {
		var q *drive.FilesListCall
		if isSharedDrive {

			q = srv.Files.List().Q(parentQuery).IncludeTeamDriveItems(true).DriveId(sharedDriveId).IncludeItemsFromAllDrives(true).SupportsTeamDrives(true).Corpora("drive").SupportsAllDrives(true).
				Fields("nextPageToken, files(id, name)")
			if err != nil {
				log.Fatalf("Unable to retrieve files: %v", err)
			}
		} else {
			//r, err = srv.Files.List().IncludeItemsFromAllDrives(true).SupportsAllDrives(true).PageSize(10).
			//	Fields("nextPageToken, files(id, name)").Do()
			q = srv.Files.List().Q(parentQuery).IncludeItemsFromAllDrives(true).SupportsAllDrives(true).
				Fields("nextPageToken, files(id, name)")

			if err != nil {
				log.Fatalf("Unable to retrieve files: %v", err)
			}
		}
		if pageToken != "" {
			q = q.PageToken(pageToken)
		}
		r, err := q.Do()
		if err != nil {
			fmt.Printf("An error occurred: %v\n", err)
			return files, err
		}
		files = append(files, r.Files...)
		pageToken = r.NextPageToken
		if pageToken == "" {
			break
		}
	}

	fmt.Println("Files:")
	if len(files) == 0 {
		fmt.Println("No files found.")
	} else {
		for _, i := range files {
			fmt.Printf("%s (%s)\n", i.Name, i.Id)
		}
	}
	return files, nil
}

/*
We get this:
but your computer or network may be sending automated queries. To protect our users, we can't process your request right now.
 https://support.google.com/websearch/answer/86640

<html><head><meta http-equiv="content-type" content="text/html; charset=utf-8"/><title>
Sorry...</title><style> body { font-family: verdana, arial, sans-serif; background-color: #fff; color: #000; }</style></head><body><div><table><tr><td><b><font face=sans-serif size=10><font color=#4285f4>G</font><font color=#ea4335>o</font><font color=#fbbc05>o</font><font color=#4285f4>g</font><font color=#34a853>l</font><font color=#ea4335>e</font></font></b></td><td style="text-align: left; vertical-align: bottom; padding-bottom: 15px; width: 50%"><div style="border-bottom: 1px solid #dfdfdf;">Sorry...</div></td></tr></table></div><div style="margin-left: 4em;"><h1>We're sorry...</h1><p>... but your computer or network may be sending automated queries. To protect our users, we can't process your request right now.</p></div><div style="margin-left: 4em;">See <a href="https://support.google.com/websearch/answer/86640">Google Help</a> for more information.<br/><br/></div><div style="text-align: center; border-top: 1px solid #dfdfdf;"><a href="https://www.google.com">Google Home</a></div></body></html>

*/
func GetFileFromGDrive(srv *drive.Service, fileId string) (*drive.File, string, error) {

	var fileContents string
	file, err := srv.Files.Get(fileId).Do()
	if e, ok := err.(*googleapi.Error); ok {
		log.Printf("Unable to retrieve info about file %s: %v", fileId, e)
		return file, "", e
	}
	count := 0
	for {
		fileResp, err := srv.Files.Get(fileId).Download()
		if e, ok := err.(*googleapi.Error); ok {
			if e.Code == 403 {
				if count > 10 {
					log.Printf("403 10 times, giving up : %v", e)
					return file, "", err
				} else {
					count = count + 1
					log.Printf("403 waiting : %v", e)
					time.Sleep(2 * time.Second)
				}

			} else {
				log.Printf("Unable to retrieve file %s: %v", file.Name, e)
				return file, "", e
			}
		} else {
			b, err := ioutil.ReadAll(fileResp.Body)
			if err != nil {
				log.Printf("Unable to convert downloaded fil%s: %v", file.Name, err)
				return file, "", err
			}
			fileContents = string(b)
			break
		}
	}
	return file, fileContents, nil
}

func GetFromGDrive(mc *minio.Client, v1 *viper.Viper) (string, error) {
	bucketName, err := configTypes.GetBucketName(v1) //miniocfg["bucket"] //   get the top level bucket for all of gleaner operations from config file
	// get the sitegraph entry from config file
	var domains []Sources
	//err := v1.UnmarshalKey("sitegraphs", &domains)

	sources, err := configTypes.GetSources(v1)
	if err != nil {
		log.Error(err)
	}
	domains = configTypes.GetActiveSourceByType(sources, googleDriveType)
	//var results []*drive.File
	var results []string
	for _, s := range domains {
		//serviceJson := os.Getenv(s.CredentialsFile) // just use separate files for all credentials
		serviceJson := s.CredentialsFile
		srv, err := GetDriveCredentials(serviceJson)
		if err != nil {
			log.Printf("googledrive api key access failed: %s : %s", s.Name, err)
			continue
		}
		u, _ := url.Parse(s.URL)
		fn := filepath.Base(u.Path)
		log.Printf("reading google folder id: %s", fn)
		l, err := GetFileList(srv, fn, false, "")

		for _, f := range l {
			//results = append(results,f)
			o, err := gfileProcessing(mc, v1, srv, f, s.Name, bucketName)
			if err != nil {
				continue
			}
			results = append(results, o)
		}
	}
	var count = len(results)
	m := fmt.Sprintf("GoogleDrives %d files processed", count)
	return m, err
}

func gfileProcessing(mc *minio.Client, v1 *viper.Viper, srv *drive.Service, f *drive.File, sourceName string, bucketName string) (string, error) {
	var fileId = f.Id
	_, contents, err := GetFileFromGDrive(srv, fileId)
	if err != nil {
		fmt.Printf("error with reading  JSON '%s' from google drive:%s ", f.Name, sourceName)
		return fileId, err
	}

	// TODO, how do we quickly validate the JSON-LD files to make sure it is at least formatted well

	sha := common.GetSHA(contents) // Don't normalize big files..

	// Upload the file
	log.Printf("  file %s downloaded. Uploading to %s: %s", f.Name, bucketName, sourceName)

	objectName := fmt.Sprintf("summoned/%s/%s.jsonld", sourceName, sha)
	_, err = graph.LoadToMinio(contents, bucketName, objectName, mc)
	if err != nil {
		return objectName, err
	}
	log.Printf(" file %s uploaded to %s. Uploaded : %s", f.Name, bucketName, sourceName)
	// mill the json-ld to nq and upload to minio
	// we bypass graph.GraphNG which does a time consuming blank node fix which is not required
	// when dealing with a single large file.
	// log.Print("Milling graph")
	//graph.GraphNG(mc, fmt.Sprintf("summoned/%s/", domains[k].Name), v1)
	proc, options := common.JLDProc(v1) // Make a common proc and options to share with the upcoming go funcs
	rdf, err := common.JLD2nq(contents, proc, options)
	if err != nil {
		return "", err
	}

	log.Printf("Processed files being uploaded to %s: %s", bucketName, sourceName)
	milledName := fmt.Sprintf("milled/%s/%s.rdf", sourceName, sha)
	_, err = graph.LoadToMinio(rdf, bucketName, milledName, mc)
	if err != nil {
		return objectName, err
	}
	log.Printf("Processed files Upload to %s complete: %s", milledName, sourceName)

	// build prov
	// log.Print("Building prov")
	err = StoreProvNG(v1, mc, sourceName, sha, sourceName, "summoned")
	if err != nil {
		return objectName, err
	}

	log.Printf("Loaded: %d", len(contents))
	return objectName, err
}
