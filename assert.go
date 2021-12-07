package jsonassert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"
	"sort"
)

var nilVal = reflect.ValueOf(nil)

type Testing interface {
	Error(args ...interface{})
	Errorf(format string, args ...interface{})
	Helper()
}

// StructCheck is a convenience function for calling Equal. It is useful for verifying that the
// struct(s) you've created to receive JSON data in your application can losslessly encode and
// decode that JSON data. StructCheck will:
//   1. Open and read the text from an input JSON file
//   2. Encode the text in the JSON file to the result map, struct or slice
//   3. Decode the result map, struct, or slice back to JSON
//   4. Compare the input JSON text with the output JSON text using the Equal function
func StructCheck(t Testing, filename string, result interface{}) {
	t.Helper()
	var originalText, encodedText bytes.Buffer

	isMapType, err := resultArgCheck(result)
	if err != nil {
		t.Error(err)
		return
	}

	f, err := os.Open(filename)
	if err != nil {
		t.Error(err)
		return
	}
	defer f.Close()

	r := io.TeeReader(f, &originalText) // save original text to buffer while decoding JSON to result
	if err := json.NewDecoder(r).Decode(result); err != nil {
		t.Errorf("error decoding json in %s: %v", filename, err)
		return
	}

	json.NewEncoder(&encodedText).Encode(result)
	if isMapType {
		notifyErrors(t, filename, EqualMap(originalText.Bytes(), encodedText.Bytes()))
	} else {
		notifyErrors(t, filename, EqualSlice(originalText.Bytes(), encodedText.Bytes()))
	}
}

// EqualMap takes as its input two JSON byte slices and causes tests to fail as appropriate
// if the JSON doesn't match. Values are considered to be equivalent if they are:
//   1. Both the same data type and values match
//   2. One value is nil and the other is the default value for its data type. This is done since Go may
//      not serialize JSON the same way as other languages. Go is very strict and consistent, but the JSON
//      you have may not be. To make that easier, these values are considered equal:
//      	a. "" and nil
//      	b. 0.0 and nil
//      	c. false and nil
//      	d. empty slice and nil
func EqualMap(json1, json2 []byte) []error {
	json1Map, err1 := getJSONMap(json1)
	json2Map, err2 := getJSONMap(json2)
	if err1 != nil || err2 != nil {
		var errors []error
		if err1 != nil {
			errors = append(errors, fmt.Errorf("error unmarshalling json1: %v", err1))
		}
		if err2 != nil {
			errors = append(errors, fmt.Errorf("error unmarshalling json2: %v", err2))
		}
		return errors
	}
	return compareMaps("", json1Map, json2Map)
}

// EqualSlice takes as its input two JSON byte slices and causes tests to fail as appropriate
// if the JSON doesn't match. Values are considered to be equivalent if they are:
//   1. Both the same data type and values match
//   2. One value is nil and the other is the default value for its data type. This is done since Go may
//      not serialize JSON the same way as other languages. Go is very strict and consistent, but the JSON
//      you have may not be. To make that easier, these values are considered equal:
//      	a. "" and nil
//      	b. 0.0 and nil
//      	c. false and nil
//      	d. empty slice and nil
func EqualSlice(json1, json2 []byte) []error {
	json1Slice, err1 := getJSONSlice(json1)
	json2Slice, err2 := getJSONSlice(json2)
	if err1 != nil || err2 != nil {
		var errors []error
		if err1 != nil {
			errors = append(errors, fmt.Errorf("error unmarshalling json1: %v", err1))
		}
		if err2 != nil {
			errors = append(errors, fmt.Errorf("error unmarshalling json2: %v", err2))
		}
		return errors
	}
	return compareSlices("", json1Slice, json2Slice)
}

func notifyErrors(t Testing, filename string, errors []error) {
	if len(errors) > 0 {
		t.Errorf("*** %d errors in %s", len(errors), filename)
	}
	for _, err := range errors {
		t.Error(err)
	}
}

func resultArgCheck(result interface{}) (bool, error) {
	resT := reflect.TypeOf(result)
	resKind := resT.Kind()
	var elemKind reflect.Kind
	if resKind == reflect.Ptr {
		elemKind = resT.Elem().Kind()
	}
	isMapType := elemKind == reflect.Struct || elemKind == reflect.Map
	if resKind != reflect.Ptr || elemKind != reflect.Struct && elemKind != reflect.Map && elemKind != reflect.Slice && elemKind != reflect.Array {
		return isMapType, fmt.Errorf("invalid argument: result must be a pointer to a struct, slice, or map, but got %T", result)
	}
	return isMapType, nil
}

func getJSONMap(text []byte) (map[string]interface{}, error) {
	jsonMap := make(map[string]interface{})
	return jsonMap, json.Unmarshal(text, &jsonMap)
}

func getJSONSlice(text []byte) ([]interface{}, error) {
	jsonSlice := []interface{}{}
	return jsonSlice, json.Unmarshal(text, &jsonSlice)
}

func compareMaps(location string, map1, map2 map[string]interface{}) []error {
	var errors []error
	for _, key := range keys(map1) {
		errors = append(errors, compareValues(getLocation(location, key), map1[key], map2[key])...)
	}
	for _, key := range keys(map2) {
		value1, ok := map1[key]
		if !ok { // matched values were checked in the first loop, so only check missing ones here
			errors = append(errors, compareValues(getLocation(location, key), value1, map2[key])...)
		}
	}
	return errors
}

func getLocation(location, key string) string {
	if location == "" {
		return key
	}
	return fmt.Sprintf("%s.%s", location, key)
}

func keys(v map[string]interface{}) []string {
	keys := []string{}
	for key := range v {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func compareValues(location string, value1, value2 interface{}) []error {
	switch v1 := value1.(type) {
	case bool:
		if !boolEqual(v1, value2) {
			return []error{notifyError(location, value1, value2)}
		}
	case float64:
		if !floatEqual(v1, value2) {
			return []error{notifyError(location, value1, value2)}
		}
	case map[string]interface{}:
		v2, ok := value2.(map[string]interface{})
		if value2 != nil && !ok {
			return []error{notifyError(location, value1, value2)}
		}
		return compareMaps(location, v1, v2)
	case string:
		if !stringEqual(v1, value2) {
			return []error{notifyError(location, value1, value2)}
		}
	case nil:
		if !isEmpty(value2) {
			return []error{notifyError(location, value1, value2)}
		}
	default:
		return compareSlices(location, value1, value2)
	}
	return nil
}

func notifyError(location string, value1, value2 interface{}) error {
	return fmt.Errorf("%s mismatch. %v vs. %v", location, quoteString(value1), quoteString(value2))
}

func quoteString(v interface{}) string {
	strV, ok := v.(string)
	if ok {
		return fmt.Sprintf("%q", strV)
	}
	return fmt.Sprintf("%v", v)
}

func boolEqual(value1 bool, value2 interface{}) bool {
	return value1 == value2 || !value1 && value2 == nil
}

func floatEqual(value1 float64, value2 interface{}) bool {
	return value1 == value2 || value1 == 0.0 && value2 == nil
}

func isEmpty(value interface{}) bool {
	if value == "" || value == nil || value == 0.0 || value == false {
		return true
	}
	if mapV, ok := value.(map[string]interface{}); ok {
		return isMapEmpty(mapV)
	}
	rv := reflect.ValueOf(value)
	return rv.Kind() == reflect.Slice && rv.Len() == 0
}

func isMapEmpty(value map[string]interface{}) bool {
	for _, key := range keys(value) {
		if !isEmpty(value[key]) {
			return false
		}
	}
	return true
}

func stringEqual(value1 string, value2 interface{}) bool {
	return value1 == value2 || value1 == "" && value2 == nil
}

func compareSlices(location string, value1, value2 interface{}) []error {
	rv1 := reflect.ValueOf(value1)
	rv2 := reflect.ValueOf(value2)
	if rv1.Kind() != reflect.Slice || (rv2.Kind() != reflect.Slice && rv2 != nilVal) {
		return []error{notifyError(location, value1, value2)}
	}
	len1 := sliceLen(rv1)
	if len1 != sliceLen(rv2) {
		return []error{notifyError(location, value1, value2)}
	}
	if len1 == 0 {
		return nil
	}

	var errors []error
	for i := 0; i < len1; i++ {
		errors = append(errors, compareValues(fmt.Sprintf("%s[%d]", location, i), rv1.Index(i).Interface(), rv2.Index(i).Interface())...)
	}
	return errors
}

func sliceLen(v reflect.Value) int {
	if v == nilVal || v.IsNil() || v.Kind() != reflect.Slice {
		return 0
	}
	return v.Len()
}
