package main

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"flag"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"log"
	enginecheck "rmazur.io/engine-check"
	"sort"
)

var (
	apiVersion = flag.String("api-version", "1.35", "docker API version to use")
	overlaysDir = flag.String("overlays-dir", "/var/lib/docker/overlay2", "overlays directory base path")

	skipUnused = flag.Bool("skip-unused", false, "don't print unused overlays")
	showUsed = flag.Bool("show-used", false, "print overlays usage summary")
	showAll = flag.Bool("show-all", false, "print all overlay IDs")
)

func main()  {
	sum := sha1.Sum([]byte("hello"))
	fmt.Println(hex.EncodeToString(sum[:]))
	flag.Parse()

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithVersion(*apiVersion))
	if err != nil {
		log.Fatalf("Cannot init Docker client: %s", err)
	}

	ctx := context.Background()
	usageData := make(usageInfo)

	images, err := cli.ImageList(ctx, types.ImageListOptions{All: true})
	if err != nil {
		log.Fatalf("Failed to list all images: %s", err)
	}
	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{All: true})
	if err != nil {
		log.Fatalf("Failed to list all containers: %s", err)
	}

	for _, imgSummary := range images {
		if image, _, err := cli.ImageInspectWithRaw(ctx, imgSummary.ID); err == nil {
			usageData.addImage(image, enginecheck.OverlaysFromGraphDriver(image.GraphDriver))
		} else {
			log.Printf("Failed to inspect image %s: %s", imgSummary.ID, err)
		}
	}

	for _, cSummary := range containers {
		if container, err := cli.ContainerInspect(ctx, cSummary.ID); err == nil {
			usageData.addContainer(container, enginecheck.OverlaysFromGraphDriver(container.GraphDriver))
		} else {
			log.Printf("Failed to inspect container %s: %s", cSummary.ID, err)
		}
	}

	if *showUsed {
		fmt.Println("Overlays Usage Summary")
		fmt.Println("======================")
		for _, oui := range usageData {
			fmt.Println(oui.String())
		}
	}
	allIds, err := enginecheck.AllOverlays(*overlaysDir)
	if !*skipUnused {
		if err != nil {
			log.Fatalf("Failed to read overlays directory %s: %s", *overlaysDir, err)
		}

		fmt.Println("Unused overlays")
		fmt.Println("===============")
		for _, id := range usageData.selectUnused(allIds) {
			fmt.Println(id)
		}
	}
	if *showAll {
		fmt.Println("All overlays")
		fmt.Println("============")
		sorted := allIds.Slice()
		sort.Strings(sorted)
		for _, id := range sorted {
			fmt.Println(id)
		}
	}
}
