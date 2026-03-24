package subscription

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/devilzzcpp/agregator-zzxx/common"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// @Summary      Создать подписку
// @Tags         subscriptions
// @Accept       json
// @Produce      json
// @Param        input  body      CreateSubscriptionRequest  true  "Данные подписки"
// @Success      201    {object}  SubscriptionResponse
// @Failure      400    {object}  common.ValidationErrorResponseDoc
// @Failure      409    {object}  common.ErrorResponse
// @Failure      500    {object}  common.ErrorResponse
// @Router       /subscriptions [post]
func (h *Handler) Create(c *gin.Context) {
	var req CreateSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "CreateSubscriptionRequest.UserID") && strings.Contains(errMsg, "uuid") {
			common.JSONValidationError(c, map[string]string{"user_id": "должен быть UUID"})
			return
		}

		common.JSONValidationError(c, map[string]string{"request": "некорректные входные данные"})
		return
	}

	resp, err := h.service.Create(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, ErrConflict) {
			common.JSONError(c, http.StatusConflict, err.Error())
			return
		}
		if errors.Is(err, ErrInvalidDate) || errors.Is(err, ErrBadRequest) {
			errMsg := err.Error()
			if strings.Contains(errMsg, "price") {
				common.JSONValidationError(c, map[string]string{"price": errMsg})
				return
			}
			if strings.Contains(errMsg, "start_date") {
				common.JSONValidationError(c, map[string]string{"start_date": errMsg})
				return
			}
			if strings.Contains(errMsg, "end_date") {
				common.JSONValidationError(c, map[string]string{"end_date": errMsg})
				return
			}

			common.JSONValidationError(c, map[string]string{"request": errMsg})
			return
		}

		if common.Logger != nil {
			common.Logger.Error("subscription create: внутренняя ошибка", zap.Error(err))
		}
		common.JSONError(c, http.StatusInternalServerError, "внутренняя ошибка сервера")
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// @Summary      Список подписок
// @Tags         subscriptions
// @Produce      json
// @Param        user_id       query     string  false  "UUID пользователя"
// @Param        service_name  query     string  false  "Название сервиса"
// @Param        limit         query     int     false  "Лимит (по умолчанию без ограничений)"
// @Param        offset        query     int     false  "Смещение"
// @Success      200    {array}   SubscriptionResponse
// @Failure      400    {object}  common.ErrorResponse
// @Failure      500    {object}  common.ErrorResponse
// @Router       /subscriptions [get]
func (h *Handler) List(c *gin.Context) {
	var q ListSubscriptionsQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		common.JSONError(c, http.StatusBadRequest, "некорректные query параметры")
		return
	}

	resp, err := h.service.List(c.Request.Context(), q)
	if err != nil {
		if common.Logger != nil {
			common.Logger.Error("subscription list: внутренняя ошибка", zap.Error(err))
		}
		common.JSONError(c, http.StatusInternalServerError, "внутренняя ошибка сервера")
		return
	}

	c.JSON(http.StatusOK, resp)
}

// @Summary      Получить подписку по ID
// @Tags         subscriptions
// @Produce      json
// @Param        id   path      string  true  "UUID подписки"
// @Success      200  {object}  SubscriptionResponse
// @Failure      404  {object}  common.ErrorResponse
// @Failure      500  {object}  common.ErrorResponse
// @Router       /subscriptions/{id} [get]
func (h *Handler) GetByID(c *gin.Context) {
	id := c.Param("id")

	resp, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			common.JSONError(c, http.StatusNotFound, err.Error())
			return
		}

		if common.Logger != nil {
			common.Logger.Error("subscription getByID: внутренняя ошибка", zap.Error(err))
		}
		common.JSONError(c, http.StatusInternalServerError, "внутренняя ошибка сервера")
		return
	}

	c.JSON(http.StatusOK, resp)
}

// @Summary      Обновить подписку (частичное обновление)
// @Description  Все поля опциональны. end_date можно передать как null для сброса.
// @Tags         subscriptions
// @Accept       json
// @Produce      json
// @Param        id     path      string                    true  "UUID подписки"
// @Param        input  body      UpdateSubscriptionRequest  true  "Обновляемые поля"
// @Success      200    {object}  SubscriptionResponse
// @Failure      400    {object}  common.ValidationErrorResponseDoc
// @Failure      404    {object}  common.ErrorResponse
// @Failure      409    {object}  common.ErrorResponse
// @Failure      500    {object}  common.ErrorResponse
// @Router       /subscriptions/{id} [put]
func (h *Handler) Update(c *gin.Context) {
	id := c.Param("id")
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		common.JSONError(c, http.StatusBadRequest, "некорректное тело запроса")
		return
	}
	c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

	var req UpdateSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "UpdateSubscriptionRequest.Price") || strings.Contains(errMsg, ".price") {
			common.JSONValidationError(c, map[string]string{"price": "должен быть числом"})
			return
		}
		if strings.Contains(errMsg, "UpdateSubscriptionRequest.UserID") || strings.Contains(errMsg, ".user_id") {
			common.JSONValidationError(c, map[string]string{"user_id": "должен быть UUID"})
			return
		}
		if strings.Contains(errMsg, "UpdateSubscriptionRequest.StartDate") || strings.Contains(errMsg, ".start_date") {
			common.JSONValidationError(c, map[string]string{"start_date": "ожидается строка в формате MM-YYYY"})
			return
		}
		if strings.Contains(errMsg, "UpdateSubscriptionRequest.EndDate") || strings.Contains(errMsg, ".end_date") {
			common.JSONValidationError(c, map[string]string{"end_date": "ожидается строка в формате MM-YYYY или null"})
			return
		}

		common.JSONValidationError(c, map[string]string{"request": "некорректные входные данные"})
		return
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil {
		common.JSONError(c, http.StatusBadRequest, "некорректный JSON")
		return
	}

	endDateProvided := false
	endDateSetToNull := false
	if value, ok := raw["end_date"]; ok {
		endDateProvided = true
		if string(value) == "null" {
			endDateSetToNull = true
		}
	}

	resp, err := h.service.Update(c.Request.Context(), id, req, endDateProvided, endDateSetToNull)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			common.JSONError(c, http.StatusNotFound, err.Error())
			return
		}
		if errors.Is(err, ErrConflict) {
			common.JSONError(c, http.StatusConflict, err.Error())
			return
		}
		if errors.Is(err, ErrInvalidDate) || errors.Is(err, ErrBadRequest) {
			errMsg := err.Error()
			if strings.Contains(errMsg, "price") {
				common.JSONValidationError(c, map[string]string{"price": errMsg})
				return
			}
			if strings.Contains(errMsg, "start_date") {
				common.JSONValidationError(c, map[string]string{"start_date": errMsg})
				return
			}
			if strings.Contains(errMsg, "end_date") {
				common.JSONValidationError(c, map[string]string{"end_date": errMsg})
				return
			}
			if strings.Contains(errMsg, "user_id") || strings.Contains(errMsg, "UUID") || strings.Contains(errMsg, "uuid") {
				common.JSONValidationError(c, map[string]string{"user_id": errMsg})
				return
			}

			common.JSONValidationError(c, map[string]string{"request": errMsg})
			return
		}

		if common.Logger != nil {
			common.Logger.Error("subscription update: внутренняя ошибка", zap.Error(err))
		}
		common.JSONError(c, http.StatusInternalServerError, "внутренняя ошибка сервера")
		return
	}

	c.JSON(http.StatusOK, resp)
}

// @Summary      Удалить подписку
// @Tags         subscriptions
// @Param        id   path  string  true  "UUID подписки"
// @Success      204
// @Failure      404  {object}  common.ErrorResponse
// @Failure      500  {object}  common.ErrorResponse
// @Router       /subscriptions/{id} [delete]
func (h *Handler) Delete(c *gin.Context) {
	id := c.Param("id")

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		if errors.Is(err, ErrNotFound) {
			common.JSONError(c, http.StatusNotFound, err.Error())
			return
		}

		if common.Logger != nil {
			common.Logger.Error("subscription delete: внутренняя ошибка", zap.Error(err))
		}
		common.JSONError(c, http.StatusInternalServerError, "внутренняя ошибка сервера")
		return
	}

	c.Status(http.StatusNoContent)
}

// @Summary      Суммарная стоимость подписок за период
// @Description  Считает помесячно: каждый активный месяц прибавляет price. Границы включительны.
// @Tags         subscriptions
// @Produce      json
// @Param        from          query     string  true   "Начало периода, формат MM-YYYY"
// @Param        to            query     string  true   "Конец периода, формат MM-YYYY"
// @Param        user_id       query     string  false  "Фильтр по UUID пользователя"
// @Param        service_name  query     string  false  "Фильтр по названию сервиса"
// @Success      200  {object}  SubscriptionsTotalResponse
// @Failure      400  {object}  common.ErrorResponse
// @Failure      500  {object}  common.ErrorResponse
// @Router       /subscriptions/total [get]
func (h *Handler) Total(c *gin.Context) {
	fromStr := c.Query("from")
	toStr := c.Query("to")

	if fromStr == "" || toStr == "" {
		common.JSONError(c, http.StatusBadRequest, "параметры from и to обязательны")
		return
	}

	resp, err := h.service.Total(
		c.Request.Context(),
		c.Query("user_id"),
		c.Query("service_name"),
		fromStr,
		toStr,
	)
	if err != nil {
		if errors.Is(err, ErrInvalidDate) || errors.Is(err, ErrBadRequest) {
			common.JSONError(c, http.StatusBadRequest, err.Error())
			return
		}

		if common.Logger != nil {
			common.Logger.Error("subscription total: внутренняя ошибка", zap.Error(err))
		}
		common.JSONError(c, http.StatusInternalServerError, "внутренняя ошибка сервера")
		return
	}

	c.JSON(http.StatusOK, resp)
}
