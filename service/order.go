package service

import (
	"gorm.io/gorm"
	"groupbuy/model"
)

type ProductSummaryItem struct {
	ProductID   uint64  `json:"product_id"`
	ProductName string  `json:"product_name"`
	Category    string  `json:"category"`
	Unit        string  `json:"unit"`
	TotalQty    uint32  `json:"total_qty"`
	TotalAmount float64 `json:"total_amount"`
}

func CalcProductSummary(db *gorm.DB, date string) ([]ProductSummaryItem, error) {
	type agg struct {
		ProductID   uint64
		TotalQty    uint32
		TotalAmount float64
	}

	var aggs []agg
	err := db.Table("orders").
		Select("product_id, SUM(quantity) as total_qty, SUM(total_price) as total_amount").
		Where("order_date = ? AND status IN ?", date, model.ValidOrderStatuses()).
		Group("product_id").
		Scan(&aggs).Error
	if err != nil {
		return nil, err
	}

	var results []ProductSummaryItem
	for _, a := range aggs {
		var pName, pCat, pUnit string
		db.Table("products").Select("name, category, unit").Where("id = ?", a.ProductID).Row().Scan(&pName, &pCat, &pUnit)
		results = append(results, ProductSummaryItem{
			ProductID:   a.ProductID,
			ProductName: pName,
			Category:    pCat,
			Unit:        pUnit,
			TotalQty:    a.TotalQty,
			TotalAmount: a.TotalAmount,
		})
	}

	return results, nil
}

type PickupSlipItem struct {
	OrderID       uint64  `json:"order_id"`
	OrderNo       string  `json:"order_no"`
	ProductName   string  `json:"product_name"`
	Unit          string  `json:"unit"`
	Quantity      uint32  `json:"quantity"`
	UnitPrice     float64 `json:"unit_price"`
	TotalPrice    float64 `json:"total_price"`
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

func CalcPickupSlips(db *gorm.DB, date string) ([]PickupSlip, error) {
	var points []model.PickupPoint
	db.Find(&points)

	var slips []PickupSlip
	for _, pp := range points {
		var orders []model.Order
		db.Preload("Product").
			Where("order_date = ? AND pickup_point_id = ? AND status IN ?", date, pp.ID, model.ValidOrderStatuses()).
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

	return slips, nil
}
