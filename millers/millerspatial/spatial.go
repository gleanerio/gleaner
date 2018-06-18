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
	geojson "github.com/paulmach/go.geojson"

	"earthcube.org/Project418/gleaner/utils"
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
		log.Printf("About to run #%d in a goroutine\n", k)
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
	var skipSchemaOrgLine = false
	var skipSchemaOrgCircle = true
	var skipSchemaOrgBox = false

	// TODO:
	//  implement a different spatial frame
	sfr := framing.SpatialFrame(jsonld)

	// connection client
	client := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:9851",
	})

	var idToUse string

	featureCollection := geojson.NewFeatureCollection()

	// bleve indexes on the URL of the landing page
	// spatial  indexes on the URL -> schema:URL  -> schema:ID   (which ever it finds last...)
	// graph  indexes on  @ID  if not present..  it's a blank node...

	for _, frame := range sfr {
		log.Printf("\nFRAME: %+v\n", frame)
		// check that URL and ID are URLs and use ID preferred over URL
		// TODO..  clean up this bit-o-crap(tm) code block....
		if isValidURL(url) {
			idToUse = url
		}
		// if isValidURL(frame.URL) {
		// 	idToUse = frame.URL
		// }
		// if isValidURL(frame.ID) {
		// 	idToUse = frame.ID
		// }

		if idToUse == "" {
			log.Printf("ERROR:  we have no ID in spatial indexer to use")
			break
		}

	SpatialCoverages:
		for _, coverage := range frame.SpatialCoverages {
			//log.Printf("\nCOVERAGE: %+v\n", coverage)

			// if GeoJSON was provided in spatial frame, index that.
			for _, subjectOf := range coverage.SubjectOf {
				if subjectOf.FileFormat == "application/vnd.geo+json" && subjectOf.Text != "" {
					if !skipGeoJSON {
						//log.Printf("Found GeoJSON: %s", subjectOf.Text)
						geojson, err := addGeoJSON(featureCollection, idToUse, []byte(subjectOf.Text))
						if err == nil {
							log.Printf("Indexed GEOJSON: %s", geojson)
							// Skip to the next
							continue SpatialCoverages
						} else {
							// log.Println("Could not index the GeoJSON:", err)
							log.Printf("ERROR: id:%s  on the geojson: %v", idToUse, err)
						}
					}
				}
			}

		SpatialCoverageGeo:
			for _, geo := range coverage.Geos {
				switch geo.Type {

				case "GeoCoordinates":
					if geo.Longitude != "" && geo.Latitude != "" {
						//log.Printf("Point: %s,%s", geo.Longitude, geo.Latitude)
						geojson, err := addSchemaOrgPoint(featureCollection, idToUse, geo.Longitude, geo.Latitude)
						if err == nil {
							log.Printf("Indexed schema:lat,lon %s", geojson)
							// Skip to the next
							continue SpatialCoverageGeo
						} else {
							// log.Println("Could not index the Point:", err)
							log.Printf("ERROR: id:%s  on geom POINT: %v", idToUse, err)
						}
					}
					break

				case "GeoShape":
					// Process schema:polygon
					if !skipSchemaOrgPolygon {
						if geo.Polygon != "" {
							//log.Printf("Found schema:polygon %s", coverage.Geo.Polygon)
							geojson, err := addSchemaOrgPolygon(featureCollection, idToUse, geo.Polygon)
							if err == nil {
								log.Printf("Indexed schema:polygon %s", geojson)
								// Skip to the next
								continue SpatialCoverageGeo
							} else {
								// log.Println("Could not index the Polygon:", err)
								log.Printf("ERROR: id:%s  on geom POLYGON: %v", idToUse, err)
							}
						}
					}

					// Process schema:line
					if !skipSchemaOrgLine {
						if geo.Line != "" {
							//log.Printf("Found schema:line %s", coverage.Geo.Box)
							geojson, err := addSchemaOrgLine(featureCollection, idToUse, geo.Line)
							if err == nil {
								log.Printf("Indexed schema:line %s", geojson)
								// Skip to the next
								continue SpatialCoverageGeo
							} else {
								// log.Println("Could not index the Line:", err)
								log.Printf("ERROR: id:%s  on geom LINE: %v", idToUse, err)
							}
						}
					}

					if !skipSchemaOrgCircle {
						if geo.Circle != "" {
							//log.Printf("Circle: %s", coverage.Geo.Circle)
							geojson, err := addSchemaOrgCircle(featureCollection, idToUse, geo.Circle)
							if err == nil {
								log.Printf("Indexed schema:circle %s", geojson)
								// Skip to the next
								continue SpatialCoverageGeo
							} else {
								// log.Println("Could not index the Circle:", err)
								log.Printf("ERROR: id:%s  on geom CIRCLE: %v", idToUse, err)
							}
						}
					}

					// Process schema:box
					if !skipSchemaOrgBox {
						if geo.Box != "" {
							//log.Printf("Found schema:box %s", geo.Box)
							geojson, err := addSchemaOrgBox(featureCollection, idToUse, geo.Box)
							if err == nil {
								log.Printf("Indexed schema:box %s", geojson)
								// Skip to the next
								continue SpatialCoverageGeo
							} else {
								// log.Println("Could not index the Box:", err)
								log.Printf("ERROR: id:%s  on geom BOX: %v", idToUse, err)
							}
						}
					}
					break
				}
			}
			// End of SpatialCoverageGeo loop
		}
		// End of SpatialCoverages loop

		fc, err := featureCollection.MarshalJSON()
		if err != nil {
			log.Println("Could not marshall the GeoJSON for %s", idToUse)
		} else {
			geojson, err := processGeoJSON(client, idToUse, string(fc))
			if err == nil {
				log.Printf("Indexed geometries %s", geojson)
			} else {
				log.Println("Could not index the GeoJSON: %s -> %s", fc, err)
			}
		}
	}
	client.Close()

	return fmt.Sprintf("SPATIALINDEXER: done with client report: %s", url)
}

/**
 * Insert some GeoJSON into the Redis client with the given identifier
 */
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

/**
 * Given some GeoJSON string, add it to a GeoJSON FeatureCollection
 */
func addGeoJSON(fc *geojson.FeatureCollection, idToUse string, gjb []byte) (string, error) {
	geom, err := geojson.UnmarshalGeometry(gjb)
	if err != nil {
		return "", err
	}

	switch geom.Type {
	case "Point", "MultiPoint", "LineString", "MultiLineString", "Polygon", "MultiPolygon", "GeometryCollection":
		fc.AddFeature(geojson.NewFeature(geom))
		break

	case "Feature":
		feature, ferr := geojson.UnmarshalFeature(gjb)
		if ferr != nil {
			return "", err
		}
		fc.AddFeature(feature)
		break

	case "FeatureCollection":
		fcoll, fcerr := geojson.UnmarshalFeatureCollection(gjb)
		if fcerr != nil {
			return "", fcerr
		}
		for _, feature := range fcoll.Features {
			fc.AddFeature(feature)
		}
		break

	default:
		return "", errors.New("UNKNOWN GeoJSON Type: " + string(geom.Type))
	}

	return string(gjb), nil
}

/**
 * Given a schema:polygon value, add it to a GeoJSON FeatureCollection
 */
func addSchemaOrgPolygon(fc *geojson.FeatureCollection, idToUse string, polygon string) (string, error) {
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
	return addGeoJSON(fc, idToUse, []byte(geojson))
}

/**
 * Given a schema:box value, add it to a GeoJSON FeatureCollection
 */
func addSchemaOrgBox(fc *geojson.FeatureCollection, idToUse string, box string) (string, error) {
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

	// Convert the points to floats
	southwest_lon, err := strconv.ParseFloat(lowerleftpoint[0], 64)
	if err != nil {
		return "", err
	}
	southwest_lat, err := strconv.ParseFloat(lowerleftpoint[1], 64)
	if err != nil {
		return "", err
	}
	northeast_lon, err := strconv.ParseFloat(upperrightpoint[0], 64)
	if err != nil {
		return "", err
	}
	northeast_lat, err := strconv.ParseFloat(upperrightpoint[1], 64)
	if err != nil {
		return "", err
	}

	// Build the GeoJSON
	json := jsonbuilder.Object()
	json.Set("type", "Polygon").Set("coordinates", jsonbuilder.Array(jsonbuilder.Array()))
	southwest := json.Enter("coordinates").Enter(0).Push(jsonbuilder.Array(southwest_lon, southwest_lat))
	json.Enter("coordinates").Set(0, southwest)
	northwest := json.Enter("coordinates").Enter(0).Push(jsonbuilder.Array(northeast_lon, southwest_lat))
	json.Enter("coordinates").Set(0, northwest)
	northeast := json.Enter("coordinates").Enter(0).Push(jsonbuilder.Array(northeast_lon, northeast_lat))
	json.Enter("coordinates").Set(0, northeast)
	southeast := json.Enter("coordinates").Enter(0).Push(jsonbuilder.Array(southwest_lon, northeast_lat))
	json.Enter("coordinates").Set(0, southeast)
	closing_point := json.Enter("coordinates").Enter(0).Push(jsonbuilder.Array(southwest_lon, southwest_lat))
	json.Enter("coordinates").Set(0, closing_point)
	geojson := json.Marshal()
	return addGeoJSON(fc, idToUse, []byte(geojson))
}

/**
 * Given a schema:line value, add it to a GeoJSON FeatureCollection
 */
func addSchemaOrgLine(fc *geojson.FeatureCollection, idToUse string, line string) (string, error) {
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
	geojson := json.Marshal()
	return addGeoJSON(fc, idToUse, []byte(geojson))
}

/**
 * Given a schema:circle value, add it to a GeoJSON FeatureCollection
 */
func addSchemaOrgCircle(fc *geojson.FeatureCollection, idToUse string, circle string) (string, error) {
	return "", errors.New("Circle is not supported in Tile38 or GeoJSON")

}

/**
 * Given a latitude and longitude as strings, add it to a GeoJSON FeatureCollection
 */
func addSchemaOrgPoint(fc *geojson.FeatureCollection, idToUse string, lon string, lat string) (string, error) {
	x, err := strconv.ParseFloat(lon, 64)
	if err != nil {
		return "", err
	}
	y, err := strconv.ParseFloat(lat, 64)
	if err != nil {
		return "", err
	}

	json := jsonbuilder.Object()
	json.Set("type", "Point").Set("coordinates", jsonbuilder.Array(x, y))
	geojson := json.Marshal()
	//log.Println("POINT: ", geojson)
	return addGeoJSON(fc, idToUse, []byte(geojson))
}

func isValidURL(toTest string) bool {
	_, err := url.ParseRequestURI(toTest)
	if err != nil {
		return false
	}
	return true
}
