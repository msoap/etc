package main

import (
	"bytes"
	"cmp"
	_ "embed"
	"fmt"
	"html/template"
	"maps"
	"math"
	"os"
	"slices"
	"strconv"
	"time"
)

//go:embed html_report.tmpl
var htmlReportTemplate string

// maxSeriesColors - number of categorical colors, more series are folded into "Other"
const maxSeriesColors = 8

var providerTitles = map[string]string{"openai": "OpenAI", "gemini": "Gemini"}

type htmlRow struct {
	Provider    string // provider title
	Name        string // provider title, model name or date
	Color       string // css class of series color
	HasRequests bool
	Share       float64 // % of total cost of all providers
	Bar         float64 // % of the largest cost in the table
	Stats
}

// CostPerMTokens - average cost of 1M tokens (input + output)
func (r htmlRow) CostPerMTokens() float64 {
	tokens := r.InputTokens + r.OutputTokens
	if tokens == 0 {
		return 0
	}
	return r.Cost / float64(tokens) * 1e6
}

type htmlDayProvider struct {
	htmlRow
	Models []htmlRow
}

type htmlDay struct {
	htmlRow
	Providers []htmlDayProvider
}

type heatCell struct {
	Date  string
	Cost  float64
	Level int // 0 (no cost) .. heatLevels
}

type heatRow struct {
	htmlRow
	Cells []heatCell
}

const heatLevels = 6

type chartSeg struct {
	Y, H, Cost        float64
	Key, Color, Title string // Key - legend item key for filtering by click on legend
}

// otherKey - chart key of models folded into "Other"
const otherKey = "other"

// chartModelKey returns key of model for chart legend, models without own color are folded into "Other"
func chartModelKey(m htmlRow) string {
	if m.Color == "s-other" {
		return otherKey
	}
	return "m:" + m.Provider + "/" + m.Name
}

func chartModelLabel(m htmlRow, multiProvider bool) string {
	if multiProvider {
		return m.Name + " (" + m.Provider + ")"
	}
	return m.Name
}

type chartColumn struct {
	X, W, HitX, HitW    float64
	Title               string
	ByProvider, ByModel []chartSeg
}

// chartTip - data of the day for JS tooltip of the chart
type chartTip struct {
	Date         string    `json:"date"`
	Weekday      string    `json:"weekday"` // Mon, Tue, ...
	Cost         float64   `json:"cost"`
	HasRequests  bool      `json:"hasRequests"`
	Requests     int64     `json:"requests"`
	InputTokens  int64     `json:"inputTokens"`
	OutputTokens int64     `json:"outputTokens"`
	Providers    []tipItem `json:"providers"`
	Models       []tipItem `json:"models"`
}

type tipItem struct {
	Key   string  `json:"key"` // same as in chart legend
	Name  string  `json:"name"`
	Color string  `json:"color"`
	Cost  float64 `json:"cost"`
}

type chartTick struct {
	Pos   float64
	Label string
}

type legendItem struct {
	Key, Name, Color string
}

type chartView struct {
	Width, Height, Left, Right, Top, Bottom float64
	YTicks, XLabels                         []chartTick
	Columns                                 []chartColumn
	Tips                                    []chartTip
	ProviderLegend, ModelLegend             []legendItem
}

type htmlView struct {
	Generated, StartDate, EndDate string
	Days                          int
	MultiProvider                 bool
	Total                         htmlRow
	AvgPerDay                     float64
	PeakDay                       htmlRow
	Providers                     []htmlRow
	Models                        []htmlRow
	DayRows                       []htmlDay
	Dates                         []string
	HeatProviders, HeatModels     []heatRow
	Chart                         chartView
}

func providerTitle(provider string) string {
	return cmp.Or(providerTitles[provider], provider)
}

func seriesColor(index, count int) string {
	if count > maxSeriesColors && index >= maxSeriesColors-1 {
		return "s-other"
	}
	return "s" + strconv.Itoa(index+1)
}

func modelKey(provider, model string) string {
	return provider + "\x00" + model
}

func pct(value, total float64) float64 {
	if total <= 0 {
		return 0
	}
	return value / total * 100
}

// niceCeil rounds up to 1, 2, 2.5, 5 x 10^n for chart axis
func niceCeil(v float64) float64 {
	if v <= 0 {
		return 1
	}
	exp := math.Pow(10, math.Floor(math.Log10(v)))
	for _, nice := range []float64{1, 2, 2.5, 5, 10} {
		if v <= nice*exp*(1+1e-9) {
			return nice * exp
		}
	}
	return 10 * exp
}

func buildHTMLView(reports []*Report, now time.Time) htmlView {
	view := htmlView{
		Generated:     now.UTC().Format("2006-01-02 15:04 UTC"),
		MultiProvider: len(reports) > 1,
		Total:         htmlRow{Name: "Total", Share: 100},
	}

	dateSet := map[string]bool{}
	for _, r := range reports {
		for date := range r.Days {
			dateSet[date] = true
		}
	}
	view.Dates = slices.Sorted(maps.Keys(dateSet))
	view.Days = len(view.Dates)
	if len(view.Dates) > 0 {
		view.StartDate, view.EndDate = view.Dates[0], view.Dates[len(view.Dates)-1]
	}

	// providers and models totals
	totals := make([]*DayStats, len(reports))
	for i, r := range reports {
		totals[i] = r.total()
		view.Total.Stats.add(totals[i].Stats)
		view.Total.HasRequests = view.Total.HasRequests || r.HasRequests
	}

	maxProviderCost, maxModelCost := 0.0, 0.0
	for i, r := range reports {
		view.Providers = append(view.Providers, htmlRow{
			Provider:    providerTitle(r.Provider),
			Name:        providerTitle(r.Provider),
			Color:       seriesColor(i, len(reports)),
			HasRequests: r.HasRequests,
			Share:       pct(totals[i].Cost, view.Total.Cost),
			Stats:       totals[i].Stats,
		})
		maxProviderCost = max(maxProviderCost, totals[i].Cost)

		for model, s := range totals[i].Models {
			view.Models = append(view.Models, htmlRow{
				Provider:    providerTitle(r.Provider),
				Name:        model,
				HasRequests: r.HasRequests,
				Share:       pct(s.Cost, view.Total.Cost),
				Stats:       *s,
			})
			maxModelCost = max(maxModelCost, s.Cost)
		}
	}
	for i := range view.Providers {
		view.Providers[i].Bar = pct(view.Providers[i].Cost, maxProviderCost)
	}

	// color follows the model: assigned once by total cost rank
	slices.SortStableFunc(view.Models, func(a, b htmlRow) int {
		return cmp.Or(cmp.Compare(b.Cost, a.Cost), cmp.Compare(a.Provider, b.Provider), cmp.Compare(a.Name, b.Name))
	})
	modelColors := map[string]string{}
	for i := range view.Models {
		m := &view.Models[i]
		m.Color = seriesColor(i, len(view.Models))
		m.Bar = pct(m.Cost, maxModelCost)
		modelColors[modelKey(m.Provider, m.Name)] = m.Color
	}

	// days
	maxDayCost := 0.0
	for _, date := range view.Dates {
		day := htmlDay{htmlRow: htmlRow{Name: date, HasRequests: view.Total.HasRequests}}
		for i, r := range reports {
			dayStats := r.Days[date]
			if dayStats == nil {
				dayStats = &DayStats{Models: map[string]*Stats{}}
			}
			day.Stats.add(dayStats.Stats)

			provider := htmlDayProvider{htmlRow: htmlRow{
				Provider:    providerTitle(r.Provider),
				Name:        providerTitle(r.Provider),
				Color:       view.Providers[i].Color,
				HasRequests: r.HasRequests,
				Stats:       dayStats.Stats,
			}}
			for _, model := range sortedModels(dayStats.Models) {
				provider.Models = append(provider.Models, htmlRow{
					Provider:    providerTitle(r.Provider),
					Name:        model,
					Color:       modelColors[modelKey(providerTitle(r.Provider), model)],
					HasRequests: r.HasRequests,
					Stats:       *dayStats.Models[model],
				})
			}
			day.Providers = append(day.Providers, provider)
		}
		maxDayCost = max(maxDayCost, day.Cost)
		view.DayRows = append(view.DayRows, day)
	}

	for i := range view.DayRows {
		day := &view.DayRows[i]
		day.Share = pct(day.Cost, view.Total.Cost)
		day.Bar = pct(day.Cost, maxDayCost)
		if day.Cost > view.PeakDay.Cost {
			view.PeakDay = day.htmlRow
		}
		for j := range day.Providers {
			p := &day.Providers[j]
			p.Share, p.Bar = pct(p.Cost, view.Total.Cost), pct(p.Cost, maxDayCost)
			for k := range p.Models {
				p.Models[k].Share, p.Models[k].Bar = pct(p.Models[k].Cost, view.Total.Cost), pct(p.Models[k].Cost, maxDayCost)
			}
		}
	}
	if view.Days > 0 {
		view.AvgPerDay = view.Total.Cost / float64(view.Days)
	}

	view.HeatProviders, view.HeatModels = buildHeatmaps(view)
	view.Chart = buildChart(view, maxDayCost)

	return view
}

// buildHeatmaps returns provider x day and model x day cost matrices
func buildHeatmaps(view htmlView) (providers, models []heatRow) {
	providerCells := map[string]map[string]float64{} // provider -> date -> cost
	modelCells := map[string]map[string]float64{}    // model key -> date -> cost
	maxProvider, maxModel := 0.0, 0.0
	for _, day := range view.DayRows {
		for _, p := range day.Providers {
			if providerCells[p.Name] == nil {
				providerCells[p.Name] = map[string]float64{}
			}
			providerCells[p.Name][day.Name] = p.Cost
			maxProvider = max(maxProvider, p.Cost)
			for _, m := range p.Models {
				key := modelKey(m.Provider, m.Name)
				if modelCells[key] == nil {
					modelCells[key] = map[string]float64{}
				}
				modelCells[key][day.Name] = m.Cost
				maxModel = max(maxModel, m.Cost)
			}
		}
	}

	row := func(base htmlRow, costs map[string]float64, maxCost float64) heatRow {
		result := heatRow{htmlRow: base}
		for _, date := range view.Dates {
			cost := costs[date]
			level := 0
			if cost > 0 {
				level = min(heatLevels, max(1, int(math.Ceil(cost/maxCost*heatLevels))))
			}
			result.Cells = append(result.Cells, heatCell{Date: date, Cost: cost, Level: level})
		}
		return result
	}

	for _, p := range view.Providers {
		providers = append(providers, row(p, providerCells[p.Name], maxProvider))
	}
	for _, m := range view.Models {
		models = append(models, row(m, modelCells[modelKey(m.Provider, m.Name)], maxModel))
	}
	return providers, models
}

// buildChart returns geometry of daily stacked column chart (by provider and by model)
func buildChart(view htmlView, maxDayCost float64) chartView {
	chart := chartView{Width: 960, Height: 280, Left: 56, Right: 12, Top: 16, Bottom: 32}
	plotW := chart.Width - chart.Left - chart.Right
	plotH := chart.Height - chart.Top - chart.Bottom
	baseline := chart.Top + plotH

	yMax := niceCeil(maxDayCost)
	const yTicks = 4
	for i := range yTicks + 1 {
		value := yMax * float64(i) / yTicks
		chart.YTicks = append(chart.YTicks, chartTick{
			Pos:   baseline - value/yMax*plotH,
			Label: strconv.FormatFloat(value, 'f', -1, 64),
		})
	}

	for _, p := range view.Providers {
		chart.ProviderLegend = append(chart.ProviderLegend, legendItem{Key: p.Name, Name: p.Name, Color: p.Color})
	}
	hasOther := false
	for _, m := range view.Models {
		if m.Color == "s-other" {
			hasOther = true
			continue
		}
		chart.ModelLegend = append(chart.ModelLegend, legendItem{
			Key:   chartModelKey(m),
			Name:  chartModelLabel(m, view.MultiProvider),
			Color: m.Color,
		})
	}
	if hasOther {
		chart.ModelLegend = append(chart.ModelLegend, legendItem{Key: otherKey, Name: "Other", Color: "s-other"})
	}

	// stacking order of models is the same for all days: by total cost rank
	modelOrder := map[string]int{}
	for i, m := range view.Models {
		modelOrder[chartModelKey(m)] = i
	}

	n := max(1, len(view.DayRows))
	step := plotW / float64(n)
	barW := min(step*0.7, 48)
	labelEvery := int(math.Ceil(float64(n) / 12))

	stack := func(date string, items []chartSeg) []chartSeg {
		var segs []chartSeg
		y := baseline
		for _, item := range items {
			if item.Cost <= 0 {
				continue
			}
			item.H = item.Cost / yMax * plotH
			y -= item.H
			item.Y = y
			item.Title = fmt.Sprintf("%s · %s: %.4f", date, item.Title, item.Cost)
			segs = append(segs, item)
		}
		return segs
	}

	for i, day := range view.DayRows {
		x := chart.Left + float64(i)*step + (step-barW)/2
		column := chartColumn{
			X: x, W: barW,
			HitX: chart.Left + float64(i)*step, HitW: step,
			Title: fmt.Sprintf("%s: %.4f", day.Name, day.Cost),
		}
		chart.Tips = append(chart.Tips, buildChartTip(day, view.MultiProvider))

		var providers, models []chartSeg
		other := chartSeg{Key: otherKey, Title: "Other", Color: "s-other"}
		var dayModels []htmlRow
		for _, p := range day.Providers {
			providers = append(providers, chartSeg{Key: p.Name, Title: p.Name, Color: p.Color, Cost: p.Cost})
			for _, m := range p.Models {
				if m.Color == "s-other" {
					other.Cost += m.Cost
				} else {
					dayModels = append(dayModels, m)
				}
			}
		}
		slices.SortStableFunc(dayModels, func(a, b htmlRow) int {
			return cmp.Compare(modelOrder[chartModelKey(a)], modelOrder[chartModelKey(b)])
		})
		for _, m := range dayModels {
			models = append(models, chartSeg{Key: chartModelKey(m), Title: chartModelLabel(m, view.MultiProvider), Color: m.Color, Cost: m.Cost})
		}
		models = append(models, other)

		column.ByProvider = stack(day.Name, providers)
		column.ByModel = stack(day.Name, models)
		chart.Columns = append(chart.Columns, column)

		if i%labelEvery == 0 {
			chart.XLabels = append(chart.XLabels, chartTick{Pos: x + barW/2, Label: day.Name[2:]}) // YY-MM-DD
		}
	}

	return chart
}

// buildChartTip returns all providers and models of the day (without folding into "Other"), sorted by cost
func buildChartTip(day htmlDay, multiProvider bool) chartTip {
	tip := chartTip{
		Date:         day.Name,
		Cost:         day.Cost,
		HasRequests:  day.HasRequests,
		Requests:     day.Requests,
		InputTokens:  day.InputTokens,
		OutputTokens: day.OutputTokens,
		Providers:    []tipItem{},
		Models:       []tipItem{},
	}
	if date, err := time.Parse(time.DateOnly, day.Name); err == nil {
		tip.Weekday = date.Weekday().String()[:3]
	}
	for _, p := range day.Providers {
		if p.Cost > 0 {
			tip.Providers = append(tip.Providers, tipItem{Key: p.Name, Name: p.Name, Color: p.Color, Cost: p.Cost})
		}
		for _, m := range p.Models {
			if m.Cost > 0 {
				tip.Models = append(tip.Models, tipItem{Key: chartModelKey(m), Name: chartModelLabel(m, multiProvider), Color: m.Color, Cost: m.Cost})
			}
		}
	}
	byCost := func(a, b tipItem) int { return cmp.Compare(b.Cost, a.Cost) }
	slices.SortStableFunc(tip.Providers, byCost)
	slices.SortStableFunc(tip.Models, byCost)
	return tip
}

// formatInt formats integer with thousands separators: 1234567 -> 1,234,567
func formatInt(n int64) string {
	s := strconv.FormatInt(n, 10)
	sign := ""
	if n < 0 {
		sign, s = "-", s[1:]
	}
	var out []byte
	for i := range len(s) {
		if i > 0 && (len(s)-i)%3 == 0 {
			out = append(out, ',')
		}
		out = append(out, s[i])
	}
	return sign + string(out)
}

var htmlTemplateFuncs = template.FuncMap{
	"cost":  func(v float64) string { return fmt.Sprintf("%.4f", v) },
	"int":   formatInt,
	"pct":   func(v float64) string { return fmt.Sprintf("%.1f%%", v) },
	"num":   func(v float64) string { return strconv.FormatFloat(v, 'f', 2, 64) },
	"raw":   func(v float64) string { return strconv.FormatFloat(v, 'f', -1, 64) },
	"width": func(v float64) template.CSS { return template.CSS(fmt.Sprintf("width:%.2f%%", v)) },
	"add":   func(a, b float64) float64 { return a + b },
	"sub":   func(a, b float64) float64 { return a - b },
}

func writeHTMLReport(fileName string, reports []*Report) {
	tmpl := template.Must(template.New("report").Funcs(htmlTemplateFuncs).Parse(htmlReportTemplate))

	var buf bytes.Buffer
	err := tmpl.Execute(&buf, buildHTMLView(reports, time.Now()))
	handleError(err, "Error rendering HTML report")

	err = os.WriteFile(fileName, buf.Bytes(), 0o644)
	handleError(err, "Error writing HTML report")
}
