package stats

import (
	"time"

	"github.com/google/uuid"
)

type GeneralStatsResponse struct {
	TotalCouponsCreated   int64   `json:"total_coupons_created"`
	TotalCouponsActivated int64   `json:"total_coupons_activated"`
	ActivationRate        float64 `json:"activation_rate"`
	ActivePartnersCount   int64   `json:"active_partners_count"`
	TotalPartnersCount    int64   `json:"total_partners_count"`
	LastUpdated           string  `json:"last_updated"`
}

type PartnerStatsResponse struct {
	PartnerID            uuid.UUID `json:"partner_id"`
	PartnerName          string    `json:"partner_name"`
	TotalCoupons         int64     `json:"total_coupons"`
	ActivatedCoupons     int64     `json:"activated_coupons"`
	UnusedCoupons        int64     `json:"unused_coupons"`
	BrandedSitePurchases int64     `json:"branded_site_purchases"`
	ActivationRate       float64   `json:"activation_rate"`
	LastActivity         *string   `json:"last_activity"`
}

type PartnerListStatsResponse struct {
	Partners []PartnerStatsResponse `json:"partners"`
	Total    int64                  `json:"total"`
}

type StatsFiltersRequest struct {
	PartnerID *uuid.UUID `query:"partner_id"`
	DateFrom  *string    `query:"date_from"`
	DateTo    *string    `query:"date_to"`
	Period    *string    `query:"period"`
}

type TimeSeriesStatsResponse struct {
	Period string                 `json:"period"`
	Data   []TimeSeriesStatsPoint `json:"data"`
}

type TimeSeriesStatsPoint struct {
	Date             string `json:"date"`
	CouponsCreated   int64  `json:"coupons_created"`
	CouponsActivated int64  `json:"coupons_activated"`
	CouponsPurchased int64  `json:"coupons_purchased"`
	NewPartnersCount int64  `json:"new_partners_count"`
}

type SystemHealthResponse struct {
	Status                string  `json:"status"` // healthy, warning, critical
	DatabaseStatus        string  `json:"database_status"`
	RedisStatus           string  `json:"redis_status"`
	ImageProcessingQueue  int64   `json:"image_processing_queue"`
	AverageProcessingTime float64 `json:"average_processing_time"`
	ErrorRate             float64 `json:"error_rate"`
	Uptime                string  `json:"uptime"`
	LastUpdated           string  `json:"last_updated"`
}

type RealTimeStatsResponse struct {
	Timestamp                time.Time `json:"timestamp"`
	ActiveUsers              int64     `json:"active_users"`
	CouponsActivatedLast5Min int64     `json:"coupons_activated_last_5_min"`
	ImagesProcessingNow      int64     `json:"images_processing_now"`
	SystemLoad               float64   `json:"system_load"`
}

type CouponsByStatusResponse struct {
	New       int64 `json:"new"`
	Activated int64 `json:"activated"`
	Used      int64 `json:"used"`
	Completed int64 `json:"completed"`
}

type CouponsBySizeResponse struct {
	Size21x30 int64 `json:"size_21x30"`
	Size30x40 int64 `json:"size_30x40"`
	Size40x40 int64 `json:"size_40x40"`
	Size40x50 int64 `json:"size_40x50"`
	Size40x60 int64 `json:"size_40x60"`
	Size50x70 int64 `json:"size_50x70"`
}

type CouponsByStyleResponse struct {
	Grayscale int64 `json:"grayscale"`
	SkinTones int64 `json:"skin_tones"`
	PopArt    int64 `json:"pop_art"`
	MaxColors int64 `json:"max_colors"`
}

type TopPartnersResponse struct {
	Partners []TopPartnerItem `json:"partners"`
}

type TopPartnerItem struct {
	PartnerID        uuid.UUID `json:"partner_id"`
	PartnerName      string    `json:"partner_name"`
	ActivatedCoupons int64     `json:"activated_coupons"`
	TotalCoupons     int64     `json:"total_coupons"`
	ActivationRate   float64   `json:"activation_rate"`
}
