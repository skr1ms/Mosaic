package marketplace

import (
	"fmt"
	"strings"
)

// Service provides marketplace operations
type Service struct {
	partnerRepo PartnerRepository
	validator   *Validator
}

// NewService creates a new marketplace service
func NewService(partnerRepo PartnerRepository) *Service {
	return &Service{
		partnerRepo: partnerRepo,
		validator:   NewValidator(),
	}
}

// GenerateProductURL generates marketplace product URL for a partner's product
func (s *Service) GenerateProductURL(req *ProductURLRequest) (*ProductURLResponse, error) {
	// Validate request
	if err := s.validator.ValidateProductURLRequest(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Get partner
	partner, err := s.partnerRepo.GetByID(req.PartnerID)
	if err != nil {
		return nil, fmt.Errorf("partner not found: %w", err)
	}

	response := &ProductURLResponse{
		URL:         "",
		SKU:         "",
		HasArticle:  false,
		PartnerName: partner.GetBrandName(),
		Marketplace: req.Marketplace,
		Size:        req.Size,
		Style:       req.Style,
	}

	// Try to get specific article
	article, err := s.partnerRepo.GetArticleBySizeStyle(req.PartnerID, req.Size, req.Style, string(req.Marketplace))
	if err == nil && article != nil && article.GetSKU() != "" {
		// We have a specific article - generate URL
		response.SKU = article.GetSKU()
		response.HasArticle = true
		response.URL = s.generateURLFromArticle(partner, article, req.Size, req.Style, req.Marketplace)
	} else {
		// No specific article - use general marketplace link
		response.URL = s.getGeneralMarketplaceLink(partner, req.Marketplace)
		response.HasArticle = response.URL != ""
	}

	return response, nil
}

// CheckProductAvailability checks if a product is available on marketplace
func (s *Service) CheckProductAvailability(req *ProductAvailabilityRequest) (*ProductAvailabilityResponse, error) {
	// Validate request
	if err := s.validator.ValidateProductAvailabilityRequest(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	response := &ProductAvailabilityResponse{
		Marketplace: req.Marketplace,
		SKU:         req.SKU,
		Available:   false,
		ProductURL:  "",
	}

	// If we have partner info, try to get article
	if req.PartnerID != nil && req.Size != "" && req.Style != "" {
		partner, err := s.partnerRepo.GetByID(*req.PartnerID)
		if err == nil {
			article, err := s.partnerRepo.GetArticleBySizeStyle(*req.PartnerID, req.Size, req.Style, string(req.Marketplace))
			if err == nil && article != nil && article.GetSKU() != "" {
				response.SKU = article.GetSKU()
				response.Available = true
				response.ProductURL = s.generateURLFromArticle(partner, article, req.Size, req.Style, req.Marketplace)
			} else {
				// No specific article, try general link
				generalURL := s.getGeneralMarketplaceLink(partner, req.Marketplace)
				if generalURL != "" {
					response.ProductURL = generalURL
					response.Available = true
				}
			}
		}
	} else if req.SKU != "" {
		// If we only have SKU, generate direct link
		response.SKU = req.SKU
		response.Available = true
		response.ProductURL = s.generateDirectURL(req.SKU, req.Marketplace)
	}

	return response, nil
}

// generateURLFromArticle generates URL using article data and partner templates
func (s *Service) generateURLFromArticle(partner Partner, article Article, size, style string, marketplace Marketplace) string {
	sku := article.GetSKU()

	switch marketplace {
	case MarketplaceOzon:
		return s.generateOzonURL(partner, sku, size, style)
	case MarketplaceWildberries:
		return s.generateWildberriesURL(partner, sku, size, style)
	default:
		return ""
	}
}

// generateOzonURL generates URL for Ozon marketplace
func (s *Service) generateOzonURL(partner Partner, sku, size, style string) string {
	// Try to use partner template first
	if template := partner.GetOzonLinkTemplate(); template != "" {
		link := template
		link = strings.Replace(link, "{sku}", sku, -1)
		link = strings.Replace(link, "{size}", size, -1)
		link = strings.Replace(link, "{style}", style, -1)
		return link
	}

	// Use standard Ozon pattern
	return fmt.Sprintf("https://www.ozon.ru/product/%s", sku)
}

// generateWildberriesURL generates URL for Wildberries marketplace
func (s *Service) generateWildberriesURL(partner Partner, sku, size, style string) string {
	// Try to use partner template first
	if template := partner.GetWildberriesLinkTemplate(); template != "" {
		link := template
		link = strings.Replace(link, "{sku}", sku, -1)
		link = strings.Replace(link, "{size}", size, -1)
		link = strings.Replace(link, "{style}", style, -1)
		return link
	}

	// Use standard Wildberries pattern
	return fmt.Sprintf("https://www.wildberries.ru/catalog/%s/detail.aspx", sku)
}

// generateDirectURL generates direct URL from SKU without partner data
func (s *Service) generateDirectURL(sku string, marketplace Marketplace) string {
	switch marketplace {
	case MarketplaceOzon:
		return fmt.Sprintf("https://www.ozon.ru/product/%s", sku)
	case MarketplaceWildberries:
		return fmt.Sprintf("https://www.wildberries.ru/catalog/%s/detail.aspx", sku)
	default:
		return ""
	}
}

// getGeneralMarketplaceLink returns general marketplace link for partner
func (s *Service) getGeneralMarketplaceLink(partner Partner, marketplace Marketplace) string {
	switch marketplace {
	case MarketplaceOzon:
		return partner.GetOzonLink()
	case MarketplaceWildberries:
		return partner.GetWildberriesLink()
	default:
		return ""
	}
}

func (s *Service) GetValidator() *Validator {
	return s.validator
}

func (s *Service) GetSupportedMarketplaces() []Marketplace {
	return SupportedMarketplaces
}

func (s *Service) GetValidSizes() []string {
	return ValidSizes
}

func (s *Service) GetValidStyles() []string {
	return ValidStyles
}
