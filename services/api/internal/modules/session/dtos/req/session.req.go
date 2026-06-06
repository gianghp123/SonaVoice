package req

type FinalizeSessionReq struct {
	SessionID   string `json:"session_id"`
	ActualUsage int    `json:"actual_usage"`
}
