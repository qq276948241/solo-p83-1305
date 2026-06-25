package model

import (
	"time"

	"gorm.io/gorm"
)

const (
	StatusPending   int8 = 1
	StatusSorted    int8 = 2
	StatusPickedUp  int8 = 3
	StatusCancelled int8 = 4
)

func ValidOrderStatuses() []int8 {
	return []int8{StatusPending, StatusSorted, StatusPickedUp}
}

type Supplier struct {
	ID          uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string         `gorm:"type:varchar(100);not null" json:"name"`
	StallNumber string         `gorm:"type:varchar(50);uniqueIndex;not null" json:"stall_number"`
	Contact     string         `gorm:"type:varchar(50);default:''" json:"contact"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Supplier) TableName() string { return "suppliers" }

type Product struct {
	ID         uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	Name       string         `gorm:"type:varchar(100);not null" json:"name"`
	Category   string         `gorm:"type:varchar(50);default:''" json:"category"`
	Unit       string         `gorm:"type:varchar(20);not null;default:'份'" json:"unit"`
	Price      float64        `gorm:"type:decimal(10,2);not null;default:0" json:"price"`
	SupplierID uint64         `gorm:"not null;index" json:"supplier_id"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`

	Supplier Supplier `gorm:"foreignKey:SupplierID" json:"supplier,omitempty"`
}

func (Product) TableName() string { return "products" }

type PickupPoint struct {
	ID        uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      string         `gorm:"type:varchar(100);not null" json:"name"`
	Address   string         `gorm:"type:varchar(255);default:''" json:"address"`
	Contact   string         `gorm:"type:varchar(50);default:''" json:"contact"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (PickupPoint) TableName() string { return "pickup_points" }

type Order struct {
	ID            uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	OrderNo       string         `gorm:"type:varchar(32);uniqueIndex;not null" json:"order_no"`
	ProductID     uint64         `gorm:"not null;index" json:"product_id"`
	PickupPointID uint64         `gorm:"not null;index" json:"pickup_point_id"`
	Quantity      uint32         `gorm:"type:int unsigned;not null;default:1" json:"quantity"`
	UnitPrice     float64        `gorm:"type:decimal(10,2);not null" json:"unit_price"`
	TotalPrice    float64        `gorm:"type:decimal(10,2);not null" json:"total_price"`
	OrderDate     string         `gorm:"type:date;not null;index" json:"order_date"`
	CustomerName  string         `gorm:"type:varchar(50);default:''" json:"customer_name"`
	CustomerPhone string         `gorm:"type:varchar(20);default:''" json:"customer_phone"`
	Status        int8           `gorm:"type:tinyint;not null;default:1;index" json:"status"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`

	Product     Product     `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	PickupPoint PickupPoint `gorm:"foreignKey:PickupPointID" json:"pickup_point,omitempty"`
}

func (Order) TableName() string { return "orders" }
