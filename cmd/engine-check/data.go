package main

import (
	"fmt"
	"github.com/docker/docker/api/types"
	enginecheck "rmazur.io/engine-check"
	"sort"
	"strings"
)

const minIdLength = 12

type overlayUsageInfo struct {
	overlayId      string
	imageIds       []string
	imageNames     []string
	containerIds   []string
	containerNames []string
}

func (oui *overlayUsageInfo) String() string {
	if len(oui.containerNames) != len(oui.containerIds) {
		panic(fmt.Errorf("inconsistent containers data for %s: %#v", oui.overlayId, oui))
	}
	if len(oui.imageNames) != len(oui.imageIds) {
		panic(fmt.Errorf("inconsistent images data for %s: %#v", oui.overlayId, oui))
	}

	return oui.overlayId + " " +
		concatNames("images", oui.imageNames, oui.imageIds) + "; " +
		concatNames("containers", oui.containerNames, oui.containerIds)
}

func concatNames(prefix string, names []string, ids []string) string {
	res := prefix + ":"
	for i, name := range names {
		if len(name) > 0 {
			res += " " + name
		} else {
			res += " " + strings.TrimPrefix(ids[i], "sha256:")[0:minIdLength]
		}
	}
	return res
}

type usageInfo map[string]*overlayUsageInfo

func (ui usageInfo) lookup(overlayId string) *overlayUsageInfo {
	oui, present := ui[overlayId]
	if !present {
		oui = &overlayUsageInfo{overlayId: overlayId}
		ui[overlayId] = oui
	}
	return oui
}

func (ui usageInfo) addImage(image types.ImageInspect, overlays enginecheck.OverlayIdSet) {
	for overlayId := range overlays {
		oui := ui.lookup(overlayId)
		oui.imageIds = append(oui.imageIds, image.ID)
		name := ""
		if len(image.RepoTags) > 0 {
			name = image.RepoTags[0]
		}
		oui.imageNames = append(oui.imageNames, name)
	}
}

func (ui usageInfo) addContainer(container types.ContainerJSON, overlays enginecheck.OverlayIdSet) {
	for overlayId := range overlays {
		oui := ui.lookup(overlayId)
		oui.containerIds = append(oui.containerIds, container.ID)
		oui.containerNames = append(oui.containerNames, container.Name)
	}
}

func (ui usageInfo) selectUnused(ids enginecheck.OverlayIdSet) []string {
	var res []string
	for id := range ids {
		if _, present := ui[id]; !present {
			res = append(res, id)
		}
	}
	sort.Strings(res)
	return res
}
