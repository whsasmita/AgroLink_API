package handlers

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/whsasmita/AgroLink_API/config"
	"github.com/whsasmita/AgroLink_API/dto"
	"github.com/whsasmita/AgroLink_API/models"
	"github.com/whsasmita/AgroLink_API/services"
	"github.com/whsasmita/AgroLink_API/utils"
)

type ProductHandler struct {
	productService services.ProductService
}

func NewProductHandler(s services.ProductService) *ProductHandler {
	return &ProductHandler{productService: s}
}

// CreateProduct menangani pembuatan produk baru oleh petani.
func (h *ProductHandler) CreateProduct(c *gin.Context) {
	var input dto.CreateProductInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid input data", err)
		return
	}

	// Ambil pengguna saat ini dari context
	currentUser := c.MustGet("user").(*models.User)
	if currentUser.Farmer == nil {
		utils.ErrorResponse(c, http.StatusForbidden, "Forbidden: Only farmers can create products", nil)
		return
	}

	product, err := h.productService.CreateProduct(input, currentUser.Farmer.UserID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create product", err)
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Product created successfully", product)
}

// GetAllProducts menangani pengambilan semua produk (publik).
func (h *ProductHandler) GetAllProducts(c *gin.Context) {
	products, err := h.productService.GetAllProducts()
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve products", err)
		return
	}
	utils.SuccessResponse(c, http.StatusOK, "Products retrieved successfully", products)
}

func (h *ProductHandler) GetMyProducts(c *gin.Context) {
	// Ambil pengguna saat ini dari context
	currentUser := c.MustGet("user").(*models.User)
	
	// Validasi bahwa pengguna adalah seorang petani
	if currentUser.Farmer == nil {
		utils.ErrorResponse(c, http.StatusForbidden, "Forbidden: Only farmers can view their products", nil)
		return
	}

	products, err := h.productService.GetMyProducts(currentUser.Farmer.UserID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve products", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Your products retrieved successfully", products)
}


// GetProductByID menangani pengambilan satu produk berdasarkan ID (publik).
func (h *ProductHandler) GetProductByID(c *gin.Context) {
	productID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid product ID format", err)
		return
	}

	product, err := h.productService.GetProductByID(productID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Product not found", nil)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Product retrieved successfully", product)
}

// UpdateProduct menangani pembaruan produk oleh petani pemilik.
func (h *ProductHandler) UpdateProduct(c *gin.Context) {
	productID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid product ID format", err)
		return
	}

	var input dto.UpdateProductInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid input data", err)
		return
	}

	currentUser := c.MustGet("user").(*models.User)
	if currentUser.Farmer == nil {
		utils.ErrorResponse(c, http.StatusForbidden, "Forbidden: Only farmers can update products", nil)
		return
	}

	updatedProduct, err := h.productService.UpdateProduct(productID, input, currentUser.Farmer.UserID)
	if err != nil {
		if strings.Contains(err.Error(), "forbidden") {
			utils.ErrorResponse(c, http.StatusForbidden, err.Error(), nil)
			return
		}
		if strings.Contains(err.Error(), "not found") {
			utils.ErrorResponse(c, http.StatusNotFound, err.Error(), nil)
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update product", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Product updated successfully", updatedProduct)
}

func (h *ProductHandler) UploadImage(c *gin.Context) {
    // Ambil pengguna saat ini untuk menamai file
    currentUser := c.MustGet("user").(*models.User)

    // Ambil file dari form-data dengan key "image"
    file, err := c.FormFile("image")
    if err != nil {
        utils.ErrorResponse(c, http.StatusBadRequest, "Image file not provided", err)
        return
    }

    // === Validasi ===
    if file.Size > 2*1024*1024 { // 2MB
        utils.ErrorResponse(c, http.StatusRequestEntityTooLarge, "File size exceeds 2MB limit", nil)
        return
    }
    ext := filepath.Ext(file.Filename)
    if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
        utils.ErrorResponse(c, http.StatusBadRequest, "Invalid file type. Only JPG, JPEG, PNG are allowed.", nil)
        return
    }

    // Buat nama file yang unik untuk menghindari konflik
    newFileName := fmt.Sprintf("%s-%d%s", currentUser.ID, time.Now().UnixNano(), ext)
    
    // Tentukan direktori penyimpanan: public/uploads/products/
    uploadDir := filepath.Join("public", "uploads", "products")
    savePath := filepath.Join(uploadDir, newFileName)

    // Simpan file
    if err := c.SaveUploadedFile(file, savePath); err != nil {
        utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to save file", err)
        return
    }

    // [PERBAIKAN] Buat URL yang bisa diakses publik (Menggunakan /static/uploads/products/)
    // Asumsi: Server Go melayani folder 'public' melalui rute '/static'
    appUrl := config.AppConfig_.App.APP_URL
    publicURL := fmt.Sprintf("%s/static/uploads/products/%s", appUrl, newFileName) // <-- Menggunakan path yang konsisten
    
    utils.SuccessResponse(c, http.StatusOK, "File uploaded successfully", gin.H{
        "url": publicURL,
    })
}

// DeleteProduct menangani penghapusan produk oleh petani pemilik.
func (h *ProductHandler) DeleteProduct(c *gin.Context) {
	productID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid product ID format", err)
		return
	}

	currentUser := c.MustGet("user").(*models.User)
	if currentUser.Farmer == nil {
		utils.ErrorResponse(c, http.StatusForbidden, "Forbidden: Only farmers can delete products", nil)
		return
	}

	err = h.productService.DeleteProduct(productID, currentUser.Farmer.UserID)
	if err != nil {
		if strings.Contains(err.Error(), "forbidden") {
			utils.ErrorResponse(c, http.StatusForbidden, err.Error(), nil)
			return
		}
		if strings.Contains(err.Error(), "not found") {
			utils.ErrorResponse(c, http.StatusNotFound, err.Error(), nil)
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete product", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Product deleted successfully", nil)
}