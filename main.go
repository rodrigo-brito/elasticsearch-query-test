package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gocarina/gocsv"
	"github.com/spf13/cobra"
)

type elasticHit struct {
	Score          float64                `json:"_score"`
	Index          string                 `json:"_index"`
	Type           string                 `json:"_type"`
	ID             string                 `json:"_id"`
	UID            string                 `json:"_uid"`
	Routing        string                 `json:"_routing"`
	Parent         string                 `json:"_parent"`
	Version        interface{}            `json:"_version"`
	Sort           interface{}            `json:"sort"`
	Highlight      interface{}            `json:"highlight"`
	Source         map[string]interface{} `json:"_source"`
	Fields         interface{}            `json:"fields"`
	Explanation    interface{}            `json:"_explanation"`
	MatchedQueries interface{}            `json:"matched_queries"`
	InnerHits      interface{}            `json:"inner_hits"`
}

type expectation struct {
	SearchTerm     string `csv:"search_term"`
	ResultField    string `csv:"result_field"`
	ResultValue    string `csv:"result_value"`
	ResultPosition int    `csv:"result_position"`
	Descritption   string `csv:"description"`
}

func getExpectations(expectationFile string) ([]*expectation, error) {
	clientsFile, err := os.OpenFile(expectationFile, os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		return nil, err
	}
	defer clientsFile.Close()

	var expectations []*expectation

	if err := gocsv.UnmarshalFile(clientsFile, &expectations); err != nil {
		return nil, err
	}

	return expectations, nil
}

func checkResult(endpoint string, expectation *expectation) (bool, string, error) {
	client := new(http.Client)
	req, err := http.NewRequest("GET", endpoint, nil)

	q := req.URL.Query()
	q.Add("q", expectation.SearchTerm)
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		return false, "", err
	}
	defer resp.Body.Close()

	var results []*elasticHit
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&results); err != nil {
		return false, "", err
	}

	if len(results) < expectation.ResultPosition {
		return false, "", fmt.Errorf("searching: '%s', exepect position %d, but got only %d results",
			expectation.SearchTerm, expectation.ResultPosition, len(results))
	}

	result := results[expectation.ResultPosition]
	value, ok := result.Source[expectation.ResultField]
	if !ok {
		return false, "", fmt.Errorf("searching: '%s', exepect field '%s', but not found on type '%s'",
			expectation.SearchTerm, expectation.ResultField, result.Type)
	}

	return fmt.Sprint(value) == expectation.ResultValue, fmt.Sprint(value), nil
}

func run(baseURL, expectationFile string) {
	expectations, err := getExpectations(expectationFile)
	if err != nil {
		log.Fatal(err)
	}

	startTime := time.Now()
	total := 0
	for _, expectation := range expectations {
		ok, result, err := checkResult(baseURL, expectation)
		if err != nil {
			fmt.Printf("parse error: %s\n", err)
			continue
		}

		if ok {
			total++
			fmt.Print("OK: ")
		} else {
			fmt.Print("FAIL: ")
		}

		fmt.Printf("searching '%s', expect '%s' got '%s'\n", expectation.SearchTerm,
			expectation.ResultValue, result)
	}
	fmt.Println("---------------")
	fmt.Printf("Spend time: %.4f sec\n", time.Now().Sub(startTime).Seconds())
	fmt.Printf("Accuracy %d/%d (%d%%)", total, len(expectations), total*100/len(expectations))
}

func main() {
	var (
		 expectationFile string
		 baseURL string
	)

	rootCmd := &cobra.Command{
		Use:   "elasticsearch-query-test",
		Short: "Simple tool to test ElasticSearch query accuracy",
		Run: func(cmd *cobra.Command, args []string) {
			run(baseURL, expectationFile)
		},
	}

	rootCmd.PersistentFlags().StringVarP(&expectationFile, "expectation", "e",
		"expectations.csv", "expectation file eg. dir/custom_expect.csv")
	rootCmd.PersistentFlags().StringVarP(&baseURL, "url", "u",
		"http://localhost", "search base url eg. http://myhost.com/search")

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
