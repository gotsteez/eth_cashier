package ethcashier

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

const CMC_API_URL = "https://pro-api.coinmarketcap.com/v1/cryptocurrency/quotes/latest"

type CMCClient struct {
	apiKey string
}

// Response structures for CoinMarketCap API
type CMCResponse struct {
	Data map[string]CryptoCurrency `json:"data"`
}

type CryptoCurrency struct {
	Quote map[string]PriceQuote `json:"quote"`
}

type PriceQuote struct {
	Price float64 `json:"price"`
}

func NewCMCClient(apiKey string) *CMCClient {
	return &CMCClient{
		apiKey: apiKey,
	}
}

// GetEthereumPrice returns the current price of Ethereum in USD
func (c *CMCClient) GetEthereumPrice() (float64, error) {
	req, err := http.NewRequest("GET", CMC_API_URL, nil)
	if err != nil {
		return 0, fmt.Errorf("error creating request: %v", err)
	}

	q := req.URL.Query()
	q.Add("symbol", "ETH")
	q.Add("convert", "USD")
	req.URL.RawQuery = q.Encode()

	req.Header.Add("X-CMC_PRO_API_KEY", c.apiKey)
	req.Header.Add("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("error reading response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var response CMCResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return 0, fmt.Errorf("error parsing response: %v", err)
	}

	ethData, exists := response.Data["ETH"]
	if !exists {
		return 0, fmt.Errorf("ethereum data not found in response")
	}

	usdQuote, exists := ethData.Quote["USD"]
	if !exists {
		return 0, fmt.Errorf("USD quote not found in response")
	}

	return usdQuote.Price, nil
}
