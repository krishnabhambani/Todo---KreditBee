package controllers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/todo-app/backend/services"
)

type GroupController struct {
	todoService services.TodoService
}

func NewGroupController(service services.TodoService) *GroupController {
	return &GroupController{todoService: service}
}

type GroupInput struct {
	Title       string     `json:"title" binding:"required"`
	Description string     `json:"description"`
	DueDate     *time.Time `json:"due_date"`
}

func (ctrl *GroupController) CreateGroup(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "User not authenticated",
			"data":    nil,
		})
		return
	}

	var input GroupInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Title is required",
			"data":    nil,
		})
		return
	}

	group, err := ctrl.todoService.CreateGroup(c.Request.Context(), input.Title, input.Description, input.DueDate, userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Group created successfully",
		"data":    group,
	})
}

func (ctrl *GroupController) GetGroups(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "User not authenticated",
			"data":    nil,
		})
		return
	}

	search := c.Query("search")
	status := c.Query("status")
	sortParam := c.Query("sort")
	groups, err := ctrl.todoService.GetGroups(c.Request.Context(), userID.(uint), search, status, sortParam)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to fetch groups",
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Groups fetched successfully",
		"data":    groups,
	})
}

func (ctrl *GroupController) GetGroupByID(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "User not authenticated",
			"data":    nil,
		})
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid Group ID",
			"data":    nil,
		})
		return
	}

	group, err := ctrl.todoService.GetGroupByID(c.Request.Context(), uint(id), userID.(uint))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Group fetched successfully",
		"data":    group,
	})
}

func (ctrl *GroupController) UpdateGroup(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "User not authenticated",
			"data":    nil,
		})
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid Group ID",
			"data":    nil,
		})
		return
	}

	var input GroupInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Title is required",
			"data":    nil,
		})
		return
	}

	group, err := ctrl.todoService.UpdateGroup(c.Request.Context(), uint(id), input.Title, input.Description, input.DueDate, userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Group updated successfully",
		"data":    group,
	})
}

func (ctrl *GroupController) DeleteGroup(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "User not authenticated",
			"data":    nil,
		})
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid Group ID",
			"data":    nil,
		})
		return
	}

	err = ctrl.todoService.DeleteGroup(c.Request.Context(), uint(id), userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Group deleted successfully",
		"data":    nil,
	})
}
