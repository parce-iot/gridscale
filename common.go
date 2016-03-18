package gridscale

type objectWrapper struct {
	IPs *map[string]restIP `json:"ips,omitempty"`
	IP  *restIP            `json:"ip,omitempty"`

	Networks *map[string]restNetwork `json:"networks,omitempty"`
	Network  *restNetwork            `json:"network,omitempty"`

	Storages *map[string]restStorage `json:"storages,omitempty"`
	Storage  *restStorage            `json:"storage,omitempty"`

	SSHKeys *map[string]SSHKey `json:"sshkeys,omitempty"`
	SSHKey  *SSHKey            `json:"sshkey,omitempty"`

	Servers *map[string]Server `json:"servers,omitempty"`
	Server  *Server            `json:"server,omitempty"`

	Templates *map[string]Template `json:"templates,omitempty"`
}
