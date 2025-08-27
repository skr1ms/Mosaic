package chat

type wsEnvelope struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}
