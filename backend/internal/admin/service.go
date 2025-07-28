package admin

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/skr1ms/mosaic/internal/coupon"
	"github.com/skr1ms/mosaic/internal/image"
	"github.com/skr1ms/mosaic/internal/partner"
	"github.com/skr1ms/mosaic/pkg/bcrypt"
	"github.com/skr1ms/mosaic/pkg/randomCouponCode"
	"github.com/skr1ms/mosaic/pkg/updatePartnerData"
)

// AdminServiceDeps содержит зависимости для AdminService
type AdminServiceDeps struct {
	AdminRepository   *AdminRepository
	PartnerRepository *partner.PartnerRepository
	CouponRepository  *coupon.CouponRepository
	ImageRepository   *image.ImageRepository
}

// AdminService содержит бизнес-логику для админской части
type AdminService struct {
	deps *AdminServiceDeps
}

// NewAdminService создает новый экземпляр AdminService
func NewAdminService(deps *AdminServiceDeps) *AdminService {
	return &AdminService{
		deps: deps,
	}
}

// processSocialLinks обрабатывает ссылки на социальные сети
func (s *AdminService) processSocialLinks(telegram, whatsapp string) (string, string) {
	telegramLink := ""
	whatsappLink := ""

	if telegram != "" {
		if strings.HasPrefix(telegram, "https://t.me/") {
			telegramLink = telegram
		} else {
			telegramLink = "https://t.me/" + telegram
		}
	}

	if whatsapp != "" {
		if strings.HasPrefix(whatsapp, "https://wa.me/") {
			whatsappLink = whatsapp
		} else {
			whatsappLink = "https://wa.me/" + whatsapp
		}
	}

	return telegramLink, whatsappLink
}

// CreateAdmin создает нового администратора
func (s *AdminService) CreateAdmin(req CreateAdminRequest) (*Admin, error) {
	log := zerolog.Ctx(context.Background())
	// Проверяем, существует ли администратор с таким логином
	existingAdmin, err := s.deps.AdminRepository.GetByLogin(req.Login)
	if err == nil && existingAdmin != nil {
		log.Error().Str("login", req.Login).Msg("Admin already exists")
		return nil, fmt.Errorf("admin already exists: %w", err)
	}

	// Хешируем пароль
	hashedPassword, err := bcrypt.HashPassword(req.Password)
	if err != nil {
		log.Error().Err(err).Msg("Failed to hash password")
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Создаем администратора
	admin := &Admin{
		Login:    req.Login,
		Password: hashedPassword,
	}

	if err := s.deps.AdminRepository.Create(admin); err != nil {
		log.Error().Err(err).Msg("Failed to create admin")
		return nil, fmt.Errorf("failed to create admin: %w", err)
	}

	return admin, nil
}

// GetAdmins возвращает список всех администраторов
func (s *AdminService) GetAdmins() ([]*Admin, error) {
	log := zerolog.Ctx(context.Background())
	admins, err := s.deps.AdminRepository.GetAll()
	if err != nil {
		log.Error().Err(err).Msg("Failed to find all admins")
		return nil, fmt.Errorf("failed to find all admins: %w", err)
	}
	return admins, nil
}

// ChangePassword изменяет пароль администратора
func (s *AdminService) ChangePassword(adminID uuid.UUID, req ChangePasswordRequest) error {
	log := zerolog.Ctx(context.Background())
	// Получаем администратора
	admin, err := s.deps.AdminRepository.GetByID(adminID)
	if err != nil {
		log.Error().Err(err).Str("admin_id", adminID.String()).Msg("Failed to find admin by ID")
		return fmt.Errorf("failed to find admin by ID: %w", err)
	}

	// Проверяем текущий пароль
	if !bcrypt.CheckPassword(req.CurrentPassword, admin.Password) {
		log.Error().Str("admin_id", adminID.String()).Msg("Invalid password")
		return fmt.Errorf("invalid password")
	}

	// Хешируем новый пароль
	hashedPassword, err := bcrypt.HashPassword(req.NewPassword)
	if err != nil {
		log.Error().Err(err).Str("admin_id", adminID.String()).Msg("Failed to hash password")
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Обновляем пароль
	if err := s.deps.AdminRepository.UpdatePassword(adminID, hashedPassword); err != nil {
		log.Error().Err(err).Str("admin_id", adminID.String()).Msg("Failed to change password")
		return fmt.Errorf("failed to change password: %w", err)
	}

	return nil
}

// GetDashboardData возвращает данные для дашборда администратора
func (s *AdminService) GetDashboardData() (map[string]interface{}, error) {
	log := zerolog.Ctx(context.Background())
	result := make(map[string]interface{})

	// Общая статистика по купонам
	allCoupons, err := s.deps.CouponRepository.GetAll(context.Background())
	if err != nil {
		log.Error().Err(err).Msg("Failed to find all coupons")
		return nil, fmt.Errorf("failed to find all coupons: %w", err)
	}

	// Подсчитываем статистику купонов
	totalCoupons := len(allCoupons)
	usedCoupons := 0
	purchasedCoupons := 0
	for _, c := range allCoupons {
		if c.Status == "used" {
			usedCoupons++
		}
		if c.IsPurchased {
			purchasedCoupons++
		}
	}

	result["coupons"] = map[string]int{
		"total":     totalCoupons,
		"used":      usedCoupons,
		"purchased": purchasedCoupons,
		"new":       totalCoupons - usedCoupons,
	}

	// Статистика по партнерам
	allPartners, err := s.deps.PartnerRepository.GetAll(context.Background())
	if err != nil {
		log.Error().Err(err).Msg("Failed to find all partners")
		return nil, fmt.Errorf("failed to find all partners: %w", err)
	}

	activePartners := 0
	for _, p := range allPartners {
		if p.Status == "active" {
			activePartners++
		}
	}

	result["partners"] = map[string]int{
		"total":  len(allPartners),
		"active": activePartners,
	}

	// Последние активированные купоны
	recentCoupons, err := s.deps.CouponRepository.GetRecentActivated(context.Background(), 10)
	if err == nil {
		result["recent_activations"] = recentCoupons
	}

	// Статистика по задачам обработки изображений
	allImages, err := s.deps.ImageRepository.GetAll(context.Background())
	if err == nil {
		processingImages := 0
		completedImages := 0
		failedImages := 0

		for _, img := range allImages {
			switch img.Status {
			case "processing":
				processingImages++
			case "completed":
				completedImages++
			case "failed":
				failedImages++
			}
		}

		result["image_processing"] = map[string]int{
			"total":      len(allImages),
			"processing": processingImages,
			"completed":  completedImages,
			"failed":     failedImages,
		}
	}

	return result, nil
}

// GetPartners возвращает список партнеров с фильтрацией и поиском
func (s *AdminService) GetPartners(search, status string) ([]*partner.Partner, error) {
	log := zerolog.Ctx(context.Background())
	var partners []*partner.Partner
	var err error

	if search == "" && status == "" {
		partners, err = s.deps.PartnerRepository.GetAll(context.Background())
	} else {
		// Если есть поисковые фильтры, используем поиск
		partners, err = s.deps.PartnerRepository.Search(context.Background(), search, status)
	}

	if err != nil {
		log.Error().Err(err).Str("search", search).Str("status", status).Msg("Failed to find all partners")
		return nil, fmt.Errorf("failed to find all partners: %w", err)
	}

	return partners, nil
}

// CreatePartner создает нового партнера
func (s *AdminService) CreatePartner(req partner.CreatePartnerRequest) (*partner.Partner, error) {
	log := zerolog.Ctx(context.Background())
	// Проверяем уникальность логина
	if _, err := s.deps.PartnerRepository.GetByLogin(context.Background(), req.Login); err == nil {
		log.Error().Str("login", req.Login).Msg("Partner already exists")
		return nil, fmt.Errorf("partner already exists: %w", err)
	}

	// Проверяем уникальность домена
	if _, err := s.deps.PartnerRepository.GetByDomain(context.Background(), req.Domain); err == nil {
		log.Error().Str("domain", req.Domain).Msg("Partner already exists")
		return nil, fmt.Errorf("partner already exists: %w", err)
	}

	// Хешируем пароль
	hashedPassword, err := bcrypt.HashPassword(req.Password)
	if err != nil {
		log.Error().Err(err).Msg("Failed to hash password")
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Генерируем уникальный код партнера
	partnerCode, err := s.generateUniquePartnerCode()
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate partner code")
		return nil, fmt.Errorf("failed to generate partner code: %w", err)
	}

	// Обрабатываем ссылки на социальные сети
	telegramLink, whatsappLink := s.processSocialLinks(req.Telegram, req.Whatsapp)

	// Создаем партнера
	newPartner := &partner.Partner{
		PartnerCode:     partnerCode,
		Login:           req.Login,
		Password:        hashedPassword,
		Domain:          req.Domain,
		BrandName:       req.BrandName,
		LogoURL:         req.LogoURL,
		OzonLink:        req.OzonLink,
		WildberriesLink: req.WildberriesLink,
		Email:           req.Email,
		Address:         req.Address,
		Phone:           req.Phone,
		Telegram:        req.Telegram,
		Whatsapp:        req.Whatsapp,
		TelegramLink:    telegramLink,
		WhatsappLink:    whatsappLink,
		AllowSales:      req.AllowSales,
		Status:          req.Status,
	}

	if err := s.deps.PartnerRepository.Create(context.Background(), newPartner); err != nil {
		log.Error().Err(err).Msg("Failed to create partner")
		return nil, fmt.Errorf("failed to create partner: %w", err)
	}

	return newPartner, nil
}

// generateUniquePartnerCode генерирует уникальный 4-значный код партнера
func (s *AdminService) generateUniquePartnerCode() (string, error) {
	log := zerolog.Ctx(context.Background())
	// Получаем всех партнеров и находим максимальный код
	partners, err := s.deps.PartnerRepository.GetAll(context.Background())
	if err != nil {
		log.Error().Err(err).Msg("Failed to find all partners")
		return "", fmt.Errorf("failed to find all partners: %w", err)
	}

	maxCode := 0
	for _, p := range partners {
		var code int
		if _, err := strconv.Atoi(p.PartnerCode); err == nil {
			if code > maxCode {
				maxCode = code
			}
		}
	}

	// Если максимальный код равен 9999, возвращаем ошибку, иначе инкрементируем на 1
	newCode := maxCode + 1
	if newCode > 9999 {
		log.Error().Msg("Max partner code reached")
		return "", fmt.Errorf("max partner code reached")
	}

	return fmt.Sprintf("%04d", newCode), nil
}

// GetPartner возвращает партнера по ID
func (s *AdminService) GetPartner(id uuid.UUID) (*partner.Partner, error) {
	log := zerolog.Ctx(context.Background())
	p, err := s.deps.PartnerRepository.GetByID(context.Background(), id)
	if err != nil {
		log.Error().Err(err).Str("partner_id", id.String()).Msg("Partner not found")
		return nil, fmt.Errorf("partner not found: %w", err)
	}
	return p, nil
}

// UpdatePartner обновляет информацию о партнере
func (s *AdminService) UpdatePartner(id uuid.UUID, req partner.UpdatePartnerRequest) (*partner.Partner, error) {
	log := zerolog.Ctx(context.Background())
	// Получаем существующего партнера
	existingPartner, err := s.deps.PartnerRepository.GetByID(context.Background(), id)
	if err != nil {
		log.Error().Err(err).Str("partner_id", id.String()).Msg("Partner not found")
		return nil, fmt.Errorf("partner not found: %w", err)
	}

	// Проверяем уникальность логина (если он изменяется)
	if req.Login != nil && *req.Login != existingPartner.Login {
		if _, err := s.deps.PartnerRepository.GetByLogin(context.Background(), *req.Login); err == nil {
			log.Error().Str("login", *req.Login).Msg("Partner already exists")
			return nil, fmt.Errorf("partner already exists: %w", err)
		}
		existingPartner.Login = *req.Login
	}

	// Проверяем уникальность домена (если он изменяется)
	if req.Domain != nil && *req.Domain != existingPartner.Domain {
		if _, err := s.deps.PartnerRepository.GetByDomain(context.Background(), *req.Domain); err == nil {
			log.Error().Str("domain", *req.Domain).Msg("Partner already exists")
			return nil, fmt.Errorf("partner already exists: %w", err)
		}
		existingPartner.Domain = *req.Domain
	}

	// Обновляем пароль (если указан)
	if req.Password != nil {
		hashedPassword, err := bcrypt.HashPassword(*req.Password)
		if err != nil {
			log.Error().Err(err).Msg("Failed to hash password")
			return nil, fmt.Errorf("failed to hash password: %w", err)
		}
		existingPartner.Password = hashedPassword
	}

	// Обновляем остальные поля
	updatePartnerData.UpdatePartnerData(existingPartner, &req)

	// Обрабатываем ссылки на социальные сети
	if req.Telegram != nil || req.Whatsapp != nil {
		telegramLink, whatsappLink := s.processSocialLinks(
			existingPartner.Telegram,
			existingPartner.Whatsapp,
		)
		existingPartner.TelegramLink = telegramLink
		existingPartner.WhatsappLink = whatsappLink
	}

	// Сохраняем изменения
	if err := s.deps.PartnerRepository.Update(context.Background(), existingPartner); err != nil {
		log.Error().Err(err).Str("partner_id", id.String()).Msg("Failed to update partner")
		return nil, fmt.Errorf("failed to update partner: %w", err)
	}

	return existingPartner, nil
}

// BlockPartner блокирует партнера
func (s *AdminService) BlockPartner(id uuid.UUID) error {
	log := zerolog.Ctx(context.Background())
	p, err := s.deps.PartnerRepository.GetByID(context.Background(), id)
	if err != nil {
		log.Error().Err(err).Str("partner_id", id.String()).Msg("Partner not found")
		return fmt.Errorf("partner not found: %w", err)
	}

	p.Status = "blocked"
	if err := s.deps.PartnerRepository.Update(context.Background(), p); err != nil {
		log.Error().Err(err).Str("partner_id", id.String()).Msg("Failed to block partner")
		return fmt.Errorf("failed to block partner: %w", err)
	}

	// Блокируем все купоны партнера
	if err := s.deps.CouponRepository.UpdateStatusByPartnerID(context.Background(), id, true); err != nil {
		log.Error().Err(err).Str("partner_id", id.String()).Msg("Failed to block coupons")
		return fmt.Errorf("failed to block coupons: %w", err)
	}

	return nil
}

// UnblockPartner разблокирует партнера
func (s *AdminService) UnblockPartner(id uuid.UUID) error {
	log := zerolog.Ctx(context.Background())
	p, err := s.deps.PartnerRepository.GetByID(context.Background(), id)
	if err != nil {
		log.Error().Err(err).Str("partner_id", id.String()).Msg("Partner not found")
		return fmt.Errorf("partner not found: %w", err)
	}

	p.Status = "active"
	if err := s.deps.PartnerRepository.Update(context.Background(), p); err != nil {
		log.Error().Err(err).Str("partner_id", id.String()).Msg("Failed to unblock partner")
		return fmt.Errorf("failed to unblock partner: %w", err)
	}

	// Разблокируем все купоны партнера
	if err := s.deps.CouponRepository.UpdateStatusByPartnerID(context.Background(), id, false); err != nil {
		log.Error().Err(err).Str("partner_id", id.String()).Msg("Failed to unblock coupons")
		return fmt.Errorf("failed to unblock coupons: %w", err)
	}

	return nil
}

// DeletePartner удаляет партнера
func (s *AdminService) DeletePartner(id uuid.UUID) error {
	log := zerolog.Ctx(context.Background())
	if err := s.deps.PartnerRepository.DeleteWithCoupons(context.Background(), id); err != nil {
		log.Error().Err(err).Str("partner_id", id.String()).Msg("Failed to delete partner")
		return fmt.Errorf("failed to delete partner: %w", err)
	}
	return nil
}

// GetPartnerStatistics возвращает статистику партнера
func (s *AdminService) GetPartnerStatistics(id uuid.UUID) (map[string]interface{}, error) {
	log := zerolog.Ctx(context.Background())
	// Проверяем существование партнера
	_, err := s.deps.PartnerRepository.GetByID(context.Background(), id)
	if err != nil {
		log.Error().Err(err).Str("partner_id", id.String()).Msg("Partner not found")
		return nil, fmt.Errorf("partner not found: %w", err)
	}

	// Получаем статистику купонов
	stats, err := s.deps.CouponRepository.GetStatistics(context.Background(), &id)
	if err != nil {
		log.Error().Err(err).Str("partner_id", id.String()).Msg("Failed to get partner statistics")
		return nil, fmt.Errorf("failed to get partner statistics: %w", err)
	}

	return map[string]interface{}{
		"coupon_statistics": stats,
	}, nil
}

// GetCoupons возвращает список купонов с фильтрацией
func (s *AdminService) GetCoupons(code, status, size, style string, partnerID *uuid.UUID) ([]*coupon.Coupon, error) {
	log := zerolog.Ctx(context.Background())
	coupons, err := s.deps.CouponRepository.Search(context.Background(), code, status, size, style, partnerID)
	if err != nil {
		log.Error().Err(err).Interface("filters", map[string]interface{}{
			"code": code, "status": status, "size": size, "style": style, "partner_id": partnerID,
		}).Msg("Failed to get coupons")
		return nil, fmt.Errorf("failed to get coupons: %w", err)
	}
	return coupons, nil
}

// GetCouponsPaginated возвращает купоны с пагинацией
func (s *AdminService) GetCouponsPaginated(code, status, size, style string, partnerID *uuid.UUID, page, limit int) ([]*coupon.Coupon, int64, error) {
	log := zerolog.Ctx(context.Background())
	coupons, total, err := s.deps.CouponRepository.SearchWithPagination(context.Background(), code, status, size, style, partnerID, page, limit)
	if err != nil {
		log.Error().Err(err).Interface("filters", map[string]interface{}{
			"code": code, "status": status, "size": size, "style": style, "partner_id": partnerID, "page": page, "limit": limit,
		}).Msg("Failed to get coupons")
		return nil, 0, fmt.Errorf("failed to get coupons: %w", err)
	}
	return coupons, int64(total), nil
}

// CreateCoupons создает купоны
func (s *AdminService) CreateCoupons(req coupon.CreateCouponRequest) ([]*coupon.Coupon, error) {
	log := zerolog.Ctx(context.Background())
	// Определяем код партнера
	partnerCode := "0000" // По умолчанию для собственных купонов

	if req.PartnerID != uuid.Nil {
		partner, err := s.deps.PartnerRepository.GetByID(context.Background(), req.PartnerID)
		if err != nil {
			log.Error().Err(err).Str("partner_id", req.PartnerID.String()).Msg("Partner not found")
			return nil, fmt.Errorf("partner not found: %w", err)
		}
		partnerCode = partner.PartnerCode
	}

	// Генерируем купоны
	coupons := make([]*coupon.Coupon, req.Count)

	for i := 0; i < req.Count; i++ {
		// Генерируем уникальный код купона
		code, err := randomCouponCode.GenerateUniqueCouponCode(partnerCode, s.deps.CouponRepository)
		if err != nil {
			log.Error().Err(err).Str("partner_code", partnerCode).Int("attempt", i).Msg("Failed to create coupons")
			return nil, fmt.Errorf("failed to create coupons: %w", err)
		}

		coupons[i] = &coupon.Coupon{
			Code:      code,
			PartnerID: req.PartnerID,
			Size:      string(req.Size),
			Style:     string(req.Style),
			Status:    string(coupon.StatusNew),
		}
	}

	// Сохраняем купоны в базе данных
	if err := s.deps.CouponRepository.CreateBatch(context.Background(), coupons); err != nil {
		log.Error().Err(err).Int("count", req.Count).Str("partner_code", partnerCode).Msg("Failed to create coupons")
		return nil, fmt.Errorf("failed to create coupons: %w", err)
	}

	return coupons, nil
}

// ExportCoupons экспортирует купоны в CSV
func (s *AdminService) ExportCoupons(code, status, size, style string, partnerID *uuid.UUID) (string, error) {
	log := zerolog.Ctx(context.Background())
	// Получаем купоны для экспорта
	coupons, err := s.deps.CouponRepository.Search(context.Background(), code, status, size, style, partnerID)
	if err != nil {
		log.Error().Err(err).Interface("filters", map[string]interface{}{
			"code": code, "status": status, "size": size, "style": style, "partner_id": partnerID,
		}).Msg("Failed to export coupons")
		return "", fmt.Errorf("failed to export coupons: %w", err)
	}

	if len(coupons) == 0 {
		log.Warn().Interface("filters", map[string]interface{}{
			"code": code, "status": status, "size": size, "style": style, "partner_id": partnerID,
		}).Msg("No coupons found for export")
		return "", fmt.Errorf("no coupons found for export")
	}

	// Формируем CSV
	var csvBuilder strings.Builder
	csvBuilder.WriteString("Code,Partner ID,Size,Style,Status,Is Purchased,Purchase Email,Created At,Used At\n")

	for _, c := range coupons {
		usedAt := ""
		if c.UsedAt != nil {
			usedAt = c.UsedAt.Format("2006-01-02 15:04:05")
		}

		purchaseEmail := ""
		if c.PurchaseEmail != nil {
			purchaseEmail = *c.PurchaseEmail
		}

		csvBuilder.WriteString(fmt.Sprintf("%s,%s,%s,%s,%s,%t,%s,%s,%s\n",
			c.Code,
			c.PartnerID.String(),
			c.Size,
			c.Style,
			c.Status,
			c.IsPurchased,
			purchaseEmail,
			c.CreatedAt.Format("2006-01-02 15:04:05"),
			usedAt,
		))
	}

	return csvBuilder.String(), nil
}

// ExportPartnerCoupons экспортирует купоны партнера
func (s *AdminService) ExportPartnerCoupons(partnerID uuid.UUID) (string, error) {
	return s.ExportCoupons("", "", "", "", &partnerID)
}

// BatchDeleteCoupons удаляет купоны по списку ID
func (s *AdminService) BatchDeleteCoupons(couponIDs []uuid.UUID) (int64, error) {
	log := zerolog.Ctx(context.Background())
	if len(couponIDs) == 0 {
		return 0, fmt.Errorf("bad request")
	}

	deleted, err := s.deps.CouponRepository.BatchDelete(context.Background(), couponIDs)
	if err != nil {
		log.Error().Err(err).Msg("Failed to delete coupons")
		return 0, fmt.Errorf("failed to delete coupons: %w", err)
	}

	return deleted, nil
}

// GetCoupon возвращает купон по ID
func (s *AdminService) GetCoupon(id uuid.UUID) (*coupon.Coupon, error) {
	log := zerolog.Ctx(context.Background())
	c, err := s.deps.CouponRepository.GetByID(context.Background(), id)
	if err != nil {
		log.Error().Err(err).Str("coupon_id", id.String()).Msg("Coupon not found")
		return nil, fmt.Errorf("coupon not found: %w", err)
	}
	return c, nil
}

// ResetCoupon сбрасывает купон в исходное состояние
func (s *AdminService) ResetCoupon(id uuid.UUID) error {
	log := zerolog.Ctx(context.Background())
	if err := s.deps.CouponRepository.Reset(context.Background(), id); err != nil {
		log.Error().Err(err).Str("coupon_id", id.String()).Msg("Failed to reset coupon")
		return fmt.Errorf("failed to reset coupon: %w", err)
	}
	return nil
}

// DeleteCoupon удаляет купон
func (s *AdminService) DeleteCoupon(id uuid.UUID) error {
	log := zerolog.Ctx(context.Background())
	if err := s.deps.CouponRepository.Delete(context.Background(), id); err != nil {
		log.Error().Err(err).Str("coupon_id", id.String()).Msg("Failed to delete coupon")
		return fmt.Errorf("failed to delete coupon: %w", err)
	}
	return nil
}

// GetStatistics возвращает общую статистику
func (s *AdminService) GetStatistics() (map[string]interface{}, error) {
	log := zerolog.Ctx(context.Background())
	stats, err := s.deps.CouponRepository.GetStatistics(context.Background(), nil)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get statistics")
		return nil, fmt.Errorf("failed to get statistics: %w", err)
	}

	return map[string]interface{}{
		"coupon_statistics": stats,
	}, nil
}

// GetPartnersStatistics возвращает статистику по партнерам
func (s *AdminService) GetPartnersStatistics() (map[string]interface{}, error) {
	log := zerolog.Ctx(context.Background())
	partners, err := s.deps.PartnerRepository.GetAll(context.Background())
	if err != nil {
		log.Error().Err(err).Msg("Failed to get partners")
		return nil, fmt.Errorf("failed to get partners: %w", err)
	}

	result := make(map[string]interface{})
	partnerStats := make([]map[string]interface{}, 0, len(partners))

	for _, p := range partners {
		stats, err := s.deps.CouponRepository.GetStatistics(context.Background(), &p.ID)
		if err != nil {
			continue // Пропускаем партнеров с ошибкой статистики
		}

		partnerStats = append(partnerStats, map[string]interface{}{
			"partner_id":   p.ID,
			"brand_name":   p.BrandName,
			"partner_code": p.PartnerCode,
			"status":       p.Status,
			"statistics":   stats,
		})
	}

	result["partners"] = partnerStats
	return result, nil
}

// GetSystemStatistics возвращает системную статистику
func (s *AdminService) GetSystemStatistics() (map[string]interface{}, error) {
	result := make(map[string]interface{})

	// Статистика по купонам
	couponStats, err := s.deps.CouponRepository.GetStatistics(context.Background(), nil)
	if err == nil {
		result["coupons"] = couponStats
	}

	// Статистика по партнерам
	partners, err := s.deps.PartnerRepository.GetAll(context.Background())
	if err == nil {
		activePartners := 0
		for _, p := range partners {
			if p.Status == "active" {
				activePartners++
			}
		}
		result["partners"] = map[string]interface{}{
			"total":  len(partners),
			"active": activePartners,
		}
	}

	// Статистика по обработке изображений
	images, err := s.deps.ImageRepository.GetAll(context.Background())
	if err == nil {
		processingImages := 0
		completedImages := 0
		failedImages := 0

		for _, img := range images {
			switch img.Status {
			case "processing":
				processingImages++
			case "completed":
				completedImages++
			case "failed":
				failedImages++
			}
		}

		result["image_processing"] = map[string]interface{}{
			"total":      len(images),
			"processing": processingImages,
			"completed":  completedImages,
			"failed":     failedImages,
		}
	}

	return result, nil
}

// GetAnalytics возвращает аналитику
func (s *AdminService) GetAnalytics() (map[string]interface{}, error) {
	log := zerolog.Ctx(context.Background())
	result := make(map[string]interface{})

	// Получаем все купоны для анализа
	allCoupons, err := s.deps.CouponRepository.GetAll(context.Background())
	if err != nil {
		log.Error().Err(err).Msg("Failed to get coupons")
		return nil, fmt.Errorf("failed to get coupons: %w", err)
	}

	// Анализ по дням
	dailyStats := make(map[string]int)
	sizeStats := make(map[string]int)
	styleStats := make(map[string]int)

	for _, c := range allCoupons {
		// Статистика по дням создания
		day := c.CreatedAt.Format("2006-01-02")
		dailyStats[day]++

		// Статистика по размерам
		sizeStats[c.Size]++

		// Статистика по стилям
		styleStats[c.Style]++
	}

	result["daily_creation"] = dailyStats
	result["size_distribution"] = sizeStats
	result["style_distribution"] = styleStats

	return result, nil
}

// GetDashboardStatistics возвращает статистику для дашборда
func (s *AdminService) GetDashboardStatistics() (map[string]interface{}, error) {
	return s.GetDashboardData()
}

// GetAllImages возвращает все задачи обработки изображений
func (s *AdminService) GetAllImages() ([]*image.Image, error) {
	log := zerolog.Ctx(context.Background())
	images, err := s.deps.ImageRepository.GetAll(context.Background())
	if err != nil {
		log.Error().Err(err).Msg("Failed to get images")
		return nil, fmt.Errorf("failed to get images: %w", err)
	}
	return images, nil
}

// GetImageDetails возвращает детали задачи обработки изображения
func (s *AdminService) GetImageDetails(id uuid.UUID) (*image.Image, error) {
	log := zerolog.Ctx(context.Background())
	task, err := s.deps.ImageRepository.GetByID(context.Background(), id)
	if err != nil {
		log.Error().Err(err).Str("image_id", id.String()).Msg("Image not found")
		return nil, fmt.Errorf("image not found: %w", err)
	}
	return task, nil
}

// DeleteImageTask удаляет задачу обработки изображения
func (s *AdminService) DeleteImageTask(id uuid.UUID) error {
	log := zerolog.Ctx(context.Background())
	if err := s.deps.ImageRepository.Delete(context.Background(), id); err != nil {
		log.Error().Err(err).Str("image_id", id.String()).Msg("Failed to delete image")
		return fmt.Errorf("failed to delete image: %w", err)
	}
	return nil
}

// RetryImageTask повторно запускает обработку изображения
func (s *AdminService) RetryImageTask(id uuid.UUID) error {
	log := zerolog.Ctx(context.Background())
	task, err := s.deps.ImageRepository.GetByID(context.Background(), id)
	if err != nil {
		log.Error().Err(err).Str("image_id", id.String()).Msg("Image not found")
		return fmt.Errorf("image not found: %w", err)
	}

	// Сбрасываем статус задачи для повторной обработки
	task.Status = "queued"
	task.ErrorMessage = nil
	task.StartedAt = nil
	task.CompletedAt = nil

	if err := s.deps.ImageRepository.Update(context.Background(), task); err != nil {
		log.Error().Err(err).Str("image_id", id.String()).Msg("Failed to update image")
		return fmt.Errorf("failed to update image: %w", err)
	}

	return nil
}
