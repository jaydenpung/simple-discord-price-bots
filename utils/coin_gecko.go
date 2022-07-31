package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

const (
	GeckoURL = "https://api.coingecko.com/api/v3/coins/%s"
)

type CurrentPrice struct {
	USD float64 `json:"usd"`
	BTC float64 `json:"btc"`
}

type TotalValueLocked struct {
	USD float64 `json:"usd"`
	BTC float64 `json:"btc"`
}

type MarketData struct {
	CurrentPrice            CurrentPrice     `json:"current_price"`
	MarketCap               CurrentPrice     `json:"market_cap"`
	TotalValueLocked        TotalValueLocked `json:"total_value_locked"`
	PriceChangePercent      float64          `json:"price_change_percentage_24h"`
	PriceChangeCurrency     CurrentPrice     `json:"price_change_24h_in_currency"`
	MarketCapChangePercent  float64          `json:"market_cap_change_percentage_24h"`
	MarketCapChangeCurrency CurrentPrice     `json:"market_cap_change_24h_in_currency"`
	TotalSupply             float64          `json:"total_supply"`
	CirculatingSupply       float64          `json:"circulating_supply"`
}

// The following is the API response gecko gives
type GeckoPriceResults struct {
	ID         string     `json:"id"`
	Symbol     string     `json:"symbol"`
	Name       string     `json:"name"`
	MarketData MarketData `json:"market_data"`
}

// GetCryptoPrice retrieves the price of a given ticker using the coin gecko API
func GetCryptoPrice(ticker string) (GeckoPriceResults, error) {
	var price GeckoPriceResults

	reqURL := fmt.Sprintf(GeckoURL, ticker)
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

	if resp.StatusCode == 429 {
		return price, errors.New("being rate limited by coingecko")
	}

	results, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return price, err
	}
	err = json.Unmarshal(results, &price)
	if err != nil {
		return price, err
	}

	return price, nil
}
