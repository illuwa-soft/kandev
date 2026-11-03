package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kandev/kandev/internal/notifications/models"
	"github.com/kandev/kandev/internal/notifications/store"
	"github.com/kandev/kandev/internal/notifications/templates"
)

type templateRequest struct {
	Version          int     `json:"version"`
	Title            string  `json:"title"`
	Body             string  `json:"body"`
	ExpectedRevision *string `json:"expected_revision"`
	Sample           string  `json:"sample"`
}

func (r templateRequest) template() *models.NotificationTemplate {
	return &models.NotificationTemplate{Version: r.Version, Title: r.Title, Body: r.Body}
}
func (r templateRequest) revision() string {
	if r.ExpectedRevision == nil {
		return ""
	}
	return *r.ExpectedRevision
}

func readTemplateRequest(c *gin.Context) (templateRequest, bool) {
	c.Header("Cache-Control", "no-store")
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 128*1024)
	decoder := json.NewDecoder(c.Request.Body)
	decoder.DisallowUnknownFields()
	var request templateRequest
	err := decoder.Decode(&request)
	if err == nil {
		var extra any
		if decoder.Decode(&extra) != io.EOF {
			err = templates.ErrInvalid
		}
	}
	if err != nil {
		var tooLarge *http.MaxBytesError
		status := http.StatusBadRequest
		if errors.As(err, &tooLarge) {
			status = http.StatusRequestEntityTooLarge
		}
		c.JSON(status, gin.H{"error": "invalid_template_request"})
		return request, false
	}
	return request, true
}

func templateError(c *gin.Context, err error) {
	c.Header("Cache-Control", "no-store")
	switch {
	case errors.Is(err, store.ErrProviderNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "provider not found"})
	case errors.Is(err, store.ErrTemplateConflict):
		c.JSON(http.StatusConflict, gin.H{"error": "template_conflict"})
	case errors.Is(err, templates.ErrInvalid), errors.Is(err, store.ErrNotApprise):
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_template"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "template_storage_error"})
	}
}

func (h *Handlers) httpTemplateCatalog(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	result, err := h.controller.TemplateCatalog(c.Request.Context(), c.Param("id"))
	if err != nil {
		templateError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}
func (h *Handlers) httpSaveTemplate(c *gin.Context) {
	r, ok := readTemplateRequest(c)
	if !ok {
		return
	}
	result, err := h.controller.SaveTemplate(c.Request.Context(), c.Param("id"), c.Param("event"), r.revision(), r.template())
	if err != nil {
		templateError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}
func (h *Handlers) httpDeleteTemplate(c *gin.Context) {
	r, ok := readTemplateRequest(c)
	if !ok {
		return
	}
	err := h.controller.DeleteTemplate(c.Request.Context(), c.Param("id"), c.Param("event"), r.revision())
	if err != nil {
		templateError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
func (h *Handlers) httpPreviewTemplate(c *gin.Context) {
	r, ok := readTemplateRequest(c)
	if !ok {
		return
	}
	result, err := h.controller.PreviewTemplate(c.Request.Context(), c.Param("id"), c.Param("event"), r.Sample, r.template())
	if err != nil {
		templateError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}
