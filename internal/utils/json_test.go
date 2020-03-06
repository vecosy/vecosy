package utils

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNormalizeMap(t *testing.T) {
	check := assert.New(t)
	input := map[interface{}]interface{}{
		"str1":   "str1",
		1:        1,
		int64(2): 2,
		1.1:      1.1,
		true:     true,
		false:    false,
	}
	expected := map[string]interface{}{
		"str1":  "str1",
		"1":     1,
		"2":     2,
		"1.1":   1.1,
		"true":  true,
		"false": false,
	}
	output, err := NormalizeMap(input)
	check.NoError(err)
	check.EqualValues(expected, output)
}
