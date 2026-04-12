package internal

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Subscribe(c *gin.Context) {
	var req SubscribeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid input: email and repo are required"})
		return
	}

	err := h.service.Subscribe(c.Request.Context(), req)
	if err == nil {
		c.JSON(http.StatusOK, MessageResponse{Message: "subscription successful, confirmation email sent"})
		return
	}

	switch {
	case errors.Is(err, ErrInvalidRepoFormat):
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	case errors.Is(err, ErrRepoNotFound):
		c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
	case errors.Is(err, ErrAlreadySubscribed):
		c.JSON(http.StatusConflict, ErrorResponse{Error: err.Error()})
	case errors.Is(err, ErrRateLimited):
		c.JSON(http.StatusServiceUnavailable, ErrorResponse{Error: "service temporarily unavailable, try again later"})
	default:
		log.Printf("subscribe error: %v", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
	}
}

func (h *Handler) Unsubscribe(c *gin.Context) {
	token := c.Param("token")

	err := h.service.Unsubscribe(c.Request.Context(), token)
	if err == nil {
		c.JSON(http.StatusOK, MessageResponse{Message: "unsubscribed successfully"})
		return
	}

	switch {
	case errors.Is(err, ErrTokenNotFound):
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "token not found"})
	default:
		log.Printf("unsubscribe error: %v", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
	}
}

func (h *Handler) GetSubscriptions(c *gin.Context) {
	email := c.Query("email")

	subs, err := h.service.GetSubscriptions(c.Request.Context(), email)
	if err == nil {
		c.JSON(http.StatusOK, subs)
		return
	}

	switch {
	case errors.Is(err, ErrInvalidEmail):
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	default:
		log.Printf("get subscriptions error: %v", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
	}
}

func (h *Handler) ConfirmSubscription(c *gin.Context) {
	token := c.Param("token")

	err := h.service.ConfirmSubscription(c.Request.Context(), token)
	if err == nil {
		c.JSON(http.StatusOK, MessageResponse{Message: "subscription confirmed successfully"})
		return
	}

	switch {
	case errors.Is(err, ErrTokenNotFound):
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "token not found"})
	default:
		log.Printf("confirm error: %v", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
	}
}
