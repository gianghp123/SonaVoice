package res

type CreateSessionRes struct {
	ID                  string               `json:"id"`
	MaxDuration         int64                `json:"max_duration"`
	WebRTCConnectionRes *WebRTCConnectionRes `json:"webrtc_connection"`
}
