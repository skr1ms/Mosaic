package payment

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type PaymentRepository struct {
	db *bun.DB
}

var _ PaymentRepositoryInterface = (*PaymentRepository)(nil)

func NewPaymentRepository(db *bun.DB) *PaymentRepository {
	return &PaymentRepository{db: db}
}

// CreateOrder создает новый заказ
func (r *PaymentRepository) CreateOrder(ctx context.Context, order *Order) error {
	order.CreatedAt = time.Now()
	order.UpdatedAt = time.Now()
	_, err := r.db.NewInsert().Model(order).Exec(ctx)
	if err != nil {
		return fmt.Errorf("error creating order: %w", err)
	}
	return nil
}

// GetOrderByNumber возвращает заказ по номеру
func (r *PaymentRepository) GetOrderByNumber(ctx context.Context, orderNumber string) (*Order, error) {
	order := &Order{}
	err := r.db.NewSelect().
		Model(order).
		Where("order_number = ?", orderNumber).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return order, nil
}

// GetOrderByID возвращает заказ по ID
func (r *PaymentRepository) GetOrderByID(ctx context.Context, id uuid.UUID) (*Order, error) {
	order := &Order{}
	err := r.db.NewSelect().
		Model(order).
		Where("id = ?", id).
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting order by number: %w", err)
	}
	return order, nil
}

// UpdateOrderStatus обновляет статус заказа
func (r *PaymentRepository) UpdateOrderStatus(ctx context.Context, orderNumber string, status string, alfaBankOrderID *string) error {
	query := r.db.NewUpdate().
		Model((*Order)(nil)).
		Set("status = ?", status).
		Set("updated_at = ?", time.Now()).
		Where("order_number = ?", orderNumber)

	if alfaBankOrderID != nil {
		query = query.Set("alfabank_order_id = ?", *alfaBankOrderID)
	}

	if status == OrderStatusPaid {
		query = query.Set("paid_at = ?", time.Now())
	}

	_, err := query.Exec(ctx)
	if err != nil {
		return fmt.Errorf("error updating order status: %w", err)
	}
	return nil
}

// UpdateOrderPaymentURL обновляет URL платежа для заказа
func (r *PaymentRepository) UpdateOrderPaymentURL(ctx context.Context, orderNumber string, paymentURL string) error {
	_, err := r.db.NewUpdate().
		Model((*Order)(nil)).
		Set("payment_url = ?", paymentURL).
		Set("updated_at = ?", time.Now()).
		Where("order_number = ?", orderNumber).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("error updating order payment URL: %w", err)
	}
	return nil
}

// UpdateOrderCoupon обновляет купон заказа
func (r *PaymentRepository) UpdateOrderCoupon(ctx context.Context, orderNumber string, couponID uuid.UUID) error {
	_, err := r.db.NewUpdate().
		Model((*Order)(nil)).
		Set("coupon_id = ?", couponID).
		Set("updated_at = ?", time.Now()).
		Where("order_number = ?", orderNumber).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("error updating order coupon: %w", err)
	}
	return nil
}

// GetOrdersByEmail возвращает список заказов по email пользователя
func (r *PaymentRepository) GetOrdersByEmail(ctx context.Context, email string, limit int) ([]Order, error) {
	var orders []Order
	err := r.db.NewSelect().
		Model(&orders).
		Where("user_email = ?", email).
		Order("created_at DESC").
		Limit(limit).
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting orders by email: %w", err)
	}
	return orders, nil
}

// GetOrdersByPartner возвращает список заказов по ID партнера
func (r *PaymentRepository) GetOrdersByPartner(ctx context.Context, partnerID uuid.UUID, limit int) ([]Order, error) {
	var orders []Order
	err := r.db.NewSelect().
		Model(&orders).
		Where("partner_id = ?", partnerID).
		Order("created_at DESC").
		Limit(limit).
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting orders by partner: %w", err)
	}
	return orders, nil
}

// GetOrdersCountByStatus возвращает количество заказов по статусу
func (r *PaymentRepository) GetOrdersCountByStatus(ctx context.Context, status string) (int, error) {
	count, err := r.db.NewSelect().
		Model((*Order)(nil)).
		Where("status = ?", status).
		Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("error getting orders count by status: %w", err)
	}
	return count, nil
}

// GetOrdersCountByPartner возвращает количество заказов по ID партнера и статусу
func (r *PaymentRepository) GetOrdersCountByPartner(ctx context.Context, partnerID uuid.UUID, status string) (int, error) {
	query := r.db.NewSelect().
		Model((*Order)(nil)).
		Where("partner_id = ?", partnerID)

	if status != "" {
		query = query.Where("status = ?", status)
	}

	count, err := query.Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("error getting orders count by partner: %w", err)
	}
	return count, nil
}
