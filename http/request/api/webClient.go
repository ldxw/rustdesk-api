package api

// SharedPeerRequest webclient 分享访问请求
type SharedPeerRequest struct {
	ShareToken string `json:"share_token" validate:"required"`
}
