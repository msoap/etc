/*
	habr.com-get-favorites - get favorites from habrahabr sites

	Install:
		go get -u github.com/msoap/etc/habr.com-get-favorites

	Usage:
		habr.com-get-favorites [user_name] > habr-favorites.txt
*/
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/msoap/html2data"
)

var hosts = []string{"https://habr.com", "https://geektimes.com"}

type item struct {
	url   string
	title string
}

func main() {
	userName := os.Getenv("USER")
	if len(os.Args) == 2 {
		userName = os.Args[1]
	}

	for _, host := range hosts {
		url := fmt.Sprintf("%s/users/%s/favorites/", host, userName)
		result, err := getFromURL(host, url)
		if err != nil {
			log.Printf("failed to parse %s: %s", url, err)
		}

		fmt.Printf("%s favorites for %s\n-----------------------------------\n", host, userName)
		for i := len(result) - 1; i >= 0; i-- {
			fmt.Printf("%s %s\n", result[i].url, result[i].title)
		}
		fmt.Println()
	}
}

func getFromURL(host, habrUrl string) ([]item, error) {
	result := []item{}
	doc := html2data.FromURL(habrUrl)

	// parse links
	links, err := doc.GetDataNested("h2.post__title", map[string]string{
		"title": "a.post__title_link",
		"url":   "a.post__title_link:attr(href)",
	})
	if err != nil {
		return nil, err
	}

	for _, row := range links {
		if len(row["title"]) == 0 || len(row["url"]) == 0 {
			return nil, fmt.Errorf("failed to parse link: %+v", row)
		}
		result = append(result, item{
			title: row["title"][0],
			url:   row["url"][0],
		})
	}

	// parse next page
	nextPage, err := doc.GetDataSingle("div.page__footer > ul > li a[id=next_page]:attr(href)")
	if err != nil {
		return nil, err
	}

	if len(nextPage) > 0 {
		nextItems, err := getFromURL(host, host+nextPage)
		if err != nil {
			return nil, err
		}
		result = append(result, nextItems...)
	}

	return result, nil
}
