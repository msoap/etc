// GO111MODULE=off go get -u github.com/msoap/etc/gt-cli
// ln -s $GOPATH/bin/gt-cli $GOPATH/bin/gt-cli-en # for English as destination
package main

import (
	"bufio"
	"encoding/json"
	"flag"
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
	userAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:101.0) Gecko/20100101 Firefox/101.0"
	baseURL   = "https://translate.googleapis.com/translate_a/single?client=gtx&sl=auto&tl=%s&dt=t&q=%s"
)

var (
	reBin     = regexp.MustCompile(`gt-cli-([a-z]{2})$`)
	reCommas  = regexp.MustCompile(`,+`)
	reLangCmd = regexp.MustCompile(`^/([a-z]{2})`)
)

func main() {
	help := flag.Bool("help", false, "show help")
	chat := flag.Bool("chat", false, "chat mode")
	flag.Parse()

	if *help {
		fmt.Printf("usage: %s text | -chat\n", os.Args[0])
		return
	}

	if *chat {
		chatMode()
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

	if text == "" {
		return
	}

	var to string
	if len(os.Args) > 0 {
		to = getLang(os.Args[0])
	}
	if to == "" {
		to = detectLang(text)
	}

	fmt.Println(translate(to, text))
}

func detectLang(text string) string {
	if lang := reLangCmd.FindStringSubmatch(text); len(lang) == 2 {
		return lang[1]
	}

	to := "ru"
	for _, r := range text {
		if unicode.In(r, unicode.Cyrillic) {
			to = "en"
			break
		}
	}

	return to
}

func chatMode() {
	for {
		fmt.Print("> ")
		text, err := bufio.NewReader(os.Stdin).ReadString('\n')
		errCheck(err)

		text = strings.TrimSpace(text)
		if text == "" {
			continue
		}

		if text == "/exit" {
			break
		}
		to := detectLang(text)
		text = strings.TrimSpace(reLangCmd.ReplaceAllString(text, ""))

		fmt.Println(translate(to, text))
	}
}

func translate(to string, text string) string {
	urlGT := fmt.Sprintf(baseURL, to, url.QueryEscape(text))
	resultRaw, err := getHTTP(urlGT)
	errCheck(err)
	if len(os.Getenv("DEBUG")) > 0 {
		fmt.Printf("url: %s\nresult: %s\n", urlGT, resultRaw)
	}

	resultRaw = reCommas.ReplaceAllString(resultRaw, ",")
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

	return strings.Join(trTexts, "")
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
