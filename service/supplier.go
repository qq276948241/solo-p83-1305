package service

import (
	"gorm.io/gorm"
	"groupbuy/model"
)

type SupplierReconciliation struct {
	SupplierID   uint64            `json:"supplier_id"`
	SupplierName string            `json:"supplier_name"`
	StallNumber  string            `json:"stall_number"`
	TotalAmount  float64           `json:"total_amount"`
	TotalQty     uint32            `json:"total_qty"`
	Details      []SupplierDetail  `json:"details"`
}

type SupplierDetail struct {
	ProductID   uint64  `json:"product_id"`
	ProductName string  `json:"product_name"`
	Quantity    uint32  `json:"quantity"`
	TotalAmount float64 `json:"total_amount"`
}

func CalcReconciliation(db *gorm.DB, date string) ([]SupplierReconciliation, error) {
	type productAgg struct {
		ProductID   uint64
		TotalQty    uint32
		TotalAmount float64
	}

	var aggs []productAgg
	if err := db.Table("orders").
		Select("product_id, SUM(quantity) as total_qty, SUM(total_price) as total_amount").
		Where("order_date = ? AND status IN ?", date, model.ValidOrderStatuses()).
		Group("product_id").
		Scan(&aggs).Error; err != nil {
		return nil, err
	}

	supplierMap := make(map[uint64]*SupplierReconciliation)

	for _, agg := range aggs {
		var productName string
		var supplierID uint64
		db.Table("products").Select("name, supplier_id").Where("id = ?", agg.ProductID).Row().Scan(&productName, &supplierID)

		if supplierMap[supplierID] == nil {
			var sName, sStall string
			db.Table("suppliers").Select("name, stall_number").Where("id = ?", supplierID).Row().Scan(&sName, &sStall)
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

	return results, nil
}
