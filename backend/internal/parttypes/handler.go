package parttypes

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/partflow/smart-store/pkg/response"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes registers part types routes
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	partTypes := router.Group("/part-types")
	{
		partTypes.GET("", h.ListPartTypes)
		partTypes.POST("", h.CreatePartType)
		partTypes.GET("/:id", h.GetPartType)
		partTypes.PUT("/:id", h.UpdatePartType)
		partTypes.DELETE("/:id", h.DeletePartType)
		partTypes.GET("/:id/specifications", h.GetTypeSpecifications)
	}

	specifications := router.Group("/specifications")
	{
		specifications.GET("", h.ListSpecifications)
		specifications.POST("", h.CreateSpecification)
	}

	typeSpecs := router.Group("/type-specifications")
	{
		typeSpecs.POST("", h.LinkSpecification)
		typeSpecs.DELETE("/:part_type_id/:specification_id", h.UnlinkSpecification)
	}

	itemSpecs := router.Group("/item-specifications")
	{
		itemSpecs.GET("/:item_id", h.GetItemSpecifications)
		itemSpecs.PUT("/:item_id", h.UpdateItemSpecifications)
	}
}

// ListPartTypes lists all part types
func (h *Handler) ListPartTypes(c *gin.Context) {
	types, err := h.service.ListPartTypes(c.Request.Context())
	if err != nil {
		response.InternalError(c, "Failed to list part types")
		return
	}

	response.OK(c, types, "Part types retrieved successfully")
}

// CreatePartType creates a new part type
func (h *Handler) CreatePartType(c *gin.Context) {
	var req CreatePartTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	partType, err := h.service.CreatePartType(c.Request.Context(), &req)
	if err != nil {
		response.InternalError(c, "Failed to create part type")
		return
	}

	response.Created(c, partType, "Part type created successfully")
}

// GetPartType gets a part type by ID
func (h *Handler) GetPartType(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid part type ID")
		return
	}

	partType, err := h.service.GetPartType(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, "Part type not found")
		return
	}

	response.OK(c, partType, "Part type retrieved successfully")
}

// UpdatePartType updates a part type
func (h *Handler) UpdatePartType(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid part type ID")
		return
	}

	var req UpdatePartTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	partType, err := h.service.UpdatePartType(c.Request.Context(), id, &req)
	if err != nil {
		response.InternalError(c, "Failed to update part type")
		return
	}

	response.OK(c, partType, "Part type updated successfully")
}

// DeletePartType deletes a part type
func (h *Handler) DeletePartType(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid part type ID")
		return
	}

	if err := h.service.DeletePartType(c.Request.Context(), id); err != nil {
		response.InternalError(c, "Failed to delete part type")
		return
	}

	response.OK(c, nil, "Part type deleted successfully")
}

// GetTypeSpecifications gets specifications for a part type
func (h *Handler) GetTypeSpecifications(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid part type ID")
		return
	}

	specs, err := h.service.GetTypeSpecifications(c.Request.Context(), id)
	if err != nil {
		response.InternalError(c, "Failed to get type specifications")
		return
	}

	response.OK(c, specs, "Type specifications retrieved successfully")
}

// ListSpecifications lists all available specifications
func (h *Handler) ListSpecifications(c *gin.Context) {
	specs, err := h.service.ListSpecifications(c.Request.Context())
	if err != nil {
		response.InternalError(c, "Failed to list specifications")
		return
	}

	response.OK(c, specs, "Specifications retrieved successfully")
}

// CreateSpecification creates a new specification
func (h *Handler) CreateSpecification(c *gin.Context) {
	var req CreateSpecificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	spec, err := h.service.CreateSpecification(c.Request.Context(), &req)
	if err != nil {
		response.InternalError(c, "Failed to create specification")
		return
	}

	response.Created(c, spec, "Specification created successfully")
}

// LinkSpecification links a specification to a part type
func (h *Handler) LinkSpecification(c *gin.Context) {
	var req LinkSpecificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	link, err := h.service.LinkSpecification(c.Request.Context(), &req)
	if err != nil {
		response.InternalError(c, "Failed to link specification")
		return
	}

	response.Created(c, link, "Specification linked successfully")
}

// UnlinkSpecification unlinks a specification from a part type
func (h *Handler) UnlinkSpecification(c *gin.Context) {
	partTypeID, err := uuid.Parse(c.Param("part_type_id"))
	if err != nil {
		response.BadRequest(c, "Invalid part type ID")
		return
	}

	specificationID, err := uuid.Parse(c.Param("specification_id"))
	if err != nil {
		response.BadRequest(c, "Invalid specification ID")
		return
	}

	if err := h.service.UnlinkSpecification(c.Request.Context(), partTypeID, specificationID); err != nil {
		response.InternalError(c, "Failed to unlink specification")
		return
	}

	response.OK(c, nil, "Specification unlinked successfully")
}

// GetItemSpecifications gets specifications for an inventory item
func (h *Handler) GetItemSpecifications(c *gin.Context) {
	itemID, err := uuid.Parse(c.Param("item_id"))
	if err != nil {
		response.BadRequest(c, "Invalid item ID")
		return
	}

	specs, err := h.service.GetItemSpecifications(c.Request.Context(), itemID)
	if err != nil {
		response.InternalError(c, "Failed to get item specifications")
		return
	}

	response.OK(c, specs, "Item specifications retrieved successfully")
}

// UpdateItemSpecifications updates specifications for an inventory item
func (h *Handler) UpdateItemSpecifications(c *gin.Context) {
	itemID, err := uuid.Parse(c.Param("item_id"))
	if err != nil {
		response.BadRequest(c, "Invalid item ID")
		return
	}

	var req UpdateItemSpecsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	if err := h.service.UpdateItemSpecifications(c.Request.Context(), itemID, &req); err != nil {
		response.InternalError(c, "Failed to update item specifications")
		return
	}

	response.OK(c, nil, "Item specifications updated successfully")
}