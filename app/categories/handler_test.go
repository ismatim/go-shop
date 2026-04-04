package categories

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mytheresa/go-hiring-challenge/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockCategoryRepository is a mock type for the CategoryRepository interface
type MockCategoryRepository struct {
	mock.Mock
}

func (m *MockCategoryRepository) All() ([]models.Category, error) {
	args := m.Called()
	return args.Get(0).([]models.Category), args.Error(1)
}

func (m *MockCategoryRepository) Create(cat *models.Category) error {
	args := m.Called(cat)
	return args.Error(0)
}

func TestHandleList(t *testing.T) {
	t.Run("successful list retrieval", func(t *testing.T) {
		repo := new(MockCategoryRepository)
		handler := NewCategoryHandler(repo)

		mockCats := []models.Category{
			{Code: "clothing", Name: "Clothing"},
			{Code: "shoes", Name: "Shoes"},
		}

		repo.On("All").Return(mockCats, nil)

		req := httptest.NewRequest("GET", "/categories", nil)
		w := httptest.NewRecorder()

		handler.HandleList(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp []models.Category
		err := json.NewDecoder(w.Body).Decode(&resp)
		assert.NoError(t, err)
		assert.Len(t, resp, 2)
		assert.Equal(t, "clothing", resp[0].Code)
	})

	t.Run("repository error during list", func(t *testing.T) {
		repo := new(MockCategoryRepository)
		handler := NewCategoryHandler(repo)

		repo.On("All").Return([]models.Category{}, errors.New("db error"))

		req := httptest.NewRequest("GET", "/categories", nil)
		w := httptest.NewRecorder()

		handler.HandleList(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Failed to fetch categories")
	})
}

func TestHandleCreate(t *testing.T) {
	t.Run("successful category creation", func(t *testing.T) {
		repo := new(MockCategoryRepository)
		handler := NewCategoryHandler(repo)

		newCat := models.Category{Code: "tech", Name: "Technology"}

		// We use mock.MatchedBy to ensure the pointer data matches our expectations
		repo.On("Create", mock.MatchedBy(func(c *models.Category) bool {
			return c.Code == "tech" && c.Name == "Technology"
		})).Return(nil)

		body, _ := json.Marshal(newCat)
		req := httptest.NewRequest("POST", "/categories", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		handler.HandleCreate(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		repo.AssertExpectations(t)
	})

	t.Run("invalid JSON body", func(t *testing.T) {
		repo := new(MockCategoryRepository)
		handler := NewCategoryHandler(repo)

		req := httptest.NewRequest("POST", "/categories", bytes.NewBufferString("{invalid json}"))
		w := httptest.NewRecorder()

		handler.HandleCreate(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid JSON body")
	})

	t.Run("repository error during create", func(t *testing.T) {
		repo := new(MockCategoryRepository)
		handler := NewCategoryHandler(repo)

		newCat := models.Category{Code: "fail", Name: "Failure"}
		repo.On("Create", mock.Anything).Return(errors.New("db failure"))

		body, _ := json.Marshal(newCat)
		req := httptest.NewRequest("POST", "/categories", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		handler.HandleCreate(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Could not save category")
	})
}
