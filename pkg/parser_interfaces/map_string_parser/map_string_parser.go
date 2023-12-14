// Package map_string_parser is used for recursive parsing some nested
// `map[string]interface` data into some struct with tags that contains
// possible paths to specific nested data fields.
// Path can be force specified by tag `kind`.
// For use case examples, see the `map_string_parser_test.go` file.
package map_string_parser

import (
	"reflect"
	"strings"
)

// A MapParserState contains state for package.
type MapParserState[T any] struct {
	// Source is a nested map[string]interface{} data for parsing
	Source map[string]interface{}
	// Target is the resulting object, where will be saved parsed data
	Target *T
	// kind is the optional qualifying parameter
	// for determine target field that has the same kind tag value
	kind string
}

// NewMapStringParser setup Source data, Target struct and optional kind value
func NewMapStringParser[T any](source map[string]interface{}, kind string) *MapParserState[T] {
	return &MapParserState[T]{
		Source: source,
		Target: new(T),
		kind:   kind,
	}
}

// Parse is the function that directly does parsing
func (s *MapParserState[T]) Parse() {
	value := reflect.ValueOf(s.Target)
	numFields := value.Elem().NumField()
	structType := value.Elem().Type()
	// loop CoinData and checking tag are equal param key
	for i := 0; i < numFields; i++ {
		tagKind := structType.Field(i).Tag.Get("kind") // path tag value
		// check if kind not empty
		if len(tagKind) > 0 && tagKind != s.kind {
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
			// check if a path has point sign `.`, therefore, has multiple parts
			var parts []string
			if strings.Contains(path, ".") {
				parts = strings.Split(path, ".")
			} else {
				parts = append(parts, path)
			}
			for name, paramValue := range s.Source {
				// parts[0] is a current item name (e.g., coin)
				if parts[0] == name {
					if paramValue, ok := paramValue.(map[string]interface{}); ok {
						tail := parts[1:]
						deepValue := getDeepParam(tail, paramValue)
						if deepValue != nil {
							field := structType.Field(i)
							value.Elem().FieldByName(field.Name).Set(reflect.ValueOf(deepValue))
						}
					}
					if paramValue, ok := paramValue.(string); ok {
						field := structType.Field(i)
						value.Elem().FieldByName(field.Name).Set(reflect.ValueOf(paramValue))
						break breakPathsLoop
					}
				}
			}
		}
	}
}

// getDeepParam recursive func for go through a path to value
func getDeepParam(tail []string, paramValue map[string]interface{}) interface{} {
	for pvKey, pvValue := range paramValue {
		// TODO can be added max depth level
		if pvKey == tail[0] {
			if subVal, ok := pvValue.(map[string]interface{}); ok {
				newTail := tail[1:]
				return getDeepParam(newTail, subVal)
			} else {
				return pvValue
			}
		}
	}
	return nil
}
