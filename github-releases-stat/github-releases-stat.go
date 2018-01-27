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

	go get -u github.com/msoap/etc/github-releases-stat

Example:

	github-releases-stat -summary coreos etcd

Source:
	https://github.com/msoap/etc/tree/master/github-releases-stat

*/
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"time"

	"github.com/google/go-github/github"
)

// ItemsPerPage - github pagination size
const ItemsPerPage = 10

// AssetOut - one asset with stat
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

func printOneRelease(release ReleaseOut) {
	fmt.Printf("%s (%s) - %s\n", release.Name, release.PublishedAt, release.HTMLURL)
	for i, assets := range release.Assets {
		fmt.Printf("  %d. %-35s: %d\n", i, assets.Name, assets.DownloadCount)
	}
}

func getOneRelease(release *github.RepositoryRelease) (result ReleaseOut) {
	result.Name = *release.Name
	result.PublishedAt = (*release.PublishedAt).String()
	result.HTMLURL = *release.HTMLURL
	result.TagName = *release.TagName
	result.Assets = []AssetOut{}
	sort.SliceStable(release.Assets, func(a int, b int) bool { return *release.Assets[a].DownloadCount > *release.Assets[b].DownloadCount })
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
	releases := []*github.RepositoryRelease{}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	if getAll {
		nextPage := 1
		for nextPage != 0 {
			releasesChunk, response, err := client.Repositories.ListReleases(ctx, flag.Args()[0], flag.Args()[1], &github.ListOptions{Page: nextPage, PerPage: ItemsPerPage})
			if err != nil {
				log.Fatal(err)
			}
			releases = append(releases, releasesChunk...)
			nextPage = response.NextPage
		}

	} else {

		release, _, err := client.Repositories.GetLatestRelease(ctx, flag.Args()[0], flag.Args()[1])
		if err != nil {
			log.Fatal(err)
		}
		releases = append(releases, release)

	}

	allDownloads := 0
	jsonAllOut := struct {
		AllDownloads int          `json:"all_downloads"`
		Releases     []ReleaseOut `json:"releases"`
	}{}
	for _, release := range releases {
		if getJSON {
			releaseOut := getOneRelease(release)
			jsonAllOut.Releases = append(jsonAllOut.Releases, releaseOut)
		} else {
			printOneRelease(getOneRelease(release))
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
