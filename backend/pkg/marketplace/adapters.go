package marketplace

import (
	"context"

	"github.com/google/uuid"
	"github.com/skr1ms/mosaic/internal/partner"
)

type PartnerAdapter struct {
	partner *partner.Partner
}

func NewPartnerAdapter(p *partner.Partner) *PartnerAdapter {
	return &PartnerAdapter{partner: p}
}

func (p *PartnerAdapter) GetID() uuid.UUID {
	return p.partner.ID
}

func (p *PartnerAdapter) GetBrandName() string {
	return p.partner.BrandName
}

func (p *PartnerAdapter) GetOzonLink() string {
	return p.partner.OzonLink
}

func (p *PartnerAdapter) GetOzonLinkTemplate() string {
	return p.partner.OzonLinkTemplate
}

func (p *PartnerAdapter) GetWildberriesLink() string {
	return p.partner.WildberriesLink
}

func (p *PartnerAdapter) GetWildberriesLinkTemplate() string {
	return p.partner.WildberriesLinkTemplate
}

type ArticleAdapter struct {
	article *partner.PartnerArticle
}

func NewArticleAdapter(a *partner.PartnerArticle) *ArticleAdapter {
	return &ArticleAdapter{article: a}
}

func (a *ArticleAdapter) GetSKU() string {
	return a.article.SKU
}

func (a *ArticleAdapter) GetSize() string {
	return a.article.Size
}

func (a *ArticleAdapter) GetStyle() string {
	return a.article.Style
}

func (a *ArticleAdapter) GetMarketplace() string {
	return a.article.Marketplace
}

type PartnerRepositoryInterface interface {
	GetByID(ctx context.Context, id uuid.UUID) (*partner.Partner, error)
	GetArticleBySizeStyle(ctx context.Context, partnerID uuid.UUID, size, style, marketplace string) (*partner.PartnerArticle, error)
}

type PartnerRepositoryAdapter struct {
	repo PartnerRepositoryInterface
}

func NewPartnerRepositoryAdapter(repo PartnerRepositoryInterface) *PartnerRepositoryAdapter {
	return &PartnerRepositoryAdapter{repo: repo}
}

func (r *PartnerRepositoryAdapter) GetByID(partnerID uuid.UUID) (Partner, error) {
	p, err := r.repo.GetByID(context.Background(), partnerID)
	if err != nil {
		return nil, err
	}
	return NewPartnerAdapter(p), nil
}

func (r *PartnerRepositoryAdapter) GetArticleBySizeStyle(partnerID uuid.UUID, size, style, marketplace string) (Article, error) {
	a, err := r.repo.GetArticleBySizeStyle(context.Background(), partnerID, size, style, marketplace)
	if err != nil {
		return nil, err
	}
	return NewArticleAdapter(a), nil
}
