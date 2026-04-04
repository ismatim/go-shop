package catalog

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mytheresa/go-hiring-challenge/models"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// --- Helper function to create decimal from string for testing ---
func d(s string) decimal.Decimal {
	v, _ := decimal.NewFromString(s)
	return v
}

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) GetFiltered(filter models.ProductFilter) ([]models.Product, int64, error) {
	args := m.Called(filter)
	return args.Get(0).([]models.Product), int64(args.Int(1)), args.Error(2)
}

func (m *MockRepository) GetByCode(code string) (*models.Product, error) {
	args := m.Called(code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Product), args.Error(1)
}

func (m *MockRepository) GetAllProducts() ([]models.Product, error) {
	args := m.Called()
	return args.Get(0).([]models.Product), args.Error(1)
}

func TestHandleGet(t *testing.T) {
	t.Run("successful response with default pagination", func(t *testing.T) {
		repo := new(MockRepository)
		handler := NewCatalogHandler(repo)

		mockProducts := []models.Product{
			{
				Code:     "PROD1",
				Price:    d("10.50"),
				Category: models.Category{Name: "Shoes"},
			},
			{
				Code:     "PROD2",
				Price:    d("30.50"),
				Category: models.Category{Name: "Shoes"},
			},
		}

		// Expectations: Default limit 10, offset 0
		repo.On("GetFiltered", models.ProductFilter{Offset: 0, Limit: 10}).
			Return(mockProducts, 2, nil)

		req := httptest.NewRequest("GET", "/catalog", nil)
		w := httptest.NewRecorder()

		handler.HandleGet(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		var resp CatalogResponse
		err := json.NewDecoder(w.Body).Decode(&resp)
		assert.NoError(t, err)
		assert.Equal(t, 2, resp.Total)
		assert.Equal(t, "PROD1", resp.Products[0].Code)
		assert.True(t, d("10.50").Equal(resp.Products[0].Price))
	})

	t.Run("handles custom pagination parameters", func(t *testing.T) {
		repo := new(MockRepository)
		handler := NewCatalogHandler(repo)

		repo.On("GetFiltered", models.ProductFilter{Offset: 20, Limit: 5}).
			Return([]models.Product{}, 100, nil)

		req := httptest.NewRequest("GET", "/catalog?offset=20&limit=5", nil)
		w := httptest.NewRecorder()

		handler.HandleGet(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		repo.AssertExpectations(t)
	})

	t.Run("returns 500 on repository error", func(t *testing.T) {
		repo := new(MockRepository)
		handler := NewCatalogHandler(repo)

		repo.On("GetFiltered", mock.Anything).
			Return([]models.Product{}, 0, errors.New("db error"))

		req := httptest.NewRequest("GET", "/catalog", nil)
		w := httptest.NewRecorder()

		handler.HandleGet(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestHandleGetByCode(t *testing.T) {
	t.Run("successful details with price inheritance", func(t *testing.T) {
		repo := new(MockRepository)
		handler := NewCatalogHandler(repo)

		prodPrice := d("100.00")
		mockProduct := &models.Product{
			Code:     "PROD_DETAIL",
			Price:    prodPrice,
			Category: models.Category{Name: "Clothing"},
			Variants: []models.Variant{
				{SKU: "V1", Name: "Small", Price: decimal.Zero}, // Should inherit 100.00
				{SKU: "V2", Name: "Large", Price: d("120.50")},  // Should keep 120.50
			},
		}

		repo.On("GetByCode", "PROD_DETAIL").Return(mockProduct, nil)

		// Create request and manually set PathValue (Go 1.22+ feature)
		req := httptest.NewRequest("GET", "/catalog/PROD_DETAIL", nil)
		req.SetPathValue("code", "PROD_DETAIL")
		w := httptest.NewRecorder()

		handler.HandleGetByCode(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp ProductDTO
		json.NewDecoder(w.Body).Decode(&resp)

		// Assert Variant 1 inherited price
		assert.Equal(t, "100.00", resp.Variants[0].Price.StringFixed(2))
		// Assert Variant 2 kept its own price
		assert.Equal(t, "120.5", resp.Variants[1].Price.String())
	})

	t.Run("returns 404 when product not found", func(t *testing.T) {
		repo := new(MockRepository)
		handler := NewCatalogHandler(repo)

		repo.On("GetByCode", "NONEXISTENT").Return(nil, errors.New("not found"))

		req := httptest.NewRequest("GET", "/catalog/NONEXISTENT", nil)
		req.SetPathValue("code", "NONEXISTENT")
		w := httptest.NewRecorder()

		handler.HandleGetByCode(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}
