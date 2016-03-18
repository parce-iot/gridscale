package gridscale

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
)

type restStorage struct {
	ID                 string                             `json:"object_uuid"`
	ParentID           string                             `json:"parent_uuid"`
	LastUsedTemplateID *string                            `json:"last_used_template"`
	Name               string                             `json:"name"`
	Status             string                             `json:"status"`
	Labels             []string                           `json:"labels"`
	Capacity           int                                `json:"capacity"`
	Snapshots          []StorageSnapshot                  `json:"snapshots"`
	CurrentPrice       decimal.Decimal                    `json:"current_price"`
	Relations          map[string][]StorageServerRelation `json:"relations"`
	CreateTime         time.Time                          `json:"create_time"`
	ChangeTime         time.Time                          `json:"change_time"`
	LocationName       string                             `json:"location_name"`
	LocationIata       string                             `json:"location_iata"`
	LocationID         string                             `json:"location_uuid"`
	LocationCountry    string                             `json:"location_country"`
}

func (r *restStorage) AsStorage() Storage {
	s := Storage{}
	s.ID = r.ID
	s.ParentID = r.ParentID
	s.LastUsedTemplateID = r.LastUsedTemplateID
	s.Name = r.Name
	s.Status = r.Status
	s.Labels = r.Labels
	s.Capacity = r.Capacity
	s.Snapshots = r.Snapshots
	s.CurrentPrice = r.CurrentPrice
	s.Relations = r.Relations["servers"]
	s.CreateTime = r.CreateTime
	s.ChangeTime = r.ChangeTime
	s.LocationID = r.LocationID

	return s
}

// StorageSnapshot holds information of a storage snapshot.
type StorageSnapshot struct {
	ID                 string    `json:"object_uuid"`
	Name               string    `json:"object_name"`
	Capacity           string    `json:"object_capacity"`
	LastUsedTemplateID *string   `json:"last_used_template,omitempty"`
	CreateTime         time.Time `json:"create_time"`
}

// StorageServerRelation holds information about a server connected to a storage.
type StorageServerRelation struct {
	ServerID        string    `json:"object_uuid"`
	ServerName      string    `json:"object_name"`
	RelationCreated time.Time `json:"create_time"`
	Lun             int       `json:"lun"`
	Bus             int       `json:"bus"`
	BootDevice      bool      `json:"bootdevice"`
	Controller      int       `json:"controller"`
	Target          int       `json:"target"`
}

// Storage holds information about a storage.
type Storage struct {
	ID                 string
	ParentID           string
	LastUsedTemplateID *string
	Name               string
	Status             string
	Labels             []string
	Capacity           int
	Snapshots          []StorageSnapshot
	CurrentPrice       decimal.Decimal
	Relations          []StorageServerRelation
	CreateTime         time.Time
	ChangeTime         time.Time
	LocationID         string
}

// StorageTemplateParameters holds parameters to create a storage volume from a template
type StorageTemplateParameters struct {
	Hostname     string   `json:"hostname"`
	Password     string   `json:"password,omitempty"`
	PasswordType string   `json:"password_type,omitempty"` // plain
	TemplateID   string   `json:"template_uuid"`
	SSHKeyIDs    []string `json:"sshkeys,omitempty"`
}

type restCreateStorageRequest struct {
	Labels     []string                   `json:"labels,omitempty"`
	Name       string                     `json:"name"`
	LocationID string                     `json:"location_uuid"`
	Capacity   int                        `json:"capacity"`
	Template   *StorageTemplateParameters `json:"template,omitempty"`
}

type restCreateStorageResponse struct {
	ID string `json:"object_uuid"`
}

// CreateStorage creates a new storage volume.
func (c *Client) CreateStorage(locationID string, name string, capacity int, template *StorageTemplateParameters, labels []string) (*Storage, error) {
	rr := restCreateStorageRequest{}
	rr.Labels = labels
	rr.LocationID = locationID
	rr.Name = name
	rr.Capacity = capacity
	rr.Template = template

	reqBody, err := json.Marshal(rr)
	if err != nil {
		return nil, err
	}

	resp, err := c.post("/objects/storages", reqBody)
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	body := buf.Bytes()
	if resp.StatusCode == 202 {
		respInfo := restCreateStorageResponse{}
		err = json.Unmarshal(body, &respInfo)
		if err != nil {
			return nil, err
		}

		sto, err := c.GetStorage(respInfo.ID)
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

	if resp.StatusCode == 400 {
		return nil, fmt.Errorf("Input data incomplete or invalid: %s", body)
	}

	return nil, fmt.Errorf("Unknown Error (%d): %s", resp.StatusCode, body)
}

// GetStorages returns a list of all storages.
func (c *Client) GetStorages() ([]Storage, error) {
	resp, err := c.get("/objects/storages")
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	body := buf.Bytes()

	if resp.StatusCode == 404 {
		return nil, fmt.Errorf("Storage not found: %s", body)
	}

	if resp.StatusCode == 401 {
		return nil, fmt.Errorf("Not authorized: %s", body)
	}

	if resp.StatusCode == 204 {
		// empty response, no networks to return
		return nil, fmt.Errorf("Storage not found: %s", body)
	}

	wrpr := objectWrapper{}
	err = json.Unmarshal(body, &wrpr)
	if err != nil {
		return nil, err
	}

	strgs := []Storage{}
	for _, strg := range *wrpr.Storages {
		strgs = append(strgs, strg.AsStorage())
	}

	return strgs, nil
}

// GetStorage returns information about a storage identified by its ID.
func (c *Client) GetStorage(storageID string) (*Storage, error) {
	resp, err := c.get("/objects/storages/" + storageID)
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	body := buf.Bytes()

	if resp.StatusCode == 404 {
		return nil, fmt.Errorf("Storage not found: %s", body)
	}

	if resp.StatusCode == 401 {
		return nil, fmt.Errorf("Not authorized: %s", body)
	}

	if resp.StatusCode == 204 {
		// empty response, no networks to return
		return nil, fmt.Errorf("Storage not found: %s", body)
	}

	wrpr := objectWrapper{}
	err = json.Unmarshal(body, &wrpr)
	if err != nil {
		return nil, err
	}

	net := wrpr.Storage.AsStorage()

	return &net, nil
}

// DeleteStorage deletes a storage or answers with an error message.
func (c *Client) DeleteStorage(storageID string) error {
	resp, err := c.delete("/objects/storages/"+storageID, nil)
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
		return fmt.Errorf("Storage not found: %s", body)
	}

	if resp.StatusCode == 403 {
		return fmt.Errorf("Storage is still in use: %s", body)
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

// UpdateStorageName updates a storage's name identified by the network id.
func (c *Client) UpdateStorageName(storageID string, name string) error {
	p := patchStorageRequest{}
	p.Name = name
	return c.patchStorage(storageID, &p)
}

// UpdateStorageLabels changes the labels of a storage identified by its id.
func (c *Client) UpdateStorageLabels(storageID string, labels []string) error {
	p := patchStorageRequest{}
	p.Labels = labels
	return c.patchStorage(storageID, &p)
}

// UpdateStorageCapacity changes the capacity of a storage identified by its id.
func (c *Client) UpdateStorageCapacity(storageID string, capacity int) error {
	p := patchStorageRequest{}
	p.Capacity = &capacity
	return c.patchStorage(storageID, &p)
}

type patchStorageRequest struct {
	Labels   []string `json:"labels,omitempty"`
	Name     string   `json:"name,omitempty"`
	Capacity *int     `json:"capacity,omitempty"`
}

func (c *Client) patchStorage(storageID string, p *patchStorageRequest) error {
	reqBody, err := json.Marshal(p)
	if err != nil {
		return err
	}

	resp, err := c.patch("/objects/storages/"+storageID, reqBody)
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
