package gridscale

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// SSHKey holds information about an ssh key.
type SSHKey struct {
	ID        string   `json:"object_uuid"`
	Name      string   `json:"name"`
	Labels    []string `json:"labels"`
	PublicKey string   `json:"sshkey"`
}

type restCreateSSHKeyRequest struct {
	Labels    []string `json:"labels,omitempty"`
	Name      string   `json:"name"`
	PublicKey string   `json:"sshkey"`
}

type restCreateSSHKeyResponse struct {
	ID string `json:"object_uuid"`
}

// AddSSHKey adds a new ssh key.
func (c *Client) AddSSHKey(name string, publicKey string, labels []string) (*SSHKey, error) {
	rReq := restCreateSSHKeyRequest{}
	rReq.Name = name
	rReq.Labels = labels
	rReq.PublicKey = publicKey

	reqBody, err := json.Marshal(rReq)
	if err != nil {
		return nil, err
	}

	resp, err := c.post("/objects/sshkeys", reqBody)
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	body := buf.Bytes()

	if resp.StatusCode == 202 {
		respInfo := restCreateSSHKeyResponse{}
		err = json.Unmarshal(body, &respInfo)
		if err != nil {
			return nil, err
		}

		net, err := c.GetSSHKey(respInfo.ID)
		if err != nil {
			return nil, err
		}

		return net, nil
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

// GetSSHKeys returns a list of all ssh keys.
func (c *Client) GetSSHKeys() ([]SSHKey, error) {
	resp, err := c.get("/objects/sshkeys")
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == 401 {
		return nil, fmt.Errorf("Not authorized")
	}

	if resp.StatusCode == 204 {
		// empty response, no keys to return
		return []SSHKey{}, nil
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	body := buf.Bytes()

	wrpr := objectWrapper{}
	err = json.Unmarshal(body, &wrpr)
	if err != nil {
		return nil, err
	}

	keys := []SSHKey{}
	for _, key := range *wrpr.SSHKeys {
		keys = append(keys, key)
	}

	return keys, nil
}

// GetSSHKey returns information about an ssh key identified by its ID.
func (c *Client) GetSSHKey(sshKeyID string) (*SSHKey, error) {
	resp, err := c.get("/objects/sshkeys/" + sshKeyID)
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	body := buf.Bytes()

	if resp.StatusCode == 404 {
		return nil, fmt.Errorf("SSHKey not found: %s", body)
	}

	if resp.StatusCode == 401 {
		return nil, fmt.Errorf("Not authorized: %s", body)
	}

	if resp.StatusCode == 204 {
		// empty response, no ssh keys to return
		return nil, fmt.Errorf("SSHKey not found: %s", body)
	}

	wrpr := objectWrapper{}
	err = json.Unmarshal(body, &wrpr)
	if err != nil {
		return nil, err
	}

	return wrpr.SSHKey, nil
}

// DeleteSSHKey deletes an SSH key or answers with an error message.
func (c *Client) DeleteSSHKey(sshkeyID string) error {
	resp, err := c.delete("/objects/sshkeys/"+sshkeyID, nil)
	if err != nil {
		return err
	}

	if resp.StatusCode == 424 {
		return fmt.Errorf("Action not possible - ssh key is in wrong status")
	}

	if resp.StatusCode == 404 {
		return fmt.Errorf("SSHKey not found")
	}

	if resp.StatusCode == 403 {
		return fmt.Errorf("SSHKey still in use")
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

// UpdateSSHKeyName updates an SSH key name identified by the SSH key id.
func (c *Client) UpdateSSHKeyName(sshkeyID string, name string) error {
	p := patchSSHKeyRequest{}
	p.Name = name
	return c.patchSSHKey(sshkeyID, &p)
}

// UpdateSSHKeyLabels changes the labels of an SSH key identified by its id.
func (c *Client) UpdateSSHKeyLabels(sshkeyID string, labels []string) error {
	p := patchSSHKeyRequest{}
	p.Labels = labels
	return c.patchSSHKey(sshkeyID, &p)
}

// UpdateSSHKeyPublicKey updates an SSH public key.
func (c *Client) UpdateSSHKeyPublicKey(sshkeyID string, publicKey string) error {
	p := patchSSHKeyRequest{}
	p.PublicKey = publicKey
	return c.patchSSHKey(sshkeyID, &p)
}

type patchSSHKeyRequest struct {
	Labels    []string `json:"labels,omitempty"`
	Name      string   `json:"name,omitempty"`
	PublicKey string   `json:"sshkey,omitempty"`
}

func (c *Client) patchSSHKey(sshKeyID string, p *patchSSHKeyRequest) error {
	reqBody, err := json.Marshal(p)
	if err != nil {
		return err
	}

	resp, err := c.patch("/objects/sshkeys/"+sshKeyID, reqBody)
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
