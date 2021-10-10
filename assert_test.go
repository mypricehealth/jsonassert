package jsonassert

import (
	"fmt"
	"testing"
)

var jsonComplete = `{
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

var jsonNulls = `{
	"num": 1,
	"num-empty": null,
	"str": "2",
	"str-empty": null,
	"b-true": true,
	"b-false": null,
	"arr": ["1", "2", "3"],
	"arr-empty": null,
	"obj": { "a": "val", "b": "val2" },
	"obj-empty": null
}`

var jsonMissingStrings = `{
	"num": 1,
	"num-empty": 0.0,
	"str-empty": "",
	"b-true": true,
	"b-false": false,
	"arr": ["1", "2"],
	"arr-empty": [],
	"obj": { "a": "val" },
	"obj-empty": {}
}`

var jsonNoEmptyValues = `{
	"num": 1,
	"str": "2",
	"b-true": true,
	"arr": ["1", "2", "3"],
	"obj": { "a": "val", "b": "val2" }
}`

var jsonNewDataTypes = `{
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
		name           string
		json1          string
		json2          string
		expectedErrors []error
	}{
		{"same string", jsonComplete, jsonComplete, nil},
		{"null on the right", jsonComplete, jsonNulls, nil},
		{"null on the left", jsonNulls, jsonComplete, nil},
		{"missing on the right", jsonComplete, jsonMissingStrings,
			[]error{
				fmt.Errorf(`arr mismatch. [1 2 3] vs. [1 2]`),
				fmt.Errorf(`obj.b mismatch. "val2" vs. <nil>`),
				fmt.Errorf(`str mismatch. "2" vs. <nil>`),
			}},
		{"missing on the left", jsonMissingStrings, jsonComplete,
			[]error{
				fmt.Errorf(`arr mismatch. [1 2] vs. [1 2 3]`),
				fmt.Errorf(`obj.b mismatch. <nil> vs. "val2"`),
				fmt.Errorf(`str mismatch. <nil> vs. "2"`),
			}},
		{"missing empty on the right", jsonComplete, jsonNoEmptyValues, nil},
		{"missing empty on the left", jsonNoEmptyValues, jsonComplete, nil},
		{"totally different values", jsonComplete, jsonNewDataTypes,
			[]error{
				fmt.Errorf(`arr mismatch. [1 2 3] vs. map[a:val b:val2]`),
				fmt.Errorf(`arr-empty mismatch. [] vs. map[]`),
				fmt.Errorf(`b-false mismatch. false vs. "false"`),
				fmt.Errorf(`b-true mismatch. true vs. "true"`),
				fmt.Errorf(`num mismatch. 1 vs. "1"`),
				fmt.Errorf(`num-empty mismatch. 0 vs. ""`),
				fmt.Errorf(`obj mismatch. map[a:val b:val2] vs. [1 2 3]`),
				fmt.Errorf(`obj-empty mismatch. map[] vs. []`),
				fmt.Errorf(`str mismatch. "2" vs. 2`),
				fmt.Errorf(`str-empty mismatch. "" vs. 0`),
			}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeT := &fakeTester{}
			Equal(fakeT, []byte(tt.json1), []byte(tt.json2))
			if len(tt.expectedErrors) != len(fakeT.errors) {
				t.Errorf("different number of errors. wanted: %d, got %d", len(tt.expectedErrors), len(fakeT.errors))
			}
			for i := 0; i < len(tt.expectedErrors) && i < len(fakeT.errors); i++ {
				expected := tt.expectedErrors[i]
				actual := fakeT.errors[i]
				if expected.Error() != actual.Error() {
					t.Errorf("error mismatch[%d]. want: %s, got: %s", i, expected, actual)
				}
			}
		})
	}
}

type fakeTester struct {
	errors []error
}

func (t *fakeTester) Error(args ...interface{}) {
	t.errors = append(t.errors, fmt.Errorf(fmt.Sprint(args...)))
}
func (t *fakeTester) Errorf(format string, args ...interface{}) {
	t.errors = append(t.errors, fmt.Errorf(format, args...))
}
