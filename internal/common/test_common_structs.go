package common

type jsonexpectations struct {
	name            string
	json            map[string]string
	IdentifierType  string `default:JsonSha`
	IdentifierPaths string
	expected        string
	expectedPath    string
	errorExpected   bool `default:false`
	ignore          bool `default:false`
}
