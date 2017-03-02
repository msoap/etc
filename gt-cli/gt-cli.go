package main

import (
	"encoding/json"
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

const userAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_2) AppleWebKit/602.4.8 (KHTML, like Gecko) Version/10.0.3 Safari/602.4.7"

var (
	baseURL = strings.Join([]string{"https:", "//", "translate", ".goog", "leap", "is.", "com/trans", "late_a/sin", "gle?client=g", "tx&sl=auto&tl=%s&dt=t&q=%s"}, "")
	// baseURL = "http://localhost:8080/t?tl%s&q=%s"
)

func main() {
	if len(os.Args) >= 2 && os.Args[1] == "--help" {
		fmt.Printf("usage: %s text\n", os.Args[0])
		return
	}

	var text string
	if isPipe(os.Stdin) {
		textBytes, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			fmt.Println(err)
			return
		}
		text = string(textBytes)
	} else {
		text = strings.Join(os.Args[1:], " ")
	}

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

	resultRaw = regexp.MustCompile(`,+`).ReplaceAllString(resultRaw, ",")
	result := []interface{}{}
	err = json.Unmarshal([]byte(resultRaw), &result)
	if err != nil {
		fmt.Println(err)
		return
	}

	trTexts := []string{}
	if len(result) > 0 && len(result[0].([]interface{})) > 0 {
		for _, item := range result[0].([]interface{}) {
			if trText, ok := item.([]interface{})[0].(string); ok && trText != "" {
				trTexts = append(trTexts, trText)
			}
		}
	}

	fmt.Println(strings.Join(trTexts, ""))
}

func getHTTP(url string) (string, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	request.Header.Set("User-Agent", userAgent)
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

func isPipe(std *os.File) bool {
	if stdoutStat, err := std.Stat(); err != nil || (stdoutStat.Mode()&os.ModeCharDevice) == 0 {
		return true
	}

	return false
}
