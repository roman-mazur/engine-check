package enginecheck // import "rmazur.io/engine-check"

import (
	"github.com/docker/docker/api/types"
	"io/ioutil"
	"path/filepath"
	"regexp"
)

const overlay2 = "overlay2"

var pathOverlayIdRegex = regexp.MustCompile(".+/overlay2/([a-z0-9]+).*")
var fileNameOverlayIdRegex = regexp.MustCompile("^([a-z0-9]+).*")

// OverlayIdSet represents a set of overlay IDs.
type OverlayIdSet map[string]struct{}

// Slice creates a string slice that contains all IDs in the set.
func (s OverlayIdSet) Slice() []string {
	res := make([]string, len(s))
	index := 0
	for id := range s {
		res[index] = id
		index++
	}
	return res
}

func (s OverlayIdSet) add(id string) {
	s[id] = struct{}{}
}

// OverlaysFromGraphDriver extracts a set of OverlayFS overlay IDs mentioned in the GraphDriver info.
func OverlaysFromGraphDriver(gd types.GraphDriverData) OverlayIdSet {
	if gd.Name != overlay2 {
		return nil
	}
	res := make(OverlayIdSet)
	for _, str := range gd.Data {
		for _, ref := range filepath.SplitList(str) {
			if match := pathOverlayIdRegex.FindStringSubmatch(ref); match != nil {
				res.add(match[1])
			}
		}
	}
	return res
}

// AllOverlays returns all overlay IDs found in the specified directory (like /var/lib/docker/overlay2).
func AllOverlays(basePath string) (OverlayIdSet, error) {
	files, err := ioutil.ReadDir(basePath)
	if err != nil {
		return nil, err
	}
	res := make(OverlayIdSet)
	for _, info := range files {
		if match := fileNameOverlayIdRegex.FindStringSubmatch(info.Name()); match != nil {
			if match[1] != "l" {
				res.add(match[1])
			}
		}
	}
	return res, nil
}
