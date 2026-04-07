package handlers

import (
	"net/http"
	"pvz-service/internal/api"
	"pvz-service/internal/middleware"
	"pvz-service/internal/service"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	svc *service.Service
}

func NewHandler(svc *service.Service) *Handler {
	return &Handler{
		svc: svc,
	}
}

// DummyLogin - использует middleware.GenerateToken
func (h *Handler) DummyLogin(c *gin.Context) {
	var req struct {
		Role string `json:"role"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid request"})
		return
	}

	var role api.UserRole
	switch req.Role {
	case "employee":
		role = api.UserRoleEmployee
	case "moderator":
		role = api.UserRoleModerator
	default:
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid role"})
		return
	}

	// Используем middleware для генерации токена
	token, err := middleware.GenerateToken(uuid.New(), role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.String(http.StatusOK, token)
}

// Register
func (h *Handler) Register(c *gin.Context) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Role     string `json:"role"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid request"})
		return
	}

	var role api.UserRole
	switch req.Role {
	case "employee":
		role = api.UserRoleEmployee
	case "moderator":
		role = api.UserRoleModerator
	default:
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid role"})
		return
	}

	user, err := h.svc.Register(c.Request.Context(), req.Email, req.Password, role)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, user)
}

// Login - использует middleware.GenerateToken
func (h *Handler) Login(c *gin.Context) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid request"})
		return
	}

	user, err := h.svc.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": err.Error()})
		return
	}

	// Используем middleware для генерации токена
	token, err := middleware.GenerateToken(*user.Id, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.String(http.StatusOK, token)
}

// CreatePVZ
func (h *Handler) CreatePVZ(c *gin.Context) {
	var req struct {
		City string `json:"city"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid request"})
		return
	}

	pvz, err := h.svc.CreatePVZ(c.Request.Context(), api.PVZCity(req.City))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, pvz)
}

// GetPVZList
func (h *Handler) GetPVZList(c *gin.Context) {
	var startDate, endDate *time.Time

	if sd := c.Query("startDate"); sd != "" {
		t, err := time.Parse(time.RFC3339, sd)
		if err == nil {
			startDate = &t
		}
	}

	if ed := c.Query("endDate"); ed != "" {
		t, err := time.Parse(time.RFC3339, ed)
		if err == nil {
			endDate = &t
		}
	}

	page := 1
	if p := c.Query("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	limit := 10
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
			if limit > 30 {
				limit = 30
			}
		}
	}

	pvzs, err := h.svc.GetPVZWithDetails(c.Request.Context(), startDate, endDate, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, pvzs)
}

// CreateReception
func (h *Handler) CreateReception(c *gin.Context) {
	var req struct {
		PvzId string `json:"pvzId"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid request"})
		return
	}

	pvzID, err := uuid.Parse(req.PvzId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid pvzId"})
		return
	}

	reception, err := h.svc.CreateReception(c.Request.Context(), pvzID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, reception)
}

// AddProduct
func (h *Handler) AddProduct(c *gin.Context) {
	var req struct {
		Type  string `json:"type"`
		PvzId string `json:"pvzId"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid request"})
		return
	}

	pvzID, err := uuid.Parse(req.PvzId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid pvzId"})
		return
	}

	var productType api.ProductType
	switch req.Type {
	case "электроника":
		productType = api.ProductTypeЭлектроника
	case "одежда":
		productType = api.ProductTypeОдежда
	case "обувь":
		productType = api.ProductTypeОбувь
	default:
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid product type"})
		return
	}

	product, err := h.svc.AddProduct(c.Request.Context(), pvzID, productType)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, product)
}

// CloseReception
func (h *Handler) CloseReception(c *gin.Context) {
	pvzID, err := uuid.Parse(c.Param("pvzId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid pvzId"})
		return
	}

	reception, err := h.svc.CloseReception(c.Request.Context(), pvzID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, reception)
}

// DeleteLastProduct
func (h *Handler) DeleteLastProduct(c *gin.Context) {
	pvzID, err := uuid.Parse(c.Param("pvzId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid pvzId"})
		return
	}

	if err := h.svc.DeleteLastProduct(c.Request.Context(), pvzID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "product deleted"})
}
