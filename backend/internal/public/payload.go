package public

import (
	"github.com/google/uuid"
)

type SendEmailRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type PurchaseCouponRequest struct {
	Size         string  `json:"size" validate:"required,oneof=21x30 30x40 40x40 40x50 40x60 50x70"`
	Style        string  `json:"style" validate:"required,oneof=grayscale skin_tones pop_art max_colors"`
	Email        string  `json:"email" validate:"required,email"`
	PaymentToken string  `json:"payment_token" validate:"required"`
	Amount       float64 `json:"amount" validate:"required"`
}

type CouponInfoResponse struct {
	ID     uuid.UUID `json:"id"`
	Code   string    `json:"code"`
	Size   string    `json:"size"`
	Style  string    `json:"style"`
	Status string    `json:"status"`
	Valid  bool      `json:"valid"`
}

type ImageUploadResponse struct {
	Message  string    `json:"message"`
	ImageID  uuid.UUID `json:"image_id"`
	NextStep string    `json:"next_step"`
}

type ProcessingStatusResponse struct {
	ID          uuid.UUID `json:"id"`
	Status      string    `json:"status"`
	Progress    int       `json:"progress"`
	Error       *string   `json:"error,omitempty"`
	StartedAt   *string   `json:"started_at,omitempty"`
	CompletedAt *string   `json:"completed_at,omitempty"`
}

type PartnerInfoResponse struct {
	BrandName       string `json:"brand_name"`
	Domain          string `json:"domain"`
	LogoURL         string `json:"logo_url"`
	OzonLink        string `json:"ozon_link"`
	WildberriesLink string `json:"wildberries_link"`
	Email           string `json:"email"`
	Phone           string `json:"phone"`
	Address         string `json:"address"`
	TelegramLink    string `json:"telegram_link"`
	WhatsappLink    string `json:"whatsapp_link"`
	AllowPurchases  bool   `json:"allow_purchases"`
}

type SizeInfo struct {
	Size  string `json:"size"`
	Title string `json:"title"`
	Price int    `json:"price"`
}

type StyleInfo struct {
	Style       string `json:"style"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

type GenerateProductURLRequest struct {
	Marketplace string `json:"marketplace" validate:"required,oneof=ozon wildberries"`
	Style       string `json:"style" validate:"required,oneof=grayscale skin_tones pop_art max_colors"`
	Size        string `json:"size" validate:"required,oneof='21x30' '30x40' '40x40' '40x50' '40x60' '50x70'"`
}

type GenerateProductURLResponse struct {
	URL         string `json:"url"`
	SKU         string `json:"sku"`
	HasArticle  bool   `json:"has_article"`
	PartnerName string `json:"partner_name"`
	Marketplace string `json:"marketplace"`
	Size        string `json:"size"`
	Style       string `json:"style"`
}

type GenerateAllPreviewsRequest struct {
	ImageID string `json:"image_id" validate:"required,uuid"`
	Size    string `json:"size" validate:"required,oneof=21x30 30x40 40x40 40x50 40x60 50x70"`
	UseAI   bool   `json:"use_ai"`
}

type GenerateAllPreviewsResponse struct {
	Previews []PreviewInfo `json:"previews"`
	Total    int           `json:"total"`
	ImageID  string        `json:"image_id"`
}

type PreviewInfo struct {
	ID       string `json:"id"`
	URL      string `json:"url"`
	Style    string `json:"style"`
	Contrast string `json:"contrast"`
	Label    string `json:"label"`
	IsAI     bool   `json:"is_ai"`
}

type SearchSchemaPageRequest struct {
	ImageID    string `json:"image_id" validate:"required,uuid"`
	PageNumber int    `json:"page_number" validate:"required,min=1"`
}

type SearchSchemaPageResponse struct {
	PageURL    string `json:"page_url"`
	PageNumber int    `json:"page_number"`
	TotalPages int    `json:"total_pages"`
	Found      bool   `json:"found"`
}

type ReactivateCouponRequest struct {
	Code  string `json:"code" validate:"required,len=12"`
	Email string `json:"email,omitempty" validate:"omitempty,email"`
}

type ReactivateCouponResponse struct {
	ActivatedAt  string `json:"activated_at"`
	PreviewURL   string `json:"preview_url"`
	StonesCount  int    `json:"stones_count"`
	ArchiveURL   string `json:"archive_url"`
	CanDownload  bool   `json:"can_download"`
	CanSendEmail bool   `json:"can_send_email"`
	PageCount    int    `json:"page_count"`
}

type MarketplaceStatusRequest struct {
	Marketplace string `json:"marketplace" validate:"required,oneof=ozon wildberries"`
	SKU         string `json:"sku" validate:"required"`
}

type MarketplaceStatusResponse struct {
	Available   bool   `json:"available"`
	Marketplace string `json:"marketplace"`
	SKU         string `json:"sku"`
	ProductURL  string `json:"product_url"`
}

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