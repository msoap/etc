// go get -u github.com/msoap/etc/yt-cli
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
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
	Code    int      `json:"code"`
	Lang    string   `json:"lang"`
	Text    []string `json:"text"`
	Message string   `json:"message"`
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

	urlYT := fmt.Sprintf(baseURL, "translate", url.QueryEscape(ytKey), url.QueryEscape(lang), url.QueryEscape(text))
	result := translateResult{}
	err = callAPI(urlYT, &result)
	errCheck(err)

	switch result.Code {
	case 200:
		fmt.Println(strings.Join(result.Text, ""))
	case 401:
		fmt.Printf("Error: %s\n", result.Message)
	default:
		fmt.Printf("Unknown result code: %d, message: %s\n", result.Code, result.Message)
	}
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

func callAPI(url string, result interface{}) error {
	client := &http.Client{
		Timeout: timeOut * time.Second,
	}

	response, err := client.Get(url)
	if err != nil {
		return err
	}

	defer response.Body.Close()

	var apiJSON io.Reader = response.Body

	if os.Getenv("DEBUG") != "" {
		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return err
		}
		fmt.Printf("url: %s\nresult: %s\n", url, string(body))
		apiJSON = bytes.NewReader(body)
	}

	if err := json.NewDecoder(apiJSON).Decode(result); err != nil {
		return err
	}

	return nil
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
