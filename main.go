package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"groupbuy/config"
	"groupbuy/handler"
	"groupbuy/middleware"
	"groupbuy/model"
)

func main() {
	cfg := config.Load()

	db, err := gorm.Open(mysql.Open(cfg.DSN()), &gorm.Config{})
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}
	fmt.Println("数据库连接成功")

	db.AutoMigrate(&model.Supplier{}, &model.Product{}, &model.PickupPoint{}, &model.Order{})

	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})

	api := r.Group("/api")
	api.Use(middleware.TokenAuth(cfg))
	{
		orderH := handler.OrderHandler{DB: db}
		api.POST("/orders", orderH.Create)
		api.GET("/orders", orderH.List)
		api.PATCH("/orders/:id/status", orderH.UpdateStatus)

		sortingH := handler.SortingHandler{DB: db}
		api.GET("/sorting/summary", sortingH.Summary)

		pickupH := handler.PickupHandler{DB: db}
		api.GET("/pickup/slips", pickupH.List)

		supplierH := handler.SupplierHandler{DB: db}
		api.GET("/supplier/reconciliation", supplierH.Reconciliation)
		api.GET("/suppliers", supplierH.List)
		api.GET("/products", supplierH.ListProducts)
		api.GET("/pickup-points", supplierH.ListPickupPoints)
	}

	addr := ":" + cfg.ServerPort
	fmt.Printf("服务启动在 %s\n", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}
