package partner

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/skr1ms/mosaic/internal/coupon"
	"github.com/skr1ms/mosaic/pkg/jwt"
)

type PartnerServiceDeps struct {
	PartnerRepository PartnerRepositoryInterface
	CouponRepository  CouponRepositoryInterface
	ImageRepository   ImageRepositoryInterface
	S3Client          S3ClientInterface
	Recaptcha         RecaptchaInterface
	JwtService        JWTInterface
	MailSender        MailerInterface
	Config            ConfigInterface
}

type PartnerService struct {
	deps *PartnerServiceDeps
}

func (s *PartnerService) PartnerLogin(login, password string) (*Partner, *jwt.TokenPair, error) {
	return nil, nil, fmt.Errorf("not implemented in PartnerService: use auth service")
}

func NewPartnerService(deps *PartnerServiceDeps) *PartnerService {
	return &PartnerService{
		deps: deps,
	}
}

// Repository access methods
func (s *PartnerService) GetPartnerRepository() PartnerRepositoryInterface {
	return s.deps.PartnerRepository
}

func (s *PartnerService) GetCouponRepository() CouponRepositoryInterface {
	return s.deps.CouponRepository
}

func (s *PartnerService) ExportCoupons(partnerID uuid.UUID, status string, format string) ([]byte, string, string, error) {
	coupons, err := s.deps.PartnerRepository.GetPartnerCouponsForExport(context.Background(), partnerID, status)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to fetch coupons: %w", err)
	}

	if len(coupons) == 0 {
		return nil, "", "", fmt.Errorf("no coupons found")
	}

	format = strings.ToLower(strings.TrimSpace(format))
	if format != "txt" && format != "csv" {
		format = "csv"
	}

	if format == "txt" {
		var b strings.Builder
		for _, c := range coupons {
			b.WriteString(fmt.Sprintf("Code: %s\n", c.CouponCode))
			b.WriteString(fmt.Sprintf("Partner Status: %s\n", c.PartnerStatus))
			b.WriteString(fmt.Sprintf("Coupon Status: %s\n", c.CouponStatus))
			b.WriteString(fmt.Sprintf("Size: %s\n", c.Size))
			b.WriteString(fmt.Sprintf("Style: %s\n", c.Style))
			b.WriteString(fmt.Sprintf("Created At: %s\n", c.CreatedAt.Format("2006-01-02 15:04:05")))
			b.WriteString("---\n")
		}
		filename := fmt.Sprintf("partner_coupons_%s.txt", time.Now().Format("20060102_150405"))
		return []byte(b.String()), filename, "text/plain; charset=utf-8", nil
	}

	var buf bytes.Buffer
	buf.Write([]byte{0xEF, 0xBB, 0xBF})
	buf.WriteString("sep=,\r\n")
	w := csv.NewWriter(&buf)
	w.Comma = ','
	w.UseCRLF = true
	_ = w.Write([]string{"Coupon Code", "Partner Status", "Coupon Status", "Size", "Style", "Created At"})
	for _, c := range coupons {
		_ = w.Write([]string{
			c.CouponCode,
			c.PartnerStatus,
			c.CouponStatus,
			c.Size,
			c.Style,
			c.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	w.Flush()
	filename := fmt.Sprintf("partner_coupons_%s.csv", time.Now().Format("20060102_150405"))
	return buf.Bytes(), filename, "text/csv; charset=utf-8", nil
}

// DeletePartnerWithCoupons deletes partner and all their coupons
func (s *PartnerService) DeletePartnerWithCoupons(ctx context.Context, partnerID uuid.UUID) error {
	err := s.deps.PartnerRepository.DeleteWithCoupons(ctx, partnerID)
	if err != nil {
		return fmt.Errorf("failed to delete partner: %w", err)
	}

	return nil
}

// GetComparisonStatistics returns comparison statistics with other partners
func (s *PartnerService) GetComparisonStatistics(ctx context.Context, partnerID uuid.UUID) (map[string]any, error) {

	if s.deps == nil || s.deps.CouponRepository == nil || s.deps.PartnerRepository == nil {
		return nil, fmt.Errorf("service dependencies not initialized")
	}

	topUsed, err := s.deps.CouponRepository.GetTopActivatedByPartner(ctx, 100)
	if err != nil {
		topUsed = []coupon.PartnerCount{}
	}

	topPurchased, err := s.deps.CouponRepository.GetTopPurchasedByPartner(ctx, 100)
	if err != nil {
		topPurchased = []coupon.PartnerCount{}
	}

	ensureEntry := func(list []coupon.PartnerCount, current coupon.PartnerCount) []coupon.PartnerCount {
		if list == nil {
			list = []coupon.PartnerCount{}
		}
		for _, e := range list {
			if e.PartnerID == current.PartnerID {
				return list
			}
		}
		return append(list, current)
	}

	partnerObj, err := s.deps.PartnerRepository.GetByID(ctx, partnerID)
	currentCode := ""
	currentBrand := ""
	if err == nil && partnerObj != nil {
		currentCode = partnerObj.PartnerCode
		currentBrand = partnerObj.BrandName
	}

	currentUsedCount, err := s.deps.CouponRepository.CountActivatedByPartner(ctx, partnerID)
	if err != nil {
		currentUsedCount = 0
	}

	currentPurchasedCount, err := s.deps.CouponRepository.CountBrandedPurchasesByPartner(ctx, partnerID)
	if err != nil {
		currentPurchasedCount = 0
	}

	currentUsed := coupon.PartnerCount{PartnerID: partnerID, PartnerCode: currentCode, BrandName: currentBrand, Count: currentUsedCount}

	currentPurchased := coupon.PartnerCount{PartnerID: partnerID, PartnerCode: currentCode, BrandName: currentBrand, Count: currentPurchasedCount}

	topUsed = ensureEntry(topUsed, currentUsed)
	topPurchased = ensureEntry(topPurchased, currentPurchased)

	type scored struct {
		item coupon.PartnerCount
		diff int64
	}
	filterClosest := func(list []coupon.PartnerCount, pivot int64) []coupon.PartnerCount {
		if list == nil {
			return []coupon.PartnerCount{}
		}

		var scoredList []scored
		for _, e := range list {
			d := e.Count - pivot
			if d < 0 {
				d = -d
			}
			if d <= 50 {
				scoredList = append(scoredList, scored{item: e, diff: d})
			}
		}

		if len(scoredList) == 0 {
			return []coupon.PartnerCount{}
		}

		sort.Slice(scoredList, func(i, j int) bool {
			if scoredList[i].diff == scoredList[j].diff {
				return scoredList[i].item.Count > scoredList[j].item.Count
			}
			return scoredList[i].diff < scoredList[j].diff
		})

		if len(scoredList) > 10 {
			scoredList = scoredList[:10]
		}

		out := make([]coupon.PartnerCount, 0, len(scoredList))
		for _, s := range scoredList {
			out = append(out, s.item)
		}
		return out
	}

	usedFiltered := filterClosest(topUsed, currentUsedCount)
	purchasedFiltered := filterClosest(topPurchased, currentPurchasedCount)

	return map[string]any{
		"used":      usedFiltered,
		"purchased": purchasedFiltered,
		"me": map[string]any{
			"partner_code": currentCode,
			"brand_name":   currentBrand,
		},
	}, nil
}

// DownloadCouponMaterials downloads materials of redeemed partner coupon from S3
func (s *PartnerService) DownloadCouponMaterials(id uuid.UUID) ([]byte, string, error) {
	coupon, err := s.deps.CouponRepository.GetByID(context.Background(), id)
	if err != nil {
		return nil, "", fmt.Errorf("coupon not found: %w", err)
	}

	if coupon.Status != "used" && coupon.Status != "completed" {
		return nil, "", fmt.Errorf("coupon must be used or completed to download materials")
	}

	imageRecord, err := s.deps.ImageRepository.GetByCouponID(context.Background(), id)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get image for coupon: %w", err)
	}

	if imageRecord.SchemaS3Key == nil {
		return nil, "", fmt.Errorf("no schema ZIP archive found for coupon")
	}

	reader, err := s.deps.S3Client.DownloadFile(context.Background(), *imageRecord.SchemaS3Key)
	if err != nil {
		return nil, "", fmt.Errorf("failed to download ZIP archive: %w", err)
	}
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read ZIP archive: %w", err)
	}

	filename := fmt.Sprintf("coupon_%s_materials.zip", coupon.Code)
	return data, filename, nil
}

// InitializeArticleGrid creates empty article grid for partner
func (s *PartnerService) InitializeArticleGrid(partnerID uuid.UUID) error {
	return s.deps.PartnerRepository.InitializeArticleGrid(context.Background(), partnerID)
}

// GetArticleGrid gets article grid for partner
func (s *PartnerService) GetArticleGrid(partnerID uuid.UUID) (map[string]map[string]map[string]string, error) {
	return s.deps.PartnerRepository.GetArticleGrid(context.Background(), partnerID)
}

// UpdateArticleSKU updates article in grid cell
func (s *PartnerService) UpdateArticleSKU(partnerID uuid.UUID, size, style, marketplace, sku string) error {
	return s.deps.PartnerRepository.UpdateArticleSKU(context.Background(), partnerID, size, style, marketplace, sku)
}

// GetArticleBySizeStyle gets article by size, style and marketplace
func (s *PartnerService) GetArticleBySizeStyle(partnerID uuid.UUID, size, style, marketplace string) (*PartnerArticle, error) {
	return s.deps.PartnerRepository.GetArticleBySizeStyle(context.Background(), partnerID, size, style, marketplace)
}

// GenerateProductLink generates a product link by size, style and marketplace
func (s *PartnerService) GenerateProductLink(partnerID uuid.UUID, size, style, marketplace string) string {
	// Get partner
	partner, err := s.deps.PartnerRepository.GetByID(context.Background(), partnerID)
	if err != nil {
		return ""
	}

	// Find article in grid
	article, err := s.deps.PartnerRepository.GetArticleBySizeStyle(context.Background(), partnerID, size, style, marketplace)
	if err != nil || article == nil || article.SKU == "" {
		// If article not found, return general link
		switch marketplace {
		case MarketplaceOzon:
			return partner.OzonLink
		case MarketplaceWildberries:
			return partner.WildberriesLink
		default:
			return ""
		}
	}

	var link string
	switch marketplace {
	case MarketplaceOzon:
		link = s.generateOzonProductLink(partner, article, size, style)
	case MarketplaceWildberries:
		link = s.generateWildberriesProductLink(partner, article, size, style)
	default:
		return ""
	}

	return link
}

// generateOzonProductLink generates link for Ozon
func (s *PartnerService) generateOzonProductLink(partner *Partner, article *PartnerArticle, size, style string) string {
	if partner.OzonLinkTemplate != "" {
		link := partner.OzonLinkTemplate
		link = strings.Replace(link, "{sku}", article.SKU, -1)
		link = strings.Replace(link, "{size}", size, -1)
		link = strings.Replace(link, "{style}", style, -1)
		return link
	}

	return fmt.Sprintf("https://www.ozon.ru/product/%s", article.SKU)
}

// generateWildberriesProductLink generates link for Wildberries
func (s *PartnerService) generateWildberriesProductLink(partner *Partner, article *PartnerArticle, size, style string) string {
	if partner.WildberriesLinkTemplate != "" {
		link := partner.WildberriesLinkTemplate
		link = strings.Replace(link, "{sku}", article.SKU, -1)
		link = strings.Replace(link, "{size}", size, -1)
		link = strings.Replace(link, "{style}", style, -1)
		return link
	}

	return fmt.Sprintf("https://www.wildberries.ru/catalog/%s/detail.aspx", article.SKU)
}

// ValidateURLTemplate validates URL template for marketplace
func ValidateURLTemplate(template, marketplace string) error {
	if template == "" {
		return fmt.Errorf("template cannot be empty")
	}

	if !strings.HasPrefix(template, "http://") && !strings.HasPrefix(template, "https://") {
		return fmt.Errorf("template must start with http:// or https://")
	}

	if !strings.Contains(template, "{sku}") {
		return fmt.Errorf("template must contain {sku} placeholder")
	}

	switch marketplace {
	case MarketplaceOzon:
		if !strings.Contains(template, "ozon.ru") {
			return fmt.Errorf("ozon template must contain 'ozon.ru' domain")
		}
	case MarketplaceWildberries:
		if !strings.Contains(template, "wildberries.ru") && !strings.Contains(template, "wb.ru") {
			return fmt.Errorf("wildberries template must contain 'wildberries.ru' or 'wb.ru' domain")
		}
	default:
		return fmt.Errorf("unsupported marketplace: %s", marketplace)
	}

	return nil
}

// GetDefaultURLTemplates returns standard templates for marketplaces
func GetDefaultURLTemplates() map[string]string {
	return map[string]string{
		MarketplaceOzon:        "https://www.ozon.ru/product/{sku}",
		MarketplaceWildberries: "https://www.wildberries.ru/catalog/{sku}/detail.aspx",
	}
}
