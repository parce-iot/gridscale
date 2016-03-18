package gridscale

import (
	"os"
	"testing"
)

func TestCreateRenameDeleteNetwork(t *testing.T) {
	userID := os.Getenv("GRIDSCALE_USERID")
	apiToken := os.Getenv("GRIDSCALE_APITOKEN")
	endpoint := "https://api.gridscale.io"
	c, _ := NewClient(userID, apiToken, endpoint)

	n, err := c.CreateNetwork("45ed677b-3702-4b36-be2a-a2eab9827950", "foobar", true, nil)
	if err != nil {
		t.Errorf("Could not create network: %s", err)
		return
	}

	defer c.DeleteNetwork(n.ID)

	err = c.UpdateNetworkName(n.ID, "bambaz")
	if err != nil {
		t.Errorf("Could not update network name (%s): %s", n.ID, err)
		return
	}

	nets, err := c.GetNetworks()
	if err != nil {
		t.Errorf("Could not get list of networks: %s\n", err)
		return
	}

	found := false
	for _, net := range nets {
		if net.ID == n.ID {
			found = true
			if net.Name != "bambaz" {
				t.Errorf("Changing network name from 'foobar' to 'bambaz' did not work!")
				return
			}
		}
	}
	if !found {
		t.Errorf("Could not find a newly created network '%s' in returned network list!", n.ID)
		return
	}

	err = c.DeleteNetwork(n.ID)
	if err != nil {
		t.Errorf("Could not delete network: %s", err)
		return
	}

	nets, err = c.GetNetworks()
	if err != nil {
		t.Errorf("Could not get list of networks: %s\n", err)
		return
	}

	for _, net := range nets {
		if net.ID == n.ID {
			t.Errorf("Could not delete network %s", n.ID)
		}
	}
}
