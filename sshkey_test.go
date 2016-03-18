package gridscale

import (
	"os"
	"testing"
)

func TestCreateRenameDeleteSSHKey(t *testing.T) {
	userID := os.Getenv("GRIDSCALE_USERID")
	apiToken := os.Getenv("GRIDSCALE_APITOKEN")
	endpoint := "https://api.gridscale.io"
	c, _ := NewClient(userID, apiToken, endpoint)

	pub := "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDR+SXM8pSFsbjyDoilWfwchkve3cvTa7Q1sA8J92XXAf44iYtFPHS+gvrul5zSOEvgLvAlk2C7rGn0RjUm7Ap+uFCms/fCCiLVv2uTVpkru6xY+t/nUYu0F2woKrk3QmIp5UkFrXQ8CXCe6V+fY9su1t5BDRExl+O1ZlGnn7TZxFO6UAV6bdWXf+ovrdb2AlUSUOtWj7oYK42RNZfYiiUkQ7X6SDAJDab6xeX7056kYn240WJtKtTqVKda9eTQK/oSWIw327IgQ3xgo0tWaVy54h5O1UlTWgSDA1WHkDrcukbxR5jzEU5EZRQhvGuvnuVFpRN6Zi0Qd7rwlmcMU8XR user@example.com"

	newKey, err := c.AddSSHKey("testkey", pub, nil)
	if err != nil {
		t.Errorf("Cannot add SSH key: %s\n", err)
		return
	}
	defer c.DeleteSSHKey(newKey.ID)

	keys, err := c.GetSSHKeys()
	if err != nil {
		t.Errorf("Cannot read key list: %s\n", err)
		return
	}

	found := false
	for _, key := range keys {
		if key.ID == newKey.ID {
			found = true
		}
	}
	if !found {
		t.Errorf("Could not find a added ssh key %s in returned key list!", newKey.ID)
		return
	}

	err = c.DeleteSSHKey(newKey.ID)
	if err != nil {
		t.Errorf("Cannot delete SSH Key %s: %s", newKey.ID, err)
		return
	}

	keys, err = c.GetSSHKeys()
	if err != nil {
		t.Errorf("Cannot read key list: %s\n", err)
		return
	}

	for _, key := range keys {
		if key.ID == newKey.ID {
			t.Errorf("SSH key %s should have been deleted but is still in the list!", newKey.ID)
			return
		}
	}
}
