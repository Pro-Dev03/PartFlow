package products

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

// RegisterRoutes registers products routes
func RegisterRoutes(router *gin.RouterGroup, db *sqlx.DB) {
	repo := NewRepository(db)
	service := NewService(repo)
	handler := NewHandler(service)

	// Category routes
	categories := router.Group("/categories")
	{
		categories.POST("", handler.CreateCategory)
		categories.GET("/:id", handler.GetCategory)
		categories.GET("", handler.ListCategories)
		categories.PUT("/:id", handler.UpdateCategory)
		categories.DELETE("/:id", handler.DeleteCategory)
	}

	// Brand routes
	brands := router.Group("/brands")
	{
		brands.POST("", handler.CreateBrand)
		brands.GET("/:id", handler.GetBrand)
		brands.GET("", handler.ListBrands)
		brands.PUT("/:id", handler.UpdateBrand)
		brands.DELETE("/:id", handler.DeleteBrand)
	}

	// Product routes
	products := router.Group("/products")
	{
		products.POST("", handler.CreateProduct)
		products.GET("/:id", handler.GetProduct)
		products.GET("/barcode/:barcode", handler.GetProductByBarcode)
		products.GET("", handler.ListProducts)
		products.PUT("/:id", handler.UpdateProduct)
		products.DELETE("/:id", handler.DeleteProduct)
		products.POST("/:id/archive", handler.ArchiveProduct)
		products.POST("/:id/barcode", handler.GenerateBarcode)
		products.GET("/:id/stock", handler.GetProductStock)
		products.GET("/search", handler.SearchProducts)
	}
}