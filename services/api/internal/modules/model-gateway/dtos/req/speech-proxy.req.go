package req

type StartConnectionReq struct {
	EnableDefaultIceServers bool                `json:"enable_default_ice_servers"`
	Body                    StartConnectionBody `json:"body"`
}

type StartConnectionBody struct {
	UserID      string `json:"user_id"`
	SessionID   string `json:"session_id"`
	MaxDuration int64  `json:"max_duration"`
}
