// go build -ldflags '-s -w' -trimpath -o "$(go env GOPATH)/bin/gt-cli" gt-cli/gt-cli.go
// rlwrap gt-cli -chat # for history and input editing
// ln -s $GOPATH/bin/gt-cli $GOPATH/bin/gt-cli-en # for English as destination
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"
	"unicode"
)

const (
	userAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/18.1 Safari/605.1.15"
	baseURL   = "https://translate.googleapis.com/translate_a/single"
	timeOut   = 10 * time.Second
)

var (
	reBin     = regexp.MustCompile(`gt-cli-([a-z]{2})$`)
	reCommas  = regexp.MustCompile(`,+`)
	reLangCmd = regexp.MustCompile(`^/([a-z]{2}) `)
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
		textBytes, err := io.ReadAll(os.Stdin)
		if err != nil {
			log.Fatalf("read error: %s", err)
		}
		text = string(textBytes)
	} else {
		text = strings.Join(os.Args[1:], " ")
	}

	if text == "" {
		return
	}

	to := getLangByBin()
	if to == "" {
		to = detectLang(text)
	}

	tr, err := translate(to, text)
	if err != nil {
		log.Fatalf("translate error: %s", err)
	}
	fmt.Println(tr)
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
	defaultTo := getLangByBin()
	prevTr := ""
	for {
		fmt.Print("❱❱ ")
		text, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("read error: %s", err)
			continue
		}

		text = strings.TrimSpace(text)
		if text == "" {
			continue
		}

		if text == "/exit" {
			break
		}

		if (text == "/prev" || text == "^") && prevTr != "" {
			text = prevTr
		}

		to := defaultTo
		if to == "" {
			to = detectLang(text)
		}

		text = strings.TrimSpace(reLangCmd.ReplaceAllString(text, ""))

		tr, err := translate(to, text)
		if err != nil {
			log.Printf("translate error: %s", err)
			continue
		}
		prevTr = tr

		fmt.Println("\033[33m" + tr + "\033[0m")
	}
}

func translate(to string, text string) (string, error) {
	params := url.Values{}
	params.Set("client", "gtx")
	params.Set("sl", "auto")
	params.Set("tl", to)
	params.Set("dt", "t")
	params.Set("q", text)
	urlGT := baseURL + "?" + params.Encode()

	resultRaw, err := getHTTP(urlGT)
	if err != nil {
		return "", err
	}

	if len(os.Getenv("DEBUG")) > 0 {
		fmt.Printf("url: %s\nresult: %s\n", urlGT, resultRaw)
	}

	resultRaw = reCommas.ReplaceAllString(resultRaw, ",")
	result := []interface{}{}
	err = json.Unmarshal([]byte(resultRaw), &result)
	if err != nil {
		return "", err
	}

	trTexts := []string{}
	if len(result) > 0 && len(result[0].([]interface{})) > 0 {
		for _, item := range result[0].([]interface{}) {
			if trText, ok := item.([]interface{})[0].(string); ok && trText != "" {
				trTexts = append(trTexts, trText)
			}
		}
	}

	return strings.Join(trTexts, ""), nil
}

func getLangByBin() string {
	if len(os.Args) == 0 {
		return ""
	}
	if m := reBin.FindStringSubmatch(os.Args[0]); len(m) == 2 {
		return m[1]
	}

	return ""
}

func getHTTP(url string) (string, error) {
	client := &http.Client{
		Timeout: timeOut,
	}

	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}

	request.Header.Set("User-Agent", userAgent)

	response, err := client.Do(request)
	if err != nil {
		return "", err
	}

	defer func() {
		if err := response.Body.Close(); err != nil {
			log.Printf("response.Body.Close error: %s", err)
		}
	}()

	body, err := io.ReadAll(response.Body)
	if err != nil {
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
