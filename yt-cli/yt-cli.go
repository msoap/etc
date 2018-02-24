// go get -u github.com/msoap/etc/yt-cli
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
	"unicode"
)

const (
	tokenVarName = "YT_KEY"
	baseURL      = "https://translate.yandex.net/api/v1.5/tr.json/%s?key=%s&lang=%s&text=%s"
	timeOut      = 10
)

type translateResult struct {
	Code int      `json:"code"`
	Lang string   `json:"lang"`
	Text []string `json:"text"`
}

func main() {
	ytKey, text, err := getInputData()
	if err != nil {
		errCheck(err)
	}

	lang := "en-ru"
	for _, r := range text {
		if unicode.In(r, unicode.Cyrillic) {
			lang = "ru-en"
			break
		}
	}

	urlYT := fmt.Sprintf(baseURL, "translate", ytKey, lang, url.QueryEscape(text))
	resultRaw, err := getHTTP(urlYT)
	errCheck(err)
	if len(os.Getenv("DEBUG")) > 0 {
		fmt.Printf("url: %s\nresult: %s\n", urlYT, string(resultRaw))
	}

	result := translateResult{}
	err = json.Unmarshal(resultRaw, &result)
	errCheck(err)

	fmt.Println(strings.Join(result.Text, ""))
}

func getInputData() (ytKey string, text string, err error) {
	if len(os.Args) >= 2 && os.Args[1] == "--help" {
		fmt.Printf("usage: %s text\n", os.Args[0])
		os.Exit(0)
	}

	if os.Getenv(tokenVarName) == "" {
		return "", "", fmt.Errorf("%s required", tokenVarName)
	}
	ytKey = os.Getenv(tokenVarName)

	if isPipe(os.Stdin) {
		textBytes, err := ioutil.ReadAll(os.Stdin)
		errCheck(err)
		text = string(textBytes)
	} else {
		text = strings.Join(os.Args[1:], " ")
	}

	return ytKey, text, nil
}

func getHTTP(url string) ([]byte, error) {
	client := &http.Client{
		Timeout: timeOut * time.Second,
	}

	response, err := client.Get(url)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	if err := response.Body.Close(); err != nil {
		return nil, err
	}

	return body, nil
}

func isPipe(std *os.File) bool {
	if stdoutStat, err := std.Stat(); err != nil || (stdoutStat.Mode()&os.ModeCharDevice) == 0 {
		return true
	}

	return false
}

func errCheck(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
