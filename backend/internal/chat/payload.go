package chat

type wsEnvelope struct {
	Type string `json:"type"`
	Data any    `json:"data"`
}
