package model

type Item struct {
	ID             uint   `gorm:"primaryKey"`
	Name           string `gorm:"not null"`
	Stock          int    `gorm:"not null"`
	AvailableStock int    `gorm:"not null"`
	ReservedStock  int    `gorm:"not null"`
}

type ProcessedOrder struct {
	ID      uint `gorm:"primaryKey"`
	OrderID uint `gorm:"not null"`
}
