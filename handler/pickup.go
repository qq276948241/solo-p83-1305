package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"groupbuy/service"
)

type PickupHandler struct {
	DB *gorm.DB
}

func (h *PickupHandler) List(c *gin.Context) {
	date := c.Query("date")
	if date == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "date 参数必填"})
		return
	}

	results, err := service.CalcPickupSlips(h.DB, date)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "提货单查询失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"date": date, "data": results})
}
