package main

import (
	"context"

	"github.com/google/uuid"
	"github.com/skr1ms/mosaic/internal/coupon"
	"github.com/skr1ms/mosaic/internal/image"
	"github.com/skr1ms/mosaic/internal/partner"
	"github.com/skr1ms/mosaic/internal/payment"
	"github.com/skr1ms/mosaic/pkg/randomCouponCode"
)

type PartnerRepositoryAdapter struct {
	partnerRepo *partner.PartnerRepository
}

func NewPartnerRepositoryAdapter(partnerRepo *partner.PartnerRepository) *PartnerRepositoryAdapter {
	return &PartnerRepositoryAdapter{
		partnerRepo: partnerRepo,
	}
}

func (a *PartnerRepositoryAdapter) GetByID(ctx context.Context, id uuid.UUID) (*coupon.Partner, error) {
	p, err := a.partnerRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &coupon.Partner{
		ID:          p.ID,
		PartnerCode: p.PartnerCode,
		Domain:      p.Domain,
		BrandName:   p.BrandName,
	}, nil
}

func (a *PartnerRepositoryAdapter) GetByPartnerCode(ctx context.Context, code string) (*coupon.Partner, error) {
	p, err := a.partnerRepo.GetByPartnerCode(ctx, code)
	if err != nil {
		return nil, err
	}
	return &coupon.Partner{
		ID:          p.ID,
		PartnerCode: p.PartnerCode,
		Domain:      p.Domain,
		BrandName:   p.BrandName,
	}, nil
}

type CouponRepositoryAdapter struct {
	couponRepo *coupon.CouponRepository
}

func NewCouponRepositoryAdapter(couponRepo *coupon.CouponRepository) *CouponRepositoryAdapter {
	return &CouponRepositoryAdapter{
		couponRepo: couponRepo,
	}
}

func (a *CouponRepositoryAdapter) GetByID(ctx context.Context, id uuid.UUID) (*image.Coupon, error) {
	c, err := a.couponRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return &image.Coupon{
		ID:          c.ID,
		Code:        c.Code,
		Size:        string(c.Size),
		Style:       string(c.Style),
		Status:      c.Status,
		UserEmail:   c.UserEmail,
		CompletedAt: c.CompletedAt,
		StonesCount: c.StonesCount,
	}, nil
}

func (a *CouponRepositoryAdapter) GetByCode(ctx context.Context, code string) (*image.Coupon, error) {
	c, err := a.couponRepo.GetByCode(ctx, code)
	if err != nil {
		return nil, err
	}

	return &image.Coupon{
		ID:          c.ID,
		Code:        c.Code,
		Size:        string(c.Size),
		Style:       string(c.Style),
		Status:      c.Status,
		UserEmail:   c.UserEmail,
		CompletedAt: c.CompletedAt,
		StonesCount: c.StonesCount,
	}, nil
}

func (a *CouponRepositoryAdapter) Update(ctx context.Context, imgCoupon *image.Coupon) error {
	c := &coupon.Coupon{
		ID:          imgCoupon.ID,
		Code:        imgCoupon.Code,
		Size:        imgCoupon.Size,
		Style:       imgCoupon.Style,
		Status:      imgCoupon.Status,
		StonesCount: imgCoupon.StonesCount,
	}
	return a.couponRepo.Update(ctx, c)
}

type RandomCouponCodeGeneratorImpl struct{}

func (r *RandomCouponCodeGeneratorImpl) GenerateUniqueCouponCode(partnerCode string, repo payment.CouponRepositoryInterface) (string, error) {
	return randomCouponCode.GenerateUniqueCouponCode(partnerCode, &RandomCouponRepoAdapter{repo: repo})
}

type RandomCouponRepoAdapter struct {
	repo payment.CouponRepositoryInterface
}

func (a *RandomCouponRepoAdapter) CodeExists(ctx context.Context, code string) (bool, error) {
	return a.repo.CodeExists(ctx, code)
}
