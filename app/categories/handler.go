// Package categories contains the HTTP handlers for category-related operations.
package categories

import (
	"encoding/json"
	"net/http"

	"github.com/mytheresa/go-hiring-challenge/app/api"
	"github.com/mytheresa/go-hiring-challenge/models"
)

// CategoryRepository defines the contract for category operations
type CategoryRepository interface {
	All() ([]models.Category, error)
	Create(cat *models.Category) error
}

type CategoryHandler struct {
	repo CategoryRepository
}

func NewCategoryHandler(r CategoryRepository) *CategoryHandler {
	return &CategoryHandler{repo: r}
}

func (h *CategoryHandler) HandleList(w http.ResponseWriter, r *http.Request) {
	cats, err := h.repo.All()
	if err != nil {
		api.ErrorResponse(w, http.StatusInternalServerError, "Failed to fetch categories")
		return
	}
	api.OKResponse(w, cats)
}

func (h *CategoryHandler) HandleCreate(w http.ResponseWriter, r *http.Request) {
	var cat models.Category
	if err := json.NewDecoder(r.Body).Decode(&cat); err != nil {
		api.ErrorResponse(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}

	if err := h.repo.Create(&cat); err != nil {
		api.ErrorResponse(w, http.StatusInternalServerError, "Could not save category")
		return
	}
	api.OKResponse(w, cat)
}
