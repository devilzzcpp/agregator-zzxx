package subscription

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Create(c *gin.Context) {
	var req CreateSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err.Error()))
		return
	}

	resp, err := h.service.Create(c.Request.Context(), req)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, resp)
}

func (h *Handler) List(c *gin.Context) {
	var q ListSubscriptionsQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err.Error()))
		return
	}

	resp, err := h.service.List(c.Request.Context(), q)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *Handler) GetByID(c *gin.Context) {
	id := c.Param("id")

	resp, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *Handler) Update(c *gin.Context) {
	id := c.Param("id")
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("invalid request body"))
		return
	}
	c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

	var req UpdateSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err.Error()))
		return
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("invalid json"))
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
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *Handler) Delete(c *gin.Context) {
	id := c.Param("id")

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		handleError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) Total(c *gin.Context) {
	fromStr := c.Query("from")
	toStr := c.Query("to")

	if fromStr == "" || toStr == "" {
		c.JSON(http.StatusBadRequest, errorResponse("from and to query params are required"))
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
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func handleError(c *gin.Context, err error) {
	if errors.Is(err, ErrNotFound) {
		c.JSON(http.StatusNotFound, errorResponse(err.Error()))
		return
	}
	if errors.Is(err, ErrConflict) {
		c.JSON(http.StatusConflict, errorResponse(err.Error()))
		return
	}
	if errors.Is(err, ErrInvalidDate) || errors.Is(err, ErrBadRequest) {
		c.JSON(http.StatusBadRequest, errorResponse(err.Error()))
		return
	}
	c.JSON(http.StatusInternalServerError, errorResponse("internal server error"))
}

func errorResponse(msg string) gin.H {
	return gin.H{"error": msg}
}
