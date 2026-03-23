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
// Для каждого месяца диапазона [q.From, q.To] суммируются цены активных подписок,
// затем результаты складываются — т.е. подписка за N месяцев считается N раз.
func (r *Repository) Total(ctx context.Context, q TotalSubscriptionsQuery) (int, error) {
	// Нормализуем на первое число месяца
	from := time.Date(q.From.Year(), q.From.Month(), 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(q.To.Year(), q.To.Month(), 1, 0, 0, 0, 0, time.UTC)

	total := 0
	current := from

	for !current.After(to) {
		var monthTotal int

		dbq := r.db.WithContext(ctx).
			Model(&models.Subscription{}).
			Where("start_date <= ? AND (end_date IS NULL OR end_date >= ?)", current, current)

		if q.UserID != "" {
			dbq = dbq.Where("user_id = ?", q.UserID)
		}
		if q.ServiceName != "" {
			dbq = dbq.Where("service_name = ?", q.ServiceName)
		}

		result := dbq.Select("COALESCE(SUM(price), 0)").Scan(&monthTotal)
		if result.Error != nil {
			return 0, fmt.Errorf("repository: total calculation failed for %s: %w", current.Format("01-2006"), result.Error)
		}

		total += monthTotal
		current = current.AddDate(0, 1, 0)
	}

	return total, nil
}
