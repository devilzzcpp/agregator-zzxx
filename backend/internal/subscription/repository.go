package subscription

import (
	"context"
	"fmt"
	"time"

	"github.com/devilzzcpp/agregator-zzxx/internal/models"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, sub *models.Subscription) (*models.Subscription, error) {
	result := r.db.WithContext(ctx).Create(sub)
	if result.Error != nil {
		return nil, fmt.Errorf("repository: create subscription failed: %w", result.Error)
	}

	return sub, nil
}

func (r *Repository) GetByID(ctx context.Context, id string) (*models.Subscription, error) {
	var sub models.Subscription
	result := r.db.WithContext(ctx).First(&sub, "id = ?", id)
	if result.Error != nil {
		return nil, fmt.Errorf("repository: get subscription failed: %w", result.Error)
	}

	return &sub, nil
}

func (r *Repository) List(ctx context.Context, q ListSubscriptionsQuery) ([]models.Subscription, error) {
	var subs []models.Subscription
	query := r.db.WithContext(ctx).Model(&models.Subscription{})

	if q.UserID != "" {
		query = query.Where("user_id = ?", q.UserID)
	}
	if q.ServiceName != "" {
		query = query.Where("service_name = ?", q.ServiceName)
	}
	if q.Limit > 0 {
		query = query.Limit(q.Limit)
	}
	if q.Offset > 0 {
		query = query.Offset(q.Offset)
	}

	result := query.Find(&subs)
	if result.Error != nil {
		return nil, fmt.Errorf("repository: list subscriptions failed: %w", result.Error)
	}

	return subs, nil
}

func (r *Repository) Update(ctx context.Context, sub *models.Subscription) (*models.Subscription, error) {
	result := r.db.WithContext(ctx).Save(sub)
	if result.Error != nil {
		return nil, fmt.Errorf("repository: update subscription failed: %w", result.Error)
	}

	return sub, nil
}

func (r *Repository) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(&models.Subscription{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("repository: delete subscription failed: %w", result.Error)
	}

	return nil
}

// HasPeriodOverlap проверяет пересечение периодов для user_id + service_name.
// excludeID используется при update, чтобы исключить текущую запись из проверки.
func (r *Repository) HasPeriodOverlap(
	ctx context.Context,
	userID, serviceName, excludeID string,
	startDate time.Time,
	endDate *time.Time,
) (bool, error) {
	query := r.db.WithContext(ctx).
		Model(&models.Subscription{}).
		Where("user_id = ? AND service_name = ?", userID, serviceName)

	if excludeID != "" {
		query = query.Where("id <> ?", excludeID)
	}

	query = query.Where(
		"start_date <= COALESCE(?, DATE '9999-12-31') AND ? <= COALESCE(end_date, DATE '9999-12-31')",
		endDate,
		startDate,
	)

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return false, fmt.Errorf("repository: overlap check failed: %w", err)
	}

	return count > 0, nil
}

// Total считает суммарную стоимость подписок за период.
// Один запрос через GENERATE_SERIES: PostgreSQL разворачивает все месяцы
// диапазона [from, to] и суммирует price активных в каждом месяце подписок.
func (r *Repository) Total(ctx context.Context, q TotalSubscriptionsQuery) (int, error) {
	from := time.Date(q.From.Year(), q.From.Month(), 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(q.To.Year(), q.To.Month(), 1, 0, 0, 0, 0, time.UTC)

	query := `
		SELECT COALESCE(SUM(s.price), 0)
		FROM generate_series(?::date, ?::date, '1 month'::interval) AS m(month)
		JOIN subscriptions s
		  ON s.start_date <= m.month
		 AND (s.end_date IS NULL OR s.end_date >= m.month)
		WHERE 1=1`
	args := []interface{}{from, to}

	if q.UserID != "" {
		query += " AND s.user_id = ?"
		args = append(args, q.UserID)
	}
	if q.ServiceName != "" {
		query += " AND s.service_name = ?"
		args = append(args, q.ServiceName)
	}

	var total int
	if err := r.db.WithContext(ctx).Raw(query, args...).Scan(&total).Error; err != nil {
		return 0, fmt.Errorf("repository: total calculation failed: %w", err)
	}

	return total, nil
}
