package gridscale

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
)

// Server holds information about a server.
type Server struct {
	ID              string                      `json:"object_uuid"`
	Name            string                      `json:"name"`
	Status          string                      `json:"status"`
	Labels          []string                    `json:"labels"`
	Cores           int                         `json:"cores"`
	Power           bool                        `json:"power"`
	ConsoleToken    string                      `json:"console_token"`
	CurrentPrice    decimal.Decimal             `json:"current_price"`
	Relations       map[string][]ServerRelation `json:"relations"`
	CreateTime      time.Time                   `json:"create_time"`
	ChangeTime      time.Time                   `json:"change_time"`
	LocationName    string                      `json:"location_name"`
	LocationIata    string                      `json:"location_iata"`
	LocationID      string                      `json:"location_uuid"`
	LocationCountry string                      `json:"location_country"`
}

// ServerRelation holds all relations to other gridscale objects (storages, networks etc.).
type ServerRelation struct {
	Distance  []interface{}           `json:"distance"`  // TODO: what's that?
	IsoImages []interface{}           `json:"isoimages"` // TODO
	Networks  []ServerNetworkRelation `json:"networks"`
	PublicIPs []interface{}           `json:"public_ips"` // TODO
	Storages  []ServerStorageRelation `json:"storages"`
}

// ServerNetworkRelation contains information about a network connected to a server.
type ServerNetworkRelation struct {
	NetworkID   string `json:"object_uuid"`
	NetworkName string `json:"object_name"`
	MAC         string `json:"mac"`
	Ordering    int    `json:"ordering"`
}

// ServerStorageRelation contains information about a storage volume connected to a server
type ServerStorageRelation struct {
	StorageID   string `json:"object_uuid"`
	StorageName string `json:"object_name"`
	BootDevice  bool   `json:"bootdevice"`
	Bus         int    `json:"bus"`
	Capacity    int    `json:"capacity"`
	Controller  int    `json:"controller"`
	LUN         int    `json:"lun"`
}

type restCreateServerRequest struct {
	Labels     []string `json:"labels,omitempty"`
	Name       string   `json:"name"`
	LocationID string   `json:"location_uuid"`
	Memory     int      `json:"memory"`
	Cores      int      `json:"cores"`
}

type restCreateServerResponse struct {
	ID string `json:"object_uuid"`
}

// CreateServer creates a new server volume.
func (c *Client) CreateServer(locationID string, name string, cores int, memoryGB int, labels []string) (*Server, error) {
	rr := restCreateServerRequest{}
	rr.Labels = labels
	rr.LocationID = locationID
	rr.Name = name
	rr.Memory = memoryGB
	rr.Cores = cores

	reqBody, err := json.Marshal(rr)
	if err != nil {
		return nil, err
	}

	resp, err := c.post("/objects/servers", reqBody)
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	body := buf.Bytes()
	if resp.StatusCode == 202 {
		respInfo := restCreateServerResponse{}
		err = json.Unmarshal(body, &respInfo)
		if err != nil {
			return nil, err
		}

		sto, err := c.GetServer(respInfo.ID)
		if err != nil {
			return nil, err
		}

		return sto, nil
	}

	if resp.StatusCode == 415 {
		return nil, fmt.Errorf("HTTP Request failed: Unsupported Media Type: %s", body)
	}

	if resp.StatusCode == 404 {
		return nil, fmt.Errorf("Not authorized: %s", body)
	}

	if resp.StatusCode == 402 {
		return nil, fmt.Errorf("No valid payment method available: %s", body)
	}

	if resp.StatusCode == 400 {
		return nil, fmt.Errorf("Input data incomplete or invalid: %s", body)
	}

	return nil, fmt.Errorf("Unknown Error (%d): %s", resp.StatusCode, body)
}

// GetServers returns a list of all servers.
func (c *Client) GetServers() ([]Server, error) {
	resp, err := c.get("/objects/servers")
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	body := buf.Bytes()

	if resp.StatusCode == 404 {
		return nil, fmt.Errorf("Servers not found: %s", body)
	}

	if resp.StatusCode == 401 {
		return nil, fmt.Errorf("Not authorized: %s", body)
	}

	if resp.StatusCode == 204 {
		// empty response, no servers to return
		return []Server{}, nil
	}

	wrpr := objectWrapper{}
	err = json.Unmarshal(body, &wrpr)
	if err != nil {
		return nil, err
	}

	srvs := []Server{}
	for _, srv := range *wrpr.Servers {
		srvs = append(srvs, srv)
	}

	return srvs, nil
}

// GetServer returns information about a server identified by its ID.
func (c *Client) GetServer(serverID string) (*Server, error) {
	resp, err := c.get("/objects/servers/" + serverID)
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	body := buf.Bytes()

	if resp.StatusCode == 404 {
		return nil, fmt.Errorf("Server not found: %s", body)
	}

	if resp.StatusCode == 401 {
		return nil, fmt.Errorf("Not authorized: %s", body)
	}

	if resp.StatusCode == 204 {
		// empty response, no networks to return
		return nil, fmt.Errorf("Server not found: %s", body)
	}

	wrpr := objectWrapper{}
	err = json.Unmarshal(body, &wrpr)
	if err != nil {
		return nil, err
	}

	return wrpr.Server, nil
}

// DeleteServer deletes a server or answers with an error message.
func (c *Client) DeleteServer(serverID string) error {
	resp, err := c.delete("/objects/servers/"+serverID, nil)
	if err != nil {
		return err
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	body := buf.Bytes()

	if resp.StatusCode == 424 {
		return fmt.Errorf("Action not possible - network is in wrong status: %s", body)
	}

	if resp.StatusCode == 404 {
		return fmt.Errorf("Server not found: %s", body)
	}

	if resp.StatusCode == 403 {
		return fmt.Errorf("Server is still in use: %s", body)
	}

	if resp.StatusCode == 401 {
		return fmt.Errorf("Not authorized: %s", body)
	}

	if resp.StatusCode == 400 {
		return fmt.Errorf("Invalid Request: %s", body)
	}

	if resp.StatusCode == 204 { // ok
		return nil
	}

	return fmt.Errorf("Unknown Error (%d): %s", resp.StatusCode, body)
}

// UpdateServerName updates a server's name identified by the network id.
func (c *Client) UpdateServerName(serverID string, name string) error {
	p := patchServerRequest{}
	p.Name = name
	return c.patchServer(serverID, &p)
}

// UpdateServerLabels changes the labels of a server identified by its id.
func (c *Client) UpdateServerLabels(serverID string, labels []string) error {
	p := patchServerRequest{}
	p.Labels = labels
	return c.patchServer(serverID, &p)
}

// UpdateServerCores changes the server's number of cores.
func (c *Client) UpdateServerCores(serverID string, cores int) error {
	p := patchServerRequest{}
	p.Cores = &cores
	return c.patchServer(serverID, &p)
}

// UpdateServerMemory changes the server's memory [GB].
func (c *Client) UpdateServerMemory(serverID string, memoryGB int) error {
	p := patchServerRequest{}
	p.Memory = &memoryGB
	return c.patchServer(serverID, &p)
}

func (c *Client) updateServerPowerStatus(serverID string, powerOn bool) error {
	p := map[string]bool{}
	p["power"] = powerOn

	reqBody, err := json.Marshal(p)
	if err != nil {
		return err
	}

	resp, err := c.patch("/objects/servers/"+serverID+"/power", reqBody)
	if err != nil {
		return err
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	body := buf.Bytes()

	if resp.StatusCode == 202 || resp.StatusCode == 204 {
		return nil
	}

	if resp.StatusCode == 400 {
		return fmt.Errorf("Input data incomplete or invalid: %s", body)
	}

	if resp.StatusCode == 415 {
		return fmt.Errorf("REST HTTP: Unsupported Media Type: %s", body)
	}

	if resp.StatusCode == 424 {
		return fmt.Errorf("Action not possible - object is in wrong status: %s", body)
	}

	if resp.StatusCode == 404 {
		return fmt.Errorf("Object ID not found: %s", body)
	}

	return fmt.Errorf("Unknown Error (%d): %s", resp.StatusCode, body)
}

// PowerOnServer turns a server on.
func (c *Client) PowerOnServer(serverID string) error {
	return c.updateServerPowerStatus(serverID, true)
}

// PowerOffServer turns a server off.
func (c *Client) PowerOffServer(serverID string) error {
	return c.updateServerPowerStatus(serverID, false)
}

type patchServerRequest struct {
	Labels []string `json:"labels,omitempty"`
	Name   string   `json:"name,omitempty"`
	Power  *bool    `json:"power,omitempty"`
	Cores  *int     `json:"cores,omitempty"`
	Memory *int     `json:"memory,omitempty"`
}

func (c *Client) patchServer(serverID string, p *patchServerRequest) error {
	reqBody, err := json.Marshal(p)
	if err != nil {
		return err
	}

	resp, err := c.patch("/objects/servers/"+serverID, reqBody)
	if err != nil {
		return err
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	body := buf.Bytes()

	if resp.StatusCode == 202 || resp.StatusCode == 204 {
		return nil
	}

	if resp.StatusCode == 400 {
		return fmt.Errorf("Input data incomplete or invalid: %s", body)
	}

	if resp.StatusCode == 415 {
		return fmt.Errorf("REST HTTP: Unsupported Media Type: %s", body)
	}

	if resp.StatusCode == 424 {
		return fmt.Errorf("Action not possible - object is in wrong status: %s", body)
	}

	if resp.StatusCode == 404 {
		return fmt.Errorf("Object ID not found: %s", body)
	}

	return fmt.Errorf("Unknown Error (%d): %s", resp.StatusCode, body)
}

// ConnectIPAddress maps a public IP address to a server connected to the public network.
func (c *Client) ConnectIPAddress(ipAddressID, serverID string) error {
	m := map[string]string{"object_uuid": ipAddressID}
	reqBody, err := json.Marshal(&m)
	if err != nil {
		return err
	}

	resp, err := c.post("/objects/servers/"+serverID+"/ips", reqBody)
	if err != nil {
		return err
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	body := buf.Bytes()

	if resp.StatusCode == 202 {
		return nil
	}

	if resp.StatusCode == 400 {
		return fmt.Errorf("Input data incomplete or invalid: %s", body)
	}

	if resp.StatusCode == 403 {
		return fmt.Errorf("Server and IP are in different locations: %s", body)
	}

	if resp.StatusCode == 409 {
		return fmt.Errorf("IP is already connected to server: %s", body)
	}

	if resp.StatusCode == 404 {
		return fmt.Errorf("Not found: %s", body)
	}

	return fmt.Errorf("Unknown Error (%d): %s", resp.StatusCode, body)
}

// DisconnectIPAddress disconnects an IP address from a server.
func (c *Client) DisconnectIPAddress(ipAddressID, serverID string) error {
	resp, err := c.delete("/objects/servers/"+serverID+"/ips/"+ipAddressID, nil)
	if err != nil {
		return err
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	body := buf.Bytes()

	if resp.StatusCode == 202 {
		return nil
	}

	if resp.StatusCode == 400 {
		return fmt.Errorf("Input data incomplete or invalid: %s", body)
	}

	if resp.StatusCode == 404 {
		return fmt.Errorf("Not found: %s", body)
	}

	if resp.StatusCode == 424 {
		return fmt.Errorf("Action not possible - object in wrong status: %s", body)
	}

	return fmt.Errorf("Unknown Error (%d): %s", resp.StatusCode, body)
}

type connectNetworkRequest struct {
	NetworkID string `json:"object_uuid"`
	Ordering  int    `json:"ordering"`
}

// ConnectNetwork connects a server to a network. A new ethernet device is added to the server.
func (c *Client) ConnectNetwork(networkID string, ordering int, serverID string) error {
	m := connectNetworkRequest{}
	m.NetworkID = networkID
	m.Ordering = ordering

	reqBody, err := json.Marshal(&m)
	if err != nil {
		return err
	}

	resp, err := c.post("/objects/servers/"+serverID+"/networks", reqBody)
	if err != nil {
		return err
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	body := buf.Bytes()

	if resp.StatusCode == 202 {
		return nil
	}

	if resp.StatusCode == 400 {
		return fmt.Errorf("Input data incomplete or invalid: %s", body)
	}

	if resp.StatusCode == 403 {
		return fmt.Errorf("Server and network are in different locations: %s", body)
	}

	if resp.StatusCode == 409 {
		return fmt.Errorf("Network is already connected to server: %s", body)
	}

	if resp.StatusCode == 404 {
		return fmt.Errorf("Not found: %s", body)
	}

	return fmt.Errorf("Unknown Error (%d): %s", resp.StatusCode, body)
}

// DisconnectNetwork disconnects a network from a server.
func (c *Client) DisconnectNetwork(networkID, serverID string) error {
	resp, err := c.delete("/objects/servers/"+serverID+"/networks/"+networkID, nil)
	if err != nil {
		return err
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	body := buf.Bytes()

	if resp.StatusCode == 202 {
		return nil
	}

	if resp.StatusCode == 400 {
		return fmt.Errorf("Input data incomplete or invalid: %s", body)
	}

	if resp.StatusCode == 404 {
		return fmt.Errorf("Not found: %s", body)
	}

	if resp.StatusCode == 424 {
		return fmt.Errorf("Action not possible - object in wrong status: %s", body)
	}

	return fmt.Errorf("Unknown Error (%d): %s", resp.StatusCode, body)
}

type connectStorageRequest struct {
	StorageID  string `json:"object_uuid"`
	Bootdevice bool   `json:"bootdevice"`
}

// ConnectStorage connects a storage volume to a server.
func (c *Client) ConnectStorage(storageID string, bootdevice bool, serverID string) error {
	m := connectStorageRequest{}
	m.StorageID = storageID
	m.Bootdevice = bootdevice

	reqBody, err := json.Marshal(&m)
	if err != nil {
		return err
	}

	resp, err := c.post("/objects/servers/"+serverID+"/storages", reqBody)
	if err != nil {
		return err
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	body := buf.Bytes()

	if resp.StatusCode == 202 {
		return nil
	}

	if resp.StatusCode == 400 {
		return fmt.Errorf("Input data incomplete or invalid: %s", body)
	}

	if resp.StatusCode == 403 {
		return fmt.Errorf("Server and storage are in different locations: %s", body)
	}

	if resp.StatusCode == 409 {
		return fmt.Errorf("Storage is already connected to server: %s", body)
	}

	if resp.StatusCode == 404 {
		return fmt.Errorf("Not found: %s", body)
	}

	return fmt.Errorf("Unknown Error (%d): %s", resp.StatusCode, body)

}

// DisconnectStorage disconnects a storage volume from a server.
func (c *Client) DisconnectStorage(storageID, serverID string) error {
	resp, err := c.delete("/objects/servers/"+serverID+"/storages/"+storageID, nil)
	if err != nil {
		return err
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	body := buf.Bytes()

	if resp.StatusCode == 202 {
		return nil
	}

	if resp.StatusCode == 400 {
		return fmt.Errorf("Input data incomplete or invalid: %s", body)
	}

	if resp.StatusCode == 404 {
		return fmt.Errorf("Not found: %s", body)
	}

	if resp.StatusCode == 424 {
		return fmt.Errorf("Action not possible - object in wrong status: %s", body)
	}

	return fmt.Errorf("Unknown Error (%d): %s", resp.StatusCode, body)
}

// ConnectIsoImage connects an ISO image as an optical disk drive to a virtual machine.
func (c *Client) ConnectIsoImage(isoImageID, serverID string) error {
	m := map[string]string{"object_uuid": isoImageID}
	reqBody, err := json.Marshal(&m)
	if err != nil {
		return err
	}

	resp, err := c.post("/objects/servers/"+serverID+"/isoimages", reqBody)
	if err != nil {
		return err
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	body := buf.Bytes()

	if resp.StatusCode == 202 {
		return nil
	}

	if resp.StatusCode == 400 {
		return fmt.Errorf("Input data incomplete or invalid: %s", body)
	}

	if resp.StatusCode == 403 {
		return fmt.Errorf("Server and IP are in different locations: %s", body)
	}

	if resp.StatusCode == 409 {
		return fmt.Errorf("IP is already connected to server: %s", body)
	}

	if resp.StatusCode == 404 {
		return fmt.Errorf("Not found: %s", body)
	}

	return fmt.Errorf("Unknown Error (%d): %s", resp.StatusCode, body)
}

// DisconnectIsoImage disconnects an ISO image from a virtual machine.
func (c *Client) DisconnectIsoImage(isoImageID, serverID string) error {
	resp, err := c.delete("/objects/servers/"+serverID+"/isoimages/"+isoImageID, nil)
	if err != nil {
		return err
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	body := buf.Bytes()

	if resp.StatusCode == 202 {
		return nil
	}

	if resp.StatusCode == 400 {
		return fmt.Errorf("Input data incomplete or invalid: %s", body)
	}

	if resp.StatusCode == 404 {
		return fmt.Errorf("Not found: %s", body)
	}

	if resp.StatusCode == 424 {
		return fmt.Errorf("Action not possible - object in wrong status: %s", body)
	}

	return fmt.Errorf("Unknown Error (%d): %s", resp.StatusCode, body)
}
