package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"golang.org/x/oauth2/google"
)

type CostAmount struct {
	Value    probablyFloat `json:"value"`
	Currency string        `json:"currency"`
}

type CostResult struct {
	Amount CostAmount `json:"amount"`
}

type Bucket struct {
	StartTime int64        `json:"start_time"`
	Results   []CostResult `json:"results"`
}

type CostsResponse struct {
	Data []Bucket `json:"data"`
}

type BigQueryResponse struct {
	JobComplete bool `json:"jobComplete"`
	Rows        []struct {
		F []struct {
			V string `json:"v"`
		} `json:"f"`
	} `json:"rows"`
}

type DayCost struct {
	Date string
	Cost float64
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
	" SUM(cost) + SUM(IFNULL((SELECT SUM(c.amount) FROM UNNEST(credits) c), 0)) AS net_cost" +
	" FROM `%s.%s.gcp_billing_export_v1_*`" +
	" WHERE service.description IN ('Gemini API', 'Generative Language API')" +
	" AND DATE(usage_start_time) >= @start_date" +
	" GROUP BY usage_date ORDER BY usage_date"

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

func fetchOpenAI(key string, startDate time.Time, days int) []DayCost {
	url := fmt.Sprintf("https://api.openai.com/v1/organization/costs?start_time=%d&limit=%d", startDate.Unix(), days+1)

	req, err := http.NewRequest("GET", url, nil)
	handleError(err, "Error creating request")

	body := doRequest(req, key)

	var costsResp CostsResponse
	err = json.Unmarshal(body, &costsResp)
	handleError(err, "Error parsing response: "+string(body))

	result := make([]DayCost, 0, len(costsResp.Data))
	for _, bucket := range costsResp.Data {
		date := time.Unix(bucket.StartTime, 0).UTC().Format(time.DateOnly)
		cost := 0.0
		for _, result := range bucket.Results {
			cost += float64(result.Amount.Value)
		}
		result = append(result, DayCost{Date: date, Cost: cost})
	}

	return result
}

// fetchGemini reads net cost (after credits) from Cloud Billing export in BigQuery, auth via Application Default Credentials
func fetchGemini(project, dataset string, startDate time.Time, days int) []DayCost {
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
		"timeoutMs": 50000,
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

	costByDate := map[string]float64{}
	for _, row := range bqResp.Rows {
		if len(row.F) != 2 {
			log.Fatalf("Error: unexpected BigQuery row: %v", row)
		}
		cost, err := strconv.ParseFloat(row.F[1].V, 64)
		handleError(err, "Error parsing cost")
		costByDate[row.F[0].V] = cost
	}

	// billing export has no rows for days without usage, fill them with zeros
	result := make([]DayCost, 0, days)
	for i := range days {
		date := startDate.AddDate(0, 0, i).Format(time.DateOnly)
		result = append(result, DayCost{Date: date, Cost: costByDate[date]})
	}

	return result
}

func main() {
	days := flag.Int("days", 7, "Number of days to retrieve")
	provider := flag.String("provider", "openai", "Provider: openai or gemini")
	apiKey := flag.String("key", "", "OpenAI API key (or set OPENAI_ADMIN_KEY env var)")
	gcpProject := flag.String("gcp-project", os.Getenv("GOOGLE_CLOUD_PROJECT"), "GCP project with BigQuery billing export (default from GOOGLE_CLOUD_PROJECT env var or credentials)")
	bqDataset := flag.String("bq-dataset", "billing_export", "BigQuery dataset with billing export")
	total := flag.Bool("total", false, "Show total cost")
	flag.Parse()

	startDate := time.Now().UTC().AddDate(0, 0, -*days+1).Truncate(24 * time.Hour)

	var costs []DayCost
	switch *provider {
	case "openai":
		key := *apiKey
		if key == "" {
			key = os.Getenv("OPENAI_ADMIN_KEY")
		}
		if key == "" {
			log.Fatal("Error: OpenAI API key not provided. Use -key flag or set OPENAI_ADMIN_KEY env var")
		}
		costs = fetchOpenAI(key, startDate, *days)
	case "gemini":
		costs = fetchGemini(*gcpProject, *bqDataset, startDate, *days)
	default:
		log.Fatalf("Error: unknown provider %q, use openai or gemini", *provider)
	}

	totalCost := 0.0
	for _, day := range costs {
		totalCost += day.Cost
		fmt.Printf("%s %.4f\n", day.Date, day.Cost)
	}
	if *total {
		fmt.Printf("Total: %.4f\n", totalCost)
	}
}
