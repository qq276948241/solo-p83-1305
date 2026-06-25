package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"groupbuy/model"
)

type PickupHandler struct {
	DB *gorm.DB
}

type PickupSlipItem struct {
	OrderID     uint64  `json:"order_id"`
	OrderNo     string  `json:"order_no"`
	ProductName string  `json:"product_name"`
	Unit        string  `json:"unit"`
	Quantity    uint32  `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
	TotalPrice  float64 `json:"total_price"`
	CustomerName  string  `json:"customer_name"`
	CustomerPhone string  `json:"customer_phone"`
}

type PickupSlip struct {
	PickupPointID   uint64           `json:"pickup_point_id"`
	PickupPointName string           `json:"pickup_point_name"`
	Address         string           `json:"address"`
	Items           []PickupSlipItem `json:"items"`
	GrandTotal      float64          `json:"grand_total"`
}

func (h *PickupHandler) List(c *gin.Context) {
	date := c.Query("date")
	if date == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "date 参数必填"})
		return
	}

	var points []model.PickupPoint
	h.DB.Find(&points)

	var slips []PickupSlip
	for _, pp := range points {
		var orders []model.Order
		h.DB.Preload("Product").
			Where("order_date = ? AND pickup_point_id = ?", date, pp.ID).
			Find(&orders)

		if len(orders) == 0 {
			continue
		}

		slip := PickupSlip{
			PickupPointID:   pp.ID,
			PickupPointName: pp.Name,
			Address:         pp.Address,
		}

		var grandTotal float64
		for _, o := range orders {
			slip.Items = append(slip.Items, PickupSlipItem{
				OrderID:       o.ID,
				OrderNo:       o.OrderNo,
				ProductName:   o.Product.Name,
				Unit:          o.Product.Unit,
				Quantity:      o.Quantity,
				UnitPrice:     o.UnitPrice,
				TotalPrice:    o.TotalPrice,
				CustomerName:  o.CustomerName,
				CustomerPhone: o.CustomerPhone,
			})
			grandTotal += o.TotalPrice
		}
		slip.GrandTotal = grandTotal
		slips = append(slips, slip)
	}

	c.JSON(http.StatusOK, gin.H{"date": date, "data": slips})
}
