package models

import (
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type ProductFilter struct {
	CategoryCode  string
	PriceLessThan *decimal.Decimal
	Offset        int
	Limit         int
}

type ProductsRepository struct {
	db *gorm.DB
}

func NewProductsRepository(db *gorm.DB) *ProductsRepository {
	return &ProductsRepository{
		db: db,
	}
}

func (r *ProductsRepository) GetAllProducts() ([]Product, error) {
	var products []Product
	if err := r.db.Preload("Variants").Find(&products).Error; err != nil {
		return nil, err
	}
	return products, nil
}

func (r *ProductsRepository) GetByCode(code string) (*Product, error) {
	var product Product
	err := r.db.Preload("Category").Preload("Variants").Where("code = ?", code).First(&product).Error
	return &product, err
}

func (r *ProductsRepository) GetFiltered(filter ProductFilter) ([]Product, int64, error) {
	var products []Product
	var total int64

	// Define the base logic Filters
	query := r.db.Model(&Product{}).
		Preload("Category").
		Preload("Variants").
		Scopes(
			FilterByCategory(filter.CategoryCode),
			FilterPriceLess(filter.PriceLessThan),
		)

	// Get the total count based on filtered results
	// We use a Session here so the Count logic doesn't "pollute" the Find logic
	if err := query.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get the paginated results
	err := query.Scopes(Paginate(filter.Offset, filter.Limit)).Find(&products).Error

	return products, total, err
}

func FilterByCategory(code string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if code == "" {
			return db
		}
		return db.Joins("JOIN categories ON categories.id = products.category_id").
			Where("categories.code = ?", code)
	}
}

func FilterPriceLess(price *decimal.Decimal) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if price == nil {
			return db
		}
		return db.Where("products.price < ?", *price)
	}
}

func Paginate(offset, limit int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Offset(offset).Limit(limit)
	}
}
