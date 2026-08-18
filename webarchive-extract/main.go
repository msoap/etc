// Command webarchive-extract unpacks Safari .webarchive files into a plain
// HTML file plus a folder of assets.
//
//	webarchive-extract *.webarchive
//
// It writes <archive name>.html into the -out directory. If -out ends in
// .html (or .htm) it is taken as the exact output file instead, and the embed
// directory is created beside it. By default every embedded stylesheet,
// script, image and font is folded into that file as a data: URI, so the
// result is a single self-contained page. With -inline-max, assets at or
// above the given size are written to ./embed instead and referenced from
// there. Whatever the page already carried inline - <style> and <script>
// blocks, data: URIs - is left exactly as it was.
//
// Set WEBARCHIVE_EMBED_DIR to use a directory other than "embed".
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

const (
	keyMain     = "WebMainResource"
	keySubs     = "WebSubresources"
	keySubframe = "WebSubframeArchives"

	// embedEnv overrides the name of the assets directory.
	embedEnv     = "WEBARCHIVE_EMBED_DIR"
	defaultEmbed = "embed"
)

// embedDirName is the assets directory, relative to -out.
func embedDirName() string {
	if name := strings.TrimSpace(os.Getenv(embedEnv)); name != "" {
		if filepath.IsAbs(name) {
			warnf("%s must be relative to -out, ignoring %q", embedEnv, name)
			return defaultEmbed
		}
		return filepath.Clean(name)
	}
	return defaultEmbed
}

// namesHTMLFile reports whether -out is an exact output file rather than a
// directory. Such a path fixes the HTML filename and puts the embed directory
// alongside it.
func namesHTMLFile(out string) bool {
	switch strings.ToLower(filepath.Ext(out)) {
	case ".html", ".htm":
		return true
	}
	return false
}

// split turns -out into the directory to write into and, when it names an
// exact file, the HTML filename to use.
func (o options) split() (dir, htmlName string) {
	if namesHTMLFile(o.out) {
		return filepath.Dir(o.out), filepath.Base(o.out)
	}
	return o.out, ""
}

type options struct {
	out         string
	embedName   string
	inlineMax   sizeLimit
	dataURIs    bool
	rewrite     bool
	manifest    bool
	list        bool
	forcePlutil bool
	verbose     bool
}

// inlineAsset reports whether an asset of n bytes belongs in the HTML as a
// data: URI rather than in a file of its own.
func (o options) inlineAsset(n int) bool {
	return o.rewrite && o.inlineMax.covers(n)
}

// sizeLimit is the -inline-max value: a byte count, or unlimited.
type sizeLimit struct {
	n         int64
	unlimited bool
}

func (s sizeLimit) covers(n int) bool { return s.unlimited || int64(n) < s.n }

func (s sizeLimit) String() string {
	if s.unlimited {
		return "all"
	}
	return strconv.FormatInt(s.n, 10)
}

var sizeSuffix = []struct {
	suffix string
	mult   int64
}{
	{"gb", 1 << 30}, {"g", 1 << 30},
	{"mb", 1 << 20}, {"m", 1 << 20},
	{"kb", 1 << 10}, {"k", 1 << 10},
	{"b", 1},
}

func (s *sizeLimit) Set(v string) error {
	v = strings.ToLower(strings.TrimSpace(v))
	switch v {
	case "all", "inf", "unlimited", "-1":
		*s = sizeLimit{unlimited: true}
		return nil
	case "none", "off":
		*s = sizeLimit{}
		return nil
	}
	mult := int64(1)
	for _, suf := range sizeSuffix {
		if rest, ok := strings.CutSuffix(v, suf.suffix); ok {
			v, mult = strings.TrimSpace(rest), suf.mult
			break
		}
	}
	f, err := strconv.ParseFloat(v, 64)
	if err != nil || f < 0 {
		return fmt.Errorf("want a size like 200k, 1.5M, 0 or all")
	}
	*s = sizeLimit{n: int64(f * float64(mult))}
	return nil
}

func main() {
	o := options{embedName: embedDirName(), inlineMax: sizeLimit{unlimited: true}}
	flag.StringVar(&o.out, "out", ".", "directory to write the HTML file into, or an exact path\n"+
		"ending in .html - then the embed directory is created\n"+
		"next to that file")
	flag.Var(&o.inlineMax, "inline-max", "embed assets smaller than this in the HTML as data: URIs,\n"+
		"write bigger ones to files (\"all\" inlines everything, 0 writes\n"+
		"everything to files; accepts 200k, 1.5M, ...)")
	flag.BoolVar(&o.dataURIs, "data-uris", false, "extract data: URIs the page already had into files")
	flag.BoolVar(&o.rewrite, "rewrite", true, "rewrite asset references to the embed directory")
	flag.BoolVar(&o.manifest, "manifest", true, "write embed/manifest.json mapping files to original URLs")
	flag.BoolVar(&o.list, "list", false, "list archive contents and exit without writing anything")
	flag.BoolVar(&o.forcePlutil, "plutil", false, "always decode via plutil instead of the built-in parser")
	flag.BoolVar(&o.verbose, "v", false, "print every file written")
	flag.Usage = usage
	flag.Parse()

	archives := flag.Args()
	if len(archives) == 0 {
		usage()
		os.Exit(2)
	}

	// With several archives, give each one its own subfolder so that same-named
	// assets from different sites cannot collide.
	perArchive := len(archives) > 1
	if perArchive && namesHTMLFile(o.out) {
		fmt.Fprintf(os.Stderr, "-out %s names a single HTML file, but %d archives were given\n", o.out, len(archives))
		os.Exit(2)
	}

	failed := 0
	for _, path := range archives {
		if err := run(path, o, perArchive); err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", filepath.Base(path), err)
			failed++
		}
	}
	if failed > 0 {
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, `webarchive-extract - unpack Safari .webarchive files

usage: webarchive-extract [flags] archive.webarchive...

Writes "<archive name>.html" into -out, or to -out itself when that ends in
.html. By default every embedded stylesheet, script, image and font is
carried inside it as a data: URI, so the page is one self-contained file;
use -inline-max to send the larger ones to ./%s/ instead. Handles both
binary and XML (plutil-converted) archives. Inline <style>, <script> and
data: URIs the page already had are left as they are.

environment:
  %-14s assets directory, next to the HTML file (default %q)

flags:
`, embedDirName(), embedEnv, defaultEmbed)
	flag.PrintDefaults()
}

func run(archivePath string, o options, perArchive bool) error {
	root, err := loadPlist(archivePath, o.forcePlutil)
	if err != nil {
		return err
	}
	ar := &archive{}
	if err := ar.collect(root, 0); err != nil {
		return err
	}
	if len(ar.main.Data) == 0 {
		return fmt.Errorf("no %s data found - is this a .webarchive?", keyMain)
	}

	base := strings.TrimSuffix(filepath.Base(archivePath), filepath.Ext(archivePath))
	if o.list {
		ar.list(base)
		return nil
	}

	// Generated names (inline blocks, data: URIs) are prefixed with the page
	// slug only when archives share one embed folder; a per-archive subfolder
	// already carries that name.
	embedRel, prefix := o.embedName, slug(base)+"-"
	if perArchive {
		embedRel, prefix = filepath.Join(o.embedName, slug(base)), ""
	}
	outDir, htmlName := o.split()
	if htmlName == "" {
		htmlName = base + ".html"
	}
	e := &extractor{
		opts:     o,
		outDir:   outDir,
		embedRel: filepath.ToSlash(embedRel),
		embedDir: filepath.Join(outDir, embedRel),
		prefix:   prefix,
		used:     map[string]bool{},
	}
	return e.extract(ar, htmlName)
}

// ------------------------------------------------------------------ archives

type resource struct {
	URL       string
	MIME      string
	Encoding  string
	FrameName string
	Data      []byte
}

type archive struct {
	main  resource
	subs  []resource
	byURL map[string][]int // URL -> indices into subs, for duplicate detection
}

// collect walks the plist, flattening subframe archives into the subresource
// list so that iframe sources get extracted and rewritten like anything else.
func (a *archive) collect(v any, depth int) error {
	if depth > 32 {
		return fmt.Errorf("subframe archives nested more than 32 levels deep")
	}
	m, ok := v.(map[string]any)
	if !ok {
		return fmt.Errorf("archive root is not a dictionary")
	}
	if r, ok := toResource(m[keyMain]); ok {
		if depth == 0 {
			a.main = r
		} else {
			a.addSub(r)
		}
	}
	if arr, ok := m[keySubs].([]any); ok {
		for _, item := range arr {
			if r, ok := toResource(item); ok && !a.duplicate(r) {
				a.addSub(r)
			}
		}
	}
	if arr, ok := m[keySubframe].([]any); ok {
		for _, item := range arr {
			if err := a.collect(item, depth+1); err != nil {
				return err
			}
		}
	}
	return nil
}

// addSub appends a subresource, indexing it by URL for duplicate detection.
func (a *archive) addSub(r resource) {
	if r.URL != "" {
		if a.byURL == nil {
			a.byURL = map[string][]int{}
		}
		a.byURL[r.URL] = append(a.byURL[r.URL], len(a.subs))
	}
	a.subs = append(a.subs, r)
}

// duplicate reports whether an identical resource was already collected.
// Subframe archives routinely repeat the assets of their parent page. Only
// same-URL resources are compared, so a big archive does not turn this into a
// quadratic sweep over every payload.
func (a *archive) duplicate(r resource) bool {
	if r.URL == "" {
		return false
	}
	for _, i := range a.byURL[r.URL] {
		if bytes.Equal(a.subs[i].Data, r.Data) {
			return true
		}
	}
	return false
}

func toResource(v any) (resource, bool) {
	m, ok := v.(map[string]any)
	if !ok {
		return resource{}, false
	}
	r := resource{
		URL:       plistString(m["WebResourceURL"]),
		MIME:      plistString(m["WebResourceMIMEType"]),
		Encoding:  plistString(m["WebResourceTextEncodingName"]),
		FrameName: plistString(m["WebResourceFrameName"]),
	}
	if b, ok := m["WebResourceData"].([]byte); ok {
		r.Data = b
	}
	return r, r.URL != "" || len(r.Data) > 0
}

func plistString(v any) string {
	s, _ := v.(string)
	return s
}

func (a *archive) list(base string) {
	fmt.Printf("%s\n", base)
	fmt.Printf("  main  %-28s %8s  %s\n", a.main.MIME, humanSize(len(a.main.Data)), a.main.URL)
	for _, r := range a.subs {
		fmt.Printf("  sub   %-28s %8s  %s\n", r.MIME, humanSize(len(r.Data)), r.URL)
	}
	fmt.Printf("  %d subresource(s), %s total\n", len(a.subs), humanSize(a.totalSize()))
}

func (a *archive) totalSize() int {
	n := len(a.main.Data)
	for _, r := range a.subs {
		n += len(r.Data)
	}
	return n
}

// ----------------------------------------------------------------- extractor

type extractor struct {
	opts     options
	outDir   string
	embedRel string // embed path relative to the HTML file, slash-separated
	embedDir string // embed path on disk
	prefix   string // filename prefix for generated (inline / data:) files

	rw     *rewriter
	assets []*asset

	used         map[string]bool // lowercased asset filenames already taken
	embedCreated bool
	written      []entry
	inlined      int
	dataFiles    int
}

type entry struct {
	File  string `json:"file"`
	URL   string `json:"url,omitempty"`
	MIME  string `json:"mime,omitempty"`
	Bytes int    `json:"bytes"`
	Kind  string `json:"kind"`
}

// resolveState tracks the recursion in extractor.resolve: a stylesheet has to
// be finalised before it can be encoded as a data: URI, and it may in turn
// reference other assets.
type resolveState int

const (
	pending resolveState = iota
	resolving
	resolved
)

type asset struct {
	res    resource
	name   string // filename it would take in the embed directory
	inline bool   // travels inside the HTML as a data: URI
	body   []byte // final content, after its own references are rewritten
	uri    string // body as a data: URI, encoded on first use
	state  resolveState
}

// resolve finalises one asset's content, recursing into whatever it references.
func (e *extractor) resolve(i int) {
	a := e.assets[i]
	switch a.state {
	case resolved:
		return
	case resolving:
		// Reference cycle (a.css @imports b.css and back). Neither can be a
		// data: URI, because each would have to contain the other.
		a.inline = false
		return
	}
	a.state = resolving
	defer func() { a.state = resolved }()

	a.body = a.res.Data
	if !isTextual(a.res.MIME, a.name) {
		return
	}
	text := decodeText(a.body, a.res.Encoding, a.res.URL)
	if e.opts.rewrite {
		// A stylesheet inlined as a data: URI has no base URL, so relative
		// references inside it cannot resolve. If it points at anything that
		// stays on disk, keep the stylesheet on disk too.
		if a.inline {
			for _, dep := range e.rw.matches(text) {
				e.resolve(dep)
				if !e.assets[dep].inline {
					a.inline = false
					break
				}
			}
		}
		text = e.rw.apply(text, "", e.ref) // siblings share one directory
	}
	if e.opts.dataURIs && !a.inline {
		text = e.extractDataURIs(text, "")
	}
	a.body = []byte(text)
}

// ref renders a reference to asset i as seen from a file in prefix.
func (e *extractor) ref(i int, prefix string) string {
	e.resolve(i)
	a := e.assets[i]
	if a.inline {
		if a.uri == "" {
			// Encoded once: a page routinely names the same asset many times.
			a.uri = dataURI(a.res.MIME, a.name, a.body)
		}
		return a.uri
	}
	if prefix != "" {
		return prefix + "/" + a.name
	}
	return a.name
}

func (e *extractor) extract(ar *archive, htmlName string) error {
	if err := os.MkdirAll(e.outDir, 0o755); err != nil {
		return err
	}

	// Decide the fate of every subresource up front, so that a complete rewrite
	// table exists before any content is touched.
	e.rw = newRewriter(ar.main.URL)
	e.assets = make([]*asset, len(ar.subs))
	for i, r := range ar.subs {
		e.assets[i] = &asset{res: r, name: e.assetName(r), inline: e.opts.inlineAsset(len(r.Data))}
		if e.opts.rewrite {
			e.rw.add(r.URL, i)
		}
	}
	for i := range e.assets {
		e.resolve(i)
	}

	// Only assets that stay external are written out; the inlined ones travel
	// inside the HTML as data: URIs.
	for _, a := range e.assets {
		if a.inline {
			e.inlined++
			continue
		}
		if err := e.writeAsset(a.name, a.body, entry{URL: a.res.URL, MIME: a.res.MIME, Kind: "subresource"}); err != nil {
			return err
		}
	}

	// Main document.
	html := decodeText(ar.main.Data, ar.main.Encoding, ar.main.URL)
	if e.opts.rewrite {
		html = e.rw.apply(html, e.embedRel, e.ref)
	}
	if e.opts.dataURIs {
		html = e.extractDataURIs(html, e.embedRel)
	}
	htmlPath := filepath.Join(e.outDir, htmlName)
	if err := os.WriteFile(htmlPath, []byte(html), 0o644); err != nil {
		return err
	}

	if e.opts.manifest && e.embedCreated {
		if err := e.writeManifest(); err != nil {
			return err
		}
	}

	fmt.Printf("%s  (%s)\n", htmlPath, humanSize(len(html)))
	if e.opts.verbose {
		for _, a := range e.assets {
			if a.inline {
				fmt.Printf("  inlined  %s  (%s)\n", a.name, humanSize(len(a.body)))
			}
		}
		for _, w := range e.written {
			fmt.Printf("  %s/%s  (%s, %s)\n", e.embedRel, w.File, humanSize(w.Bytes), w.Kind)
		}
	}
	fmt.Printf("  %d asset(s): %d inlined as data: URIs", len(ar.subs), e.inlined)
	if n := len(ar.subs) - e.inlined; n > 0 {
		fmt.Printf(", %d written to %s/", n, e.embedRel)
	}
	if e.dataFiles > 0 {
		fmt.Printf(", %d existing data: URI(s) extracted", e.dataFiles)
	}
	fmt.Println()
	return nil
}

func (e *extractor) writeAsset(name string, data []byte, meta entry) error {
	if !e.embedCreated {
		if err := os.MkdirAll(e.embedDir, 0o755); err != nil {
			return err
		}
		e.embedCreated = true
	}
	if err := os.WriteFile(filepath.Join(e.embedDir, name), data, 0o644); err != nil {
		return err
	}
	meta.File = name
	meta.Bytes = len(data)
	e.written = append(e.written, meta)
	return nil
}

func (e *extractor) writeManifest() error {
	sorted := append([]entry(nil), e.written...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].File < sorted[j].File })
	b, err := json.MarshalIndent(sorted, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(e.embedDir, "manifest.json"), append(b, '\n'), 0o644)
}

// assetName derives a unique, filesystem-safe name for a subresource.
func (e *extractor) assetName(r resource) string {
	return e.unique(fileNameForURL(r.URL, r.MIME))
}

func (e *extractor) unique(name string) string {
	if name == "manifest.json" {
		name = "manifest-asset.json"
	}
	key := strings.ToLower(name)
	if !e.used[key] {
		e.used[key] = true
		return name
	}
	stem, ext := splitExt(name)
	for i := 2; ; i++ {
		cand := fmt.Sprintf("%s-%d%s", stem, i, ext)
		key := strings.ToLower(cand)
		if !e.used[key] {
			e.used[key] = true
			return cand
		}
	}
}

func warnf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "warning: "+format+"\n", args...)
}

func humanSize(n int) string {
	switch {
	case n < 1024:
		return fmt.Sprintf("%d B", n)
	case n < 1024*1024:
		return fmt.Sprintf("%.1f KB", float64(n)/1024)
	default:
		return fmt.Sprintf("%.1f MB", float64(n)/(1024*1024))
	}
}
