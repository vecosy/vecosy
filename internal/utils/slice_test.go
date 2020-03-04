package utils

import (
	"github.com/hashicorp/go-version"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestReverseStrings(t *testing.T) {
	check := assert.New(t)
	strSlice := []string{"a", "b", "c"}
	ReverseStrings(strSlice)
	check.Equal(strSlice, []string{"c", "b", "a"})

	strSlice = []string{"a"}
	ReverseStrings(strSlice)
	check.Equal(strSlice, []string{"a"})

	strSlice = []string{}
	ReverseStrings(strSlice)
	check.Equal(strSlice, []string{})
}

func TestReverseVersion(t *testing.T) {
	check := assert.New(t)
	v1, err := version.NewVersion("v1")
	check.NoError(err)
	v2, err := version.NewVersion("v2")
	check.NoError(err)
	v3, err := version.NewVersion("v3")
	check.NoError(err)

	verSlice := []*version.Version{v1, v2, v3}
	ReverseVersion(verSlice)
	check.Equal(verSlice, []*version.Version{v3, v2, v1})

	verSlice = []*version.Version{v1}
	ReverseVersion(verSlice)
	check.Equal(verSlice, []*version.Version{v1})

	verSlice = []*version.Version{}
	ReverseVersion(verSlice)
	check.Equal(verSlice, []*version.Version{})

}
