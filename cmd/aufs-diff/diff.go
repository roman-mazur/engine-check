package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"sort"
	"strings"
)

var idPattern = regexp.MustCompile("\\w{64}(-init)?")

func extractIds(text string) []string {
	return idPattern.FindAllString(text, -1)
}

func removeContainers(all []string) []string {
	data := make(map[string]struct{})
	containers := make(map[string]struct{})
	for _, id := range all {
		if strings.HasSuffix(id, "-init") {
			containerId := strings.TrimSuffix(id, "-init")
			containers[containerId] = struct{}{}
		} else {
			data[id] = struct{}{}
		}
	}
	for containerId := range containers {
		delete(data, containerId)
	}
	res := make([]string, len(data))
	i := 0
	for id := range data {
		res[i] = id
		i++
	}
	return res
}

func main() {

	// Put output of ls -la /var/lib/docker/aufs/diff/ into "all" file.
	allFile, e := os.Open("./all")
	if e != nil { panic(e) }
	defer allFile.Close()

	// Put output of (cd /var/lib/docker/image/aufs/layerdb/sha256/ && ls | xargs -I {} cat '{}/cache-id' -n) into "used" file.
	usedFile, e := os.Open("./used")
	if e != nil { panic(e) }
	defer usedFile.Close()

	allData, _ := ioutil.ReadAll(allFile)
	all := removeContainers(extractIds(string(allData)))

	usedData, _ := ioutil.ReadAll(usedFile)
	used := extractIds(string(usedData))

	sort.Strings(all)
	sort.Strings(used)

	log.Println("Number of non-container AUFS IDs:", len(all))
	log.Println("Number of used AUFS IDs:", len(used))
	log.Println("Count of directories to delete:", len(all) - len(used))

	fmt.Println("echo 'Start deletion...'")
	cnt := 0
	for _, id := range all {
		if i := sort.SearchStrings(used, id); i == len(used) || used[i] != id {
			fmt.Println(fmt.Sprintf("rm -rf /var/lib/docker/aufs/diff/%s", id))
			cnt++
			if cnt % 5 == 0 {
				fmt.Printf("echo 'Deleted %d directories'\n", cnt)
			}
		}
	}
	fmt.Printf("echo 'Finished deletion of %d directories'\n", cnt)
}
