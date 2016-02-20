/*
github-releases-stat - get statistic of downloads from github releases

Usage:

	github-releases-stat [options] user repo_name
	-all
		get all releases (otherwise show the latest only)
	-summary
		show summary downloads
	-json
		JSON output

Install:

	go get -u github.com/google/go-github/github
	git clone https://gist.github.com/2c91d171d004fbcd0424.git github-releases-stat
	cd github-releases-stat
	go build github-releases-stat.go

Example:

	github-releases-stat -summary coreos etcd

Source:
	https://gist.github.com/msoap/2c91d171d004fbcd0424

*/
package main

import (
	"encoding/json"
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

type AssetOut struct {
	Name          string `json:"name"`
	DownloadCount int    `json:"download_count"`
}

// ReleaseOut - struct for json
type ReleaseOut struct {
	Name        string     `json:"name"`
	PublishedAt string     `json:"published_at"`
	HTMLURL     string     `json:"html_url"`
	TagName     string     `json:"tag_name"`
	Assets      []AssetOut `json:"assets"`
}

func (assets releaseAssetList) Len() int      { return len(assets) }
func (assets releaseAssetList) Swap(i, j int) { assets[i], assets[j] = assets[j], assets[i] }
func (assets releaseAssetList) Less(i, j int) bool {
	return *assets[i].DownloadCount < *assets[j].DownloadCount
}

func printOneRelease(release *github.RepositoryRelease) {
	fmt.Printf("%s (%s) - %s\n", *release.Name, *release.PublishedAt, *release.HTMLURL)
	sort.Sort(sort.Reverse(releaseAssetList(release.Assets)))
	for i, assets := range release.Assets {
		fmt.Printf("  %d. %-35s: %d\n", i, *assets.Name, *assets.DownloadCount)
	}
}

func outOneRelease(release *github.RepositoryRelease) (result ReleaseOut) {
	result.Name = *release.Name
	result.PublishedAt = (*release.PublishedAt).String()
	result.HTMLURL = *release.HTMLURL
	result.TagName = *release.TagName
	result.Assets = []AssetOut{}
	sort.Sort(sort.Reverse(releaseAssetList(release.Assets)))
	for _, assets := range release.Assets {
		result.Assets = append(result.Assets, AssetOut{*assets.Name, *assets.DownloadCount})
	}
	return result
}

func getSummary(assets []github.ReleaseAsset) (result int) {
	for _, asset := range assets {
		result += *asset.DownloadCount
	}
	return result
}

func main() {
	getAll, summary, getJSON := false, false, false
	flag.BoolVar(&getAll, "all", false, "get all releases")
	flag.BoolVar(&summary, "summary", false, "show summary downloads")
	flag.BoolVar(&getJSON, "json", false, "JSON output")
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
	releases := []github.RepositoryRelease{}

	if getAll {
		nextPage := 1
		for nextPage != 0 {
			releasesChunk, response, err := client.Repositories.ListReleases(flag.Args()[0], flag.Args()[1], &github.ListOptions{Page: nextPage, PerPage: ItemsPerPage})
			if err != nil {
				log.Fatal(err)
			}
			releases = append(releases, releasesChunk...)
			nextPage = response.NextPage
		}

	} else {

		release, _, err := client.Repositories.GetLatestRelease(flag.Args()[0], flag.Args()[1])
		if err != nil {
			log.Fatal(err)
		}
		releases = append(releases, *release)

	}

	allDownloads := 0
	jsonAllOut := struct {
		AllDownloads int          `json:"all_downloads"`
		Releases     []ReleaseOut `json:"releases"`
	}{}
	for _, release := range releases {
		if getJSON {
			releaseOut := outOneRelease(&release)
			jsonAllOut.Releases = append(jsonAllOut.Releases, releaseOut)
		} else {
			printOneRelease(&release)
		}
		allDownloads += getSummary(release.Assets)
	}
	if summary && !getJSON {
		fmt.Printf("\n%d downloads\n", allDownloads)
	}
	if getJSON {
		jsonAllOut.AllDownloads = allDownloads
		json, err := json.MarshalIndent(jsonAllOut, "", "  ")
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(string(json))
	}
}
