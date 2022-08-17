package utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
)

const (
	PonyUrl = "https://ponyfinance-api-eu256.ondigitalocean.app/%s"
)

// The following is the API response gecko gives
type PonyApiResults struct {
	TotalValueLocked      string `json:"tvl"`
	AnnualPercentageYield string `json:"apy"`
	NetAssetValue         string `json:"nav"`
}
type PonyPriceResults struct {
	TotalValueLocked      float64 `json:"tvl"`
	AnnualPercentageYield float64 `json:"apy"`
	NetAssetValue         float64 `json:"nav"`
}

func transformResult(result PonyApiResults) PonyPriceResults {
	var price PonyPriceResults
	var err error
	price.TotalValueLocked, err = strconv.ParseFloat(result.TotalValueLocked, 64)
	if err != nil {
		log.Printf("Error converting TotalValueLocked %v", result.TotalValueLocked)
	}
	price.AnnualPercentageYield, err = strconv.ParseFloat(result.AnnualPercentageYield, 64)
	if err != nil {
		log.Printf("Error converting AnnualPercentageYield %v", result.AnnualPercentageYield)
	}
	price.NetAssetValue, err = strconv.ParseFloat(result.NetAssetValue, 64)
	if err != nil {
		log.Printf("Error converting NetAssetValue %v", result.NetAssetValue)
	}

	return price
}

func GetPonyData(ticker string) (PonyPriceResults, error) {
	var result PonyApiResults
	var price PonyPriceResults

	reqURL := fmt.Sprintf(PonyUrl, "info")
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return price, err
	}

	req.Header.Add("User-Agent", "Mozilla/5.0")
	req.Header.Add("accept", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return price, err
	}

	results, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return price, err
	}
	err = json.Unmarshal(results, &result)
	if err != nil {
		return price, err
	}

	return transformResult(result), nil
}
