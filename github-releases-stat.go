/*
github-releases-stat - get statistic of downloads from github releases

Usage:

	github-releases-stat [options] user repo_name
	-all
    	get all releases (otherwise show the latest only)
	-summary
		show summary downloads


Install:

	go get -u github.com/google/go-github/github
	go build github-releases-stat.go

*/
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"sort"

	"github.com/google/go-github/github"
)

// ItemsPerPage - github pagination size
const ItemsPerPage = 10

type releaseAssetList []github.ReleaseAsset

func (assets releaseAssetList) Len() int      { return len(assets) }
func (assets releaseAssetList) Swap(i, j int) { assets[i], assets[j] = assets[j], assets[i] }
func (assets releaseAssetList) Less(i, j int) bool {
	return *assets[i].DownloadCount < *assets[j].DownloadCount
}

func printOneRelease(release *github.RepositoryRelease) {
	fmt.Printf("%s - %s\n", *release.Name, *release.HTMLURL)
	sort.Sort(sort.Reverse(releaseAssetList(release.Assets)))
	for i, assets := range release.Assets {
		fmt.Printf("  %d. %-35s: %d\n", i, *assets.Name, *assets.DownloadCount)
	}
}

func getSummary(assets []github.ReleaseAsset) (result int) {
	for _, asset := range assets {
		result += *asset.DownloadCount
	}
	return result
}

func main() {
	getAll, summary := false, false
	flag.BoolVar(&getAll, "all", false, "get all releases")
	flag.BoolVar(&summary, "summary", false, "show summary downloads")
	flag.Usage = func() {
		fmt.Printf("%s [options] user repo_name\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(0)
	}
	flag.Parse()

	if len(flag.Args()) != 2 {
		log.Fatal("Require 'user' and 'repo' name")
	}

	client := github.NewClient(nil)

	if getAll {
		releases := []github.RepositoryRelease{}
		nextPage := 1
		for nextPage != 0 {
			releasesChunk, response, err := client.Repositories.ListReleases(flag.Args()[0], flag.Args()[1], &github.ListOptions{Page: nextPage, PerPage: ItemsPerPage})
			if err != nil {
				log.Fatal(err)
			}
			releases = append(releases, releasesChunk...)
			nextPage = response.NextPage
		}

		allDowmloads := 0
		for _, release := range releases {
			printOneRelease(&release)
			allDowmloads += getSummary(release.Assets)
		}

		if summary {
			fmt.Printf("\n%d downloads\n", allDowmloads)
		}

	} else {

		release, _, err := client.Repositories.GetLatestRelease(flag.Args()[0], flag.Args()[1])
		if err != nil {
			log.Fatal(err)
		}
		printOneRelease(release)
		if summary {
			fmt.Printf("\n%d downloads\n", getSummary(release.Assets))
		}

	}
}
