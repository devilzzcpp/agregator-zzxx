package common

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type ErrorResponse struct {
	Message string `json:"message" example:"Запись не найдена"`
}

type ValidationErrorResponse struct {
	Errors  map[string]string `json:"errors"`
	Message string            `json:"message" example:"Ошибка валидации"`
}

type ValidationErrorFieldsDoc struct {
	Request     string `json:"request,omitempty" example:"некорректные входные данные"`
	ServiceName string `json:"service_name,omitempty" example:"поле обязательно"`
	Price       string `json:"price,omitempty" example:"должен быть числом"`
	UserID      string `json:"user_id,omitempty" example:"должен быть UUID"`
	StartDate   string `json:"start_date,omitempty" example:"ожидается формат MM-YYYY"`
	EndDate     string `json:"end_date,omitempty" example:"ожидается формат MM-YYYY или null"`
	From        string `json:"from,omitempty" example:"поле обязательно"`
	To          string `json:"to,omitempty" example:"поле обязательно"`
}

type ValidationErrorResponseDoc struct {
	Errors  ValidationErrorFieldsDoc `json:"errors"`
	Message string                   `json:"message" example:"Ошибка валидации"`
}

func JSONValidationError(c *gin.Context, errs map[string]string) {
	c.JSON(http.StatusBadRequest, ValidationErrorResponse{
		Message: "Ошибка валидации",
		Errors:  errs,
	})
}

func JSONError(c *gin.Context, status int, msg string) {
	c.JSON(status, ErrorResponse{
		Message: msg,
	})
}
