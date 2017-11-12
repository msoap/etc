// Parse emoji flags from https://github.com/matiassingers/emoji-flags
// Usage:
//  go run parse_unicode_flags_json.go
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

// {
//     "code": "AD",
//     "emoji": "🇦🇩",
//     "unicode": "U+1F1E6 U+1F1E9",
//     "name": "Andorra",
//     "title": "flag for Andorra"
// },
type flagItem struct {
	Code  string `json:"code"`
	Emoji string `json:"emoji"`
	Name  string `json:"name"`
}

const jsonURL = "https://github.com/matiassingers/emoji-flags/raw/master/data.json"
const outFileName = "flags.go"
const tmpl = `package main

type flagItem struct {
	Code  string
	Emoji string
	Name  string
}

var flags = []flagItem{
%s
}
`

func main() {
	resp, err := http.Get(jsonURL)
	errCheck(err)
	defer func() {
		errCheck(resp.Body.Close())
	}()

	flags := []flagItem{}
	errCheck(json.NewDecoder(resp.Body).Decode(&flags))

	items := []string{}
	for _, item := range flags {
		row := fmt.Sprintf("%#v", item)
		row = strings.TrimPrefix(row, "main.flagItem")
		items = append(items, row+",")
	}

	goFile, err := os.Create(outFileName)
	errCheck(err)
	defer func() {
		errCheck(goFile.Close())
	}()

	fmt.Fprintf(goFile, tmpl, strings.Join(items, "\n"))
	errCheck(exec.Command("gofmt", "-w", "-s", outFileName).Run())
}

func errCheck(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
