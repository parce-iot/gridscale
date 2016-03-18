package gridscale

import (
	"encoding/json"

	"github.com/shopspring/decimal"
)

type restPrice struct {
	Prices []Price `json:"prices"`
}

// Price holds pricing information about a gridscale object type.
type Price struct {
	Type         string          `json:"type"`
	PricePerUnit decimal.Decimal `json:"price_per_unit"`
	Name         string          `json:"name"`
	ProductNo    int             `json:"product_no"`
	Currency     string          `json:"currency"`
	Unit         string          `json:"unit"`
}

// GetPrices returns a list of object prices.
func (c *Client) GetPrices() ([]Price, error) {
	resp, err := c.get("/prices")
	if err != nil {
		return nil, err
	}

	decoder := json.NewDecoder(resp.Body)

	prices := &restPrice{}
	err = decoder.Decode(prices)
	if err != nil {
		return nil, err
	}

	return prices.Prices, nil
}
