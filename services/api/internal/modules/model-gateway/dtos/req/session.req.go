package req

type CloseSessionReq struct {
	SessionID   string `json:"sessionId"`
	ActualUsage int    `json:"actualUsage"`
	MaxDuration int    `json:"maxDuration"`
}
