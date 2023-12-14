package slice_with_json_parser

import (
	"encoding/json"
	"github.com/tidwall/gjson"
	"reflect"
	"strings"
)

type SliceJsonParserState[S any, T any] struct {
	Source     []S
	KeyField   string
	ValueField string
	Target     *T
	Kind       string
}

func NewSliceJsonParserState[S any, T any](source []S, kind, keyField, valField string) *SliceJsonParserState[S, T] {
	return &SliceJsonParserState[S, T]{
		Source:     source,
		Target:     new(T),
		Kind:       kind,
		KeyField:   keyField,
		ValueField: valField,
	}
}

func (s *SliceJsonParserState[S, T]) ParseCoinDataFromAttributes() {
	// todo -- point 3
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
					keyField = "Key"
				} else {
					keyField = s.KeyField
				}
				var valueField string
				if len(s.KeyField) == 0 {
					valueField = "Value"
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
