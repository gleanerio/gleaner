package acquire

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gleanerio/gleaner/internal/common"
	configTypes "github.com/gleanerio/gleaner/internal/config"
	"github.com/gleanerio/gleaner/internal/millers/graph"
	"github.com/spf13/viper"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"

	"github.com/minio/minio-go/v7"
)

const googleDriveType = "googledrive"

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

func GetDriveCredentials(apiKey string) (srv *drive.Service, err error) {
	ctx := context.Background()

	if apiKey == "" {
		b, err := ioutil.ReadFile("configs/credentials.json")
		//b, err := ioutil.ReadFile("configs/client_secret_255488082803-v2kja4qjaonb85gp8lv59333hpnt3n45.apps.googleusercontent.com.json")

		if err != nil {
			log.Fatalf("Unable to read client secret file: %v", err)
		}

		// If modifying these scopes, delete your previously saved token.json.
		config, err := google.ConfigFromJSON(b, drive.DriveMetadataReadonlyScope)
		if err != nil {
			log.Fatalf("Unable to parse client secret file to config: %v", err)
		}
		client := getClient(config)
		srv, err = drive.NewService(ctx, option.WithHTTPClient(client))
		if err != nil {
			log.Fatalf("Unable to retrieve Drive client: %v", err)
		}
		return srv, err
	} else {
		srv, err = drive.NewService(ctx, option.WithAPIKey(apiKey))
		if err != nil {
			log.Fatalf("Unable to retrieve Drive client: %v", err)
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

func GetFileList(srv *drive.Service, parentFolder string, isSharedDrive bool, sharedDriveId string) (*drive.FileList,
	error) {
	var r *drive.FileList
	var err error
	var parentQuery = fmt.Sprintf("'%s' in parents", parentFolder)
	if isSharedDrive {
		r, err = srv.Files.List().Q(parentQuery).IncludeTeamDriveItems(true).DriveId(sharedDriveId).IncludeItemsFromAllDrives(true).SupportsTeamDrives(true).Corpora("drive").SupportsAllDrives(true).
			Fields("nextPageToken, files(id, name)").Do()
		if err != nil {
			log.Fatalf("Unable to retrieve files: %v", err)
		}
	} else {
		//r, err = srv.Files.List().IncludeItemsFromAllDrives(true).SupportsAllDrives(true).PageSize(10).
		//	Fields("nextPageToken, files(id, name)").Do()
		r, err = srv.Files.List().Q(parentQuery).IncludeItemsFromAllDrives(true).SupportsAllDrives(true).
			Fields("nextPageToken, files(id, name)").Do()

		if err != nil {
			log.Fatalf("Unable to retrieve files: %v", err)
		}
	}

	fmt.Println("Files:")
	if len(r.Files) == 0 {
		fmt.Println("No files found.")
	} else {
		for _, i := range r.Files {
			fmt.Printf("%s (%s)\n", i.Name, i.Id)
		}
	}
	return r, nil
}

func GetFileFromGDrive(srv *drive.Service, fileId string) (*drive.File, string, error) {
	file, err := srv.Files.Get(fileId).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve files: %v", err)
		return file, "", err
	}

	fileResp, err := srv.Files.Get(fileId).Download()
	if err != nil {
		log.Fatalf("Unable to retrieve files: %v", err)
		return file, "", err
	}
	b, err := ioutil.ReadAll(fileResp.Body)
	return file, string(b), nil

}

func GetFromGDrive(mc *minio.Client, v1 *viper.Viper) (string, error) {
	bucketName, err := configTypes.GetBucketName(v1) //miniocfg["bucket"] //   get the top level bucket for all of gleaner operations from config file

	// get the sitegraph entry from config file
	var domains []Sources
	//err := v1.UnmarshalKey("sitegraphs", &domains)

	sources, err := configTypes.GetSources(v1)
	if err != nil {
		log.Println(err)
	}
	domains = configTypes.GetActiveSourceByType(sources, googleDriveType)
	//var results []*drive.File
	var results []string
	for _, s := range domains {
		apiKey := os.Getenv(s.GoogleApiKeyEnv)
		if apiKey == "" {
			log.Fatalf("googledrive missing api key: %s", s.Name)
			continue
		}
		srv, err := GetDriveCredentials(apiKey)
		if err != nil {
			log.Fatalf("googledrive api key access failed: %s : %s", s.Name, err)
			continue
		}
		l, err := GetFileList(srv, s.GoogleParentFolderId, false, "")

		for _, f := range l.Files {
			//results = append(results,f)
			o, err := gfileProcessing(mc, v1, srv, f, s.Name, bucketName)
			if err != nil {
				continue
			}
			results = append(results, o)
		}
	}
	var count = len(results)
	m := fmt.Sprintf("GoogleDrives %f files processed", count)
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
