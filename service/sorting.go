package service

import (
	"fmt"

	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
	"groupbuy/model"
)

type ExportRow struct {
	PickupPointID   uint64
	PickupPointName string
	ProductName     string
	Category        string
	Unit            string
	TotalQty        uint32
}

func GeneratePickupExcel(db *gorm.DB, date string) (*excelize.File, error) {
	var rows []ExportRow
	err := db.Table("orders").
		Select("orders.pickup_point_id, pickup_points.name as pickup_point_name, "+
			"products.name as product_name, products.category, products.unit, "+
			"SUM(orders.quantity) as total_qty").
		Joins("LEFT JOIN pickup_points ON pickup_points.id = orders.pickup_point_id").
		Joins("LEFT JOIN products ON products.id = orders.product_id").
		Where("orders.order_date = ? AND orders.status IN ?", date, model.ValidOrderStatuses()).
		Group("orders.pickup_point_id, orders.product_id").
		Order("orders.pickup_point_id").
		Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("查询订单失败: %w", err)
	}

	f := excelize.NewFile()
	defaultSheet := f.GetSheetName(0)

	grouped := make(map[uint64][]ExportRow)
	pointName := make(map[uint64]string)
	for _, r := range rows {
		grouped[r.PickupPointID] = append(grouped[r.PickupPointID], r)
		pointName[r.PickupPointID] = r.PickupPointName
	}

	boldStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 12},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#E8F0FE"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})

	if len(grouped) == 0 {
		sheet := "汇总"
		f.SetSheetName(defaultSheet, sheet)
		f.SetCellValue(sheet, "A1", "当天暂无待分拣订单")
		return f, nil
	}

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

	return f, nil
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
