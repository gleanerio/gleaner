package utils

type Config struct {
	Millers struct {
		Graph   bool `json:"graph"`
		Spatial bool `json:"spatial"`
		Organic bool `json:"organic"`
		Mock    bool `json:"mock"`
	} `json:"millers"`
	Sources []struct {
		Name   string `json:"name"`
		Active bool   `json:"active"`
	} `json:"sources"`
}
