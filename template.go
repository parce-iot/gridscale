package gridscale

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

type Template struct {
	Capacity        int             `json:"capacity"`
	ChangeTime      time.Time       `json:"change_time"`
	CreateTime      time.Time       `json:"create_time"`
	CurrentPrice    decimal.Decimal `json:"current_price"`
	Description     string          `json:"description"`
	LocationCountry string          `json:"location_country"`
	LocationIATA    string          `json:"location_iata"`
	LocationName    string          `json:"location_name"`
	LocationID      string          `json:"location_uuid"`
	Name            string          `json:"name"`
	ID              string          `json:"object_uuid"`
	OSType          string          `json:"ostype"`
	Private         bool            `json:"private"`
	Status          string          `json:"status"`
	Version         string          `json:"version"`
}

// GetTemplates returns a list templates.
func (c *Client) GetTemplates() ([]Template, error) {
	resp, err := c.get("/objects/templates")
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == 401 {
		return nil, fmt.Errorf("Not authorized")
	}

	if resp.StatusCode == 204 {
		// empty response, no IPs to return
		return []Template{}, nil
	}

	decoder := json.NewDecoder(resp.Body)

	rtemps := objectWrapper{}
	err = decoder.Decode(&rtemps)
	if err != nil {
		return nil, err
	}

	temps := []Template{}
	for _, temp := range *rtemps.Templates {
		temps = append(temps, temp)
	}

	return temps, nil
}

// GetTemplateByName returns a template starting with `name`. Returns an error
// if not found or found more than one template.
func (c *Client) GetTemplateByName(name string) (*Template, error) {
	temps, err := c.GetTemplates()
	if err != nil {
		return nil, err
	}

	var ret Template
	n := 0
	for _, t := range temps {
		if strings.HasPrefix(t.Name, name) {
			n = n + 1
			ret = t
		}
	}

	if n > 1 {
		return nil, fmt.Errorf("Found more than one Template. Redefine!")
	}

	if n == 0 {
		return nil, fmt.Errorf("Could not find Template with name '%s'", name)
	}

	return &ret, nil
}
