package framing

import (
	"encoding/json"
	"log"
	"reflect"

	ld "github.com/kazarena/json-gold/ld"
	jql "github.com/xiaost/jsonport"
)

// type SpatialFrameRes struct {
// 	Description string
// 	ID          string
// 	Latitude    string
// 	Longitude   string
// 	// URL         string `json:"schema:url"`
// }

type Geo struct {
  ID        string `json:"id"`
	Type      string `json:"type"`
	Latitude  string `json:"latitude"`
	Longitude string `json:"longitude"`
	Box       string `json:"box"`
	Circle    string `json:"circle"`
	Line      string `json:"line"`
	Polygon   string `json:"polygon"`
}

type SubjectOf struct {
	ID         string `json:"id"`
	Type       string `json:"type"`
	FileFormat string `json:"schema:fileFormat"`
	Text       string `json:"text"`
}

type SpatialCoverage struct {
	ID        string `json:"id"`
	Type      string `json:"type"`
	Geos      []Geo `json:"-"`
	SubjectOf []SubjectOf `json:"subjectOf"`
}

type SpatialFrameRes struct {
	ID                  string `json:"id"`
	Type                string `json:"type"`
	URL                 string `json:"url"`
	SpatialCoverages    []SpatialCoverage `json:"-"`
}

func SpatialFrame(jsonld string) []SpatialFrameRes {
	proc := ld.NewJsonLdProcessor()
	options := ld.NewJsonLdOptions("")

	frame := map[string]interface{}{
		"@context":        "http://schema.org/",
		"@explicit":       true,
		"@type":           "Dataset",
		"@id":             "",
		"spatialCoverage": map[string]interface{}{
			"@type":           "Place",
			"geo":             map[string]interface{}{},
			"subjectOf":       map[string]interface{}{
				"@type":           "CreativeWork",
				"@explicit":       true,
				"fileFormat":      "",
				"text":            "",
			},
		},
	}

	var myInterface interface{}
	err := json.Unmarshal([]byte(jsonld), &myInterface)
	if err != nil {
		log.Println("Error when transforming JSON-LD document to interface:", err)
		return nil
	}

	// TODO review  https://www.w3.org/TR/json-ld-api/#widl-JsonLdOptions-expandContext
	// Dictionary JsonLdOptions Members
	// expandContext of type (object? or DOMString), defaulting to null
	// A context that is used to initialize the active context when expanding a document.
	framedDoc, err := proc.Frame(myInterface, frame, options) // do I need the options set in order to avoid the large context that seems to be generated?
	if err != nil {
		log.Println("Error when trying to frame document", err)
		return nil
	}

	graph := framedDoc["@graph"]
	//ld.PrintDocument("JSON-LD graph section", graph) // debug print....
	jsonm, err := json.MarshalIndent(graph, "", " ")
	if err != nil {
		log.Println("Error trying to marshal data", err)
		return nil // does this work, if linked earth fails (or any fails..)  or do I return a an empty dss struct?
	}

	sfr := make([]SpatialFrameRes, 0)
	err = json.Unmarshal(jsonm, &sfr)
	if err != nil {
		log.Println("Error trying to unmarshal data to struct", err)
	}

	// Query the JSON for what we need.
	jsonFrame, _ := jql.Unmarshal([]byte(jsonm))
	spatialCvgJson := jsonFrame.Get(0, "spatialCoverage")
	if spatialCvgJson.IsArray() {
		length, _ := spatialCvgJson.Len()
		for i := 0; i < length; i++ {
			coverage, err := parseSpatialCoverage(spatialCvgJson.Get(i))
			if err == nil {
				sfr[0].SpatialCoverages = append(sfr[0].SpatialCoverages, coverage)
			}
		}
	} else if spatialCvgJson.IsObject() {
		coverage, err := parseSpatialCoverage(spatialCvgJson)
		if err == nil {
			sfr[0].SpatialCoverages = append(sfr[0].SpatialCoverages, coverage)
		}
	}

	return sfr
}

/**
 * Create a SpatialCoverage instance
 */
func parseSpatialCoverage(spatialCvgJson jql.Json) (SpatialCoverage, error) {

	sc := SpatialCoverage{}
	id, err := spatialCvgJson.GetString("id")
	if err == nil {
		reflect.ValueOf(&sc).Elem().FieldByName("ID").SetString(id)
	}
	t, err := spatialCvgJson.GetString("type")
	if err == nil {
		reflect.ValueOf(&sc).Elem().FieldByName("Type").SetString(t)
	}

	geo := spatialCvgJson.Get("geo")
	if geo.IsArray() {
		length, _ := geo.Len()
		for i := 0; i < length; i++ {
			shape, err := parseGeo(geo.Get(i))
			if err == nil {
				sc.Geos = append(sc.Geos, shape)
			}
		}
	} else if spatialCvgJson.IsObject() {
		shape, err := parseGeo(geo)
		if err == nil {
			sc.Geos = append(sc.Geos, shape)
		}
	}

	subjectOf := spatialCvgJson.Get("subjectOf")
	if subjectOf.IsArray() {
		length, _ := subjectOf.Len()
		for i := 0; i < length; i++ {
			subject, err := parseSubjectOf(subjectOf.Get(i))
			if err == nil {
				sc.SubjectOf = append(sc.SubjectOf, subject)
			}
		}
	} else if spatialCvgJson.IsObject() {
		subject, err := parseSubjectOf(subjectOf)
		if err == nil {
			sc.SubjectOf = append(sc.SubjectOf, subject)
		}
	}

	return sc, nil
}

/**
 * Create a Geo instance
 */
func parseGeo(geometry jql.Json) (Geo, error) {

	geo := Geo{}

	val, err := geometry.GetString("id")
	if err == nil {
		reflect.ValueOf(&geo).Elem().FieldByName("ID").SetString(val)
	}

	val, err = geometry.GetString("type")
	if err == nil {
		reflect.ValueOf(&geo).Elem().FieldByName("Type").SetString(val)
	}

	val, err = geometry.GetString("latitude")
	if err == nil {
		reflect.ValueOf(&geo).Elem().FieldByName("Latitude").SetString(val)
	}

	val, err = geometry.GetString("longitude")
	if err == nil {
		reflect.ValueOf(&geo).Elem().FieldByName("Longitude").SetString(val)
	}

	val, err = geometry.GetString("box")
	if err == nil {
		reflect.ValueOf(&geo).Elem().FieldByName("Box").SetString(val)
	}

	val, err = geometry.GetString("circle")
	if err == nil {
		reflect.ValueOf(&geo).Elem().FieldByName("Circle").SetString(val)
	}

	val, err = geometry.GetString("line")
	if err == nil {
		reflect.ValueOf(&geo).Elem().FieldByName("Line").SetString(val)
	}

	val, err = geometry.GetString("polygon")
	if err == nil {
		reflect.ValueOf(&geo).Elem().FieldByName("Polygon").SetString(val)
	}

	return geo, nil
}

/**
 * Create a Geo instance
 */
func parseSubjectOf(so jql.Json) (SubjectOf, error) {
	subjectOf := SubjectOf{}

	val, err := so.GetString("id")
	if err == nil {
		reflect.ValueOf(&subjectOf).Elem().FieldByName("ID").SetString(val)
	}

	val, err = so.GetString("type")
	if err == nil {
		reflect.ValueOf(&subjectOf).Elem().FieldByName("Type").SetString(val)
	}

	val, err = so.GetString("fileFormat")
	if err == nil {
		reflect.ValueOf(&subjectOf).Elem().FieldByName("FileFormat").SetString(val)
	}

	val, err = so.GetString("text")
	if err == nil {
		reflect.ValueOf(&subjectOf).Elem().FieldByName("Text").SetString(val)
	}

	return subjectOf, nil
}

