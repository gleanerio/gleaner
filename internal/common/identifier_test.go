package common

import (
	"fmt"
	configTypes "github.com/gleanerio/gleaner/internal/config"
)

/*
This is to test various identifier
It uses a structure of expectations to run a series of individual tests with the name: testname_jsonfilename.

In the future, the JSON should probably be loaded from a file in resources_test folder.
*/

/* info on possible packages:
https://cburgmer.github.io/json-path-comparison/
using https://github.com/ohler55/ojg

test your jsonpaths here:
http://jsonpath.herokuapp.com/
There are four implementations... so you can see if one might be a little quirky
*/
import (
	"bytes"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

// jsonexpectations is in test_common_structs

var empty = []configTypes.Sources{}

// using idenfiters as a stand in for array of identifiers.

func testValidJsonPath(tests []jsonexpectations, t *testing.T) {
	for _, test := range tests {
		for i, json := range test.json {
			t.Run(fmt.Sprint(test.name, "_", i), func(t *testing.T) {
				if test.ignore {
					return
				}
				result, err := GetIdentifierByPath(test.IdentifierPaths, json)
				valStr := fmt.Sprint(result)
				assert.Equal(t, test.expected, valStr)
				assert.Nil(t, err)
			})
		}
	}

	//t.Run("@id", func(t *testing.T) {
	//
	//	result, err := GetIdentifierByPath(sources[0].IdentifierPath, jsonId)
	//	valStr := fmt.Sprint(result)
	//	assert.Equal(t, "[idenfitier]", valStr)
	//	assert.Nil(t, err)
	//})
	//t.Run(".idenfitier", func(t *testing.T) {
	//
	//	result, err := GetIdentifierByPath(sources[1].IdentifierPath, jsonId)
	//	valStr := fmt.Sprint(result)
	//	assert.Equal(t, "[doi:10.1575/1912/bco-dmo.2343.1]", valStr)
	//	assert.Nil(t, err)
	//})
	//t.Run("$.idenfitier", func(t *testing.T) {
	//
	//	result, err := GetIdentifierByPath(sources[2].IdentifierPath, jsonId)
	//	valStr := fmt.Sprint(result)
	//	assert.Equal(t, "[doi:10.1575/1912/bco-dmo.2343.1]", valStr)
	//	assert.Nil(t, err)
	//})
	// to do: test for valid JSON but invalid RDF triples
}

// test the array paths
func testValidJsonPaths(tests []jsonexpectations, t *testing.T) {
	for _, test := range tests {
		for i, json := range test.json {
			t.Run(fmt.Sprint(test.name, "_", i), func(t *testing.T) {
				if test.ignore {
					return
				}
				paths := strings.Split(test.IdentifierPaths, ",")
				result, foundPath, err := GetIdentiferByPaths(paths, json)
				valStr := fmt.Sprint(result)
				assert.Equal(t, test.expected, valStr, "expected Failed")
				assert.Equal(t, test.expectedPath, foundPath, "matched Path Failed")
				assert.Nil(t, err)
			})
		}

	}

}

/*
this tests a single path against a single json file
*/
func TestValidJsonPathInput(t *testing.T) {

	var jsonId = `{
"@id":"idenfitier",
"identifier":"doi:10.1575/1912/bco-dmo.2343.1",
"identifierArray": [	
	{
	"@type": "PropertyValue",
	"@id": "https://doi.org/10.1575/1912/bco-dmo.2343.1",
	"propertyID": "https://registry.identifiers.org/registry/doi",
	"value": "doi:10.1575/1912/bco-dmo.2343.1",
	"url": "https://doi.org/10.1575/1912/bco-dmo.2343.1"
	}
],
"identifierSArray": [	
	{
	"@type": "PropertyValue",
	"@id": "https://doi.org/10.1575/1912/bco-dmo.2343.1",
	"propertyID": "https://registry.identifiers.org/registry/doi",
	"value": "doi:10.1575/1912/bco-dmo.2343.1",
	"url": "https://doi.org/10.1575/1912/bco-dmo.2343.1"
	},
	{
	"@type": "PropertyValue",
	"@id": "https://doi.org/10.1575/1912/bco-dmo.2343.N",
	"propertyID": "https://registry.identifiers.org/registry/doi",
	"value": "doi:10.1575/1912/bco-dmo.2343.1N",
	"url": "https://doi.org/10.1575/1912/bco-dmo.2343.N"
	},
	{
	"@type": "PropertyValue",
	"@id": "https://doi.org/10.1575/1912/bco-dmo.2343.P",
	"propertyID": "https://purl.org",
	"value": "doi:10.1575/1912/bco-dmo.2343.P",
	"url": "https://doi.org/10.1575/1912/bco-dmo.2343.P"
	}
],
"identifierObj": 
	{
	"@type": "PropertyValue",
	"@id": "https://doi.org/10.1575/1912/bco-dmo.2343.1",
	"propertyID": "https://registry.identifiers.org/registry/doi",
	"value": "doi:10.1575/1912/bco-dmo.2343.1",
	"url": "https://doi.org/10.1575/1912/bco-dmo.2343.1"
	}

}`

	var tests = []jsonexpectations{
		// default
		{
			name:          "@id",
			json:          map[string]string{"jsonID": jsonId},
			errorExpected: false,

			IdentifierPaths: `$['@id']`,
			expected:        "[idenfitier]",
			expectedPath:    "$['@id']",
			ignore:          false,
		},
		//https://raw.githubusercontent.com/earthcube/GeoCODES-Metadata/main/metadata/Dataset/actualdata/earthchem2.json
		{
			name:            "@.identifier",
			json:            map[string]string{"jsonID": jsonId},
			errorExpected:   false,
			IdentifierPaths: "@.identifier",
			expected:        "[doi:10.1575/1912/bco-dmo.2343.1]",
			expectedPath:    "@.identifier",
			ignore:          false,
		},
		//https://raw.githubusercontent.com/earthcube/GeoCODES-Metadata/main/metadata/Dataset/actualdata/earthchem2.json
		{
			name:            "$.identifier",
			json:            map[string]string{"jsonID": jsonId},
			errorExpected:   false,
			IdentifierPaths: "$.identifier",
			expected:        "[doi:10.1575/1912/bco-dmo.2343.1]",
			expectedPath:    "$.identifier",
			ignore:          false,
		},
		// argo example: https://raw.githubusercontent.com/earthcube/GeoCODES-Metadata/main/metadata/Dataset/actualdata/argo.json
		{
			name:            "identifiers Array ",
			json:            map[string]string{"jsonID": jsonId},
			errorExpected:   false,
			IdentifierPaths: "$.identifierSArray[?(@.propertyID=='https://registry.identifiers.org/registry/doi')].value",
			expected:        "[doi:10.1575/1912/bco-dmo.2343.1N doi:10.1575/1912/bco-dmo.2343.1]",
			expectedPath:    "$.identifierSArray[?(@.propertyID=='https://registry.identifiers.org/registry/doi')].value",
			ignore:          false,
		},
		{
			name:          "identifier_obj",
			json:          map[string]string{"jsonID": jsonId},
			errorExpected: false,
			//	IdentifierPath: "$.identifierObj[?(@.propertyID=='https://registry.identifiers.org/registry/doi')].value",
			//IdentifierPath: "$.identifierObj.propertyID[@=='https://registry.identifiers.org/registry/doi')]",
			IdentifierPaths: "$.identifierObj.value",
			expected:        "[doi:10.1575/1912/bco-dmo.2343.1]",
			expectedPath:    "$.identifierObj.value",
			ignore:          false,
		},
		//https://raw.githubusercontent.com/earthcube/GeoCODES-Metadata/main/metadata/Dataset/actualdata/earthchem2.json
		// this will not work since the || does not work
		{
			name:            " identifier or id",
			json:            map[string]string{"jsonID": jsonId},
			errorExpected:   false,
			IdentifierPaths: "[ $.identifiers[?(@.propertyID=='https://registry.identifiers.org/registry/doi')].value || $.['@id'] ]",
			expected:        "[doi:10.1575/1912/bco-dmo.2343.1]",
			expectedPath:    "[ $.identifiers[?(@.propertyID=='https://registry.identifiers.org/registry/doi')].value || $.['@id'] ]",
			ignore:          true,
		},
		// identifier as array: https://github.com/earthcube/GeoCODES-Metadata/blob/main/metadata/Dataset/allgood/bcodmo1.json
		/* needs work
		"identifier": [

		       {
		           "@type": "PropertyValue",
		           "@id": "https://doi.org/10.1575/1912/bco-dmo.2343.1",
		           "propertyID": "https://registry.identifiers.org/registry/doi",
		           "value": "doi:10.1575/1912/bco-dmo.2343.1",
		           "url": "https://doi.org/10.1575/1912/bco-dmo.2343.1"
		       }
		   ],
		*/
		// this does not work fancy array index issues. Would be nice
		{
			name:          "identifierSArray slice",
			json:          map[string]string{"jsonID": jsonId},
			errorExpected: false,
			//IdentifierPath: "$.identifierSArray[?(@.propertyID=='https://registry.identifiers.org/registry/doi')].value[-1:]",
			//IdentifierPaths: []string{"$.identifierSArray[?(@.propertyID=='https://registry.identifiers.org/registry/doi')].value.[-1:]"},
			IdentifierPaths: "$.identifierSArray[?(@.propertyID=='https://registry.identifiers.org/registry/doi')].value[0]",
			expected:        "[doi:10.1575/1912/bco-dmo.2343.1]",
			expectedPath:    "$.identifierSArray[?(@.propertyID=='https://registry.identifiers.org/registry/doi')].value.[0]",
			ignore:          true,
		},
	}

	testValidJsonPath(tests, t)
}

func TestValidJsonPathsInput(t *testing.T) {

	// this failing the first test with just one
	var jsonId = `{
"@id":"idenfitier",
"url": "http://example.com/,"
}`
	var jsonIdentifier = `{
"@id":"idenfitier",
"url": "http://example.com/",
"identifier":"doi:10"


}`
	var jsonIdentifierObject = `{
"@id":"idenfitier",
"url": "http://example.com/",
"identifier": 
	{
	"@type": "PropertyValue",
	"@id": "https://doi.org/10.1575/1912/bco-dmo.2343.1",
	"propertyID": "https://registry.identifiers.org/registry/doi",
	"value": "doi:10.1575/1912/bco-dmo.2343.1",
	"url": "https://doi.org/10.1575/1912/bco-dmo.2343.1"
	}
}`

	var jsonIdentifierArraySingle = `{
"@id":"idenfitier",
"url": "http://example.com/",
"identifier": [	
	{
	"@type": "PropertyValue",
	"@id": "https://doi.org/10.1575/1912/bco-dmo.2343.1",
	"propertyID": "https://registry.identifiers.org/registry/doi",
	"value": "doi:10.1575/1912/bco-dmo.2343.1",
	"url": "https://doi.org/10.1575/1912/bco-dmo.2343.1"
	}
]


}`
	var jsonIdentifierArrayMultiple = `{
"@id":"idenfitier",
"url": "http://example.com/",
"identifier": [	
	{
	"@type": "PropertyValue",
	"@id": "https://doi.org/10.1575/1912/bco-dmo.2343.1",
	"propertyID": "https://registry.identifiers.org/registry/doi",
	"value": "doi:10.1575/1912/bco-dmo.2343.1",
	"url": "https://doi.org/10.1575/1912/bco-dmo.2343.1"
	},
	{
	"@type": "PropertyValue",
	"@id": "https://doi.org/10.1575/1912/bco-dmo.2343.N",
	"propertyID": "https://registry.identifiers.org/registry/doi",
	"value": "doi:10.1575/1912/bco-dmo.2343.1N",
	"url": "https://doi.org/10.1575/1912/bco-dmo.2343.N"
	},
	{
	"@type": "PropertyValue",
	"@id": "https://doi.org/10.1575/1912/bco-dmo.2343.P",
	"propertyID": "https://purl.org",
	"value": "doi:10.1575/1912/bco-dmo.2343.P",
	"url": "https://doi.org/10.1575/1912/bco-dmo.2343.P"
	}
]

}`
	var problemChildIris = `
{
 "@context": {
  "@vocab": "https://schema.org/"
 },
 "@id": "https://ds.iris.edu/ds/products/emtf/",
 "@type": "Dataset",
 "name": "Magnetotelluric Transfer Functions",
 "alternateName": [
  "EMTF"
 ],
 "description": "USArray magnetotelluric transfer functions (MT TFs) calculated at USArray MT sites installed by Oregon State University, as well as other community magnetotelluric transfer functions, are available from the \"SPUD EMTF\":http://www.iris.edu/spud/emtf repository in both XML and EDI formats. The international magnetotelluric community is invited to use the EMTF repository to archive their MT TFs. Please contact \"Anna Kelbert\":https://www.usgs.gov/staff-profiles/anna-kelbert for support; all data formats are accepted and a data citation is created upon submission of data to the searchable repository.\r\n\r\nThe XML format for electromagnetic transfer functions and related conversion software was developed at Oregon State University under an NSF award, and later improved and updated by the USGS Geomagnetism Program, with continued support from IRIS. A living software repository is available in the \"SeisCode EMTF-FCU project\":https://seiscode.iris.washington.edu/projects/emtf-fcu File Conversion Utilities project. Detailed documentation and usage examples are provided by \"Kelbert &#40;2009&#41; \":https://library.seg.org/doi/10.1190/geo2018-0679.1.",
 "url": "https://ds.iris.edu/ds/products/emtf/",
 "dateCreated": "2013-07-17T22:17:22.003",
 "dateModified": "2019-11-20T14:48:11.228",
 "includedInDataCatalog": {
  "@type": "DataCatalog",
  "@id": "https://ds.iris.edu/ds/products/",
  "url": "https://ds.iris.edu/ds/products/"
 },
 "keywords": "Transportable Array,magnetotelluric data,seismic,USArray MT,MT data,global magnetotellurics,magnetotelluric impedance,seismology,geophysics,EMTF,impedance database,US magnetotellurics",
 "author": [
  {
   "@type": "Person",
   "name": "IRIS Data Products"
  },
  {
   "@type": "Person",
   "name": "Anna Kelbert",
   "affiliation": {
    "@type": "Organization",
    "name": "Geomagnetism Program, U.S. Geological Survey"
   }
  }
 ],
 "image": "https://ds.iris.edu/media/product/emtf/images/mt_func_logo_1.png"
}`
	var problemChildOpenTopo = `

{
    "@context": {
        "@vocab": "https://schema.org/"
    },
    "@type": "Dataset",
    "version": "1.0",
    "additionalType": ["geolink:Dataset", "vivo:Dataset"],
    "name": "Quaternary Surface Ruptures along Panamint Valley Fault, CA, April-May 2018",
    "@id": "https://portal.opentopography.org/dataspace/dataset?opentopoID=OTDS.062020.32611.1",
    "description": "The digital surface models, generated from sUAS-derived images and SfM methods, show fault scarps and offset channels at six sites along the oblique-slip and strike-slip segments of the Panamint Valley Fault.  The fault cuts Late Quaternary alluvial deposits. These data were collected as part of a master\u0092s thesis. The study aims to reconstruct the displacement and rupture length of the most recent earthquake, and understand the kinematics of the geometrically complex Panamint Valley Fault.",
    "url": "https://doi.org/10.5069/G92Z13PS",
    "sameAs": "https://portal.opentopography.org/dataspace/dataset?opentopoID=OTDS.062020.32611.1",
    "fileFormat" : "Point Cloud, Raster",
    "publisher" : {
        "@type": "Organization",
        "additionalType": "geolink:Organization",
        "email": "info@opentopography.org",
        "legalName": "OpenTopography",
        "name": "OpenTopography",
        "url": "https://opentopography.org",
        "award" : "National Science Foundation under Award Numbers EAR-1948997, 1948994 & 1948857",
        "logo" : "http://www.opentopography.org/sites/opentopography.org/files/ot_transp_logo.png"
    },
    "funder": [],
    "contributor": [{
        "@type": "Person",
        "name": "Israporn Sethanant"
    },{
        "@type": "Person",
        "name": "Israporn Sethanant"
    },{
        "@type": "Person",
        "name": "Wesley Dassow"
    }],

    "variableMeasured": [{
        "@type": "PropertyValue",
        "name" : "Area",
        "value" : "588,495.73 m2"
    }

    ,{
        "@type": "PropertyValue",
        "name" : "LidarReturns",
        "value" : "421,777,492 points"
    }

    ,{
        "@type": "PropertyValue",
        "name" : "PointDensity",
        "value" : "716.7 points/m2"
    }

    ],

    "temporalCoverage": "2018-04-27/2018-05-04",

    "spatialCoverage": {
        "@type": "Place",
        "additionalProperty": [{
                "@type": "PropertyValue",
                "additionalType" : "CoordinatesSystem",
                "name" : "Horizontal Coordinates",
                "value" : "WGS 84 / UTM zone 11N [EPSG: 32611]"
        },{
                "@type": "PropertyValue",
                "additionalType" : "CoordinatesSystem",
                "name" : "Vertical Coordinates",
                "value" : ""
        }],
        "geo": {
            "@type": "GeoShape",
            "box": "35.89945649,-117.2195011 36.00975424,-117.1817643"
        }
    },
    "locationCreated" : "",
    "dateCreated": "2020-06-25 15:35:13.965967",
    "license" : "Not-Provided",
    "usageinfo" : "https://opentopography.org/usageterms",
    "isAccessibleForFree" : true,
    "citation" : "The data collection was funded by the 2017 GSA Graduate Student Research Grant. Special thanks to Chris Parrish, Chase Simpson, and Richard Slocum (Oregon State University) for the guidance of flight planning, RTK positioning survey, and Trimble 5800 set up. I would also like to acknowledge Wesley von Dassow for the help in the successful GCPs collection in the field. ",
    "keywords": "Panamint Valley Fault, Eastern California Shear Zone, earthquake, complex fault",
    "identifier": {
        "@id": "https://portal.opentopography.org/dataspace/dataset?opentopoID=OTDS.062020.32611.1",
        "@type": "PropertyValue",
        "propertyID": "opentopoID",
        "value": "OTDS.062020.32611.1"
    }
}
`
	var tests = []jsonexpectations{
		// default
		// should work for all
		{
			name: "@id",
			json: map[string]string{"jsonID": jsonId, "jsonIdentifier": jsonIdentifier,
				"jsonobjectId":                jsonIdentifierObject,
				"jsonIdentifierArraySingle":   jsonIdentifierArraySingle,
				"jsonIdentifierArrayMultiple": jsonIdentifierArrayMultiple,
			},
			errorExpected: false,

			IdentifierPaths: `$['@id']`,
			expected:        "[idenfitier]",
			expectedPath:    "$['@id']",
			ignore:          false,
		},
		//https://raw.githubusercontent.com/earthcube/GeoCODES-Metadata/main/metadata/Dataset/actualdata/earthchem2.json
		// this returns an empty set [] https://cburgmer.github.io/json-path-comparison/results/dot_notation_on_object_without_key.html
		{
			name: "$.identifier.$id",
			//json:            []string{jsonId},
			json: map[string]string{"jsonID": jsonId}, //"jsonIdentifier": jsonIdentifier,
			//"jsonobjectId": jsonIdentifierObject,
			//"jsonIdentifierArraySingle": jsonIdentifierArraySingle,
			//"jsonIdentifierArrayMultiple": jsonIdentifierArrayMultiple,

			errorExpected:   false,
			IdentifierPaths: "$.identifier.value,$.identifier,$['@id']",
			expected:        "[idenfitier]",
			expectedPath:    "$['@id']",
			ignore:          false,
		},
		{
			name: "$.identifier.$.identifier",
			//json:            []string{jsonIdentifier},
			json:            map[string]string{"jsonIdentifier": jsonIdentifier},
			errorExpected:   false,
			IdentifierPaths: "$.identifier.value,$.identifier,$['@id']",
			expected:        "[doi:10]",
			expectedPath:    "$.identifier",
			ignore:          false,
		},
		//https://raw.githubusercontent.com/earthcube/GeoCODES-Metadata/main/metadata/Dataset/actualdata/earthchem2.json
		{
			name: "$.identifierObjBracket",
			//json:            []string{jsonIdentifierObject},
			json: map[string]string{
				"jsonobjectId": jsonIdentifierObject,
			},
			errorExpected:   false,
			IdentifierPaths: "$.identifier['value'],$.identifier,$['@id']",
			expected:        "[doi:10.1575/1912/bco-dmo.2343.1]",
			expectedPath:    "$.identifier['value']",
			ignore:          false,
		},
		{
			name: "$.identifierObjDot",
			//json:            []string{jsonIdentifierObject},
			json: map[string]string{
				"jsonobjectId": jsonIdentifierObject,
			},
			errorExpected:   false,
			IdentifierPaths: "$.identifier.value,$.identifier,$['@id']",
			expected:        "[doi:10.1575/1912/bco-dmo.2343.1]",
			expectedPath:    "$.identifier.value",
			ignore:          false,
		},
		{
			name: "$.identifierObjCheck",
			//json:            []string{jsonIdentifierObject},
			json: map[string]string{
				"jsonobjectId": jsonIdentifierObject,
			},
			errorExpected:   false,
			IdentifierPaths: "$.identifier.value,$.identifier,$['@id']",
			expected:        "[doi:10.1575/1912/bco-dmo.2343.1]",
			expectedPath:    "$.identifier.value",
			ignore:          false,
		},
		//https://raw.githubusercontent.com/earthcube/GeoCODES-Metadata/main/metadata/Dataset/actualdata/earthchem2.json
		{
			name: "@.identifierArraySimple",
			//json:            []string{jsonIdentifierArraySingle},
			json: map[string]string{
				"jsonIdentifierArraySingle": jsonIdentifierArraySingle,
			},
			errorExpected:   false,
			IdentifierPaths: "$.identifier[?(@.propertyID=='https://registry.identifiers.org/registry/doi')].value,$.identifier.value,$.identifier.$['@id']",
			expected:        "[doi:10.1575/1912/bco-dmo.2343.1]",
			expectedPath:    "$.identifier[?(@.propertyID=='https://registry.identifiers.org/registry/doi')].value",
			ignore:          false,
		},

		//https://raw.githubusercontent.com/earthcube/GeoCODES-Metadata/main/metadata/Dataset/actualdata/earthchem2.json
		{
			name: "@.identifierArrayMultiple",
			//json:            []string{jsonIdentifierArrayMultiple},
			json: map[string]string{
				"jsonIdentifierArrayMultiple": jsonIdentifierArrayMultiple,
			},
			errorExpected:   false,
			IdentifierPaths: "$.identifier[?(@.propertyID=='https://registry.identifiers.org/registry/doi')].value,$.identifier.value,$.identifier,$['@id']",
			expected:        "[doi:10.1575/1912/bco-dmo.2343.1N doi:10.1575/1912/bco-dmo.2343.1]",
			expectedPath:    "$.identifier[?(@.propertyID=='https://registry.identifiers.org/registry/doi')].value",
			ignore:          false,
		},
		{
			name: "@.identifierProblemChildIris",
			//json:            []string{jsonIdentifierArrayMultiple},
			json: map[string]string{
				"problem child": problemChildIris,
			},
			errorExpected:   false,
			IdentifierPaths: "$.identifier[?(@.propertyID=='https://registry.identifiers.org/registry/doi')].value,$.identifier.value,$.identifier,$['@id']",
			expected:        "[https://ds.iris.edu/ds/products/emtf/]",
			expectedPath:    "$['@id']",
			ignore:          false,
		},
		{
			name: "@.identifierProblemChildOpenTopo",
			//json:            []string{jsonIdentifierArrayMultiple},
			json: map[string]string{
				"problem child opentopo": problemChildOpenTopo,
			},
			errorExpected:   false,
			IdentifierPaths: "$.identifier[?(@.propertyID=='https://registry.identifiers.org/registry/doi')].value,$.identifier.value,$.identifier,$['@id']",
			expected:        "[OTDS.062020.32611.1]",
			expectedPath:    "$.identifier.value",
			ignore:          false,
		},
	}
	testValidJsonPaths(tests, t)
}

func testGenerateJsonPathIdentifier(tests []jsonexpectations, t *testing.T) {

	//mock configre file
	// paths are relative to the code
	var vipercontext = []byte(`
context:
  cache: true
contextmaps:
- file: ../../configs/schemaorg-current-https.jsonld
  prefix: https://schema.org/
- file: ../../configs/schemaorg-current-https.jsonld
  prefix: http://schema.org/
sources:
- sourcetype: sitemap
  name: test
  logo: https://opentopography.org/sites/opentopography.org/files/ot_transp_logo_2.png
  url: https://opentopography.org/sitemap.xml
  headless: false
  pid: https://www.re3data.org/repository/r3d100010655
  propername: OpenTopography
  domain: http://www.opentopography.org/
  active: false
  credentialsfile: ""
  other: {}
  headlesswait: 0
  delay: 0
  IdentifierType: identifiersha
`)

	for _, test := range tests {
		for i, json := range test.json {
			// needs to be defiend in the loop, so that each run has it's own configuration.
			// otherwise changing the sources information in a multi-threaded ent has issues
			viperVal := viper.New()
			viperVal.SetConfigType("yaml")
			viperVal.ReadConfig(bytes.NewBuffer(vipercontext))
			sources, err := configTypes.GetSources(viperVal)

			if err != nil {
				assert.Fail(t, err.Error())
			}

			s := sources[0]
			s.IdentifierType = test.IdentifierType
			s.IdentifierPath = test.IdentifierPaths
			t.Run(fmt.Sprint(test.name, "_", i), func(t *testing.T) {
				if test.ignore {
					return
				}
				result, err := GenerateIdentifier(viperVal, s, json)
				//valStr := fmt.Sprint(result.uniqueId)
				assert.Equal(t, test.expected, result.UniqueId, "uuid faild")
				assert.Equal(t, test.expectedPath, result.MatchedPath, "matched path failed")
				assert.Equal(t, test.IdentifierType, result.IdentifierType, "identifier failed")
				assert.Nil(t, err)
			})
		}
	}
}
func testGenerateFileShaIdentifier(tests []jsonexpectations, t *testing.T) {

	//mock configre file
	// paths are relative to the code
	var vipercontext = []byte(`
context:
  cache: true
contextmaps:
- file: ../../configs/schemaorg-current-https.jsonld
  prefix: https://schema.org/
- file: ../../configs/schemaorg-current-https.jsonld
  prefix: http://schema.org/
sources:
- sourcetype: sitemap
  name: test
  logo: https://opentopography.org/sites/opentopography.org/files/ot_transp_logo_2.png
  url: https://opentopography.org/sitemap.xml
  headless: false
  pid: https://www.re3data.org/repository/r3d100010655
  propername: OpenTopography
  domain: http://www.opentopography.org/
  active: false
  credentialsfile: ""
  other: {}
  headlesswait: 0
  delay: 0
  IdentifierType: filesha
`)

	for _, test := range tests {
		for i, json := range test.json {
			// needs to be defiend in the loop, so that each run has it's own configuration.
			// otherwise changing the sources information in a multi-threaded ent has issues
			viperVal := viper.New()
			viperVal.SetConfigType("yaml")
			viperVal.ReadConfig(bytes.NewBuffer(vipercontext))
			sources, err := configTypes.GetSources(viperVal)

			if err != nil {
				assert.Fail(t, err.Error())
			}

			s := sources[0]
			s.IdentifierType = test.IdentifierType
			s.IdentifierPath = test.IdentifierPaths
			t.Run(fmt.Sprint(test.name, "_", i), func(t *testing.T) {
				if test.ignore {
					return
				}
				result, err := GenerateIdentifier(viperVal, s, json)
				//valStr := fmt.Sprint(result.uniqueId)
				assert.Equal(t, test.expected, result.UniqueId, "uuid failed")
				assert.Equal(t, test.expectedPath, result.MatchedPath, "matched path failed")
				assert.Equal(t, test.IdentifierType, result.IdentifierType, "identifiertype match failed")
				assert.Nil(t, err)
			})
		}
	}
}

func TestGenerateFileShaIdentifier(t *testing.T) {
	var jsonIdentifier = `{
		"@context": {
		"rdf": "http://www.w3.org/1999/02/22-rdf-syntax-ns#",
			"rdfs": "http://www.w3.org/2000/01/rdf-schema#",
			"schema": "http://schema.org/",
			"xsd": "http://www.w3.org/2001/XMLSchema#"
	    },
		"@graph": [
		      {
					"@id": "https://wifire-data.sdsc.edu/dataset/a1770ff8-1665-433c-88fb-c8e6863c61fc/resource/b01d00d2-1d64-47b8-aa5c-00410d84e6e6",
					"@type": "schema:DataDownload",
					"schema:encodingFormat": "GeoJSON",
					"schema:name": "GeoJSON",
					"schema:url": "https://gis-calema.opendata.arcgis.com/datasets/34402e97810f410db0ccd1ae345d9807_5.geojson?outSR=%7B%22latestWkid%22%3A3857%2C%22wkid%22%3A102100%7D"
				}
		]
	}
`

	var jsonIdentifierArrayMultiple = `{
"@id":"idenfitier",
"url": "http://example.com/",
"identifier": [	
	{
	"@type": "PropertyValue",
	"@id": "https://doi.org/10.1575/1912/bco-dmo.2343.1",
	"propertyID": "https://registry.identifiers.org/registry/doi",
	"value": "doi:10.1575/1912/bco-dmo.2343.1",
	"url": "https://doi.org/10.1575/1912/bco-dmo.2343.1"
	},
	{
	"@type": "PropertyValue",
	"@id": "https://doi.org/10.1575/1912/bco-dmo.2343.N",
	"propertyID": "https://registry.identifiers.org/registry/doi",
	"value": "doi:10.1575/1912/bco-dmo.2343.1N",
	"url": "https://doi.org/10.1575/1912/bco-dmo.2343.N"
	},
	{
	"@type": "PropertyValue",
	"@id": "https://doi.org/10.1575/1912/bco-dmo.2343.P",
	"propertyID": "https://purl.org",
	"value": "doi:10.1575/1912/bco-dmo.2343.P",
	"url": "https://doi.org/10.1575/1912/bco-dmo.2343.P"
	}
]

}`
	var tests = []jsonexpectations{
		// default
		// should work for all
		{
			name: "filesha",
			json: map[string]string{
				"jsonIdentifierArrayMultiple": jsonIdentifierArrayMultiple,
			},
			errorExpected:   false,
			IdentifierType:  configTypes.JsonSha,
			IdentifierPaths: "$['@id']",
			expected:        "92b87f05ee545b042a563803bc148a46506b9e89",
			expectedPath:    "",
			ignore:          false,
		},
		{
			name: "normalizedsha",
			json: map[string]string{
				"jsonidentifier": jsonIdentifier,
			},
			errorExpected:   false,
			IdentifierType:  configTypes.NormalizedJsonSha,
			IdentifierPaths: "$['@id']",
			expected:        "4b741fbebb530cb553bd07639045e569a54424c7",
			expectedPath:    "",
			ignore:          false,
		},
	}

	testGenerateFileShaIdentifier(tests, t)
}

func TestGenerateJsonPathIdentifier(t *testing.T) {
	var jsonIdentifierArrayMultiple = `{
"@id":"idenfitier",
"url": "http://example.com/",
"identifier": [	
	{
	"@type": "PropertyValue",
	"@id": "https://doi.org/10.1575/1912/bco-dmo.2343.1",
	"propertyID": "https://registry.identifiers.org/registry/doi",
	"value": "doi:10.1575/1912/bco-dmo.2343.1",
	"url": "https://doi.org/10.1575/1912/bco-dmo.2343.1"
	},
	{
	"@type": "PropertyValue",
	"@id": "https://doi.org/10.1575/1912/bco-dmo.2343.N",
	"propertyID": "https://registry.identifiers.org/registry/doi",
	"value": "doi:10.1575/1912/bco-dmo.2343.1N",
	"url": "https://doi.org/10.1575/1912/bco-dmo.2343.N"
	},
	{
	"@type": "PropertyValue",
	"@id": "https://doi.org/10.1575/1912/bco-dmo.2343.P",
	"propertyID": "https://purl.org",
	"value": "doi:10.1575/1912/bco-dmo.2343.P",
	"url": "https://doi.org/10.1575/1912/bco-dmo.2343.P"
	}
]

}`
	var tests = []jsonexpectations{
		// default
		// should work for all

		{
			name: "@id_first",
			json: map[string]string{
				"jsonIdentifierArrayMultiple": jsonIdentifierArrayMultiple,
			},
			errorExpected:   false,
			IdentifierType:  configTypes.IdentifierSha,
			IdentifierPaths: "$['@id']",
			expected:        "0fe143f05d6dbff260874a9a6e8da77243c74db0",
			expectedPath:    "$['@id']",
			ignore:          false,
		},
		{
			name: "identifier_default_path",
			json: map[string]string{
				"jsonIdentifierArrayMultiple": jsonIdentifierArrayMultiple,
			},
			errorExpected:   false,
			IdentifierType:  configTypes.IdentifierSha,
			IdentifierPaths: "",
			expected:        "e59f7f11a5615bcee6f35c92d8a2162e5b611944",
			expectedPath:    "$.identifier[?(@.propertyID=='https://registry.identifiers.org/registry/doi')].value",
			ignore:          false,
		},
	}

	testGenerateJsonPathIdentifier(tests, t)
}

func TestValidJsonPathGraphInput(t *testing.T) {

	var jsonIdentifierArrayMultiple = `{
"@id":"idenfitier",
"url": "http://example.com/",
"identifier": [	
	{
	"@type": "PropertyValue",
	"@id": "https://doi.org/10.1575/1912/bco-dmo.2343.1",
	"propertyID": "https://registry.identifiers.org/registry/doi",
	"value": "doi:10.1575/1912/bco-dmo.2343.1",
	"url": "https://doi.org/10.1575/1912/bco-dmo.2343.1"
	},
	{
	"@type": "PropertyValue",
	"@id": "https://doi.org/10.1575/1912/bco-dmo.2343.N",
	"propertyID": "https://registry.identifiers.org/registry/doi",
	"value": "doi:10.1575/1912/bco-dmo.2343.1N",
	"url": "https://doi.org/10.1575/1912/bco-dmo.2343.N"
	},
	{
	"@type": "PropertyValue",
	"@id": "https://doi.org/10.1575/1912/bco-dmo.2343.P",
	"propertyID": "https://purl.org",
	"value": "doi:10.1575/1912/bco-dmo.2343.P",
	"url": "https://doi.org/10.1575/1912/bco-dmo.2343.P"
	}
]

}`

	// from wifire-data.sdsc.edu
	var jsonLdGraph = `

           {
    "@context": {
        "rdf": "http://www.w3.org/1999/02/22-rdf-syntax-ns#",
        "rdfs": "http://www.w3.org/2000/01/rdf-schema#",
        "schema": "https://schema.org/",
        "xsd": "http://www.w3.org/2001/XMLSchema#"
    },
    "@graph": [
        {
            "@id": "_:N4c6352f015f54e27a532ac1e0d693293",
            "@type": "schema:ContactPoint",
            "schema:contactType": "customer service",
            "schema:name": "CA Governor's Office of Emergency Services",
            "schema:url": "https://wifire-data.sdsc.edu"
        },
        {
            "@id": "https://wifire-data.sdsc.edu/dataset/8fd44c38-f6d3-429c-a785-1498dfaa2a6a/resource/87accf29-5b48-49dd-b299-b0a417f5a8c6",
            "@type": "schema:DataDownload",
            "schema:encodingFormat": "KML",
            "schema:name": "KML",
            "schema:url": "https://gis-calema.opendata.arcgis.com/datasets/b426dba2dacb4d788d85ad061e14e6db_2.kml?outSR=%7B%22latestWkid%22%3A3857%2C%22wkid%22%3A102100%7D"
        },
        {
            "@id": "_:N6c68117e012c4c77ba9c82f6264c5836",
            "@type": "schema:DataCatalog",
            "schema:description": "",
            "schema:name": "WIFIRE Commons Data Catalog",
            "schema:url": "https://wifire-data.sdsc.edu"
        },
        {
            "@id": "_:N8022daf479db4a99a26bfc41fe5cef9b",
            "@type": "schema:GeoShape",
            "schema:polygon": "{\"type\": \"Polygon\", \"coordinates\": [[[-178.4436, -14.3743], [-178.4436, 71.3905], [146.0827, 71.3905], [146.0827, -14.3743], [-178.4436, -14.3743]]]}"
        },
        {
            "@id": "https://wifire-data.sdsc.edu/dataset/8fd44c38-f6d3-429c-a785-1498dfaa2a6a",
            "@type": "schema:Dataset",
            "schema:dateModified": [
                "2020-04-15T23:04:24+00:00",
                "2020-04-15T23:04:24.000Z"
            ],
            "schema:datePublished": [
                "2018-01-16T20:02:32+00:00",
                "2018-01-16T20:02:32.000Z"
            ],
            "schema:description": "<span style='font-family: &quot;Avenir Next W01&quot;, &quot;Avenir Next W00&quot;, &quot;Avenir Next&quot;, Avenir, &quot;Helvetica Neue&quot;, Helvetica, Arial, sans-serif; font-size: 17px; background-color: rgb(255, 255, 255);'>This dataset represents regions, which are part of the national field level structure to support chapters. The Regions role is administrative as well as to provide oversight and program technical support to the chapters. This Region shapefile reflects the region boundaries with the associated attribute information. Red Cross Geography Model: Counties make up chapters, chapters make up regions and regions make up divisions. There are five exceptions to the Red Cross Geography model: Middlesex County, MA, Los Angeles, Kern, Riverside and San Bernardino Counties in California which are covered by more than one chapter. (many to one). In the case of these five counties, the geometry was dissolved from zip codes.\u00a0</span>",
            "schema:distribution": [
                {
                    "@id": "https://wifire-data.sdsc.edu/dataset/8fd44c38-f6d3-429c-a785-1498dfaa2a6a/resource/e53f2226-f9b0-45a0-9ec8-6fd72b6fcfe8"
                },
                {
                    "@id": "https://wifire-data.sdsc.edu/dataset/8fd44c38-f6d3-429c-a785-1498dfaa2a6a/resource/2b9bb554-48e3-44fb-b253-b27e32a3f3d9"
                },
                {
                    "@id": "https://wifire-data.sdsc.edu/dataset/8fd44c38-f6d3-429c-a785-1498dfaa2a6a/resource/7a24ea41-4c3e-43b9-85df-880e4754d613"
                },
                {
                    "@id": "https://wifire-data.sdsc.edu/dataset/8fd44c38-f6d3-429c-a785-1498dfaa2a6a/resource/0850ffdb-8cde-4c2d-baca-5616b56675d1"
                },
                {
                    "@id": "https://wifire-data.sdsc.edu/dataset/8fd44c38-f6d3-429c-a785-1498dfaa2a6a/resource/ed9b42d2-4260-490b-be0d-1a09e138ab12"
                },
                {
                    "@id": "https://wifire-data.sdsc.edu/dataset/8fd44c38-f6d3-429c-a785-1498dfaa2a6a/resource/87accf29-5b48-49dd-b299-b0a417f5a8c6"
                }
            ],
            "schema:includedInDataCatalog": {
                "@id": "_:N6c68117e012c4c77ba9c82f6264c5836"
            },
            "schema:keywords": [
                "Data Library",
                "Boundaries and Regions",
                "HIFLD",
                "CalOES Library",
                "ESF5",
                "Basedata",
                "CalOES Data Library",
                "ESF6",
                "Emergency Services",
                "American Red Cross Regions"
            ],
            "schema:license": "https://creativecommons.org/publicdomain/zero/1.0/",
            "schema:name": "American Red Cross Regions",
            "schema:publisher": {
                "@id": "https://wifire-data.sdsc.edu/organization/0e04d99f-90be-40d7-bf0a-ddda02f1eb09"
            },
            "schema:spatialCoverage": {
                "@id": "_:N57eefc034db64b7d8b24c0b108fd2e7f"
            },
            "schema:url": "https://wifire-data.sdsc.edu/dataset/american-red-cross-regions"
        },
        {
            "@id": "https://wifire-data.sdsc.edu/dataset/8fd44c38-f6d3-429c-a785-1498dfaa2a6a/resource/2b9bb554-48e3-44fb-b253-b27e32a3f3d9",
            "@type": "schema:DataDownload",
            "schema:encodingFormat": "GeoJSON",
            "schema:name": "GeoJSON",
            "schema:url": "https://gis-calema.opendata.arcgis.com/datasets/b426dba2dacb4d788d85ad061e14e6db_2.geojson?outSR=%7B%22latestWkid%22%3A3857%2C%22wkid%22%3A102100%7D"
        },
        {
            "@id": "https://wifire-data.sdsc.edu/dataset/8fd44c38-f6d3-429c-a785-1498dfaa2a6a/resource/ed9b42d2-4260-490b-be0d-1a09e138ab12",
            "@type": "schema:DataDownload",
            "schema:encodingFormat": "HTML",
            "schema:name": "ArcGIS Hub Dataset",
            "schema:url": "https://gis-calema.opendata.arcgis.com/datasets/b426dba2dacb4d788d85ad061e14e6db_2"
        },
        {
            "@id": "https://wifire-data.sdsc.edu/organization/0e04d99f-90be-40d7-bf0a-ddda02f1eb09",
            "@type": "schema:Organization",
            "schema:contactPoint": {
                "@id": "_:N4c6352f015f54e27a532ac1e0d693293"
            },
            "schema:name": "CA Governor's Office of Emergency Services"
        },
        {
            "@id": "_:N57eefc034db64b7d8b24c0b108fd2e7f",
            "@type": "schema:Place",
            "schema:geo": {
                "@id": "_:N8022daf479db4a99a26bfc41fe5cef9b"
            }
        },
        {
            "@id": "https://wifire-data.sdsc.edu/dataset/8fd44c38-f6d3-429c-a785-1498dfaa2a6a/resource/7a24ea41-4c3e-43b9-85df-880e4754d613",
            "@type": "schema:DataDownload",
            "schema:encodingFormat": "Esri REST",
            "schema:name": "Esri Rest API",
            "schema:url": "https://services.arcgis.com/pGfbNJoYypmNq86F/ArcGIS/rest/services/ARC_Master_Geography_FY19_January/FeatureServer/2"
        },
        {
            "@id": "https://wifire-data.sdsc.edu/dataset/8fd44c38-f6d3-429c-a785-1498dfaa2a6a/resource/e53f2226-f9b0-45a0-9ec8-6fd72b6fcfe8",
            "@type": "schema:DataDownload",
            "schema:encodingFormat": "ZIP",
            "schema:name": "Shapefile",
            "schema:url": "https://gis-calema.opendata.arcgis.com/datasets/b426dba2dacb4d788d85ad061e14e6db_2.zip?outSR=%7B%22latestWkid%22%3A3857%2C%22wkid%22%3A102100%7D"
        },
        {
            "@id": "https://wifire-data.sdsc.edu/dataset/8fd44c38-f6d3-429c-a785-1498dfaa2a6a/resource/0850ffdb-8cde-4c2d-baca-5616b56675d1",
            "@type": "schema:DataDownload",
            "schema:encodingFormat": "CSV",
            "schema:name": "CSV",
            "schema:url": "https://gis-calema.opendata.arcgis.com/datasets/b426dba2dacb4d788d85ad061e14e6db_2.csv?outSR=%7B%22latestWkid%22%3A3857%2C%22wkid%22%3A102100%7D"
        }
    ]
}
   `

	var tests = []jsonexpectations{
		// default

		{
			name:          "identifieGraph Not Graph",
			json:          map[string]string{"jsonID": jsonIdentifierArrayMultiple},
			errorExpected: true,
			//IdentifierPath: "$.identifierSArray[?(@.propertyID=='https://registry.identifiers.org/registry/doi')].value[-1:]",
			IdentifierPaths: "$['@graph'][?(@['@type']=='schema:Dataset')]['@id'],$.identifier[?(@.propertyID=='https://registry.identifiers.org/registry/doi')].value,$.identifier.value,$.identifier,$['@id']",

			expected:     "[doi:10.1575/1912/bco-dmo.2343.1N doi:10.1575/1912/bco-dmo.2343.1]",
			expectedPath: "$.identifier[?(@.propertyID=='https://registry.identifiers.org/registry/doi')].value",
			ignore:       false,
		},
		// grr. Ugly since the herokuapp no longer runs: used this a hint, then raw debugging: https://cburgmer.github.io/json-path-comparison/

		// this one $['@graph]*[?(@['@type']=='schema:Dataset')]  gives false here: https://jsonpath.curiousconcept.com/
		// $['@graph']*.['@type'] returns types
		// $['@graph'].*.@id returns types
		//$.@graph*[?(@.@type=="schema:Dataset")] false bad when debuggin. cannot start with an @

		// workslocally:
		// returns nil: "$['@graph']*[?(@['@type']=='schema:Dataset')]"
		// returns full object: "$['@graph'][?(@['@type']=='schema:Dataset')]"
		// returns @id: "$['@graph'][?(@['@type']=='schema:Dataset')]['@id']"  fails at: https://jsonpath.curiousconcept.com/
		{
			name:          "identifiersGraph",
			json:          map[string]string{"jsonID": jsonLdGraph},
			errorExpected: false,
			//IdentifierPath: "$.identifierSArray[?(@.propertyID=='https://registry.identifiers.org/registry/doi')].value[-1:]",
			IdentifierPaths: "$['@graph'][?(@['@type']=='schema:Dataset')]['@id']",

			expected:     "[https://wifire-data.sdsc.edu/dataset/8fd44c38-f6d3-429c-a785-1498dfaa2a6a]",
			expectedPath: "$['@graph'][?(@['@type']=='schema:Dataset')]['@id']",
			ignore:       false,
		},
		{
			name:          "identifiersGraphLong",
			json:          map[string]string{"jsonID": jsonLdGraph},
			errorExpected: false,
			//IdentifierPath: "$.identifierSArray[?(@.propertyID=='https://registry.identifiers.org/registry/doi')].value[-1:]",
			IdentifierPaths: "$['@graph'][?(@['@type']=='schema:Dataset')]['@id'],$.identifier[?(@.propertyID=='https://registry.identifiers.org/registry/doi')].value,$.identifier.value,$.identifier,$['@id']",

			expected:     "[https://wifire-data.sdsc.edu/dataset/8fd44c38-f6d3-429c-a785-1498dfaa2a6a]",
			expectedPath: "$['@graph'][?(@['@type']=='schema:Dataset')]['@id']",
			ignore:       false,
		},
	}

	testValidJsonPaths(tests, t)
}
