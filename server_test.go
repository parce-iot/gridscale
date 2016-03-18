package gridscale

import (
	"os"
	"testing"
)

func TestCreateRenameDeleteServer(t *testing.T) {
	userID := os.Getenv("GRIDSCALE_USERID")
	apiToken := os.Getenv("GRIDSCALE_APITOKEN")
	endpoint := "https://api.gridscale.io"
	c, _ := NewClient(userID, apiToken, endpoint)

	n, err := c.CreateServer("45ed677b-3702-4b36-be2a-a2eab9827950", "foobar", 2, 4, nil)
	if err != nil {
		t.Errorf("Could not create server: %s", err)
		return
	}

	defer c.DeleteServer(n.ID)

	err = c.UpdateServerName(n.ID, "bambaz")
	if err != nil {
		t.Errorf("Could not update server name (%s): %s", n.ID, err)
		return
	}

	srvrs, err := c.GetServers()
	if err != nil {
		t.Errorf("Could not get list of servers: %s\n", err)
		return
	}

	found := false
	for _, srv := range srvrs {
		if srv.ID == n.ID {
			found = true
			if srv.Name != "bambaz" {
				t.Errorf("Changing server name from 'foobar' to 'bambaz' did not work!")
				return
			}
		}
	}
	if !found {
		t.Errorf("Could not find a newly created server '%s' in returned server list!", n.ID)
		return
	}

	err = c.DeleteServer(n.ID)
	if err != nil {
		t.Errorf("Could not delete server: %s", err)
		return
	}

	srvrs, err = c.GetServers()
	if err != nil {
		t.Errorf("Could not get list of servers: %s\n", err)
		return
	}

	for _, srv := range srvrs {
		if srv.ID == n.ID {
			t.Errorf("Could not delete server %s", n.ID)
		}
	}
}
