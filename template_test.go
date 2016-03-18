package gridscale

import (
	"os"
	"strings"
	"testing"
)

func TestGetTemplateByName(t *testing.T) {
	userID := os.Getenv("GRIDSCALE_USERID")
	apiToken := os.Getenv("GRIDSCALE_APITOKEN")
	endpoint := "https://api.gridscale.io"
	c, _ := NewClient(userID, apiToken, endpoint)

	temp, err := c.GetTemplateByName("Debian")
	if err != nil {
		t.Errorf("Could not get Template by Name 'Debian': %s", err)
		return
	}

	if !strings.HasPrefix(temp.Name, "Debian") {
		t.Errorf("Template name '%s' does not start with 'Debian'", temp.Name)
		return
	}
}
