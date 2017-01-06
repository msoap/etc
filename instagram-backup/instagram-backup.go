/*
Get token:

register app: https://www.instagram.com/developer/clients/manage/

get access token:

https://api.instagram.com/oauth/authorize/?client_id=XXXXXXXXX&redirect_uri=http://XXXXXXX/&response_type=code

curl -F 'client_id=XXXXXXX' \
    -F 'client_secret=XXXXXXX' \
    -F 'grant_type=authorization_code' \
    -F 'redirect_uri=http://XXXXXXXX' \
    -F 'code=XXXXXXXX' \
    https://api.instagram.com/oauth/access_token

run:

    INSTAGRAM_TOKEN=XXXXXX.XXXXXX.XXXXXXXX go run instagram-backup.go

install:

	go build -o ~/bin/instagram-backup instagram-backup.go
*/
package main

import (
	"fmt"
	"log"
	"net/url"
	"os"

	"github.com/yanatan16/golang-instagram/instagram"
)

func main() {
	accessTocken := os.Getenv("INSTAGRAM_TOKEN")
	if accessTocken == "" {
		log.Fatal("Need INSTAGRAM_TOKEN environment var")
	}

	inst := instagram.New("", "", accessTocken, false)
	params := url.Values{}
	recent, err := inst.GetUserRecentMedia("self", params)
	if err != nil {
		log.Fatal(err)
	}

	doneChan := make(chan bool)
	mediaIter, errChan := inst.IterateMedia(recent, doneChan)
	for media := range mediaIter {
		processMedia(media)

		if isDone(media) {
			close(doneChan)
			break
		}
	}

	if err := <-errChan; err != nil {
		log.Fatal(err)
	}
}

func processMedia(media *instagram.Media) {
	createdTime, _ := media.CreatedTime.Time()
	url := ""
	if media.Images != nil {
		url = media.Images.StandardResolution.Url
	} else if media.Videos != nil {
		url = media.Videos.StandardResolution.Url
	}

	camption := ""
	if media.Caption != nil {
		camption = media.Caption.Text
	}

	location := ""
	if media.Location != nil {
		location = fmt.Sprintf("%s (Lat: %f, Lon: %f)", media.Location.Name, media.Location.Latitude, media.Location.Longitude)
	}

	fmt.Printf("ID: %s, created: %s, url: %s, %s, %s\n", media.Id, createdTime, url, camption, location)
}

func isDone(media *instagram.Media) bool {
	return false
}
