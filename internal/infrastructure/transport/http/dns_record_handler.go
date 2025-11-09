package http

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"

	"internal-dns/internal/domain"
	"internal-dns/internal/infrastructure/transport/http/middleware"
	"internal-dns/internal/repository"
	"internal-dns/internal/usecase"
)

type DNSRecordHandler struct {
	dnsUC usecase.DNSRecordUseCase
}

func NewDNSRecordHandler(dnsUC usecase.DNSRecordUseCase) *DNSRecordHandler {
	return &DNSRecordHandler{dnsUC: dnsUC}
}

type CreateDNSRecordRequest struct {
	DomainName string `json:"domain_name" validate:"required"`
	Type       string `json:"type" validate:"required"`
	Value      string `json:"value" validate:"required"`
}

type UpdateDNSRecordRequest struct {
	DomainName string `json:"domain_name" validate:"required"`
	Type       string `json:"type" validate:"required"`
	Value      string `json:"value" validate:"required"`
}

type DNSRecordResponse struct {
	ID         int64     `json:"id"`
	UserID     int64     `json:"user_id"`
	DomainName string    `json:"domain_name"`
	Type       string    `json:"type"`
	Value      string    `json:"value"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func toDNSRecordResponse(record *domain.DNSRecord) DNSRecordResponse {
	return DNSRecordResponse{
		ID:         record.ID,
		UserID:     record.UserID,
		DomainName: record.DomainName,
		Type:       string(record.Type),
		Value:      record.Value,
		CreatedAt:  record.CreatedAt,
		UpdatedAt:  record.UpdatedAt,
	}
}

func (h *DNSRecordHandler) CreateRecord(c echo.Context) error {
	user, ok := c.Get(middleware.UserContextKey).(*domain.User)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid user context"})
	}

	var req CreateDNSRecordRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request payload"})
	}

	record, err := h.dnsUC.CreateRecord(c.Request().Context(), user.ID, req.DomainName, req.Value, domain.RecordType(req.Type))
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrDuplicateDomainName):
			return c.JSON(http.StatusConflict, map[string]string{"error": err.Error()})
		case errors.Is(err, domain.ErrInvalidDomainName), errors.Is(err, domain.ErrInvalidRecordType), errors.Is(err, domain.ErrInvalidRecordValue):
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		default:
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to create record"})
		}
	}

	return c.JSON(http.StatusCreated, toDNSRecordResponse(record))
}

func (h *DNSRecordHandler) GetRecord(c echo.Context) error {
	user, ok := c.Get(middleware.UserContextKey).(*domain.User)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid user context"})
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid record ID"})
	}

	record, err := h.dnsUC.GetRecordByID(c.Request().Context(), user.ID, id)
	if err != nil {
		if errors.Is(err, repository.ErrDNSRecordNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to get record"})
	}

	return c.JSON(http.StatusOK, toDNSRecordResponse(record))
}

func (h *DNSRecordHandler) ListRecords(c echo.Context) error {
	user, ok := c.Get(middleware.UserContextKey).(*domain.User)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid user context"})
	}

	page, _ := strconv.Atoi(c.QueryParam("page"))
	pageSize, _ := strconv.Atoi(c.QueryParam("pageSize"))

	records, total, err := h.dnsUC.ListRecordsByUser(c.Request().Context(), user.ID, page, pageSize)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to list records"})
	}

	resp := make([]DNSRecordResponse, len(records))
	for i, r := range records {
		resp[i] = toDNSRecordResponse(r)
	}

	c.Response().Header().Set("X-Total-Count", strconv.Itoa(total))
	return c.JSON(http.StatusOK, resp)
}

func (h *DNSRecordHandler) UpdateRecord(c echo.Context) error {
	user, ok := c.Get(middleware.UserContextKey).(*domain.User)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid user context"})
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid record ID"})
	}

	var req UpdateDNSRecordRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request payload"})
	}

	record, err := h.dnsUC.UpdateRecord(c.Request().Context(), user.ID, id, req.DomainName, req.Value, domain.RecordType(req.Type))
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrDNSRecordNotFound):
			return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
		case errors.Is(err, repository.ErrDuplicateDomainName):
			return c.JSON(http.StatusConflict, map[string]string{"error": err.Error()})
		case errors.Is(err, domain.ErrInvalidDomainName), errors.Is(err, domain.ErrInvalidRecordType), errors.Is(err, domain.ErrInvalidRecordValue):
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		default:
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to update record"})
		}
	}

	return c.JSON(http.StatusOK, toDNSRecordResponse(record))
}

func (h *DNSRecordHandler) DeleteRecord(c echo.Context) error {
	user, ok := c.Get(middleware.UserContextKey).(*domain.User)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid user context"})
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid record ID"})
	}

	err = h.dnsUC.DeleteRecord(c.Request().Context(), user.ID, id)
	if err != nil {
		if errors.Is(err, repository.ErrDNSRecordNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to delete record"})
	}

	return c.NoContent(http.StatusNoContent)
}

