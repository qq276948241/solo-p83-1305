package handler

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"groupbuy/model"
)

type OrderHandler struct {
	DB *gorm.DB
}

type CreateOrderReq struct {
	ProductID     uint64 `json:"product_id" binding:"required"`
	PickupPointID uint64 `json:"pickup_point_id" binding:"required"`
	Quantity      uint32 `json:"quantity" binding:"required,min=1"`
	CustomerName  string `json:"customer_name"`
	CustomerPhone string `json:"customer_phone"`
}

func (h *OrderHandler) Create(c *gin.Context) {
	var req CreateOrderReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var product model.Product
	if err := h.DB.First(&product, req.ProductID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "商品不存在"})
		return
	}

	var pp model.PickupPoint
	if err := h.DB.First(&pp, req.PickupPointID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "提货点不存在"})
		return
	}

	today := time.Now().Format("2006-01-02")
	orderNo := generateOrderNo()
	totalPrice := product.Price * float64(req.Quantity)

	order := model.Order{
		OrderNo:       orderNo,
		ProductID:     req.ProductID,
		PickupPointID: req.PickupPointID,
		Quantity:      req.Quantity,
		UnitPrice:     product.Price,
		TotalPrice:    totalPrice,
		OrderDate:     today,
		CustomerName:  req.CustomerName,
		CustomerPhone: req.CustomerPhone,
		Status:        1,
	}

	if err := h.DB.Create(&order).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建订单失败"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": order})
}

func (h *OrderHandler) List(c *gin.Context) {
	date := c.Query("date")
	status := c.Query("status")

	query := h.DB.Model(&model.Order{}).Preload("Product").Preload("PickupPoint")
	if date != "" {
		query = query.Where("order_date = ?", date)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	var orders []model.Order
	if err := query.Order("id DESC").Find(&orders).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询订单失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": orders})
}

func (h *OrderHandler) UpdateStatus(c *gin.Context) {
	id := c.Param("id")
	var body struct {
		Status int8 `json:"status" binding:"required,oneof=1 2 3 4"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.DB.Model(&model.Order{}).Where("id = ?", id).Update("status", body.Status).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新状态失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "状态已更新"})
}

func generateOrderNo() string {
	now := time.Now()
	r := rand.New(rand.NewSource(now.UnixNano()))
	return fmt.Sprintf("GB%s%04d", now.Format("20060102150405"), r.Intn(10000))
}
