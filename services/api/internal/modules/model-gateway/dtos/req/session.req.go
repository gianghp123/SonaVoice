package req

type CloseSessionReq struct {
	SessionID   string `json:"session_id"`
	ActualUsage int    `json:"actual_usage"`
}
