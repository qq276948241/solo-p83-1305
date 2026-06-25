package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
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

type exportRow struct {
	PickupPointID   uint64
	PickupPointName string
	ProductName     string
	Category        string
	Unit            string
	TotalQty        uint32
}

func (h *SortingHandler) Export(c *gin.Context) {
	date := c.Query("date")
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}

	var rows []exportRow
	err := h.DB.Table("orders").
		Select("orders.pickup_point_id, pickup_points.name as pickup_point_name, "+
			"products.name as product_name, products.category, products.unit, "+
			"SUM(orders.quantity) as total_qty").
		Joins("LEFT JOIN pickup_points ON pickup_points.id = orders.pickup_point_id").
		Joins("LEFT JOIN products ON products.id = orders.product_id").
		Where("orders.order_date = ? AND orders.status IN ?", date, []int8{1, 2}).
		Group("orders.pickup_point_id, orders.product_id").
		Order("orders.pickup_point_id").
		Scan(&rows).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "导出查询失败"})
		return
	}

	f := excelize.NewFile()
	defaultSheet := f.GetSheetName(0)

	grouped := make(map[uint64][]exportRow)
	pointName := make(map[uint64]string)
	for _, r := range rows {
		grouped[r.PickupPointID] = append(grouped[r.PickupPointID], r)
		pointName[r.PickupPointID] = r.PickupPointName
	}

	boldStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 12},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#E8F0FE"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})

	if len(grouped) == 0 {
		sheet := "汇总"
		f.SetSheetName(defaultSheet, sheet)
		f.SetCellValue(sheet, "A1", "当天暂无待分拣订单")
	} else {
		first := true
		for pid, items := range grouped {
			sheet := fmt.Sprintf("%s_%d", sanitizeSheetName(pointName[pid]), pid)
			if first {
				f.SetSheetName(defaultSheet, sheet)
				first = false
			} else {
				f.NewSheet(sheet)
			}

			header := []string{"商品名称", "规格/分类", "单位", "件数"}
			for col, h := range header {
				cell, _ := excelize.CoordinatesToCellName(col+1, 1)
				f.SetCellValue(sheet, cell, h)
			}
			f.SetCellStyle(sheet, "A1", "D1", boldStyle)

			rowNum := 2
			for _, it := range items {
				f.SetCellValue(sheet, fmt.Sprintf("A%d", rowNum), it.ProductName)
				f.SetCellValue(sheet, fmt.Sprintf("B%d", rowNum), it.Category)
				f.SetCellValue(sheet, fmt.Sprintf("C%d", rowNum), it.Unit)
				f.SetCellValue(sheet, fmt.Sprintf("D%d", rowNum), it.TotalQty)
				rowNum++
			}

			f.SetColWidth(sheet, "A", "A", 25)
			f.SetColWidth(sheet, "B", "B", 18)
			f.SetColWidth(sheet, "C", "C", 10)
			f.SetColWidth(sheet, "D", "D", 10)
		}
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

func sanitizeSheetName(s string) string {
	r := []rune(s)
	if len(r) > 20 {
		r = r[:20]
	}
	result := make([]rune, 0, len(r))
	for _, ch := range r {
		if ch == '[' || ch == ']' || ch == ':' || ch == '*' || ch == '?' || ch == '/' || ch == '\\' {
			result = append(result, '_')
		} else {
			result = append(result, ch)
		}
	}
	out := string(result)
	if out == "" {
		out = "未命名"
	}
	return out
}
