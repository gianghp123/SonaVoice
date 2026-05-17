package res

type CreateSessionRes struct {
	ID                  string               `json:"id"`
	WebRTCConnectionRes *WebRTCConnectionRes `json:"webrtc_connection"`
}
