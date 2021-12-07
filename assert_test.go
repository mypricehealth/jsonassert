package jsonassert

import (
	"fmt"
	"os"
	"testing"
)

var (
	jsonComplete       = getJSON("testdata/complete.json")
	jsonNulls          = getJSON("testdata/nulls.json")
	jsonMissingStrings = getJSON("testdata/missingStrings.json")
	jsonNoEmptyValues  = getJSON("testdata/noEmpty.json")
	jsonNewDataTypes   = getJSON("testdata/newDataTypes.json")
	sliceComplete      = getJSON("testdata/array.json")
	sliceNulls         = getJSON("testdata/arrayMissing.json")
	sliceMissing       = getJSON("testdata/arrayMissing.json")
	sliceNewDataTypes  = getJSON("testdata/arrayNewDataType.json")
)

type receiveStruct struct {
	Num      float64   `json:"num"`
	NumEmpty float64   `json:"num-empty"`
	Str      string    `json:"str"`
	StrEmpty string    `json:"str-empty"`
	BTrue    bool      `json:"b-true"`
	BFalse   bool      `json:"b-false"`
	Arr      []string  `json:"arr"`
	ArrEmpty []string  `json:"arr-empty"`
	Obj      subStruct `json:"obj"`
	ObjEmpty subStruct `json:"obj-empty"`
}

type subStruct struct {
	A string `json:"a"`
	B string `json:"b"`
}

type sliceStruct struct {
	Item1 string `json:"item1"`
	Item2 string `json:"item2"`
}

func getJSON(filename string) string {
	data, _ := os.ReadFile(filename)
	return string(data)
}

func TestStructCheck(t *testing.T) {
	tests := []struct {
		name           string
		filename       string
		result         interface{}
		expectedErrors []error
	}{
		{"everything matches", "testdata/complete.json", &receiveStruct{}, nil},
		{"nothing matches", "testdata/complete.json", &subStruct{}, []error{fmt.Errorf("*** 6 errors in testdata/complete.json"), fmt.Errorf("arr mismatch. [1 2 3] vs. <nil>"), fmt.Errorf("b-true mismatch. true vs. <nil>"), fmt.Errorf("num mismatch. 1 vs. <nil>"), fmt.Errorf(`obj.a mismatch. "val" vs. <nil>`), fmt.Errorf(`obj.b mismatch. "val2" vs. <nil>`), fmt.Errorf(`str mismatch. "2" vs. <nil>`)}},
		{"empty values all gone", "testdata/noEmpty.json", &receiveStruct{}, nil},
		{"empty values are null", "testdata/nulls.json", &receiveStruct{}, nil},
		{"bad filename", "bogus.json", &receiveStruct{}, []error{fmt.Errorf("open bogus.json: The system cannot find the file specified.")}},
		{"wrong result type", "testdata/nulls.json", &jsonComplete, []error{fmt.Errorf("invalid argument: result must be a pointer to a struct, slice, or map, but got *string")}},
		{"different data types", "testdata/newDataTypes.json", &receiveStruct{}, []error{fmt.Errorf("error decoding json in testdata/newDataTypes.json: json: cannot unmarshal string into Go struct field receiveStruct.num of type float64")}},
		{"slice", "testdata/array.json", &[]sliceStruct{}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeT := &fakeTester{}
			StructCheck(fakeT, tt.filename, tt.result)
			checkErrors(t, tt.expectedErrors, fakeT.errors)
		})
	}
}

func TestEqualMap(t *testing.T) {
	tests := []struct {
		name           string
		json1          string
		json2          string
		expectedErrors []error
	}{
		{"same string", jsonComplete, jsonComplete, nil},
		{"null on the right", jsonComplete, jsonNulls, nil},
		{"null on the left", jsonNulls, jsonComplete, nil},
		{"missing on the right", jsonComplete, jsonMissingStrings, []error{
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
		{"totally different values", jsonComplete, jsonNewDataTypes, []error{
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
		{"with children", `{"a": null}`, `{"a": {"1":"", "2":"b"}}`, []error{fmt.Errorf("a mismatch. <nil> vs. map[1: 2:b]")}},
		{"invalid file 1", `{`, jsonComplete, []error{fmt.Errorf("error unmarshalling json1: unexpected end of JSON input")}},
		{"invalid file 2", jsonComplete, `{`, []error{fmt.Errorf("error unmarshalling json2: unexpected end of JSON input")}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := EqualMap([]byte(tt.json1), []byte(tt.json2))
			checkErrors(t, tt.expectedErrors, errs)
		})
	}
}

func TestEqualSlice(t *testing.T) {
	tests := []struct {
		name           string
		json1          string
		json2          string
		expectedErrors []error
	}{
		{"same string", sliceComplete, sliceComplete, nil},
		{"null on the right", sliceComplete, sliceNulls, nil},
		{"null on the left", sliceNulls, sliceComplete, nil},
		{"missing on the right", sliceComplete, sliceMissing, nil},
		{"missing on the left", sliceMissing, sliceComplete, nil},
		{"totally different values", sliceComplete, sliceNewDataTypes, []error{
			fmt.Errorf(`[0].item1 mismatch. "" vs. 1`),
			fmt.Errorf(`[0].item2 mismatch. "value2" vs. 2`),
			fmt.Errorf(`[1].item1 mismatch. "value3" vs. 3`),
			fmt.Errorf(`[1].item2 mismatch. "" vs. 4`),
		}},
		{"with children", `[{"a": null}]`, `[{"a": {"1":"", "2":"b"}}]`, []error{fmt.Errorf("[0].a mismatch. <nil> vs. map[1: 2:b]")}},
		{"invalid file 1", `[`, sliceComplete, []error{fmt.Errorf("error unmarshalling json1: unexpected end of JSON input")}},
		{"invalid file 2", sliceComplete, `[`, []error{fmt.Errorf("error unmarshalling json2: unexpected end of JSON input")}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := EqualSlice([]byte(tt.json1), []byte(tt.json2))
			checkErrors(t, tt.expectedErrors, errs)
		})
	}
}

func checkErrors(t *testing.T, expectedErrors, actualErrors []error) {
	t.Helper()
	if len(expectedErrors) != len(actualErrors) {
		t.Errorf("different number of errors. wanted: %d, got %d", len(expectedErrors), len(actualErrors))
	}
	for i := 0; i < len(expectedErrors) || i < len(actualErrors); i++ {
		expected, actual := fmt.Errorf(""), fmt.Errorf("")
		if i < len(expectedErrors) {
			expected = expectedErrors[i]
		}
		if i < len(actualErrors) {
			actual = actualErrors[i]
		}
		if expected.Error() != actual.Error() {
			t.Errorf("error mismatch[%d]. want: %s, got: %s", i, expected, actual)
		}
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
func (t *fakeTester) Helper() {}
