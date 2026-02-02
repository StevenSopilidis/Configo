package http_server

type AddVoterRequest struct {
	ID   string `json:"id"`
	Addr string `json:"addr"`
}

type AddVoterResponse struct {
	Addr string `json:"addr"`
}
