package preview

type PreviewRequest struct {
	Size      string `json:"size"`
	Style     string `json:"style"`
	PartnerID string `json:"partner_id"`
	UserEmail string `json:"user_email"`
}

type PreviewResponse struct {
	PreviewID   string `json:"preview_id"`
	PreviewURL  string `json:"preview_url"`
	Size        string `json:"size"`
	Style       string `json:"style"`
	GeneratedAt string `json:"generated_at"`
}
