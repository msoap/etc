package main

import (
	"bytes"
	"cmp"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"maps"
	"math"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/msoap/tcg"
	"golang.org/x/oauth2/google"
)

// chartWidth - maximum width of cost bar in terminal characters
const chartWidth = 10

type CostAmount struct {
	Value    probablyFloat `json:"value"`
	Currency string        `json:"currency"`
}

type CostResult struct {
	Amount   CostAmount `json:"amount"`
	LineItem string     `json:"line_item"` // only with group_by=line_item, e.g. "gpt-4o-2024-08-06, input"
}

type UsageResult struct {
	Model            string `json:"model"` // only with group_by=model
	NumModelRequests int64  `json:"num_model_requests"`
	InputTokens      int64  `json:"input_tokens"`
	OutputTokens     int64  `json:"output_tokens"`
}

type Bucket[R any] struct {
	StartTime int64 `json:"start_time"`
	Results   []R   `json:"results"`
}

type OpenAIPage[T any] struct {
	Data     []T    `json:"data"`
	HasMore  bool   `json:"has_more"`
	NextPage string `json:"next_page"`
}

type BigQueryResponse struct {
	JobComplete bool `json:"jobComplete"`
	Rows        []struct {
		F []struct {
			V string `json:"v"`
		} `json:"f"`
	} `json:"rows"`
}

type Stats struct {
	Cost         float64
	Requests     int64
	InputTokens  int64
	OutputTokens int64
}

func (s *Stats) add(other Stats) {
	s.Cost += other.Cost
	s.Requests += other.Requests
	s.InputTokens += other.InputTokens
	s.OutputTokens += other.OutputTokens
}

type DayStats struct {
	Stats
	Models map[string]*Stats
}

func (d *DayStats) add(model string, s Stats) {
	d.Stats.add(s)
	if model == "" {
		return
	}
	if d.Models[model] == nil {
		d.Models[model] = &Stats{}
	}
	d.Models[model].add(s)
}

type Report struct {
	Provider    string
	Days        map[string]*DayStats
	HasRequests bool
	HasTokens   bool
}

func newReport(provider string, startDate time.Time, days int) *Report {
	r := &Report{Provider: provider, Days: map[string]*DayStats{}}
	// APIs have no rows for days without usage, fill them with zeros
	for i := range days {
		r.day(startDate.AddDate(0, 0, i).Format(time.DateOnly))
	}
	return r
}

func (r *Report) day(date string) *DayStats {
	if r.Days[date] == nil {
		r.Days[date] = &DayStats{Models: map[string]*Stats{}}
	}
	return r.Days[date]
}

// add adds stats for a day, model can be empty when not grouped by model
func (r *Report) add(date, model string, s Stats) {
	r.day(date).add(model, s)
}

func (r *Report) total() *DayStats {
	result := &DayStats{Models: map[string]*Stats{}}
	for _, day := range r.Days {
		result.Stats.add(day.Stats)
		for model, s := range day.Models {
			if result.Models[model] == nil {
				result.Models[model] = &Stats{}
			}
			result.Models[model].add(*s)
		}
	}
	return result
}

type probablyFloat float64 // can be float or float inside string

func (p *probablyFloat) UnmarshalJSON(data []byte) error {
	var floatValue float64
	if err := json.Unmarshal(data, &floatValue); err == nil {
		*p = probablyFloat(floatValue)
		return nil
	}

	var stringValue string
	if err := json.Unmarshal(data, &stringValue); err == nil {
		if floatValue, err := strconv.ParseFloat(stringValue, 64); err == nil {
			*p = probablyFloat(floatValue)
			return nil
		}
	}

	return fmt.Errorf("invalid value for probablyFloat: %s", data)
}

const geminiCostQuery = "SELECT CAST(DATE(usage_start_time) AS STRING) AS usage_date," +
	" sku.description AS sku," +
	" IFNULL(SUM(usage.amount), 0) AS amount," +
	" SUM(cost) + SUM(IFNULL((SELECT SUM(c.amount) FROM UNNEST(credits) c), 0)) AS net_cost" +
	" FROM `%s.%s.gcp_billing_export_v1_*`" +
	" WHERE service.description IN ('Gemini API', 'Generative Language API')" +
	" AND DATE(usage_start_time) >= @start_date" +
	" GROUP BY usage_date, sku ORDER BY usage_date"

// OpenAI usage endpoints for requests and tokens, all support group_by=model
var openAIUsageEndpoints = []string{"completions", "embeddings", "moderations", "images", "audio_speeches", "audio_transcriptions"}

// "Generate content input token count gemini 2.5 flash lite short input text"
var geminiTokenSKURe = regexp.MustCompile(`(?i)^generate content (?:.* )?(input|output) token count (.+)$`)

// words after model name in Gemini SKU description
var geminiSKUQualifiers = map[string]bool{
	"text": true, "image": true, "audio": true, "video": true,
	"short": true, "long": true, "input": true, "output": true,
	"thinking": true, "non-thinking": true, "cached": true, "batch": true,
}

// parseGeminiSKU returns model name ("gemini-2.5-flash-lite") and token direction ("input" or "output")
// for token SKUs, for other SKUs returns description as model and empty direction
func parseGeminiSKU(description string) (model, direction string) {
	m := geminiTokenSKURe.FindStringSubmatch(description)
	if m == nil {
		return description, ""
	}

	words := strings.Fields(strings.ToLower(m[2]))
	for len(words) > 1 && geminiSKUQualifiers[words[len(words)-1]] {
		words = words[:len(words)-1]
	}

	return strings.Join(words, "-"), strings.ToLower(m[1])
}

// lineItemModel returns model from OpenAI cost line item: "gpt-4o-2024-08-06, input" -> "gpt-4o-2024-08-06"
func lineItemModel(lineItem string) string {
	if model, _, ok := strings.Cut(lineItem, ", "); ok {
		return model
	}
	return cmp.Or(lineItem, "unknown")
}

func handleError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %v", msg, err)
	}
}

func doRequest(req *http.Request, token string) []byte {
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	handleError(err, "Error making request")
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	handleError(err, "Error reading response")

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	return body
}

// fetchOpenAIAll fetches all pages of OpenAI organization API
func fetchOpenAIAll[T any](key, baseURL string) []T {
	var result []T
	page := ""
	for {
		pageURL := baseURL
		if page != "" {
			pageURL += "&page=" + url.QueryEscape(page)
		}

		req, err := http.NewRequest("GET", pageURL, nil)
		handleError(err, "Error creating request")

		body := doRequest(req, key)

		var resp OpenAIPage[T]
		err = json.Unmarshal(body, &resp)
		handleError(err, "Error parsing response: "+string(body))

		result = append(result, resp.Data...)
		if !resp.HasMore || resp.NextPage == "" {
			return result
		}
		page = resp.NextPage
	}
}

func bucketDate(startTime int64) string {
	return time.Unix(startTime, 0).UTC().Format(time.DateOnly)
}

func fetchOpenAI(key string, startDate time.Time, days int, byModel, verbose bool) *Report {
	report := newReport("openai", startDate, days)

	costsURL := fmt.Sprintf("https://api.openai.com/v1/organization/costs?start_time=%d&bucket_width=1d&limit=31", startDate.Unix())
	if byModel {
		costsURL += "&group_by=line_item"
	}
	for _, bucket := range fetchOpenAIAll[Bucket[CostResult]](key, costsURL) {
		for _, result := range bucket.Results {
			model := ""
			if byModel {
				model = lineItemModel(result.LineItem)
			}
			report.add(bucketDate(bucket.StartTime), model, Stats{Cost: float64(result.Amount.Value)})
		}
	}

	if !verbose {
		return report
	}

	report.HasRequests, report.HasTokens = true, true
	for _, endpoint := range openAIUsageEndpoints {
		usageURL := fmt.Sprintf("https://api.openai.com/v1/organization/usage/%s?start_time=%d&bucket_width=1d&limit=31&group_by=model", endpoint, startDate.Unix())
		for _, bucket := range fetchOpenAIAll[Bucket[UsageResult]](key, usageURL) {
			for _, result := range bucket.Results {
				model := ""
				if byModel {
					model = cmp.Or(result.Model, "unknown")
				}
				report.add(bucketDate(bucket.StartTime), model, Stats{
					Requests:     result.NumModelRequests,
					InputTokens:  result.InputTokens,
					OutputTokens: result.OutputTokens,
				})
			}
		}
	}

	return report
}

// fetchGemini reads net cost (after credits) from Cloud Billing export in BigQuery, auth via Application Default Credentials
// Billing export has token counts (in usage.amount of token SKUs), but no request counts
func fetchGemini(project, dataset string, startDate time.Time, days int, byModel bool) *Report {
	ctx := context.Background()
	creds, err := google.FindDefaultCredentials(ctx, "https://www.googleapis.com/auth/cloud-platform")
	handleError(err, "Error finding Google credentials")

	if project == "" {
		project = creds.ProjectID
	}
	if project == "" {
		log.Fatal("Error: GCP project not provided. Use -gcp-project flag or set GOOGLE_CLOUD_PROJECT env var")
	}

	token, err := creds.TokenSource.Token()
	handleError(err, "Error getting Google access token")

	reqBody, err := json.Marshal(map[string]any{
		"query":         fmt.Sprintf(geminiCostQuery, project, dataset),
		"useLegacySql":  false,
		"parameterMode": "NAMED",
		"queryParameters": []map[string]any{{
			"name":           "start_date",
			"parameterType":  map[string]string{"type": "DATE"},
			"parameterValue": map[string]string{"value": startDate.Format(time.DateOnly)},
		}},
		"timeoutMs":  50000,
		"maxResults": 100000,
	})
	handleError(err, "Error creating query")

	url := fmt.Sprintf("https://bigquery.googleapis.com/bigquery/v2/projects/%s/queries", project)
	req, err := http.NewRequest("POST", url, bytes.NewReader(reqBody))
	handleError(err, "Error creating request")

	body := doRequest(req, token.AccessToken)

	var bqResp BigQueryResponse
	err = json.Unmarshal(body, &bqResp)
	handleError(err, "Error parsing response: "+string(body))
	if !bqResp.JobComplete {
		log.Fatal("Error: BigQuery query did not complete in time")
	}

	report := newReport("gemini", startDate, days)
	report.HasTokens = true
	for _, row := range bqResp.Rows {
		if len(row.F) != 4 {
			log.Fatalf("Error: unexpected BigQuery row: %v", row)
		}
		amount, err := strconv.ParseFloat(row.F[2].V, 64)
		handleError(err, "Error parsing usage amount")
		cost, err := strconv.ParseFloat(row.F[3].V, 64)
		handleError(err, "Error parsing cost")

		model, direction := parseGeminiSKU(row.F[1].V)
		stats := Stats{Cost: cost}
		// usage.unit for token SKUs is "requests", but amount is a token count
		switch direction {
		case "input":
			stats.InputTokens = int64(math.Round(amount))
		case "output":
			stats.OutputTokens = int64(math.Round(amount))
		}
		if !byModel {
			model = ""
		}
		report.add(row.F[0].V, model, stats)
	}

	return report
}

// costBar renders cost as horizontal bar up to chartWidth characters for maxCost,
// in 2x3 pixel mode: 1px line in the middle row of the character, 2px per character
func costBar(cost, maxCost float64) string {
	if cost <= 0 || maxCost <= 0 {
		return ""
	}

	maxPixels := chartWidth * tcg.Mode2x3.Width()
	pixels := min(maxPixels, max(1, int(math.Round(cost/maxCost*float64(maxPixels))))) // show at least 1px for any cost

	buf := tcg.NewBuffer(pixels, tcg.Mode2x3.Height())
	buf.HLine(0, tcg.Mode2x3.Height()/2, pixels, tcg.Black)

	return buf.RenderAsStrings(tcg.Mode2x3)[0]
}

// alignTable formats rows as columns separated by 2 spaces: first column to the left, others to the right
func alignTable(rows [][]string) []string {
	var widths []int
	for _, row := range rows {
		for i, cell := range row {
			if i >= len(widths) {
				widths = append(widths, 0)
			}
			widths[i] = max(widths[i], utf8.RuneCountInString(cell))
		}
	}

	lines := make([]string, len(rows))
	for n, row := range rows {
		line := make([]string, len(row))
		for i, cell := range row {
			if i == 0 {
				line[i] = fmt.Sprintf("%-*s", widths[i], cell)
			} else {
				line[i] = fmt.Sprintf("%*s", widths[i], cell)
			}
		}
		lines[n] = strings.Join(line, "  ")
	}
	return lines
}

// printTable prints rows as aligned columns, bars (if not empty) are added at the end of lines
func printTable(w io.Writer, rows [][]string, bars []string) {
	for n, line := range alignTable(rows) {
		if n < len(bars) && bars[n] != "" {
			line += "  " + bars[n]
		}
		fmt.Fprintln(w, strings.TrimRight(line, " "))
	}
}

// sortedModels returns models sorted by cost desc, then by name
func sortedModels(models map[string]*Stats) []string {
	return slices.SortedFunc(maps.Keys(models), func(a, b string) int {
		return cmp.Or(cmp.Compare(models[b].Cost, models[a].Cost), cmp.Compare(a, b))
	})
}

// dateWithWeekday returns date with two-letter weekday abbreviation: "2026-11-21" → "2026-11-21 Sa"
func dateWithWeekday(date string) string {
	t, err := time.Parse(time.DateOnly, date)
	if err != nil {
		return date
	}
	return date + " " + t.Weekday().String()[:2]
}

// printReport prints report, with chart days and models are scaled to the max day cost,
// models of total section are scaled to the total cost, total line has no bar
func printReport(w io.Writer, r *Report, byModel, verbose, showTotal, chart bool) {
	dates := slices.Sorted(maps.Keys(r.Days))

	maxDayCost := 0.0
	for _, day := range r.Days {
		maxDayCost = max(maxDayCost, day.Cost)
	}
	bar := func(cost, maxCost float64) string {
		if !chart {
			return ""
		}
		return costBar(cost, maxCost)
	}

	if !byModel && !verbose {
		totalCost := 0.0
		for _, day := range r.Days {
			totalCost += day.Cost
		}

		costWidth := 0
		if chart { // align bars
			costWidth = len(fmt.Sprintf("%.4f", maxDayCost))
		}

		for _, date := range dates {
			fmt.Fprintln(w, strings.TrimRight(fmt.Sprintf("%s %*.4f %s", dateWithWeekday(date), costWidth, r.Days[date].Cost, bar(r.Days[date].Cost, maxDayCost)), " "))
		}
		if showTotal {
			fmt.Fprintf(w, "Total: %.4f\n", totalCost)
		}
		return
	}

	row := func(name string, s *Stats) []string {
		cells := []string{name, fmt.Sprintf("%.4f", s.Cost)}
		if verbose && r.HasRequests {
			cells = append(cells, strconv.FormatInt(s.Requests, 10))
		}
		if verbose && r.HasTokens {
			cells = append(cells, strconv.FormatInt(s.InputTokens, 10), strconv.FormatInt(s.OutputTokens, 10))
		}
		return cells
	}

	var rows [][]string
	var bars []string
	if verbose {
		header := []string{"date", "cost"}
		if r.HasRequests {
			header = append(header, "requests")
		}
		if r.HasTokens {
			header = append(header, "input_tokens", "output_tokens")
		}
		rows = append(rows, header)
		bars = append(bars, "")
	}

	addDay := func(name string, day *DayStats, maxCost float64, withBar bool) {
		rows = append(rows, row(name, &day.Stats))
		if withBar {
			bars = append(bars, bar(day.Cost, maxCost))
		} else {
			bars = append(bars, "")
		}
		if byModel {
			for _, model := range sortedModels(day.Models) {
				rows = append(rows, row("  "+model, day.Models[model]))
				bars = append(bars, bar(day.Models[model].Cost, maxCost))
			}
		}
	}

	for _, date := range dates {
		addDay(dateWithWeekday(date), r.Days[date], maxDayCost, true)
	}
	if showTotal {
		// total line itself is without bar (it would be always full), its models are scaled to the total cost
		total := r.total()
		addDay("Total:", total, total.Cost, false)
	}

	printTable(w, rows, bars)
}

// providerColumns formats provider columns for printColumns: header line, line per date, total line,
// first column is cost (with bar) under provider name, with verbose requests and tokens columns are added
func providerColumns(r *Report, dates []string, maxDayCost float64, verbose, showTotal, chart bool) []string {
	stats := make([]Stats, 0, len(dates)+1)
	var total Stats
	for _, date := range dates {
		var s Stats
		if day := r.Days[date]; day != nil {
			s = day.Stats
		}
		total.add(s)
		stats = append(stats, s)
	}
	if showTotal {
		stats = append(stats, total)
	}

	costWidth := 0
	for _, s := range stats {
		costWidth = max(costWidth, len(fmt.Sprintf("%.4f", s.Cost)))
	}

	// short headers to fit several providers into terminal width
	header := []string{r.Provider}
	if verbose && r.HasRequests {
		header = append(header, "req")
	}
	if verbose && r.HasTokens {
		header = append(header, "in", "out")
	}

	rows := [][]string{header}
	for i, s := range stats {
		bar := ""
		if chart && i < len(dates) { // total line has no bar
			bar = costBar(s.Cost, maxDayCost)
		}
		row := []string{strings.TrimRight(fmt.Sprintf("%*.4f %s", costWidth, s.Cost, bar), " ")}
		if verbose && r.HasRequests {
			row = append(row, strconv.FormatInt(s.Requests, 10))
		}
		if verbose && r.HasTokens {
			row = append(row, strconv.FormatInt(s.InputTokens, 10), strconv.FormatInt(s.OutputTokens, 10))
		}
		rows = append(rows, row)
	}

	return alignTable(rows)
}

// printColumns prints reports side by side, columns of each provider are separated by "┊",
// with chart all days are scaled to the max day cost of all providers, total line has no bar
func printColumns(w io.Writer, reports []*Report, verbose, showTotal, chart bool) {
	dateSet := map[string]bool{}
	maxDayCost := 0.0
	for _, r := range reports {
		for date, day := range r.Days {
			dateSet[date] = true
			maxDayCost = max(maxDayCost, day.Cost)
		}
	}
	dates := slices.Sorted(maps.Keys(dateSet))

	// each group is lines of the same height: header, dates, total
	dateRows := [][]string{{"date"}}
	for _, date := range dates {
		dateRows = append(dateRows, []string{dateWithWeekday(date)})
	}
	if showTotal {
		dateRows = append(dateRows, []string{"Total:"})
	}
	groups := [][]string{alignTable(dateRows)}
	for _, r := range reports {
		groups = append(groups, providerColumns(r, dates, maxDayCost, verbose, showTotal, chart))
	}

	widths := make([]int, len(groups))
	for g, lines := range groups {
		for _, line := range lines {
			widths[g] = max(widths[g], utf8.RuneCountInString(line))
		}
	}

	for i := range groups[0] {
		if i == 1 {
			separator := make([]string, len(widths))
			for g, width := range widths {
				// " ┊ " between groups: 1 space on the outer side of the first and the last groups, 2 for others
				n := width + 2
				if g == 0 || g == len(widths)-1 {
					n--
				}
				separator[g] = strings.Repeat("┈", n)
			}
			fmt.Fprintln(w, strings.Join(separator, "┼"))
		}

		cells := make([]string, len(groups))
		for g, lines := range groups {
			cells[g] = fmt.Sprintf("%-*s", widths[g], lines[i])
		}
		fmt.Fprintln(w, strings.TrimRight(strings.Join(cells, " ┊ "), " "))
	}
}

func main() {
	days := flag.Int("days", 7, "Number of days to retrieve")
	providers := flag.String("provider", "openai", "Provider: openai, gemini or both comma separated: openai,gemini")
	apiKey := flag.String("key", "", "OpenAI API key (or set OPENAI_ADMIN_KEY env var)")
	gcpProject := flag.String("gcp-project", os.Getenv("GOOGLE_CLOUD_PROJECT"), "GCP project with BigQuery billing export (default from GOOGLE_CLOUD_PROJECT env var or credentials)")
	bqDataset := flag.String("bq-dataset", "billing_export", "BigQuery dataset with billing export")
	total := flag.Bool("total", false, "Show total cost")
	byModel := flag.Bool("by-model", false, "Break out costs by model")
	verbose := flag.Bool("verbose", false, "Show requests and tokens in addition to cost (when provider has them)")
	chart := flag.Bool("chart", false, "Show cost as horizontal bar chart at the end of lines")
	htmlReport := flag.String("html-report", "", "Save HTML report to file, it always contains break out by models, requests and tokens (can't be used with -by-model, -verbose, -total, -chart, -col)")
	columns := flag.Bool("col", false, "Show each provider in its own columns (can't be used with -by-model)")
	flag.Parse()
	log.SetFlags(log.LstdFlags | log.Lmsgprefix) // prefix with provider after the timestamp

	if *htmlReport != "" {
		flag.Visit(func(f *flag.Flag) {
			switch f.Name {
			case "by-model", "verbose", "total", "chart", "col":
				log.Fatalf("Error: -%s can't be used with -html-report, HTML report always contains all statistics", f.Name)
			}
		})
		*byModel, *verbose = true, true
	}
	if *columns && *byModel {
		log.Fatal("Error: -col can't be used with -by-model")
	}
	collect := *htmlReport != "" || *columns

	startDate := time.Now().UTC().AddDate(0, 0, -*days+1).Truncate(24 * time.Hour)

	var reports []*Report
	providerList := strings.Split(*providers, ",")
	for i, provider := range providerList {
		provider = strings.TrimSpace(provider)
		// all errors while fetching are fatal, show which provider failed: "2026/11/21 21:30:47 openai: API error ..."
		log.SetPrefix(provider + ": ")
		var report *Report
		switch provider {
		case "openai":
			key := cmp.Or(*apiKey, os.Getenv("OPENAI_ADMIN_KEY"))
			if key == "" {
				log.Fatal("Error: OpenAI API key not provided. Use -key flag or set OPENAI_ADMIN_KEY env var")
			}
			report = fetchOpenAI(key, startDate, *days, *byModel, *verbose)
		case "gemini":
			report = fetchGemini(*gcpProject, *bqDataset, startDate, *days, *byModel)
		default:
			log.Fatalf("Error: unknown provider %q, use openai, gemini or openai,gemini", provider)
		}
		log.SetPrefix("")

		if collect {
			reports = append(reports, report)
			continue
		}

		if len(providerList) > 1 {
			if i > 0 {
				fmt.Println()
			}
			fmt.Printf("%s:\n", provider)
		}
		printReport(os.Stdout, report, *byModel, *verbose, *total, *chart)
	}

	switch {
	case *htmlReport != "":
		writeHTMLReport(*htmlReport, reports)
		fmt.Printf("HTML report saved to %s\n", *htmlReport)
	case *columns:
		printColumns(os.Stdout, reports, *verbose, *total, *chart)
	}
}
