package req

type StartConnectionReq struct {
	EnableDefaultIceServers bool                `json:"enableDefaultIceServers"`
	Body                    StartConnectionBody `json:"body"`
}

type StartConnectionBody struct {
	UserID      string `json:"user_id"`
	SessionID   string `json:"session_id"`
	MaxDuration int64  `json:"max_duration"`
}
