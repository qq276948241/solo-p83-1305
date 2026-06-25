package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"groupbuy/service"
)

type SortingHandler struct {
	DB *gorm.DB
}

func (h *SortingHandler) Summary(c *gin.Context) {
	date := c.Query("date")
	if date == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "date 参数必填"})
		return
	}

	results, err := service.CalcProductSummary(h.DB, date)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "汇总查询失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"date": date, "data": results})
}

func (h *SortingHandler) Export(c *gin.Context) {
	date := c.Query("date")
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}

	f, err := service.GeneratePickupExcel(h.DB, date)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "导出失败"})
		return
	}

	buf, err := f.WriteToBuffer()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成 Excel 失败"})
		return
	}

	fileName := fmt.Sprintf("提货单_%s.xlsx", date)
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", buf.Bytes())
}
