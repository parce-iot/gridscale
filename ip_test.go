package gridscale

import (
	"os"
	"testing"
)

func TestCreateRenameDeleteIPv4(t *testing.T) {
	userID := os.Getenv("GRIDSCALE_USERID")
	apiToken := os.Getenv("GRIDSCALE_APITOKEN")
	endpoint := "https://api.gridscale.io"
	c, _ := NewClient(userID, apiToken, endpoint)

	newIP, err := c.CreateIPv4("45ed677b-3702-4b36-be2a-a2eab9827950", true, nil, nil)
	if err != nil {
		t.Errorf("Cannot create IPv4 address: %s\n", err)
		return
	}
	defer c.DeleteIP(newIP.ID)

	if newIP.Failover != true {
		t.Errorf("Created IP is not a Failover IP as requested")
		return
	}

	ips, err := c.GetIPs()
	if err != nil {
		t.Errorf("Cannot read IP list: %s\n", err)
		return
	}

	found := false
	for _, ip := range ips {
		if ip.IP.String() == newIP.IP.String() {
			found = true
		}
	}
	if !found {
		t.Errorf("Could not find a newly created IP address %s in returned IP address list!", newIP.IP)
		return
	}

	err = c.DeleteIP(newIP.ID)
	if err != nil {
		t.Errorf("Cannot delete IP address %s: %s", newIP.IP, err)
		return
	}

	ips, err = c.GetIPs()
	if err != nil {
		t.Errorf("Cannot read IP list: %s\n", err)
		return
	}

	for _, ip := range ips {
		if ip.IP.String() == newIP.IP.String() {
			t.Errorf("IP address %s should have been deleted but is still in the list!", newIP.IP)
			return
		}
	}
}

func TestCreateRenameDeleteIPv6(t *testing.T) {
	userID := os.Getenv("GRIDSCALE_USERID")
	apiToken := os.Getenv("GRIDSCALE_APITOKEN")
	endpoint := "https://api.gridscale.io"
	c, _ := NewClient(userID, apiToken, endpoint)

	newIP, err := c.CreateIPv6("45ed677b-3702-4b36-be2a-a2eab9827950", true, nil, nil)
	if err != nil {
		t.Errorf("Cannot create IPv4 address: %s\n", err)
		return
	}

	if newIP.Failover != true {
		t.Errorf("Created IP is not a Failover IP as requested")
		return
	}

	defer c.DeleteIP(newIP.ID)

	ips, err := c.GetIPs()
	if err != nil {
		t.Errorf("Cannot read IP list: %s\n", err)
		return
	}

	found := false
	for _, ip := range ips {
		if ip.IP.String() == newIP.IP.String() {
			found = true
		}
	}
	if !found {
		t.Errorf("Could not find a newly created IP address %s in returned IP address list!", newIP.IP)
		return
	}

	err = c.DeleteIP(newIP.ID)
	if err != nil {
		t.Errorf("Cannot delete IP address %s: %s", newIP.IP, err)
		return
	}

	ips, err = c.GetIPs()
	if err != nil {
		t.Errorf("Cannot read IP list: %s\n", err)
		return
	}

	for _, ip := range ips {
		if ip.IP.String() == newIP.IP.String() {
			t.Errorf("IP address %s should have been deleted but is still in the list!", newIP.IP)
			return
		}
	}
}
