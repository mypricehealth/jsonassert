# jsonassert

The easier way to validate equivalency of two JSON documents

## Motivation

If you've used Go to connect to someone else's API, you know the pain. You go through and build a bunch 
of structs with JSON tags on them in order to Unmarshal the JSON data into strongly-typed structs. Only
problem is that it's hard to validate that those structs correctly capture all the data in the JSON because
Go is much more consistent in its JSON serialization than many API's are. 

## How does `jsonassert` help?

Pretend you have JSON that looks like this:

```javascript
{
  "str1": "val1",
  "str2": "",
  "num1": 1.2,
  "num2": 0.0,
  "bool1": true,
  "bool2": false,
}
```

You build this struct to receive the data:

```go
type APIReceiver struct {
  Str1  string  `json:"str1"`
  Str2  string  `json:"str2"`
  Num1  float64 `json:"num1"`
  Num2  float64 `json:"num2"`
  Bool1 bool    `json:"bool1"`
  Bool2 bool    `json:"bool2"`
}
```

Without `jsonassert`, you have to connect to a mock version of the API, get back a body containing
known text, Unmarshal that to JSON and then verify that the values match for every single property
what you're expecting. It's mind-numbingly boring. 

With `jsonassert`, you can just do this:

```go
import (
  "testing"
  "github.com/mypricehealth/jsonassert"
)

func TestApiReceiver(t *testing.T) {
  jsonassert.StructCheck(t, "testdata/sampleAPIReceiverResult.json", &APIReceiver{})
}
```

`StructCheck` makes it easy by assuming that if you take a known input and run it through both the JSON 
decode and then re-encode back to JSON, the values before and after should match. No need to write a test
that validates every value on every struct. No need to just cross your fingers and hope you built your
struct correctly. `StructCheck` will:
  1. Open and read the text from an input JSON file
  2. Encode the input text to the result struct, []struct, or map[string]<your struct>
  3. Decode the result map, struct, or slice back to JSON
  4. Compare the input JSON text with the output JSON text using the `Equal` function

But, the magic is really in the `Equal` function. `Equal` takes as its input two JSON byte slices and checks
if they are equivalent. What is equivalent? All of these are equivalent for the above `APIReceiver` struct:

```javascript
// All properties are filled in. This is the easy example since Go would decode each property as you'd 
// expect and then re-encode it as you'd expect. Any JSON assertion tool will get this one right.
{
  "str1": "val1",
  "str2": "",
  "num1": 1.2,
  "num2": 0.0,
  "bool1": true,
  "bool2": false,
}

// Empty properties in JSON mean those properties will get the default value for that type in Go
// so str1 will still = "", num2 will still = 0.0, and bool2 will still equal false. 
{
  "str1": "val1",
  "num1": 1.2,
  "bool1": true,
}

// Null values in JSON mean those properties will get the default value for that type in Go
// so str1 will still = "", num2 will still = 0.0, and bool2 will still equal false
{
  "str1": "val1",
  "str2": null,
  "num1": 1.2,
  "num2": null,
  "bool1": true,
  "bool2": null,
}
```

In the `Equals` function, values are considered to be equivalent if they are:
1. Both the same data type and values match
2. One value is nil and the other is the default value for its data type. These 
   values are considered equal:
    	a. "" and nil
     	b. 0.0 and nil
     	c. false and nil
     	d. empty slice and nil

## Data types

Because `jsonassert` is specifically working with JSON data, it is only testing equivalency on standard 
JSON types. According to the JSON spec, there are 6 different types. If you are doing a custom Unmarshal
into a different data type, it may not be supported

| JSON data type | Go equivalent          |
|----------------|------------------------|
| string         | string                 |
| number         | float64                |
| object         | map[string]interface{} | 
| array          | slice (any data)       |
| bool           | bool                   | 
| null           | nil                    |    


## Usage

### Equal example
```Go
import (
  "github.com/mypricehealth/jsonassert"
)

func TestJSON(t *testing.T) {
  json1 := []byte(`{"a": "1", "b": ""}`)
  json2 := []byte(`{"a": "1"})`)
  jsonassert.Equal(t, json1, json2)
}
```

### StructCheck example
```go
import (
  "testing"
  "github.com/mypricehealth/jsonassert"
)

func TestApiReceiver(t *testing.T) {
  jsonassert.StructCheck(t, "testdata/sampleAPIReceiverResult.json", &APIReceiver{})
}
```
