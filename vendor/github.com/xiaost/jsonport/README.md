
jsonport [![GoDoc](https://godoc.org/github.com/xiaost/jsonport?status.png)](https://godoc.org/github.com/xiaost/jsonport)
====

jsonport is a simple and high performance golang package for accessing json without pain. features:
* No reflection.
* Unmarshal without struct.
* Unmarshal for the given json path only.
* 2x faster than encoding/json.

It is inspired by [jmoiron/jsonq](https://github.com/jmoiron/jsonq). Feel free to post issues or PRs, I will reply ASAP :-)


## Usage

```go
package main

import (
    "fmt"
    "github.com/xiaost/jsonport"
)

func main() {
	jsonstr := `{
		"timestamp": "1438194274",
		"users": [{"id": 1, "name": "Tom"}, {"id": 2, "name": "Peter"}],
		"keywords": ["golang", "json"]
	}`
	j, _ := Unmarshal([]byte(jsonstr))

	// Output: Tom
	fmt.Println(j.GetString("users", 0, "name"))

	// Output: [golang json]
	fmt.Println(j.Get("keywords").StringArray())

	// Output: [Tom Peter]
	names, _ := j.Get("users").EachOf("name").StringArray()
	fmt.Println(names)

	// try parse STRING as NUMBER
	j.StringAsNumber()
	// Output: 1438194274
	fmt.Println(j.Get("timestamp").Int())

	// convert NUMBER, STRING, ARRAY and OBJECT type to BOOL
	j.AllAsBool()
	// Output: false
	fmt.Println(j.GetBool("status"))

	// using Unmarshal with path which can speed up json decode
	j, _ = Unmarshal([]byte(jsonstr), "users", 1, "name")
	fmt.Println(j.String())

	// Output:
	// Tom <nil>
	// [golang json] <nil>
	// [Tom Peter]
	// 1438194274 <nil>
	// false <nil>
	// Peter <nil>

}


```

For more information on getting started with `jsonport` [check out the doc](https://godoc.org/github.com/xiaost/jsonport)
