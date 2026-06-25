package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type SortingHandler struct {
	DB *gorm.DB
}

type ProductSummary struct {
	ProductID   uint64  `json:"product_id"`
	ProductName string  `json:"product_name"`
	Category    string  `json:"category"`
	Unit        string  `json:"unit"`
	TotalQty    uint32  `json:"total_qty"`
	TotalAmount float64 `json:"total_amount"`
}

func (h *SortingHandler) Summary(c *gin.Context) {
	date := c.Query("date")
	if date == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "date 参数必填"})
		return
	}

	var results []ProductSummary
	err := h.DB.Model(&struct {
		ProductID   uint64
		TotalQty    uint32
		TotalAmount float64
	}{}).
		Table("orders").
		Select("orders.product_id, SUM(orders.quantity) as total_qty, SUM(orders.total_price) as total_amount").
		Where("order_date = ?", date).
		Group("orders.product_id").
		Scan(&results).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "汇总查询失败"})
		return
	}

	for i := range results {
		var pName, pCat, pUnit string
		h.DB.Table("products").
			Select("name, category, unit").
			Where("id = ?", results[i].ProductID).
			Row().Scan(&pName, &pCat, &pUnit)
		results[i].ProductName = pName
		results[i].Category = pCat
		results[i].Unit = pUnit
	}

	c.JSON(http.StatusOK, gin.H{"date": date, "data": results})
}
