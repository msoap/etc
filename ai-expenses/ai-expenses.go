package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
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

func handleError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %v", msg, err)
	}
}

func main() {
	days := flag.Int("days", 7, "Number of days to retrieve")
	apiKey := flag.String("key", "", "OpenAI API key (or set OPENAI_ADMIN_KEY env var)")
	total := flag.Bool("total", false, "Show total cost")
	flag.Parse()

	key := *apiKey
	if key == "" {
		key = os.Getenv("OPENAI_ADMIN_KEY")
	}
	if key == "" {
		log.Fatal("Error: OpenAI API key not provided. Use -key flag or set OPENAI_ADMIN_KEY env var")
	}

	now := time.Now().UTC()
	startTime := now.AddDate(0, 0, -*days+1).Truncate(24 * time.Hour).Unix()

	url := fmt.Sprintf("https://api.openai.com/v1/organization/costs?start_time=%d&limit=%d", startTime, *days+1)

	req, err := http.NewRequest("GET", url, nil)
	handleError(err, "Error creating request")

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", key))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	handleError(err, "Error making request")
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	handleError(err, "Error reading response")

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var costsResp CostsResponse
	err = json.Unmarshal(body, &costsResp)
	handleError(err, "Error parsing response: "+string(body))

	totalCost := 0.0
	for _, bucket := range costsResp.Data {
		date := time.Unix(bucket.StartTime, 0).UTC().Format(time.DateOnly)
		cost := 0.0
		for _, result := range bucket.Results {
			cost += float64(result.Amount.Value)
		}
		totalCost += cost
		fmt.Printf("%s %.3f\n", date, cost)
	}
	if *total {
		fmt.Printf("Total: %.3f\n", totalCost)
	}
}
