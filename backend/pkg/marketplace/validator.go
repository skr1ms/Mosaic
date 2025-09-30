package marketplace

import (
	"fmt"
	"strings"
)

// Validator provides validation functions for marketplace operations
type Validator struct{}

// NewValidator creates a new marketplace validator
func NewValidator() *Validator {
	return &Validator{}
}

// ValidateMarketplace validates if the marketplace is supported
func (v *Validator) ValidateMarketplace(marketplace string) error {
	if marketplace == "" {
		return fmt.Errorf("marketplace is required")
	}

	marketplace = strings.ToLower(marketplace)

	for _, supported := range SupportedMarketplaces {
		if string(supported) == marketplace {
			return nil
		}
	}

	return fmt.Errorf("invalid marketplace '%s'. Supported marketplaces: %v", marketplace, SupportedMarketplaces)
}

// ValidateSize validates if the product size is supported
func (v *Validator) ValidateSize(size string) error {
	if size == "" {
		return nil // Size is optional in some contexts
	}

	for _, validSize := range ValidSizes {
		if validSize == size {
			return nil
		}
	}

	return fmt.Errorf("invalid size '%s'. Valid sizes: %v", size, ValidSizes)
}

// ValidateStyle validates if the product style is supported
func (v *Validator) ValidateStyle(style string) error {
	if style == "" {
		return nil // Style is optional in some contexts
	}

	for _, validStyle := range ValidStyles {
		if validStyle == style {
			return nil
		}
	}

	return fmt.Errorf("invalid style '%s'. Valid styles: %v", style, ValidStyles)
}

// ValidateProductURLRequest validates a product URL generation request
func (v *Validator) ValidateProductURLRequest(req *ProductURLRequest) error {
	if req.PartnerID == [16]byte{} {
		return fmt.Errorf("partner_id is required")
	}

	if err := v.ValidateMarketplace(string(req.Marketplace)); err != nil {
		return err
	}

	if req.Size == "" {
		return fmt.Errorf("size is required")
	}
	if err := v.ValidateSize(req.Size); err != nil {
		return err
	}

	if req.Style == "" {
		return fmt.Errorf("style is required")
	}
	if err := v.ValidateStyle(req.Style); err != nil {
		return err
	}

	return nil
}

// ValidateProductAvailabilityRequest validates a product availability check request
func (v *Validator) ValidateProductAvailabilityRequest(req *ProductAvailabilityRequest) error {
	if err := v.ValidateMarketplace(string(req.Marketplace)); err != nil {
		return err
	}

	if req.Size != "" {
		if err := v.ValidateSize(req.Size); err != nil {
			return err
		}
	}

	if req.Style != "" {
		if err := v.ValidateStyle(req.Style); err != nil {
			return err
		}
	}

	// At least one identifier should be provided
	if req.PartnerID == nil && req.SKU == "" {
		return fmt.Errorf("either partner_id with size/style or sku must be provided")
	}

	return nil
}

// IsMarketplaceSupported checks if marketplace is supported
func (v *Validator) IsMarketplaceSupported(marketplace string) bool {
	err := v.ValidateMarketplace(marketplace)
	return err == nil
}

// IsSizeValid checks if size is valid
func (v *Validator) IsSizeValid(size string) bool {
	err := v.ValidateSize(size)
	return err == nil
}

// IsStyleValid checks if style is valid
func (v *Validator) IsStyleValid(style string) bool {
	err := v.ValidateStyle(style)
	return err == nil
}
