package res

type WebRTCConnectionRes struct {
	MaxDuration int64      `json:"maxDuration"`
	SessionID   string     `json:"sessionId"`
	IceConfig   *IceConfig `json:"iceConfig,omitempty"` // only if enableDefaultIceServers: true
}

type IceConfig struct {
	IceServers []IceServer `json:"iceServers"`
}

type IceServer struct {
	Urls       []string `json:"urls"`
	Username   string   `json:"username,omitempty"`   // for TURN servers
	Credential string   `json:"credential,omitempty"` // for TURN servers
}
