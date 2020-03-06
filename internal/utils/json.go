package utils

import (
	"fmt"
	"reflect"
	"strconv"
)

// NormalizeMap convert a map[interface{}]interface{} to a map[string]interface{}
func NormalizeMap(mp map[interface{}]interface{}) (map[string]interface{}, error) {
	var err error
	strMap := make(map[string]interface{})
	for k, v := range mp {
		var keyString string
		switch typedKey := k.(type) {
		case string:
			keyString = typedKey
		case int:
			keyString = strconv.Itoa(typedKey)
		case int64:
			keyString = strconv.FormatInt(typedKey, 10)
		case float64:
			s := strconv.FormatFloat(typedKey, 'g', -1, 32)
			switch s {
			case "+Inf":
				s = ".inf"
			case "-Inf":
				s = "-.inf"
			case "NaN":
				s = ".nan"
			}
			keyString = s
		case bool:
			if typedKey {
				keyString = "true"
			} else {
				keyString = "false"
			}
		default:
			return nil, fmt.Errorf("unsupported map key of type: %s, key: %+#v, value: %+#v",
				reflect.TypeOf(k), k, v)
		}
		if mii, ok := v.(map[interface{}]interface{}); ok {
			strMap[keyString], err = NormalizeMap(mii)
			if err != nil {
				return nil, err
			}
		} else {
			strMap[keyString] = v
		}
	}
	return strMap, nil
}
