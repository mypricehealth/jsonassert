package jsonassert

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
)

var nilVal = reflect.ValueOf(nil)

// Equal takes as its input two JSON byte slices and causes tests to fail as appropriate
// if the JSON doesn't match. Values are considered to be equivalent if they are:
//   1. Both the same data type and values match
//   2. One value is nil and the other is the default value for its data type. This is done since Go may
//      not serialize JSON the same way as other languages. Go is very strict and consistent, but the JSON
//      you have may not be. To make that easier, these values are considered equal:
//      	a. "" and nil
//      	b. 0.0 and nil
//      	c. false and nil
//      	d. empty slice and nil
func Equal(t *testing.T, json1, json2 []byte) {
	json1Map, err1 := getJSONMap(json1)
	json2Map, err2 := getJSONMap(json2)
	if err1 != nil || err2 != nil {
		if err1 != nil {
			t.Error("error unmarshalling json1: %w", err1)
		}
		if err2 != nil {
			t.Error("error unmarshalling json2: %w", err2)
		}
		return
	}
	compareMaps(t, "", json1Map, json2Map)
}

func getJSONMap(text []byte) (map[string]interface{}, error) {
	jsonMap := make(map[string]interface{})
	return jsonMap, json.Unmarshal(text, &jsonMap)
}

func compareMaps(t *testing.T, location string, map1, map2 map[string]interface{}) {
	for key, value1 := range map1 {
		compareValues(t, key, value1, map2[key])
	}
	for key, value2 := range map2 {
		value1, ok := map1[key]
		if !ok { // matched values were checked in the first loop, so only check missing ones here
			childLocation := fmt.Sprintf("%s.%s", location, key)
			if location == "" {
				childLocation = key
			}
			compareValues(t, childLocation, value1, value2)
		}
	}
}

func compareValues(t *testing.T, location string, value1, value2 interface{}) {
	switch v1 := value1.(type) {
	case bool:
		if !boolEqual(v1, value2) {
			notifyError(t, location, value1, value2)
		}
	case float64:
		if !floatEqual(v1, value2) {
			notifyError(t, location, value1, value2)
		}
	case map[string]interface{}:
		v2, ok := value2.(map[string]interface{})
		if value2 != nil && !ok {
			notifyError(t, location, value1, value2)
			return
		}
		compareMaps(t, location, v1, v2)
	case string:
		if !stringEqual(v1, value2) {
			notifyError(t, location, value1, value2)
		}
	case nil:
		if !isEmpty(value2) {
			notifyError(t, location, value1, value2)
		}
	default:
		compareSlices(t, location, value1, value2)
	}
}

func notifyError(t *testing.T, location string, value1, value2 interface{}) {
	t.Errorf("%s mismatch.  %v vs. %v", location, quoteString(value1), quoteString(value2))
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

func isEmpty(value2 interface{}) bool {
	if value2 == "" || value2 == nil || value2 == 0.0 || value2 == false || reflect.DeepEqual(value2, map[string]interface{}{}) {
		return true
	}
	rv := reflect.ValueOf(value2)
	return rv.Kind() == reflect.Slice && rv.Len() == 0
}

func stringEqual(value1 string, value2 interface{}) bool {
	return value1 == value2 || value1 == "" && value2 == nil
}

func compareSlices(t *testing.T, location string, value1, value2 interface{}) {
	rv1 := reflect.ValueOf(value1)
	rv2 := reflect.ValueOf(value2)
	if rv1.Kind() != reflect.Slice && rv2.Kind() != reflect.Slice {
		notifyError(t, location, value1, value2)
	}
	len1 := sliceLen(rv1)
	if len1 != sliceLen(rv2) {
		notifyError(t, location, value1, value2)
		return
	}
	if len1 == 0 {
		return
	}
	for i := 0; i < len1; i++ {
		compareValues(t, fmt.Sprintf("%s[%d]", location, i), rv1.Index(i).Interface(), rv2.Index(i).Interface())
	}
}

func sliceLen(v reflect.Value) int {
	if v == nilVal || v.IsNil() || v.Kind() != reflect.Slice {
		return 0
	}
	return v.Len()
}
