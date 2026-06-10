package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/todo-app/backend/services"
)

type ShareController struct {
	todoService services.TodoService
}

func NewShareController(service services.TodoService) *ShareController {
	return &ShareController{todoService: service}
}

type ShareInput struct {
	Email      string `json:"email" binding:"required,email"`
	Permission string `json:"permission" binding:"required"`
}

func (ctrl *ShareController) ShareGroup(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "User not authenticated",
			"data":    nil,
		})
		return
	}

	groupIDStr := c.Param("id")
	groupID, err := strconv.ParseUint(groupIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid Group ID",
			"data":    nil,
		})
		return
	}

	var input ShareInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Validation failed: valid email and permission level (VIEW/EDIT) are required",
			"data":    nil,
		})
		return
	}

	share, err := ctrl.todoService.ShareGroup(c.Request.Context(), uint(groupID), userID.(uint), input.Email, input.Permission)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Group shared successfully",
		"data":    share,
	})
}

func (ctrl *ShareController) RemoveShare(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "User not authenticated",
			"data":    nil,
		})
		return
	}

	groupIDStr := c.Param("id")
	groupID, err := strconv.ParseUint(groupIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid Group ID",
			"data":    nil,
		})
		return
	}

	targetUserIDStr := c.Param("userId")
	targetUserID, err := strconv.ParseUint(targetUserIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid Target User ID",
			"data":    nil,
		})
		return
	}

	err = ctrl.todoService.RemoveShare(c.Request.Context(), uint(groupID), userID.(uint), uint(targetUserID))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Collaborator removed successfully",
		"data":    nil,
	})
}

func (ctrl *ShareController) GetGroupMembers(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "User not authenticated",
			"data":    nil,
		})
		return
	}

	groupIDStr := c.Param("id")
	groupID, err := strconv.ParseUint(groupIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid Group ID",
			"data":    nil,
		})
		return
	}

	members, err := ctrl.todoService.GetGroupMembers(c.Request.Context(), uint(groupID), userID.(uint))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Collaborators list fetched successfully",
		"data":    members,
	})
}

func (ctrl *ShareController) GetSharedGroups(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "User not authenticated",
			"data":    nil,
		})
		return
	}

	status := c.Query("status")
	sortParam := c.Query("sort")
	groups, err := ctrl.todoService.GetSharedGroups(c.Request.Context(), userID.(uint), status, sortParam)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to fetch shared groups",
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Shared groups fetched successfully",
		"data":    groups,
	})
}

func (ctrl *ShareController) SearchUsers(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "User not authenticated",
			"data":    nil,
		})
		return
	}

	query := c.Query("search")
	users, err := ctrl.todoService.SearchUsers(c.Request.Context(), query, userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to search users",
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Users list fetched successfully",
		"data":    users,
	})
}
