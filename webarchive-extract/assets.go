package main

// Filename derivation, text decoding, reference rewriting, inline-block
// splitting and data: URI extraction.

import (
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"mime"
	"net/url"
	"path"
	"regexp"
	"sort"
	"strings"
	"unicode/utf16"
	"unicode/utf8"
)

// ------------------------------------------------------------------- naming

// Preferred extensions; mime.ExtensionsByType is consulted as a fallback but
// returns unhelpful picks for common web types (e.g. ".jpe" for image/jpeg).
var mimeExt = map[string]string{
	"text/html":                     ".html",
	"application/xhtml+xml":         ".xhtml",
	"text/css":                      ".css",
	"text/javascript":               ".js",
	"application/javascript":        ".js",
	"application/x-javascript":      ".js",
	"text/ecmascript":               ".js",
	"application/json":              ".json",
	"application/ld+json":           ".json",
	"text/plain":                    ".txt",
	"text/xml":                      ".xml",
	"application/xml":               ".xml",
	"image/jpeg":                    ".jpg",
	"image/png":                     ".png",
	"image/gif":                     ".gif",
	"image/webp":                    ".webp",
	"image/avif":                    ".avif",
	"image/svg+xml":                 ".svg",
	"image/x-icon":                  ".ico",
	"image/vnd.microsoft.icon":      ".ico",
	"font/woff":                     ".woff",
	"font/woff2":                    ".woff2",
	"font/ttf":                      ".ttf",
	"font/otf":                      ".otf",
	"application/font-woff":         ".woff",
	"application/font-woff2":        ".woff2",
	"application/x-font-ttf":        ".ttf",
	"application/vnd.ms-fontobject": ".eot",
	"video/mp4":                     ".mp4",
	"video/webm":                    ".webm",
	"audio/mpeg":                    ".mp3",
}

func extForMIME(mimeType string) string {
	base := strings.ToLower(strings.TrimSpace(strings.SplitN(mimeType, ";", 2)[0]))
	if ext, ok := mimeExt[base]; ok {
		return ext
	}
	if exts, err := mime.ExtensionsByType(base); err == nil && len(exts) > 0 {
		return exts[0]
	}
	return ".bin"
}

func splitExt(name string) (stem, ext string) {
	ext = path.Ext(name)
	return strings.TrimSuffix(name, ext), ext
}

// fileNameForURL turns an asset URL into a readable, flat filename. Query
// strings are folded into the name so that "css?family=Material+Icons" does not
// collapse to a bare "css".
func fileNameForURL(raw, mimeType string) string {
	name, query := "", ""
	if u, err := url.Parse(raw); err == nil {
		name = path.Base(u.Path)
		if name == "/" || name == "." || name == ".." {
			name = ""
		}
		if name == "" && u.Host != "" {
			name = u.Host
		}
		query = u.RawQuery
	}
	if name == "" {
		name = "asset"
	}
	stem, ext := splitExt(sanitize(name))
	if stem == "" {
		stem = "asset"
	}
	if query != "" {
		// "v=2&x=1" reads better as "v2-x1" than as "v-2-x-1".
		if q := sanitize(strings.ReplaceAll(query, "=", "")); q != "" {
			stem += "-" + q
		}
	}
	if ext == "" || (mimeType != "" && looksGeneric(ext)) {
		ext = extForMIME(mimeType)
	}
	if len(stem) > 96 {
		stem = stem[:96]
	}
	return strings.Trim(stem, "-.") + ext
}

// looksGeneric reports whether an extension came from a path that is really a
// route rather than a filename (".php", ".aspx" and friends).
func looksGeneric(ext string) bool {
	switch strings.ToLower(ext) {
	case ".php", ".aspx", ".asp", ".jsp", ".cgi", ".do":
		return true
	}
	return false
}

var unsafeChars = regexp.MustCompile(`[^A-Za-z0-9._-]+`)
var dashRuns = regexp.MustCompile(`-{2,}`)

func sanitize(s string) string {
	s, _ = url.PathUnescape(s)
	s = unsafeChars.ReplaceAllString(s, "-")
	s = dashRuns.ReplaceAllString(s, "-")
	return strings.Trim(s, "-.")
}

func slug(s string) string {
	s = strings.ToLower(sanitize(s))
	if s == "" {
		return "page"
	}
	if len(s) > 64 {
		s = strings.Trim(s[:64], "-.")
	}
	return s
}

// ------------------------------------------------------------ text decoding

// decodeText converts resource bytes to a UTF-8 string, honouring the few
// encodings that turn up in practice.
func decodeText(data []byte, encoding, srcURL string) string {
	switch strings.ToLower(strings.TrimSpace(encoding)) {
	case "", "utf-8", "utf8", "us-ascii", "ascii":
		if utf8.Valid(data) {
			return strings.TrimPrefix(string(data), "\uFEFF")
		}
		return latin1(data) // mislabelled; single-byte is the safest guess
	case "utf-16", "utf-16le", "utf16le":
		return decodeUTF16(data, binary.LittleEndian)
	case "utf-16be", "utf16be":
		return decodeUTF16(data, binary.BigEndian)
	case "iso-8859-1", "latin1", "windows-1252", "cp1252":
		return latin1(data)
	default:
		if utf8.Valid(data) {
			return strings.TrimPrefix(string(data), "\uFEFF")
		}
		warnf("%s: unsupported text encoding %q, writing bytes unchanged", srcURL, encoding)
		return string(data)
	}
}

func decodeUTF16(data []byte, order binary.ByteOrder) string {
	u := make([]uint16, 0, len(data)/2)
	for i := 0; i+1 < len(data); i += 2 {
		u = append(u, order.Uint16(data[i:]))
	}
	return strings.TrimPrefix(string(utf16.Decode(u)), "\uFEFF")
}

func latin1(data []byte) string {
	var b strings.Builder
	b.Grow(len(data))
	for _, c := range data {
		b.WriteRune(rune(c))
	}
	return b.String()
}

func textualMIME(mimeType string) bool {
	base := strings.ToLower(strings.TrimSpace(strings.SplitN(mimeType, ";", 2)[0]))
	if strings.HasPrefix(base, "text/") {
		return true
	}
	switch base {
	case "application/javascript", "application/x-javascript", "application/json",
		"application/ld+json", "application/xml", "image/svg+xml",
		"application/xhtml+xml":
		return true
	}
	return false
}

func isTextual(mimeType, name string) bool {
	if textualMIME(mimeType) {
		return true
	}
	switch strings.ToLower(path.Ext(name)) {
	case ".css", ".js", ".mjs", ".json", ".svg", ".html", ".htm", ".xml", ".txt":
		return true
	}
	return false
}

// --------------------------------------------------------------- rewriting

// rewriter maps the original asset URLs onto their extracted filenames. Each
// URL yields several textual forms, because a page may reference the same asset
// absolutely, protocol-relatively, by root path or relatively.
type rewriter struct {
	forms   []string       // every spelling to look for
	asset   map[string]int // spelling -> index into extractor.assets
	baseURL *url.URL
	re      *regexp.Regexp
}

func newRewriter(mainURL string) *rewriter {
	r := &rewriter{asset: map[string]int{}}
	if u, err := url.Parse(mainURL); err == nil && u.Host != "" {
		r.baseURL = u
	}
	return r
}

func (r *rewriter) add(rawURL string, idx int) {
	if rawURL == "" {
		return
	}
	for _, form := range r.spellings(rawURL) {
		if _, dup := r.asset[form]; dup {
			continue // first resource to claim a spelling keeps it
		}
		r.asset[form] = idx
		r.forms = append(r.forms, form)
	}
	r.re = nil
}

// spellings lists the textual forms of rawURL that may appear in markup.
func (r *rewriter) spellings(rawURL string) []string {
	forms := []string{rawURL}
	if u, err := url.Parse(rawURL); err == nil && u.Host != "" {
		pathQuery := u.EscapedPath()
		if u.RawQuery != "" {
			pathQuery += "?" + u.RawQuery
		}
		if len(pathQuery) > 1 {
			forms = append(forms, "//"+u.Host+pathQuery)
			if r.baseURL != nil && strings.EqualFold(u.Host, r.baseURL.Host) {
				forms = append(forms, pathQuery)
				// Same-directory relative reference, e.g. "assets/app.css".
				if dir := path.Dir(r.baseURL.EscapedPath()); dir != "." && dir != "/" {
					if rel := strings.TrimPrefix(pathQuery, dir+"/"); rel != pathQuery &&
						strings.Contains(rel, "/") {
						forms = append(forms, rel, "./"+rel)
					}
				}
			}
		}
	}
	// Markup escapes ampersands in query strings.
	for _, f := range append([]string(nil), forms...) {
		if strings.Contains(f, "&") {
			forms = append(forms, strings.ReplaceAll(f, "&", "&amp;"))
		}
	}
	return forms
}

func (r *rewriter) compile() {
	if r.re != nil || len(r.forms) == 0 {
		return
	}
	// Longest first, because alternation in Go's regexp is leftmost-first: at a
	// given position the most specific spelling must win.
	forms := append([]string(nil), r.forms...)
	sort.SliceStable(forms, func(i, j int) bool { return len(forms[i]) > len(forms[j]) })
	quoted := make([]string, len(forms))
	for i, f := range forms {
		quoted[i] = regexp.QuoteMeta(f)
	}
	r.re = regexp.MustCompile(strings.Join(quoted, "|"))
}

// matches reports which assets text refers to, without rewriting anything.
func (r *rewriter) matches(text string) []int {
	r.compile()
	if r.re == nil {
		return nil
	}
	seen := map[int]bool{}
	var out []int
	for _, m := range r.re.FindAllString(text, -1) {
		if i, ok := r.asset[m]; ok && !seen[i] {
			seen[i] = true
			out = append(out, i)
		}
	}
	return out
}

// apply replaces every known reference in text, asking ref for the replacement.
// prefix is the embed directory as seen from the referencing file: "embed" for
// the HTML document, "" for a stylesheet that lives inside the embed folder.
//
// Matching happens in one pass over the text: replacing spelling by spelling
// would let a short root-relative form ("/frame.html") match inside output that
// an earlier, longer form had already produced ("embed/frame.html").
func (r *rewriter) apply(text, prefix string, ref func(idx int, prefix string) string) string {
	r.compile()
	if r.re == nil {
		return text
	}
	return r.re.ReplaceAllStringFunc(text, func(match string) string {
		i, ok := r.asset[match]
		if !ok {
			return match
		}
		return ref(i, prefix)
	})
}

// dataURI encodes an asset for embedding directly in the document. Text types
// carry an explicit charset, since data: URIs default to US-ASCII.
func dataURI(mimeType, name string, body []byte) string {
	base := strings.TrimSpace(strings.SplitN(mimeType, ";", 2)[0])
	if base == "" {
		if base = mime.TypeByExtension(path.Ext(name)); base == "" {
			base = "application/octet-stream"
		}
		base = strings.SplitN(base, ";", 2)[0]
	}
	if textualMIME(base) {
		base += ";charset=utf-8"
	}
	return "data:" + base + ";base64," + base64.StdEncoding.EncodeToString(body)
}

// -------------------------------------------------------------- data: URIs

var dataURIRe = regexp.MustCompile(`(?i)data:([a-z0-9!#$&^_.+-]+/[a-z0-9!#$&^_.+-]+)?((?:;[a-z0-9-]+=[^;,"')\s]*)*);base64,([A-Za-z0-9+/=]+)`)

// extractDataURIs pulls inline base64 payloads out into files. prefix is the
// embed directory as seen from the file being rewritten.
func (e *extractor) extractDataURIs(text, prefix string) string {
	return dataURIRe.ReplaceAllStringFunc(text, func(match string) string {
		m := dataURIRe.FindStringSubmatch(match)
		payload, err := base64.StdEncoding.DecodeString(m[3])
		if err != nil {
			return match
		}
		mimeType := m[1]
		if mimeType == "" {
			mimeType = "text/plain"
		}
		e.dataFiles++
		name := e.unique(fmt.Sprintf("%sdata-%d%s", e.prefix, e.dataFiles, extForMIME(mimeType)))
		if err := e.writeAsset(name, payload, entry{MIME: mimeType, Kind: "data-uri"}); err != nil {
			warnf("%s: %v", name, err)
			e.dataFiles--
			return match
		}
		if prefix != "" {
			return prefix + "/" + name
		}
		return name
	})
}
