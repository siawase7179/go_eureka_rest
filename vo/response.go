package vo

type AuhResponse struct {
	Code   string `json:"code"`
	Result string `json:"result"`
}

type TokenResponse struct {
	Token  string `json:"token"`
	Expiry int    `json:"expiry"`
}

type AccountInfo struct {
	ClientId       string `json:"clientId"`
	ClientPassword string `json:"clientPassword"`
}
