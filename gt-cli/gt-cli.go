// GO111MODULE=off go get -u github.com/msoap/etc/gt-cli
// ln -s $GOPATH/bin/gt-cli $GOPATH/bin/gt-cli-en # for English as destination
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

const (
	userAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_6) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/11.1.2 Safari/605.1.15"
	baseURL   = "https://translate.googleapis.com/translate_a/single?client=gtx&sl=auto&tl=%s&dt=t&q=%s"
)

var reBin = regexp.MustCompile(`gt-cli-([a-z]{2})$`)

func main() {
	if len(os.Args) >= 2 && os.Args[1] == "--help" {
		fmt.Printf("usage: %s text\n", os.Args[0])
		return
	}

	var text string
	if isPipe(os.Stdin) {
		textBytes, err := ioutil.ReadAll(os.Stdin)
		errCheck(err)
		text = string(textBytes)
	} else {
		text = strings.Join(os.Args[1:], " ")
	}

	var to string
	if len(os.Args) > 0 {
		to = getLang(os.Args[0])
	}
	if to == "" {
		to = "ru"
		for _, r := range text {
			if unicode.In(r, unicode.Cyrillic) {
				to = "en"
				break
			}
		}
	}

	urlGT := fmt.Sprintf(baseURL, to, url.QueryEscape(text))
	resultRaw, err := getHTTP(urlGT)
	errCheck(err)
	if len(os.Getenv("DEBUG")) > 0 {
		fmt.Printf("url: %s\nresult: %s\n", urlGT, resultRaw)
	}

	resultRaw = regexp.MustCompile(`,+`).ReplaceAllString(resultRaw, ",")
	result := []interface{}{}
	err = json.Unmarshal([]byte(resultRaw), &result)
	errCheck(err)

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

func getLang(bin string) string {
	if m := reBin.FindStringSubmatch(bin); len(m) == 2 {
		return m[1]
	}

	return ""
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

func errCheck(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
