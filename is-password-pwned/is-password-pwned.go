/*
Is your password pwned?

See: https://blog.cloudflare.com/using-cloudflare-workers-to-identify-pwned-passwords/

Install:
    GO111MODULE=off go get -u github.com/msoap/etc/is-password-pwned

Usage:
    is-password-pwned 'password'
*/
package main

import (
	"crypto/sha1"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/mgutz/ansi"
	"github.com/msoap/byline"
)

const apiUrl = "https://api.pwnedpasswords.com/range/"

func main() {
	if result, err := checkPass(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	} else {
		fmt.Println(result)
	}
}

func checkPass() (string, error) {
	if len(os.Args) != 2 {
		return "", fmt.Errorf("Usage:\n  is-password-pwned 'pasword'")
	}

	hashOfPass := sha1Hex(os.Args[1])
	partOne, partTwo := hashOfPass[:5], hashOfPass[5:]

	url := apiUrl + partOne
	httpResp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to load url %q: %s\n", url, err)
	}

	count := ""
	if err := byline.NewReader(httpResp.Body).AWKMode(func(line string, fields []string, vars byline.AWKVars) (string, error) {
		if vars.NF != 2 {
			return "", fmt.Errorf("failed to parse line: %q", line)
		}
		if partTwo == fields[0] {
			count = strings.TrimSpace(fields[1])
			return "", io.EOF
		}

		return line, nil
	}).SetFS(regexp.MustCompile(":")).Discard(); err != nil {
		return "", fmt.Errorf("failed to read http stream: %s", err)
	}

	if err := httpResp.Body.Close(); err != nil {
		return "", fmt.Errorf("failed to close http stream: %s", err)
	}

	if count != "" {
		return fmt.Sprintf("Password is %s %s times", ansi.ColorFunc("red")("PWNED"), count), nil
	}

	return fmt.Sprintf("Password is %s pwned", ansi.ColorFunc("green")("NOT")), nil
}

func sha1Hex(in string) string {
	h := sha1.New()
	io.WriteString(h, in)
	return strings.ToUpper(fmt.Sprintf("%x", h.Sum(nil)))
}
