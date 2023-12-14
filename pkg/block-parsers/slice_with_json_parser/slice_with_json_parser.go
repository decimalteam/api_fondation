// Package slice_with_json_parser for parsing data from data struct
// where the first level is a slice of specific type, then each
// element has value that can contains nested json.
// For use cases, see `slice_with_json_parser_test.go` file.
package slice_with_json_parser

import (
	"encoding/json"
	"github.com/tidwall/gjson"
	"reflect"
	"strings"
)

type SliceJsonParserState[S any, T any] struct {
	Source []S // data that will be parsed
	// (optional) each element has a key field,
	// there it can be customized.
	KeyField string
	// (optional) each element has a value field,
	// there it can be customized.
	ValueField string
	// new struct that will contain a result of the parsing
	Target *T
	// kind is the optional qualifying parameter
	// for determine target field that has the same kind tag value
	Kind string
}

// NewSliceJsonParserState setup Source data, Target struct,
// and optional parameters
func NewSliceJsonParserState[S any, T any](source []S, kind, keyField, valField string) *SliceJsonParserState[S, T] {
	return &SliceJsonParserState[S, T]{
		Source:     source,
		Target:     new(T),
		Kind:       kind,
		KeyField:   keyField,
		ValueField: valField,
	}
}

// Parse is the function that directly does parsing
func (s *SliceJsonParserState[S, T]) Parse() {
	// loop Target at top level Name
	value := reflect.ValueOf(s.Target).Elem()
	numFields := value.NumField()
	structType := value.Type()

	// loop Target and checking tag are equal param key
	for i := 0; i < numFields; i++ {
		tagKind := structType.Field(i).Tag.Get("kind") // path tag value
		// check if kind not empty
		if len(tagKind) > 0 && tagKind != s.Kind {
			continue
		}
		tagMsg := structType.Field(i).Tag.Get("path") // path tag value
		// check if a string has comma, split and add all paths into an array
		var paths []string
		if strings.Contains(tagMsg, ",") {
			paths = strings.Split(tagMsg, ",")
		} else {
			paths = append(paths, tagMsg)
		}
	breakPathsLoop:
		// loop possible paths
		for _, path := range paths {
			var firstLevelPart string
			jsonPart := ""
			if strings.Contains(tagMsg, ".") {
				// split first part before first `.`
				parts := strings.SplitN(path, ".", 2)
				firstLevelPart = parts[0]
				jsonPart = parts[1]
			} else {
				firstLevelPart = path
			}

			for _, paramValue := range s.Source {
				var keyField string
				if len(s.KeyField) == 0 {
					keyField = "Key" // default key field name
				} else {
					keyField = s.KeyField
				}
				var valueField string
				if len(s.KeyField) == 0 {
					valueField = "Value" // default value field name
				} else {
					valueField = s.KeyField
				}
				paramValueValue := reflect.ValueOf(paramValue).FieldByName(valueField).String()
				paramValueKey := reflect.ValueOf(paramValue).FieldByName(keyField).String()
				field := structType.Field(i)
				if paramValueKey == firstLevelPart {
					// get value by json path
					var js map[string]interface{}
					jsonParseErr := json.Unmarshal([]byte(paramValueValue), &js)
					if jsonParseErr == nil {
						deepValue := gjson.Get(paramValueValue, jsonPart).String()
						if len(deepValue) > 0 {
							value.FieldByName(field.Name).Set(reflect.ValueOf(deepValue))
							break breakPathsLoop
						}
					} else if paramValueKey == path {
						value.FieldByName(field.Name).Set(reflect.ValueOf(paramValueValue))
						break breakPathsLoop
					}
				}
			}
		}
	}
}
