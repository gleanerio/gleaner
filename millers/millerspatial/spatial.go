package millerspatial

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/url"
	"strconv"
	"strings"
	"sync"

	"earthcube.org/Project418/crawler/framing"
	redis "gopkg.in/redis.v5"

	"github.com/coyove/jsonbuilder"
	minio "github.com/minio/minio-go"

	"earthcube.org/Project418/gleaner/millers/utils"
)

func ProcessBucketObjects(mc *minio.Client, bucketname string) {
	entries := utils.GetMillObjects(mc, bucketname)
	spatialMultiCall(entries)
}

func spatialMultiCall(e []utils.Entry) {
	// Set up the the semaphore and conccurancey
	semaphoreChan := make(chan struct{}, 20) // a blocking channel to keep concurrency under control
	defer close(semaphoreChan)
	wg := sync.WaitGroup{} // a wait group enables the main process a wait for goroutines to finish

	for k := range e {
		wg.Add(1)
		fmt.Printf("About to run #%d in a goroutine\n", k)
		go func(k int) {
			semaphoreChan <- struct{}{}

			status := SpatialIndexer(e[k].Urlval, e[k].Jld)

			wg.Done() // tell the wait group that we be done
			log.Printf("#%d done with %s", k, status)
			<-semaphoreChan
		}(k)
	}
	wg.Wait()
}

// SpatialIndexer indexes data in a spatial index
func SpatialIndexer(url, jsonld string) string {

	var skipGeoJSON = false
	var skipSchemaOrgPolygon = false
	var skipSchemaOrgBox = false
	var skipSchemaOrgLine = false
	var skipSchemaOrgCircle = true

	// TODO:
	//  implement a different spatial frame

	// generate the framed view
	sfr := framing.SpatialFrame(jsonld)

	// connection client
	client := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:9851",
	})

	var idToUse string

	var v string
	for geom := range sfr {

		// check that URL and ID are URLs and use ID preferred over URL
		// TODO..  clean up this bit-o-crap(tm) code block....
		if isValidURL(url) {
			idToUse = url
		}
		if isValidURL(sfr[geom].URL) {
			idToUse = sfr[geom].URL
		}
		if isValidURL(sfr[geom].ID) {
			idToUse = sfr[geom].ID
		}

		log.Printf("Got %s   Using %s \n", url, idToUse)

		if idToUse == "" {
			log.Printf("ERROR:  we have no ID in spatial indexer to use")
			break
		}

		// if GeoJSON was provided in spatial frame, index that.

		if sfr[geom].SpatialCoverage.SubjectOf.FileFormat == "application/vnd.geo+json" && sfr[geom].SpatialCoverage.SubjectOf.Text != "" {
			if !skipGeoJSON {
				log.Printf("Found GeoJSON: %s", sfr[geom].SpatialCoverage.SubjectOf.Text)
				geojson, err := processGeoJSON(client, idToUse, sfr[geom].SpatialCoverage.SubjectOf.Text)
				if err == nil {
					log.Printf("Indexed GEOJSON: %s", geojson)
					// Skip to the next
					continue
				} else {
					log.Println("ERROR: Could not index the GeoJSON:", err)
				}
			}
		}

		log.Printf("Geo Type: %s", sfr[geom].SpatialCoverage.Geo.Type)
		switch sfr[geom].SpatialCoverage.Geo.Type {
		case "GeoCoordinates":
			if sfr[geom].SpatialCoverage.Geo.Longitude != "" && sfr[geom].SpatialCoverage.Geo.Latitude != "" {
				log.Printf("Point: %s,%s", sfr[geom].SpatialCoverage.Geo.Longitude, sfr[geom].SpatialCoverage.Geo.Latitude)
				geojson, err := processSchemaOrgPoint(client, idToUse, sfr[geom].SpatialCoverage.Geo.Longitude, sfr[geom].SpatialCoverage.Geo.Latitude)
				if err == nil {
					log.Printf("Indexed schema:long,lat %s", geojson)
					// Skip to the next
					continue
				} else {
					log.Println("ERROR: Could not index the Point:", err)
				}
			}
			break

		case "GeoShape":
			// Process schema:polygon
			if !skipSchemaOrgPolygon {
				if sfr[geom].SpatialCoverage.Geo.Polygon != "" {
					log.Printf("Found schema:polygon %s", sfr[geom].SpatialCoverage.Geo.Polygon)
					geojson, err := processSchemaOrgPolygon(client, idToUse, sfr[geom].SpatialCoverage.Geo.Polygon)
					if err == nil {
						log.Printf("Indexed schema:polygon %s", geojson)
						// Skip to the next
						continue
					} else {
						log.Println("ERROR: Could not index the Polygon:", err)
					}
				}
			}

			// Process schema:box
			if !skipSchemaOrgBox {
				if sfr[geom].SpatialCoverage.Geo.Box != "" {
					log.Printf("Found schema:box %s", sfr[geom].SpatialCoverage.Geo.Box)
					geojson, err := processSchemaOrgBox(client, idToUse, sfr[geom].SpatialCoverage.Geo.Box)
					if err == nil {
						log.Printf("Indexed schema:box %s", geojson)
						// Skip to the next
						continue
					} else {
						log.Println("ERROR: Could not index the Box:", err)
					}
				}
			}

			// Process schema:line
			if !skipSchemaOrgLine {
				//if sfr[geom].SpatialCoverage.Geo.Line != "" {
				if sfr[geom].SpatialCoverage.Geo.Box != "" {
					log.Printf("Found schema:line %s", sfr[geom].SpatialCoverage.Geo.Box)
					geojson, err := processSchemaOrgLine(client, idToUse, sfr[geom].SpatialCoverage.Geo.Box)
					if err == nil {
						log.Printf("Indexed schema:line %s", geojson)
						// Skip to the next
						continue
					} else {
						log.Println("ERROR: Could not index the Line:", err)
					}
				}
			}

			if !skipSchemaOrgCircle {
				if sfr[geom].SpatialCoverage.Geo.Circle != "" {
					log.Printf("Circle: %s", sfr[geom].SpatialCoverage.Geo.Circle)
					geojson, err := processSchemaOrgLine(client, idToUse, sfr[geom].SpatialCoverage.Geo.Circle)
					if err == nil {
						log.Printf("Indexed schema:circle %s", geojson)
						// Skip to the next
						continue
					} else {
						log.Println("ERROR: Could not index the Circle:", err)
					}
				}
			}
			break
		}
	}

	client.Close()

	return fmt.Sprintf("SPATIALINDEXER: done with client report: %s", v)
}

func processGeoJSON(client *redis.Client, idToUse string, geojson string) (string, error) {
	// Try to unmarshall the JSON
	var gjson map[string]interface{}
	err := json.Unmarshal([]byte(geojson), &gjson)
	if err != nil {
		return "", err
	}

	cmd := redis.NewStringCmd("SET", "p418", idToUse, "OBJECT", geojson)
	err = client.Process(cmd)
	if err != nil {
		return "", err
	}

	_, err = cmd.Result()
	if err != nil {
		return "", err
	}

	return geojson, nil
}

func processSchemaOrgPolygon(client *redis.Client, idToUse string, polygon string) (string, error) {
	poly_result := strings.Split(polygon, " ")
	poly_point_count := len(poly_result)
	if poly_point_count < 4 {
		return "", errors.New("Polygon has less than 4 points. See https://schema.org/polygon")
	}

	json := jsonbuilder.Object()
	json.Set("type", "Polygon").Set("coordinates", jsonbuilder.Array(jsonbuilder.Array()))
	for i := range poly_result {
		polyCoordinate := strings.Split(poly_result[i], ",")
		x, err := strconv.ParseFloat(polyCoordinate[1], 64)
		if err != nil {
			return "", err
		}
		y, err := strconv.ParseFloat(polyCoordinate[0], 64)
		if err != nil {
			return "", err
		}

		// Add the coordinate to the GeoJSON
		poly_point := json.Enter("coordinates").Enter(0).Push(jsonbuilder.Array(x, y))
		json.Enter("coordinates").Set(0, poly_point)
	}

	geojson := json.Marshal()
	cmd := redis.NewStringCmd("SET", "p418", idToUse, "OBJECT", geojson)
	err := client.Process(cmd)
	if err != nil {
		return geojson, err
	}
	_, err = cmd.Result()
	if err != nil {
		return geojson, err
	}

	return geojson, nil
}

func processSchemaOrgBox(client *redis.Client, idToUse string, box string) (string, error) {
	// Box split on space. @see http://schema.org/box
	// box_result := strings.Split(box, " ")  // deprecated due to issues with commas with spaces..
	boxp := strings.Join(strings.Fields(box), " ")
	boxpp := strings.Replace(boxp, ", ", ",", -1)
	box_result := strings.Split(boxpp, " ")

	if len(box_result) != 2 {
		return "", errors.New("Box does not have 2 coordinates. See https://schema.org/box")
	}

	lowerleftpoint := strings.Split(box_result[0], ",")
	upperrightpoint := strings.Split(box_result[1], ",")
	if len(lowerleftpoint) != 2 || len(upperrightpoint) != 2 {
		return "", errors.New("Malformed box coordinates. See https://schema.org/box")
	}

	cmd := redis.NewStringCmd("SET", "p418", idToUse, "BOUNDS", lowerleftpoint[1], lowerleftpoint[0], upperrightpoint[1], upperrightpoint[0])
	err := client.Process(cmd)
	if err != nil {
		return box, err
	}
	_, err = cmd.Result()
	if err != nil {
		return box, err
	}

	return box, nil
}

func processSchemaOrgLine(client *redis.Client, idToUse string, line string) (string, error) {
	line_result := strings.Split(line, " ")
	line_point_count := len(line_result)
	if line_point_count < 2 {
		return "", errors.New("Line does not have at least 2 coordinates. See https://schema.org/line")
	}
	json := jsonbuilder.Object()
	json.Set("type", "LineString").Set("coordinates", jsonbuilder.Array())
	for i := range line_result {
		lineCoordinate := strings.Split(line_result[i], ",")
		x, err := strconv.ParseFloat(lineCoordinate[1], 64)
		if err != nil {
			return "", err
		}
		y, err := strconv.ParseFloat(lineCoordinate[0], 64)
		if err != nil {
			return "", err
		}
		line_point := json.Enter("coordinates").Push(jsonbuilder.Array(x, y))
		json.Set("coordinates", line_point)
	}
	linestring := json.Marshal()
	cmd := redis.NewStringCmd("SET", "p418", idToUse, "OBJECT", linestring)
	err := client.Process(cmd)
	if err != nil {
		return linestring, err
	}
	_, err = cmd.Result()
	if err != nil {
		return linestring, err
	}

	return linestring, nil
}

func processSchemaOrgCircle(client *redis.Client, idToUse string, circle string) (string, error) {
	return "", errors.New("Circle is not supported in Tile38 or GeoJSON")

}

func processSchemaOrgPoint(client *redis.Client, idToUse string, lon string, lat string) (string, error) {
	point := lon + "," + lat
	cmd := redis.NewStringCmd("SET", "p418", idToUse, "POINT", lat, lon)
	err := client.Process(cmd)
	if err != nil {
		return point, err
	}
	_, err = cmd.Result()
	if err != nil {
		return point, err
	}

	return point, nil
}

func isValidURL(toTest string) bool {
	_, err := url.ParseRequestURI(toTest)
	if err != nil {
		return false
	}
	return true
}
