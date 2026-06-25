package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"groupbuy/service"
)

type SupplierHandler struct {
	DB *gorm.DB
}

func (h *SupplierHandler) Reconciliation(c *gin.Context) {
	date := c.Query("date")
	if date == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "date 参数必填"})
		return
	}

	results, err := service.CalcReconciliation(h.DB, date)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "对账查询失败"})
		return
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
