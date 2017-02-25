package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"
	"unicode"
)

const (
	UA = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_2) AppleWebKit/602.4.8 (KHTML, like Gecko) Version/10.0.3 Safari/602.4.7"
)

var (
	parseRe = regexp.MustCompile(`\[\[\["(.+?)","`)
	baseURL = strings.Join([]string{"https:", "//", "translate", ".goog", "leap", "is.", "com/trans", "late_a/sin", "gle?client=g", "tx&sl=auto&tl=%s&dt=t&q=%s"}, "")
)

func main() {
	if len(os.Args) < 2 || os.Args[1] == "--help" {
		fmt.Printf("usage: %s text\n", os.Args[0])
		return
	}

	text := strings.Join(os.Args[1:], " ")
	to := "ru"
	for _, r := range text {
		if unicode.In(r, unicode.Cyrillic) {
			to = "en"
			break
		}
	}

	urlGT := fmt.Sprintf(baseURL, to, url.QueryEscape(text))
	resultRaw, err := getHTTP(urlGT)
	if err != nil {
		fmt.Println(err)
		return
	}
	if len(os.Getenv("DEBUG")) > 0 {
		fmt.Printf("url: %s\nresult: %s\n", urlGT, resultRaw)
	}

	result := parseRe.FindStringSubmatch(resultRaw)
	if len(result) == 2 && len(result[1]) > 0 {
		fmt.Println(result[1])
	}
}

func getHTTP(url string) (string, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	request.Header.Set("User-Agent", UA)
	response, err := client.Do(request)
	if err != nil {
		return "", err
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	if err := response.Body.Close(); err != nil {
		return "", err
	}

	return string(body), nil
}
