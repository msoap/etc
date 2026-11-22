package main

import (
	"bytes"
	"html/template"
	"strings"
	"testing"
	"time"
)

func TestNiceCeil(t *testing.T) {
	tests := map[float64]float64{0: 1, 0.7: 1, 1: 1, 1.2: 2, 2.2: 2.5, 3: 5, 7: 10, 0.034: 0.05, 120: 200}
	for in, want := range tests {
		if got := niceCeil(in); got != want {
			t.Errorf("niceCeil(%v) = %v, want %v", in, got, want)
		}
	}
}

func TestFormatInt(t *testing.T) {
	tests := map[int64]string{0: "0", 999: "999", 1000: "1,000", 1234567: "1,234,567", -1234: "-1,234"}
	for in, want := range tests {
		if got := formatInt(in); got != want {
			t.Errorf("formatInt(%d) = %q, want %q", in, got, want)
		}
	}
}

func TestPeriodLength(t *testing.T) {
	tests := []struct {
		group, period string
		want          int
	}{{"week", "2026-11-16", 7}, {"month", "2026-11", 30}, {"month", "2026-12", 31}, {"month", "2028-02", 29}}
	for _, tt := range tests {
		if got := periodLength(tt.group, tt.period); got != tt.want {
			t.Errorf("periodLength(%q, %q) = %d, want %d", tt.group, tt.period, got, tt.want)
		}
	}
}

func TestBuildHTMLView(t *testing.T) {
	gemini := newReport("gemini", time.Date(2026, 11, 20, 0, 0, 0, 0, time.UTC), 2)
	gemini.HasTokens, gemini.HasRequests, gemini.NoModelRequests = true, true, true
	gemini.add("2026-11-21", "gemini-2.5-flash", Stats{Cost: 0.5, InputTokens: 300, OutputTokens: 30})
	gemini.add("2026-11-21", "", Stats{Requests: 4})

	openai := testReport()
	openai.Provider = "openai"
	view := buildHTMLView([]*Report{openai, gemini}, time.Now())

	if view.Days != 2 || view.StartDate != "2026-11-20" || view.EndDate != "2026-11-21" {
		t.Errorf("period = %d %s %s", view.Days, view.StartDate, view.EndDate)
	}
	if view.Total.Cost != 4.25 || view.Total.Requests != 19 || view.Total.InputTokens != 1420 {
		t.Errorf("total = %+v", view.Total.Stats)
	}
	// gemini has requests per provider and day, but not per model
	if p := view.Providers[1]; !p.HasRequests || p.Requests != 4 {
		t.Errorf("gemini provider = %+v", p)
	}
	if day := view.DayRows[1].Providers[1]; !day.HasRequests || day.Requests != 4 || day.Models[0].HasRequests {
		t.Errorf("gemini on 2026-11-21 = %+v", day)
	}
	if view.PeakDay.Name != "2026-11-21" || view.PeakDay.Cost != 2.5 {
		t.Errorf("peak day = %s %v", view.PeakDay.Name, view.PeakDay.Cost)
	}
	if len(view.Models) != 3 || view.Models[0].Name != "model-b" || view.Models[0].Color != "s1" || view.Models[2].Name != "gemini-2.5-flash" ||
		!view.Models[0].HasRequests || view.Models[2].HasRequests {
		t.Errorf("models = %+v", view.Models)
	}
	// color of model is the same in all days
	if color := view.DayRows[0].Providers[0].Models[1].Color; color != "s1" {
		t.Errorf("model-b color on first day = %q, want s1", color)
	}
	if len(view.Chart.Columns) != 2 || len(view.Chart.Columns[1].ByProvider) != 2 || len(view.Chart.Columns[1].ByModel) != 2 {
		t.Errorf("chart columns = %+v", view.Chart.Columns)
	}
	// legend filter: the same key in legend, chart segments and tooltip
	if key := view.Chart.ModelLegend[0].Key; key != "m:OpenAI/model-b" ||
		view.Chart.Columns[1].ByModel[0].Key != key || view.Chart.Tips[1].Models[0].Key != key {
		t.Errorf("model keys: legend %q, segment %q, tip %q",
			key, view.Chart.Columns[1].ByModel[0].Key, view.Chart.Tips[1].Models[0].Key)
	}
	if key := view.Chart.ProviderLegend[1].Key; key != "Gemini" || view.Chart.Columns[1].ByProvider[1].Key != key {
		t.Errorf("provider keys: legend %q, segment %q", key, view.Chart.Columns[1].ByProvider[1].Key)
	}
	// "Days" table by day, week and month: both days are in one partial week and month
	if len(view.Periods) != 3 || view.Periods[0].Group != "day" || len(view.Periods[0].Rows) != 2 || view.Periods[0].Rows[0].Note != "" {
		t.Fatalf("periods = %+v", view.Periods)
	}
	for i, want := range []struct{ group, name, note string }{{"week", "2026-11-16", "2 of 7 days"}, {"month", "2026-11", "2 of 30 days"}} {
		p := view.Periods[i+1]
		if p.Group != want.group || len(p.Rows) != 1 {
			t.Fatalf("period %s = %+v", want.group, p)
		}
		row := p.Rows[0]
		if row.Name != want.name || row.Note != want.note || row.Cost != 4.25 || row.Share != 100 || row.Bar != 100 ||
			row.Providers[1].Requests != 4 || len(row.Providers[0].Models) != 2 || row.Providers[0].Models[0].Color != "s1" {
			t.Errorf("period %s row = %+v", want.group, row)
		}
	}
	if len(view.Chart.Tips) != 2 {
		t.Fatalf("chart tips = %+v", view.Chart.Tips)
	}
	if tip := view.Chart.Tips[1]; tip.Weekday != "Sat" || tip.Cost != 2.5 || len(tip.Providers) != 2 || len(tip.Models) != 2 ||
		tip.Models[0].Name != "model-b (OpenAI)" || tip.Models[1].Name != "gemini-2.5-flash (Gemini)" {
		t.Errorf("chart tip for 2026-11-21 = %+v", tip)
	}
}

func TestWriteHTMLReport(t *testing.T) {
	gemini := newReport("gemini", time.Date(2026, 11, 20, 0, 0, 0, 0, time.UTC), 2)
	gemini.add("2026-11-21", "gemini-<2.5>-flash", Stats{Cost: 0.5, InputTokens: 300})

	openai := testReport()
	openai.Provider = "openai"

	tmpl := template.Must(template.New("report").Funcs(htmlTemplateFuncs).Parse(htmlReportTemplate))
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, buildHTMLView([]*Report{openai, gemini}, time.Now())); err != nil {
		t.Fatal(err)
	}

	html := buf.String()
	for _, want := range []string{"model-a", "OpenAI", "Gemini", "gemini-&lt;2.5&gt;-flash", `id="by-model"`, "4.2500", "1,120",
		`id="group-week"`, `class="scroll g-month"`, "2026-11-16", "2 of 30 days"} {
		if !strings.Contains(html, want) {
			t.Errorf("HTML report doesn't contain %q", want)
		}
	}
	if strings.Contains(html, "ZgotmplZ") {
		t.Error("HTML report contains unsafe template value (ZgotmplZ)")
	}
}
