package payment

import (
	"context"

	"github.com/google/uuid"
	"github.com/skr1ms/mosaic/config"
	"github.com/skr1ms/mosaic/internal/coupon"
	"github.com/skr1ms/mosaic/internal/partner"
)

type PaymentRepositoryInterface interface {
	CreateOrder(ctx context.Context, order *Order) error
	GetOrderByNumber(ctx context.Context, orderNumber string) (*Order, error)
	GetOrderByID(ctx context.Context, id uuid.UUID) (*Order, error)
	UpdateOrderStatus(ctx context.Context, orderNumber string, status string, alfaBankOrderID *string) error
	UpdateOrderPaymentURL(ctx context.Context, orderNumber string, paymentURL string) error
	UpdateOrderCoupon(ctx context.Context, orderNumber string, couponID uuid.UUID) error
	GetOrdersByEmail(ctx context.Context, email string, limit int) ([]Order, error)
	GetOrdersByPartner(ctx context.Context, partnerID uuid.UUID, limit int) ([]Order, error)
	GetOrdersCountByStatus(ctx context.Context, status string) (int, error)
	GetOrdersCountByPartner(ctx context.Context, partnerID uuid.UUID, status string) (int, error)
}

type CouponRepositoryInterface interface {
	Create(ctx context.Context, coupon *coupon.Coupon) error
	GetByID(ctx context.Context, id uuid.UUID) (*coupon.Coupon, error)
	GetByCode(ctx context.Context, code string) (*coupon.Coupon, error)
	Update(ctx context.Context, coupon *coupon.Coupon) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetAll(ctx context.Context) ([]*coupon.Coupon, error)
	GetByPartnerID(ctx context.Context, partnerID uuid.UUID) ([]*coupon.Coupon, error)
	CodeExists(ctx context.Context, code string) (bool, error)
	FindAvailableCoupon(ctx context.Context, size, style string, partnerID *uuid.UUID) (*coupon.Coupon, error)
	MarkAsPurchased(ctx context.Context, id uuid.UUID, purchaseEmail string) error
}

type PartnerRepositoryInterface interface {
	Create(ctx context.Context, partner *partner.Partner) error
	GetByID(ctx context.Context, id uuid.UUID) (*partner.Partner, error)
	GetByDomain(ctx context.Context, domain string) (*partner.Partner, error)
	GetByPartnerCode(ctx context.Context, partnerCode string) (*partner.Partner, error)
	Update(ctx context.Context, partner *partner.Partner) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetAll(ctx context.Context, sortBy string, order string) ([]*partner.Partner, error)
}

type AlfaBankClientInterface interface {
	RegisterOrder(ctx context.Context, req *AlfaBankRegisterRequest) (*AlfaBankRegisterResponse, error)
	GetOrderStatus(ctx context.Context, orderID string) (*AlfaBankStatusResponse, error)
}

type PaymentServiceInterface interface {
	PurchaseCoupon(ctx context.Context, req *PurchaseCouponRequest) (*PurchaseCouponResponse, error)
	GetOrderStatus(ctx context.Context, orderNumber string) (*OrderStatusResponse, error)
	GenerateOrderNumber() string
	GetAvailableOptions() *AvailableOptionsResponse
	ProcessPaymentReturn(ctx context.Context, orderNumber string) error
	ProcessWebhookNotification(ctx context.Context, notification *PaymentNotificationRequest) error
	TestAlfaBankIntegration(ctx context.Context, req *AlfaBankRegisterRequest) (*TestIntegrationResponse, error)
}

type ConfigInterface interface {
	GetAlfaBankConfig() config.AlphaBankConfig
	GetServerConfig() config.ServerConfig
}

type RandomCouponCodeGeneratorInterface interface {
	GenerateUniqueCouponCode(partnerCode string, repo CouponRepositoryInterface) (string, error)
}

type EmailServiceInterface interface {
	SendCouponPurchaseEmail(to, couponCode, size, style string) error
}
