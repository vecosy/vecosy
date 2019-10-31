package utils

import "github.com/hashicorp/go-version"

func ReverseVersion(versions []*version.Version) {
	for i := len(versions)/2 - 1; i >= 0; i-- {
		opp := len(versions) - 1 - i
		versions[i], versions[opp] = versions[opp], versions[i]
	}
}
func ReverseStrings(versions []string) {
	for i := len(versions)/2 - 1; i >= 0; i-- {
		opp := len(versions) - 1 - i
		versions[i], versions[opp] = versions[opp], versions[i]
	}
}
