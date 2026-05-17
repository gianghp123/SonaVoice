package res

type WebRTCConnectionRes struct {
	SessionID string     `json:"session_id"`
	IceConfig *IceConfig `json:"ice_config,omitempty"` // only if enableDefaultIceServers: true
}

type IceConfig struct {
	IceServers []IceServer `json:"ice_servers"`
}

type IceServer struct {
	Urls       []string `json:"urls"`
	Username   string   `json:"username,omitempty"`   // for TURN servers
	Credential string   `json:"credential,omitempty"` // for TURN servers
}
