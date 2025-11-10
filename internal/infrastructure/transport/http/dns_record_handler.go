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
	DomainName string `json:"domainName"` // Changed to camelCase
	Type       string `json:"type"`
	Value      string `json:"value"`
}

type UpdateDNSRecordRequest struct {
	DomainName string `json:"domainName"` // Changed to camelCase
	Type       string `json:"type"`
	Value      string `json:"value"`
}

type DNSRecordResponse struct {
	ID         int64     `json:"id"`
	UserID     int64     `json:"userId"`     // Changed to camelCase
	DomainName string    `json:"domainName"` // Changed to camelCase
	Type       string    `json:"type"`
	Value      string    `json:"value"`
	CreatedAt  time.Time `json:"createdAt"`  // Changed to camelCase
	UpdatedAt  time.Time `json:"updatedAt"`  // Changed to camelCase
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

// CreateRecord godoc
// @Summary Create a DNS record
// @Description Creates a new DNS record for the authenticated user.
// @Tags dns-records
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param record body CreateDNSRecordRequest true "DNS Record"
// @Success 201 {object} DNSRecordResponse
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 409 {object} map[string]string "Duplicate domain name"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /dns-records [post]
func (h *DNSRecordHandler) CreateRecord(c echo.Context) error {
	user, ok := c.Get(string(middleware.UserContextKey)).(*domain.User)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid user in context"}) // Refined error message
	}

	var req CreateDNSRecordRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
	}

	record, err := h.dnsUC.CreateRecord(c.Request().Context(), user.ID, req.DomainName, req.Value, domain.RecordType(req.Type))
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrDuplicateDomainName):
			return c.JSON(http.StatusConflict, map[string]string{"error": "Domain name already exists"}) // Refined error message
		case errors.Is(err, domain.ErrInvalidDomainName), errors.Is(err, domain.ErrInvalidRecordType), errors.Is(err, domain.ErrInvalidRecordValue):
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		default:
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create DNS record"}) // Refined error message
		}
	}

	return c.JSON(http.StatusCreated, toDNSRecordResponse(record))
}

// GetRecord godoc
// @Summary Get a DNS record by ID
// @Description Retrieves a single DNS record by its ID.
// @Tags dns-records
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Record ID"
// @Success 200 {object} DNSRecordResponse
// @Failure 400 {object} map[string]string "Invalid ID"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Record not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /dns-records/{id} [get]
func (h *DNSRecordHandler) GetRecord(c echo.Context) error {
	user, ok := c.Get(string(middleware.UserContextKey)).(*domain.User)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid user in context"}) // Refined error message
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid record ID"}) // Refined error message
	}

	record, err := h.dnsUC.GetRecordByID(c.Request().Context(), user.ID, id)
	if err != nil {
		if errors.Is(err, repository.ErrDNSRecordNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Record not found or not owned by user"}) // Refined error message
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve DNS record"}) // Refined error message
	}

	return c.JSON(http.StatusOK, toDNSRecordResponse(record))
}

// ListRecords godoc
// @Summary List user's DNS records
// @Description Retrieves a paginated list of DNS records for the authenticated user.
// @Tags dns-records
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param pageSize query int false "Page size" default(10)
// @Success 200 {array} DNSRecordResponse
// @Header 200 {string} X-Total-Count "Total number of records"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /dns-records [get]
func (h *DNSRecordHandler) ListRecords(c echo.Context) error {
	user, ok := c.Get(string(middleware.UserContextKey)).(*domain.User)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid user in context"}) // Refined error message
	}

	page, _ := strconv.Atoi(c.QueryParam("page"))
	pageSize, _ := strconv.Atoi(c.QueryParam("pageSize"))
	if page < 1 { // Added default page/pageSize logic
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	records, total, err := h.dnsUC.ListRecordsByUser(c.Request().Context(), user.ID, page, pageSize)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to list DNS records"}) // Refined error message
	}

	var resp []DNSRecordResponse // Changed to use append for clarity
	for _, r := range records {
		resp = append(resp, toDNSRecordResponse(r))
	}

	c.Response().Header().Set("X-Total-Count", strconv.Itoa(total))
	return c.JSON(http.StatusOK, resp)
}

// UpdateRecord godoc
// @Summary Update a DNS record
// @Description Updates an existing DNS record.
// @Tags dns-records
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Record ID"
// @Param record body UpdateDNSRecordRequest true "Updated DNS Record"
// @Success 200 {object} DNSRecordResponse
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Record not found"
// @Failure 409 {object} map[string]string "Duplicate domain name"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /dns-records/{id} [put]
func (h *DNSRecordHandler) UpdateRecord(c echo.Context) error {
	user, ok := c.Get(string(middleware.UserContextKey)).(*domain.User)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid user in context"}) // Refined error message
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid record ID"}) // Refined error message
	}

	var req UpdateDNSRecordRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
	}

	record, err := h.dnsUC.UpdateRecord(c.Request().Context(), user.ID, id, req.DomainName, req.Value, domain.RecordType(req.Type))
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrDNSRecordNotFound):
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Record not found or not owned by user"}) // Refined error message
		case errors.Is(err, repository.ErrDuplicateDomainName):
			return c.JSON(http.StatusConflict, map[string]string{"error": "Domain name already exists"}) // Refined error message
		case errors.Is(err, domain.ErrInvalidDomainName), errors.Is(err, domain.ErrInvalidRecordType), errors.Is(err, domain.ErrInvalidRecordValue):
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		default:
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update DNS record"}) // Refined error message
		}
	}

	return c.JSON(http.StatusOK, toDNSRecordResponse(record))
}

// DeleteRecord godoc
// @Summary Delete a DNS record
// @Description Deletes a DNS record by its ID.
// @Tags dns-records
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Record ID"
// @Success 204 "No Content"
// @Failure 400 {object} map[string]string "Invalid ID"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Record not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /dns-records/{id} [delete]
func (h *DNSRecordHandler) DeleteRecord(c echo.Context) error {
	user, ok := c.Get(string(middleware.UserContextKey)).(*domain.User)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid user in context"}) // Refined error message
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid record ID"}) // Refined error message
	}

	err = h.dnsUC.DeleteRecord(c.Request().Context(), user.ID, id)
	if err != nil {
		if errors.Is(err, repository.ErrDNSRecordNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Record not found or not owned by user"}) // Refined error message
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete DNS record"}) // Refined error message
	}

	return c.NoContent(http.StatusNoContent)
}

