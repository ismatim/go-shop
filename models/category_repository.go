// Package models contains the data models and repositories for the application.
package models

import "gorm.io/gorm"

type CategoryRepository struct {
	db *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) *CategoryRepository {
	return &CategoryRepository{
		db: db,
	}
}

func (r *CategoryRepository) All() ([]Category, error) {
	var cats []Category

	if err := r.db.Find(&cats).Error; err != nil {
		return nil, err
	}
	return cats, nil
}

func (r *CategoryRepository) Create(cat *Category) error {
	return r.db.Create(cat).Error
}
