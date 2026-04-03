// Package catalog
package catalog

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/mytheresa/go-hiring-challenge/app/api"
	"github.com/mytheresa/go-hiring-challenge/models"
	"github.com/shopspring/decimal"
)

type CatalogResponse struct {
	Products []ProductDTO `json:"products"`
	Total    int          `json:"total"`
}

type ProductDTO struct {
	Code     string          `json:"code"`
	Price    decimal.Decimal `json:"price"`
	Category string          `json:"category"`
	Variants []VariantDTO    `json:"variants,omitempty"`
}

type VariantDTO struct {
	SKU   string          `json:"sku"`
	Name  string          `json:"name"`
	Price decimal.Decimal `json:"price"`
}

// ProductRepository defines the specific methods the handler needs.
// This allows us to swap the real DB for a Mock in unit tests.
type ProductRepository interface {
	GetFiltered(filter models.ProductFilter) ([]models.Product, int64, error)
	GetByCode(code string) (*models.Product, error)
	GetAllProducts() ([]models.Product, error)
}

type CatalogHandler struct {
	repo ProductRepository
}

func NewCatalogHandler(r ProductRepository) *CatalogHandler {
	return &CatalogHandler{
		repo: r,
	}
}

func (h *CatalogHandler) HandleGet(w http.ResponseWriter, r *http.Request) {
	offset := 0
	limit := 10
	var priceFilter *decimal.Decimal

	// Parse query parameters
	queryOffset := r.URL.Query().Get("offset")
	queryLimit := r.URL.Query().Get("limit")
	queryPriceLessThan := r.URL.Query().Get("price_less_than")

	if queryOffset != "" {
		if parsedOffset, err := strconv.Atoi(queryOffset); err == nil {
			offset = parsedOffset
		}
	}

	if queryLimit != "" {
		if parsedLimit, err := strconv.Atoi(queryLimit); err == nil {
			limit = parsedLimit
		}
	}

	if queryPriceLessThan != "" {
		parsedPrice, err := decimal.NewFromString(queryPriceLessThan)
		if err != nil {
			http.Error(w, "Invalid price_less_than value", http.StatusBadRequest)
			return
		}
		priceFilter = &parsedPrice
	}

	// Fetch products with pagination
	filter := models.ProductFilter{Offset: offset, Limit: limit, PriceLessThan: priceFilter}
	res, total, err := h.repo.GetFiltered(filter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Map response
	products := make([]ProductDTO, len(res))
	for i, p := range res {
		products[i] = ProductDTO{
			Code:     p.Code,
			Price:    p.Price,
			Category: p.Category.Name,
		}
	}

	// Return the products as a JSON response
	w.Header().Set("Content-Type", "application/json")

	response := CatalogResponse{
		Products: products,
		Total:    int(total),
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *CatalogHandler) HandleGetByCode(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")
	p, err := h.repo.GetByCode(code)
	if err != nil {
		api.ErrorResponse(w, http.StatusNotFound, "Product not found")
		return
	}

	// Variant Price Inheritance Logic
	variants := make([]VariantDTO, len(p.Variants))
	for i, v := range p.Variants {
		price := v.Price
		if v.Price.IsZero() {
			price = p.Price
		}
		variants[i] = VariantDTO{SKU: v.SKU, Name: v.Name, Price: price}
	}

	api.OKResponse(w, ProductDTO{
		Code:     p.Code,
		Price:    p.Price,
		Category: p.Category.Name,
		Variants: variants,
	})
}
