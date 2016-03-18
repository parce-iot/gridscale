package gridscale

import (
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/shopspring/decimal"
)

// IPServerRelation holds server name and uuid and creation date of a serrver/object relation
type IPServerRelation struct {
	ServerName      string    `json:"server_name"`
	ServerID        string    `json:"server_uuid"`
	RelationCreated time.Time `json:"create_time"`
}

type restIP struct {
	ID              string                        `json:"object_uuid"`
	Failover        bool                          `json:"failover"`
	Relations       map[string][]IPServerRelation `json:"relations,omitempty"`
	Labels          []string                      `json:"labels"`
	LocationName    string                        `json:"location_name"`
	IP              string                        `json:"ip"`
	Prefix          string                        `json:"prefix"`
	Family          int                           `json:"family"`
	ReverseDNS      string                        `json:"reverse_dns"`
	LocationCountry string                        `json:"location_country"`
	CurrentPrice    decimal.Decimal               `json:"current_price"`
	LocationID      string                        `json:"location_uuid"`
	LocationIata    string                        `json:"location_iata"`
}

type createIPRestRequest struct {
	Failover   bool     `json:"failover"`
	Labels     []string `json:"labels,omitempty"`
	Family     int      `json:"family"`
	ReverseDNS string   `json:"reverse_dns,omitempty"`
	LocationID string   `json:"location_uuid"`
}

type createIPRestResponse struct {
	Prefix string `json:"prefix"`
	IP     string `json:"ip"`
	ID     string `json:"object_uuid"`
}

type patchIPRestRequest struct {
	Labels     []string `json:"labels,omitempty"`
	Failover   *bool    `json:"failover,omitempty"`
	ReverseDNS string   `json:"reverse_dns,omitempty"`
}

func (rip *restIP) AsIP() IP {
	ip := IP{}
	ip.ID = rip.ID
	ip.Failover = rip.Failover
	ip.Servers = rip.Relations["servers"]
	ip.Labels = rip.Labels
	ip.LocationID = rip.LocationID
	ip.ReverseDNS = rip.ReverseDNS
	ip.IP = net.ParseIP(rip.IP)
	_, ipnet, _ := net.ParseCIDR(rip.Prefix)
	ip.Prefix = *ipnet
	ip.IPVersion = rip.Family

	return ip
}

// IP information from Gridscale
type IP struct {
	ID         string
	Failover   bool
	Servers    []IPServerRelation
	LocationID string
	ReverseDNS string
	Labels     []string
	IP         net.IP
	Prefix     net.IPNet
	IPVersion  int
}

// GetIPs returns a list of IP addresses.
func (c *Client) GetIPs() ([]IP, error) {
	resp, err := c.get("/objects/ips")
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == 401 {
		return nil, fmt.Errorf("Not authorized")
	}

	if resp.StatusCode == 204 {
		// empty response, no IPs to return
		return []IP{}, nil
	}

	decoder := json.NewDecoder(resp.Body)

	rips := objectWrapper{}
	err = decoder.Decode(&rips)
	if err != nil {
		return nil, err
	}

	ips := []IP{}
	for _, rip := range *rips.IPs {
		ips = append(ips, rip.AsIP())
	}

	return ips, nil
}

// GetIP returns IP information.
func (c *Client) GetIP(objectID string) (*IP, error) {
	resp, err := c.get("/objects/ips/" + objectID)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == 401 {
		return nil, fmt.Errorf("Not authorized")
	}

	if resp.StatusCode == 404 {
		return nil, fmt.Errorf("invalid object uuid")
	}

	if resp.StatusCode == 204 {
		return nil, fmt.Errorf("no IP information available")
	}

	decoder := json.NewDecoder(resp.Body)

	wrpr := objectWrapper{}
	err = decoder.Decode(&wrpr)
	if err != nil {
		return nil, err
	}

	if wrpr.IP == nil {
		return nil, fmt.Errorf("Empty Response")
	}

	ip := wrpr.IP.AsIP()
	return &ip, nil
}

// CreateIPv4 creates a new IPv4 address.
// locationID is the location ID (required).
// failover defines if the IP address is a failover IP.
// labels is a slice of strings (optional, set to nil if not used).
// reverseDNS is the reverse DNS entry (optional, set to nil if not used).
func (c *Client) CreateIPv4(locationID string, failover bool, labels []string, reverseDNS *string) (*IP, error) {
	return c.createIP(4, locationID, failover, labels, reverseDNS)
}

// CreateIPv6 creates a new IPv6 address.
// locationID is the location ID (required).
// failover defines if the IP address is a failover IP.
// labels is a slice of strings (optional, set to nil if not used).
// reverseDNS is the reverse DNS entry (optional, set to nil if not used).
func (c *Client) CreateIPv6(locationID string, failover bool, labels []string, reverseDNS *string) (*IP, error) {
	return c.createIP(6, locationID, failover, labels, reverseDNS)
}

func (c *Client) createIP(family int, locationID string, failover bool, labels []string, reverseDNS *string) (*IP, error) {
	rip := createIPRestRequest{}
	rip.Labels = labels
	if reverseDNS != nil {
		rip.ReverseDNS = *reverseDNS
	}
	rip.LocationID = locationID
	rip.Family = family
	rip.Failover = failover

	reqBody, err := json.Marshal(rip)
	if err != nil {
		return nil, err
	}

	resp, err := c.post("/objects/ips", reqBody)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == 202 {
		decoder := json.NewDecoder(resp.Body)

		respInfo := createIPRestResponse{}
		err = decoder.Decode(&respInfo)
		if err != nil {
			return nil, err
		}

		ip, err := c.GetIP(respInfo.ID)
		if err != nil {
			return nil, err
		}

		return ip, nil
	}

	if resp.StatusCode == 415 {
		return nil, fmt.Errorf("HTTP Request failed: Unsupported Media Type")
	}

	if resp.StatusCode == 404 {
		return nil, fmt.Errorf("Not authorized")
	}

	if resp.StatusCode == 402 {
		return nil, fmt.Errorf("No valid payment method available")
	}

	if resp.StatusCode == 400 {
		return nil, fmt.Errorf("Input data incomplete or invalid")
	}

	return nil, fmt.Errorf("Unknown Error (%d)", resp.StatusCode)
}

// UpdateIPLabels sets the IP object's labels
func (c *Client) UpdateIPLabels(ipID string, labels []string) error {
	p := patchIPRestRequest{}
	p.Labels = labels

	return c.patchIPAddress(ipID, &p)
}

// UpdateIPFailover enables or disables the IP's failover capability
func (c *Client) UpdateIPFailover(ipID string, failover bool) error {
	p := patchIPRestRequest{}
	p.Failover = &failover

	return c.patchIPAddress(ipID, &p)
}

// UpdateIPReverseDNS updates an IP's reverse DNS entry.
func (c *Client) UpdateIPReverseDNS(ipID string, reverseDNS string) error {
	p := patchIPRestRequest{}
	p.ReverseDNS = reverseDNS

	return c.patchIPAddress(ipID, &p)
}

func (c *Client) patchIPAddress(ipID string, p *patchIPRestRequest) error {
	reqBody, err := json.Marshal(p)
	if err != nil {
		return err
	}

	resp, err := c.patch("/objects/ips/"+ipID, reqBody)
	if err != nil {
		return err
	}

	if resp.StatusCode == 202 || resp.StatusCode == 204 {
		return nil
	}

	if resp.StatusCode == 400 {
		return fmt.Errorf("Input data incomplete or invalid")
	}

	if resp.StatusCode == 415 {
		return fmt.Errorf("REST HTTP: Unsupported Media Type")
	}

	if resp.StatusCode == 424 {
		return fmt.Errorf("Action not possible - object is in wrong status")
	}

	if resp.StatusCode == 402 {
		return fmt.Errorf("No valid payment method available")
	}

	if resp.StatusCode == 404 {
		return fmt.Errorf("Object ID not found")
	}

	return fmt.Errorf("Unknown Error (%d)", resp.StatusCode)
}

// DeleteIP deletes an IP address.
// This is only possible if the IP address has no relation to a server - remove the relation first!
func (c *Client) DeleteIP(objectID string) error {
	resp, err := c.delete("/objects/ips/"+objectID, nil)
	if err != nil {
		return err
	}

	if resp.StatusCode == 204 {
		return nil
	}

	if resp.StatusCode == 400 {
		return fmt.Errorf("Input data incomplete or invalid")
	}

	if resp.StatusCode == 424 {
		return fmt.Errorf("Action not possible - object is in wrong status")
	}

	if resp.StatusCode == 404 {
		return fmt.Errorf("Object ID not found")
	}

	return fmt.Errorf("Unknown Error (%d)", resp.StatusCode)
}
