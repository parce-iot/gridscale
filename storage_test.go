package gridscale

import (
	"os"
	"testing"
	"time"
)

func TestCreateRenameDeleteStorage(t *testing.T) {
	userID := os.Getenv("GRIDSCALE_USERID")
	apiToken := os.Getenv("GRIDSCALE_APITOKEN")
	endpoint := "https://api.gridscale.io"
	c, _ := NewClient(userID, apiToken, endpoint)

	s, err := c.CreateStorage("45ed677b-3702-4b36-be2a-a2eab9827950", "foobar", 10, nil, nil)
	if err != nil {
		t.Errorf("Could not create storage: %s", err)
		return
	}

	defer c.DeleteStorage(s.ID)

	for s.Status == "in-provisioning" {
		time.Sleep(time.Second)
		s, err = c.GetStorage(s.ID)
	}
	if s.Status != "active" {
		t.Errorf("Unexpected storage status after creation: %s expected %s", s.Status, "active")
	}

	err = c.UpdateStorageName(s.ID, "bambaz")
	if err != nil {
		t.Errorf("Could not update storage name (%s): %s", s.ID, err)
		return
	}

	strgs, err := c.GetStorages()
	if err != nil {
		t.Errorf("Could not get list of storages: %s\n", err)
		return
	}

	found := false
	for _, strg := range strgs {
		if strg.ID == s.ID {
			found = true
			if strg.Name != "bambaz" {
				t.Errorf("Changing storage name from 'foobar' to 'bambaz' did not work!")
				return
			}
		}
	}
	if !found {
		t.Errorf("Could not find a newly created storage '%s' in returned storage list!", s.ID)
		return
	}

	err = c.DeleteStorage(s.ID)
	if err != nil {
		t.Errorf("Could not delete storage: %s", err)
		return
	}

	strgs, err = c.GetStorages()
	if err != nil {
		t.Errorf("Could not get list of storage: %s\n", err)
		return
	}

	for _, strg := range strgs {
		if strg.ID == s.ID && strg.Status != "to-be-deleted" {
			t.Errorf("Could not delete storage %s: %s", s.ID, strg.Status)
		}
	}
}
