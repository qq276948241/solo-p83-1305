package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type SupplierHandler struct {
	DB *gorm.DB
}

type SupplierReconciliation struct {
	SupplierID   uint64  `json:"supplier_id"`
	SupplierName string  `json:"supplier_name"`
	StallNumber  string  `json:"stall_number"`
	TotalAmount  float64 `json:"total_amount"`
	TotalQty     uint32  `json:"total_qty"`
	Details      []SupplierDetail `json:"details"`
}

type SupplierDetail struct {
	ProductID   uint64  `json:"product_id"`
	ProductName string  `json:"product_name"`
	Quantity    uint32  `json:"quantity"`
	TotalAmount float64 `json:"total_amount"`
}

func (h *SupplierHandler) Reconciliation(c *gin.Context) {
	date := c.Query("date")
	if date == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "date 参数必填"})
		return
	}

	type productAgg struct {
		ProductID   uint64
		TotalQty    uint32
		TotalAmount float64
	}

	var aggs []productAgg
	h.DB.Table("orders").
		Select("product_id, SUM(quantity) as total_qty, SUM(total_price) as total_amount").
		Where("order_date = ?", date).
		Group("product_id").
		Scan(&aggs)

	supplierMap := make(map[uint64]*SupplierReconciliation)

	for _, agg := range aggs {
		var productName string
		var supplierID uint64
		h.DB.Table("products").Select("name, supplier_id").Where("id = ?", agg.ProductID).Row().Scan(&productName, &supplierID)

		if supplierMap[supplierID] == nil {
			var sName, sStall string
			h.DB.Table("suppliers").Select("name, stall_number").Where("id = ?", supplierID).Row().Scan(&sName, &sStall)
			supplierMap[supplierID] = &SupplierReconciliation{
				SupplierID:   supplierID,
				SupplierName: sName,
				StallNumber:  sStall,
			}
		}

		supplierMap[supplierID].TotalAmount += agg.TotalAmount
		supplierMap[supplierID].TotalQty += agg.TotalQty
		supplierMap[supplierID].Details = append(supplierMap[supplierID].Details, SupplierDetail{
			ProductID:   agg.ProductID,
			ProductName: productName,
			Quantity:    agg.TotalQty,
			TotalAmount: agg.TotalAmount,
		})
	}

	var results []SupplierReconciliation
	for _, v := range supplierMap {
		results = append(results, *v)
	}

	c.JSON(http.StatusOK, gin.H{"date": date, "data": results})
}

func (h *SupplierHandler) List(c *gin.Context) {
	var suppliers []struct {
		ID          uint64 `json:"id"`
		Name        string `json:"name"`
		StallNumber string `json:"stall_number"`
		Contact     string `json:"contact"`
	}
	h.DB.Table("suppliers").Find(&suppliers)
	c.JSON(http.StatusOK, gin.H{"data": suppliers})
}

func (h *SupplierHandler) ListProducts(c *gin.Context) {
	var products []struct {
		ID         uint64  `json:"id"`
		Name       string  `json:"name"`
		Category   string  `json:"category"`
		Unit       string  `json:"unit"`
		Price      float64 `json:"price"`
		SupplierID uint64  `json:"supplier_id"`
	}
	h.DB.Table("products").Find(&products)
	c.JSON(http.StatusOK, gin.H{"data": products})
}

func (h *SupplierHandler) ListPickupPoints(c *gin.Context) {
	var points []struct {
		ID      uint64 `json:"id"`
		Name    string `json:"name"`
		Address string `json:"address"`
		Contact string `json:"contact"`
	}
	h.DB.Table("pickup_points").Find(&points)
	c.JSON(http.StatusOK, gin.H{"data": points})
}
