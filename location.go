package gridscale

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// Location holds information about a gridscale Location.
type Location struct {
	ID      string                      `json:"object_uuid"`
	Name    string                      `json:"name"`
	Iata    string                      `json:"iata"`
	Country string                      `json:"country"`
}

// GetLocations returns a list of all gridscale locations.
func (c *Client) GetLocations() ([]Location, error) {
	resp, err := c.get("/objects/locations")
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	body := buf.Bytes()

	if resp.StatusCode == 404 {
		return nil, fmt.Errorf("Locations not found: %s", body)
	}

	if resp.StatusCode == 401 {
		return nil, fmt.Errorf("Not authorized: %s", body)
	}

	if resp.StatusCode == 204 {
		// empty response, no servers to return
		return []Location{}, nil
	}

	wrpr := objectWrapper{}
	err = json.Unmarshal(body, &wrpr)
	if err != nil {
		return nil, err
	}

	locs := []Location{}
	for _, loc := range *wrpr.Locations {
		locs = append(locs, loc)
	}

	return locs, nil
}
