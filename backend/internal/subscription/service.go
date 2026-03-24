package subscription

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/devilzzcpp/agregator-zzxx/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrNotFound    = errors.New("подписка не найдена")
	ErrInvalidDate = errors.New("неверный формат даты, ожидается MM-YYYY")
	ErrBadRequest  = errors.New("некорректные входные данные")
	ErrConflict    = errors.New("конфликт периода подписки")
)

const dateLayout = "01-2006"

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func parseMonth(s string) (time.Time, error) {
	t, err := time.Parse(dateLayout, s)
	if err != nil {
		return time.Time{}, fmt.Errorf("%w: %q", ErrInvalidDate, s)
	}
	return t, nil
}

func modelToResponse(sub *models.Subscription) *SubscriptionResponse {
	resp := &SubscriptionResponse{
		ID:          sub.ID,
		ServiceName: sub.ServiceName,
		Price:       sub.Price,
		UserID:      sub.UserID,
		StartDate:   sub.StartDate.Format(dateLayout),
		CreatedAt:   sub.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   sub.UpdatedAt.Format(time.RFC3339),
	}
	if sub.EndDate != nil {
		s := sub.EndDate.Format(dateLayout)
		resp.EndDate = &s
	}
	return resp
}

func (s *Service) Create(ctx context.Context, req CreateSubscriptionRequest) (*SubscriptionResponse, error) {
	if req.Price <= 0 {
		return nil, fmt.Errorf("%w: price должен быть больше 0", ErrBadRequest)
	}

	startDate, err := parseMonth(req.StartDate)
	if err != nil {
		return nil, err
	}

	sub := &models.Subscription{
		ID:          uuid.New().String(),
		ServiceName: req.ServiceName,
		Price:       req.Price,
		UserID:      req.UserID,
		StartDate:   startDate,
	}

	if req.EndDate != nil {
		endDate, err := parseMonth(*req.EndDate)
		if err != nil {
			return nil, err
		}
		if endDate.Before(startDate) {
			return nil, fmt.Errorf("%w: end_date не может быть раньше start_date", ErrBadRequest)
		}
		sub.EndDate = &endDate
	}

	overlap, err := s.repo.HasPeriodOverlap(ctx, sub.UserID, sub.ServiceName, "", sub.StartDate, sub.EndDate)
	if err != nil {
		return nil, err
	}
	if overlap {
		return nil, fmt.Errorf("%w: сервис уже существует для пользователя в этом периоде", ErrConflict)
	}

	created, err := s.repo.Create(ctx, sub)
	if err != nil {
		return nil, err
	}

	return modelToResponse(created), nil
}

func (s *Service) GetByID(ctx context.Context, id string) (*SubscriptionResponse, error) {
	sub, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return modelToResponse(sub), nil
}

func (s *Service) List(ctx context.Context, q ListSubscriptionsQuery) ([]SubscriptionResponse, error) {
	subs, err := s.repo.List(ctx, q)
	if err != nil {
		return nil, err
	}

	result := make([]SubscriptionResponse, 0, len(subs))
	for i := range subs {
		result = append(result, *modelToResponse(&subs[i]))
	}
	return result, nil
}

func (s *Service) Update(ctx context.Context, id string, req UpdateSubscriptionRequest, endDateProvided bool, endDateSetToNull bool) (*SubscriptionResponse, error) {
	sub, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	if req.ServiceName != nil {
		sub.ServiceName = *req.ServiceName
	}
	if req.Price != nil {
		if *req.Price <= 0 {
			return nil, fmt.Errorf("%w: price должен быть больше 0", ErrBadRequest)
		}
		sub.Price = *req.Price
	}
	if req.UserID != nil {
		sub.UserID = *req.UserID
	}
	if req.StartDate != nil {
		startDate, err := parseMonth(*req.StartDate)
		if err != nil {
			return nil, err
		}
		sub.StartDate = startDate
	}
	if endDateProvided {
		if endDateSetToNull {
			sub.EndDate = nil
		} else {
			if req.EndDate == nil {
				return nil, fmt.Errorf("%w: end_date должен быть строкой MM-YYYY или null", ErrBadRequest)
			}
			endDate, err := parseMonth(*req.EndDate)
			if err != nil {
				return nil, err
			}
			sub.EndDate = &endDate
		}
	}

	if sub.EndDate != nil && sub.EndDate.Before(sub.StartDate) {
		return nil, fmt.Errorf("%w: end_date не может быть раньше start_date", ErrBadRequest)
	}

	overlap, err := s.repo.HasPeriodOverlap(ctx, sub.UserID, sub.ServiceName, sub.ID, sub.StartDate, sub.EndDate)
	if err != nil {
		return nil, err
	}
	if overlap {
		return nil, fmt.Errorf("%w: сервис уже существует для пользователя в этом периоде", ErrConflict)
	}

	updated, err := s.repo.Update(ctx, sub)
	if err != nil {
		return nil, err
	}

	return modelToResponse(updated), nil
}

func (s *Service) Delete(ctx context.Context, id string) error {
	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		return err
	}
	return s.repo.Delete(ctx, id)
}

func (s *Service) Total(ctx context.Context, userID, serviceName, fromStr, toStr string) (*SubscriptionsTotalResponse, error) {
	from, err := parseMonth(fromStr)
	if err != nil {
		return nil, err
	}
	to, err := parseMonth(toStr)
	if err != nil {
		return nil, err
	}
	if to.Before(from) {
		return nil, fmt.Errorf("%w: to не может быть раньше from", ErrBadRequest)
	}

	q := TotalSubscriptionsQuery{
		UserID:      userID,
		ServiceName: serviceName,
		From:        from,
		To:          to,
	}

	total, err := s.repo.Total(ctx, q)
	if err != nil {
		return nil, err
	}

	months := 0
	current := from
	for !current.After(to) {
		months++
		current = current.AddDate(0, 1, 0)
	}

	return &SubscriptionsTotalResponse{
		TotalPrice: total,
		Currency:   "RUB",
		FromMonth:  from.Format(dateLayout),
		ToMonth:    to.Format(dateLayout),
		Months:     months,
	}, nil
}
