package jsonassert

import (
	"testing"
)

var json1 = `{
	"num": 1,
	"num-empty": 0.0,
	"str": "2",
	"str-empty": "",
	"b-true": true,
	"b-false": false,
	"arr": ["1", "2", "3"],
	"arr-empty": [],
	"obj": { "a": "val", "b": "val2" },
	"obj-empty": {}
}`

var json2 = `{
	"num": 1,
	"num-empty": 0.0,
	"str-empty": "",
	"b-true": true,
	"b-false": false,
	"arr": ["1", "2", "3"],
	"arr-empty": [],
	"obj": { "a": "val", "b": "val2" },
	"obj-empty": {}
}`

var json3 = `{
	"num": 1,
	"str": "2",
	"b-true": true,
	"arr": ["1", "2", "3"],
	"obj": { "a": "val", "b": "val2" }
}`

var json4 = `{
	"num": "1",
	"num-empty": "",
	"str": 2,
	"str-empty": 0.0,
	"b-true": "true",
	"b-false": "false",
	"arr": { "a": "val", "b": "val2" },
	"arr-empty": {},
	"obj": ["1", "2", "3"],
	"obj-empty": []
}`

func TestEqual(t *testing.T) {
	tests := []struct {
		name  string
		json1 string
		json2 string
	}{
		{"same string", json1, json1},
		{"missing on the right", json1, json2},
		{"missing on the left", json2, json1},
		{"missing empty on the right", json1, json3},
		{"missing empty on the left", json3, json1},
		{"totally different values", json1, json4},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Equal(t, []byte(tt.json1), []byte(tt.json2))
		})
	}
}
