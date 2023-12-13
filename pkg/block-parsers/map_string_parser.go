package block_parsers

import (
	"reflect"
	"strings"
)

type MapParserState[T any] struct {
	Source map[string]interface{}
	Target *T
	kind   string
}

func NewMapStringParser[T any](source map[string]interface{}, template *T, kind string) MapParserState[T] {
	return MapParserState[T]{
		Source: source,
		Target: template,
		kind:   kind,
	}
}

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
						deepValue := getDeepParamMsg(tail, paramValue)
						if deepValue != nil {
							field := structType.Field(i)
							// todo go to !!!!!!!!!!!!!!
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
func getDeepParamMsg(tail []string, paramValue map[string]interface{}) interface{} {
	for pvKey, pvValue := range paramValue {
		// TODO can be added max depth level
		if pvKey == tail[0] {
			if subVal, ok := pvValue.(map[string]interface{}); ok {
				newTail := tail[1:]
				return getDeepParamMsg(newTail, subVal)
			} else {
				return pvValue
			}
		}
	}
	return nil
}
