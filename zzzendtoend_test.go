package gridscale

import (
	"os"
	"testing"
	"time"
)

func TestZZZEndToEnd(t *testing.T) {
	userID := os.Getenv("GRIDSCALE_USERID")
	apiToken := os.Getenv("GRIDSCALE_APITOKEN")
	endpoint := "https://api.gridscale.io"
	locationID := "45ed677b-3702-4b36-be2a-a2eab9827950"
	c, _ := NewClient(userID, apiToken, endpoint)

	temp, err := c.GetTemplateByName("Debian")
	if err != nil {
		t.Errorf("Could not get Template by Name 'Debian': %s", err)
		return
	}

	storTempl := StorageTemplateParameters{}
	storTempl.Hostname = "foobar"
	storTempl.Password = "foobar"
	storTempl.PasswordType = "plain"
	storTempl.TemplateID = temp.ID

	pubnet, err := c.GetPublicNetwork()
	if err != nil {
		t.Errorf("Cannot get Public network: %s", err)
		return
	}

	stor, err := c.CreateStorage(locationID, "foobar", 10, &storTempl, nil)
	if err != nil {
		t.Errorf("Could not create storage: %s", err)
		return
	}
	defer c.DeleteStorage(stor.ID)

	for stor.Status == "in-provisioning" || stor.Status == "finalizing" {
		time.Sleep(time.Second)
		stor, err = c.GetStorage(stor.ID)
	}
	if stor.Status != "active" {
		t.Errorf("Unexpected storage status after creation: %s expected %s", stor.Status, "active")
	}

	ipv4, err := c.CreateIPv4(locationID, false, nil, nil)
	if err != nil {
		t.Errorf("Cannot create IPv4 address: %s\n", err)
		return
	}
	defer c.DeleteIP(ipv4.ID)

	srv, err := c.CreateServer(locationID, "foobar", 2, 4, nil)
	if err != nil {
		t.Errorf("Could not create server: %s", err)
		return
	}
	defer c.DeleteServer(srv.ID)

	err = c.ConnectIPAddress(ipv4.ID, srv.ID)
	if err != nil {
		t.Errorf("Could not connect IP to Server: %s", err)
	}

	err = c.ConnectStorage(stor.ID, true, srv.ID)
	if err != nil {
		t.Errorf("Could not connect storage to Server: %s", err)
	}

	err = c.ConnectNetwork(pubnet.ID, 0, srv.ID)
	if err != nil {
		t.Errorf("Could not connect public network to Server: %s", err)
	}

	err = c.PowerOnServer(srv.ID)
	if err != nil {
		t.Errorf("Could not turn on server: %s", err)
		return
	}

	time.Sleep(60 * time.Second)

	err = c.PowerOffServer(srv.ID)
	if err != nil {
		t.Errorf("Could not turn off server: %s", err)
		return
	}

	time.Sleep(10 * time.Second)

	err = c.DeleteServer(srv.ID)
	if err != nil {
		t.Errorf("Could not delete server: %s", err)
		return
	}
}
