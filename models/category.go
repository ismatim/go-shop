package models

type Category struct {
	ID   uint   `gorm:"primaryKey" json:"id"`
	Code string `gorm:"uniqueIndex;not null" json:"code"`
	Name string `gorm:"not null" json:"name"`
}

func (p *Category) TableName() string {
	return "categories"
}
