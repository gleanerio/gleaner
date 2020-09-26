// Package p contains an HTTP Cloud Function.
// package p
package p

import (
	"bytes"
	"fmt"
	"image"
	"io/ioutil"
	"log"
	"net/http"

	_ "image/jpeg"
	_ "image/png"
)

// HelloWorld is a test function
func HelloWorld(w http.ResponseWriter, r *http.Request) {

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Could not read request", http.StatusBadRequest)
	}

	m, _, err := image.Decode(bytes.NewBuffer(data))
	if err != nil {
		log.Fatal(err)
	}
	bounds := m.Bounds()

	var histogram [16][4]int
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := m.At(x, y).RGBA()
			// A color's RGBA method returns values in the range [0, 65535].
			// Shifting by 12 reduces this to the range [0, 15].
			histogram[r>>12][0]++
			histogram[g>>12][1]++
			histogram[b>>12][2]++
			histogram[a>>12][3]++
		}
	}

	var b bytes.Buffer

	// Print the results.
	b.WriteString(fmt.Sprintf("%-14s %6s %6s %6s %6s\n", "bin", "red", "green", "blue", "alpha"))
	for i, x := range histogram {
		b.WriteString(fmt.Sprintf("0x%04x-0x%04x: %6d %6d %6d %6d\n", i<<12, (i+1)<<12-1, x[0], x[1], x[2], x[3]))
	}

	fmt.Fprint(w, b.String())

}
