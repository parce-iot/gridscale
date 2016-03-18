package gridscale

import (
	"encoding/json"
	"fmt"
	"time"
)

type restNetwork struct {
	ID              string                             `json:"object_uuid"`
	Name            string                             `json:"name"`
	Status          string                             `json:"status"`
	Labels          []string                           `json:"labels"`
	PublicNet       bool                               `json:"public_net"`
	L2Security      bool                               `json:"l2security"`
	Relations       map[string][]NetworkServerRelation `json:"relations"`
	CreateTime      time.Time                          `json:"create_time"`
	ChangeTime      time.Time                          `json:"change_time"`
	LocationName    string                             `json:"location_name"`
	LocationIata    string                             `json:"location_iata"`
	LocationID      string                             `json:"location_uuid"`
	LocationCountry string                             `json:"location_country"`
}

func (n *restNetwork) AsNetwork() Network {
	net := Network{}
	net.ID = n.ID
	net.Name = n.Name
	net.Status = n.Status
	net.Labels = n.Labels
	net.PublicNet = n.PublicNet
	net.L2Security = n.L2Security
	net.Relations = n.Relations["servers"]
	net.LocationID = n.LocationID

	return net
}

// NetworkServerRelation is a relation that describes the connection between a
// network and a server including the server's MAC address.
type NetworkServerRelation struct {
	ServerName      string    `json:"object_name"`
	ServerID        string    `json:"object_uuid"`
	Ordering        int       `json:"ordering"`
	MAC             string    `json:"mac"`
	RelationCreated time.Time `json:"create_time"`
}

// Network holds information about a network.
type Network struct {
	ID         string
	Name       string
	Status     string
	Labels     []string
	PublicNet  bool
	L2Security bool
	Relations  []NetworkServerRelation
	LocationID string
}

type restCreateNetworkRequest struct {
	Labels     []string `json:"labels,omitempty"`
	Name       string   `json:"name"`
	LocationID string   `json:"location_uuid"`
	L2Security bool     `json:"l2security"`
}

type restCreateNetworkResponse struct {
	ID string `json:"object_uuid"`
}

// CreateNetwork creates a new network.
// locationID is required.
// name is required.
// labels are optional.
// returns information about the created network or an error.
func (c *Client) CreateNetwork(locationID string, name string, l2security bool, labels []string) (*Network, error) {
	rReq := restCreateNetworkRequest{}
	rReq.Labels = labels
	rReq.LocationID = locationID
	rReq.L2Security = l2security
	rReq.Name = name

	reqBody, err := json.Marshal(rReq)
	if err != nil {
		return nil, err
	}

	resp, err := c.post("/objects/networks", reqBody)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == 202 {
		decoder := json.NewDecoder(resp.Body)

		respInfo := restCreateNetworkResponse{}
		err = decoder.Decode(&respInfo)
		if err != nil {
			return nil, err
		}

		net, err := c.GetNetwork(respInfo.ID)
		if err != nil {
			return nil, err
		}

		return net, nil
	}

	if resp.StatusCode == 415 {
		return nil, fmt.Errorf("HTTP Request failed: Unsupported Media Type")
	}

	if resp.StatusCode == 404 {
		return nil, fmt.Errorf("Not authorized")
	}

	if resp.StatusCode == 400 {
		return nil, fmt.Errorf("Input data incomplete or invalid")
	}

	return nil, fmt.Errorf("Unknown Error (%d)", resp.StatusCode)
}

// GetNetworks returns a list of all networks.
func (c *Client) GetNetworks() ([]Network, error) {
	resp, err := c.get("/objects/networks")
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == 401 {
		return nil, fmt.Errorf("Not authorized")
	}

	if resp.StatusCode == 204 {
		// empty response, no networks to return
		return []Network{}, nil
	}

	decoder := json.NewDecoder(resp.Body)

	wrpr := objectWrapper{}
	err = decoder.Decode(&wrpr)
	if err != nil {
		return nil, err
	}

	nets := []Network{}
	for _, rnet := range *wrpr.Networks {
		nets = append(nets, rnet.AsNetwork())
	}

	return nets, nil
}

// GetPublicNetwork returns a reference to a public network.
func (c *Client) GetPublicNetwork() (*Network, error) {
	nets, err := c.GetNetworks()
	if err != nil {
		return nil, err
	}

	for _, net := range nets {
		if net.PublicNet {
			return &net, nil
		}
	}

	return nil, fmt.Errorf("Did not find public network")
}

// GetNetwork returns information about a network identified by its ID.
func (c *Client) GetNetwork(networkID string) (*Network, error) {
	resp, err := c.get("/objects/networks/" + networkID)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == 404 {
		return nil, fmt.Errorf("Network not found")
	}

	if resp.StatusCode == 401 {
		return nil, fmt.Errorf("Not authorized")
	}

	if resp.StatusCode == 204 {
		// empty response, no networks to return
		return nil, fmt.Errorf("Network not found")
	}

	decoder := json.NewDecoder(resp.Body)

	wrpr := objectWrapper{}
	err = decoder.Decode(&wrpr)
	if err != nil {
		return nil, err
	}

	net := wrpr.Network.AsNetwork()

	return &net, nil
}

// DeleteNetwork deletes a network or answers with an error message.
func (c *Client) DeleteNetwork(networkID string) error {
	resp, err := c.delete("/objects/networks/"+networkID, nil)
	if err != nil {
		return err
	}

	if resp.StatusCode == 424 {
		return fmt.Errorf("Action not possible - network is in wrong status")
	}

	if resp.StatusCode == 404 {
		return fmt.Errorf("Network not found")
	}

	if resp.StatusCode == 403 {
		return fmt.Errorf("Network still in use")
	}

	if resp.StatusCode == 401 {
		return fmt.Errorf("Not authorized")
	}

	if resp.StatusCode == 400 {
		return fmt.Errorf("Invalid Request")
	}

	if resp.StatusCode == 204 { // ok
		return nil
	}

	return fmt.Errorf("Unknown Error (%d)", resp.StatusCode)
}

// UpdateNetworkName updates a networks name identified by the network id.
func (c *Client) UpdateNetworkName(networkID string, name string) error {
	p := patchNetworkRequest{}
	p.Name = name
	return c.patchNetwork(networkID, &p)
}

// UpdateNetworkLabels changes the labels of a network identified by its id.
func (c *Client) UpdateNetworkLabels(networkID string, labels []string) error {
	p := patchNetworkRequest{}
	p.Labels = labels
	return c.patchNetwork(networkID, &p)
}

type patchNetworkRequest struct {
	Labels []string `json:"labels,omitempty"`
	Name   string   `json:"name,omitempty"`
}

func (c *Client) patchNetwork(networkID string, p *patchNetworkRequest) error {
	reqBody, err := json.Marshal(p)
	if err != nil {
		return err
	}

	resp, err := c.patch("/objects/networks/"+networkID, reqBody)
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

	if resp.StatusCode == 404 {
		return fmt.Errorf("Object ID not found")
	}

	return fmt.Errorf("Unknown Error")
}
