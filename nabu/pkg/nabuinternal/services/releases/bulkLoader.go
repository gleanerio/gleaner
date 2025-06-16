package releases

import (
	"fmt"
	"github.com/gleanerio/gleaner/nabu/pkg/nabuinternal/objects"
	log "github.com/sirupsen/logrus"

	"errors"
	"path"
	"strings"
	"time"

	"github.com/gleanerio/gleaner/nabu/pkg/config"

	"github.com/spf13/viper"

	"github.com/minio/minio-go/v7"
)

// BulkRelease collects the objects from a bucket to load
func BulkRelease(v1 *viper.Viper, mc *minio.Client) error {
	log.Println("Release:BulkAssembly")
	bucketName, _ := config.GetBucketName(v1)
	objCfg, _ := config.GetObjectsConfig(v1)
	pa := objCfg.Prefix

	var err error

	for p := range pa {
		sp := strings.Split(pa[p], "/")
		srcname := strings.Join(sp[1:], "__")
		spj := strings.Join(sp, "__")

		// Here we will either make this a _release.nq or a _prov.nq based on the source string.
		// TODO, should I look at the specific place in the path I expect this?
		// It is an exact match, so it should not be an issue
		name_latest := ""
		if contains(sp, "summoned") {
			name_latest = fmt.Sprintf("%s_release.nq", baseName(path.Base(srcname))) // ex: counties0_release.nq
		} else if contains(sp, "prov") {
			name_latest = fmt.Sprintf("%s_prov.nq", baseName(path.Base(srcname))) // ex: counties0_prov.nq
		} else if contains(sp, "orgs") {
			name_latest = fmt.Sprint("organizations.nq") // ex: counties0_org.nq
			fmt.Println(bucketName)
			fmt.Println(name_latest)
			err = objects.PipeCopy(v1, mc, name_latest, bucketName, "orgs", "graphs/latest") // have this function return the object name and path, easy to load and remove then
			if err != nil {
				return err
			}
			return err // just fully return from the function, no need for archive copies of the org graph
		} else {
			return errors.New("Unable to form a release graph name.  Path does not hold on of; summoned, prov or org")
		}

		// Make a release graph that will be stored in graphs/latest as {provider}_release.nq
		err = objects.PipeCopy(v1, mc, name_latest, bucketName, pa[p], "graphs/latest") // have this function return the object name and path, easy to load and remove then
		if err != nil {
			return err
		}

		// TODO  Should this be optional / controlled by flag?
		// Copy the "latest" graph just made to archive with a date
		// This means the graph in latests is a duplicate of the most recently dated version in archive/{provider}
		const layout = "2006-01-02-15-04-05"
		t := time.Now()
		// TODO  review the issue of archive and latest being hard coded.
		name := fmt.Sprintf("%s/%s/%s_%s_release.nq", "graphs/archive", srcname, baseName(path.Base(spj)), t.Format(layout))
		latest_fullpath := fmt.Sprintf("%s/%s", "graphs/latest", name_latest)
		err = objects.Copy(v1, mc, bucketName, latest_fullpath, bucketName, strings.Replace(name, "latest", "archive", 1))
	}

	return err
}

func baseName(s string) string {
	n := strings.LastIndexByte(s, '.')
	if n == -1 {
		return s
	}
	return s[:n]
}

func contains(array []string, str string) bool {
	for _, a := range array {
		if a == str {
			return true
		}
	}
	return false
}
