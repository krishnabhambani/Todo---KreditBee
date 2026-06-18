package controllers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/todo-app/backend/apperrors"
	"github.com/todo-app/backend/controllers/dto"
	"github.com/todo-app/backend/services"
)

type ShareController struct {
	todoService services.TodoService
}

func NewShareController(service services.TodoService) *ShareController {
	return &ShareController{todoService: service}
}

func (ctrl *ShareController) ShareGroup(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.Error(apperrors.NewUnauthorized("unauthorized"))
		return
	}

	groupIDStr := c.Param("id")
	groupID, err := strconv.ParseUint(groupIDStr, 10, 32)
	if err != nil {
		c.Error(apperrors.NewBadRequest("invalid group ID"))
		return
	}

	var req dto.ShareGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(apperrors.NewBadRequest("email and permission (VIEW or EDIT) are required"))
		return
	}

	share, err := ctrl.todoService.ShareGroup(c.Request.Context(), uint(groupID), userID.(uint), req)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"success": true, "message": "Group shared successfully", "data": dto.MapGroupShare(share)})
}

func (ctrl *ShareController) RemoveShare(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.Error(apperrors.NewUnauthorized("unauthorized"))
		return
	}

	groupIDStr := c.Param("id")
	groupID, err := strconv.ParseUint(groupIDStr, 10, 32)
	if err != nil {
		c.Error(apperrors.NewBadRequest("invalid group ID"))
		return
	}

	targetUserIDStr := c.Param("userId")
	targetUserID, err := strconv.ParseUint(targetUserIDStr, 10, 32)
	if err != nil {
		c.Error(apperrors.NewBadRequest("invalid target user ID"))
		return
	}

	err = ctrl.todoService.RemoveShare(c.Request.Context(), uint(groupID), userID.(uint), uint(targetUserID))
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Share removed successfully", "data": nil})
}

func (ctrl *ShareController) UpdateSharePermission(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.Error(apperrors.NewUnauthorized("unauthorized"))
		return
	}

	groupIDStr := c.Param("id")
	targetUserIDStr := c.Param("userId")

	groupID, err1 := strconv.ParseUint(groupIDStr, 10, 32)
	targetUserID, err2 := strconv.ParseUint(targetUserIDStr, 10, 32)

	if err1 != nil || err2 != nil {
		c.Error(apperrors.NewBadRequest("invalid group ID or user ID"))
		return
	}

	var req dto.UpdateShareRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(apperrors.NewBadRequest("permission (VIEW or EDIT) is required"))
		return
	}

	if err := ctrl.todoService.UpdateSharePermission(c.Request.Context(), uint(groupID), userID.(uint), uint(targetUserID), req); err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Share permission updated successfully", "data": nil})
}

func (ctrl *ShareController) GetGroupMembers(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.Error(apperrors.NewUnauthorized("unauthorized"))
		return
	}

	groupIDStr := c.Param("id")
	groupID, err := strconv.ParseUint(groupIDStr, 10, 32)
	if err != nil {
		c.Error(apperrors.NewBadRequest("invalid group ID"))
		return
	}

	members, err := ctrl.todoService.GetGroupMembers(c.Request.Context(), uint(groupID), userID.(uint))
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Group members fetched successfully", "data": dto.MapGroupShareList(members)})
}

func (ctrl *ShareController) GetSharedGroups(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.Error(apperrors.NewUnauthorized("unauthorized"))
		return
	}

	status := c.Query("status")
	sortParam := c.Query("sort")
	search := c.Query("search")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	groups, meta, err := ctrl.todoService.GetSharedGroups(c.Request.Context(), userID.(uint), search, status, sortParam, page, limit)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Shared groups fetched successfully",
		"data":    dto.MapGroupShareList(groups),
		"meta":    meta,
	})
}

func (ctrl *ShareController) SearchUsers(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.Error(apperrors.NewUnauthorized("unauthorized"))
		return
	}

	query := strings.TrimSpace(c.Query("query"))
	if query == "" {
		c.JSON(http.StatusOK, gin.H{"success": true, "data": []interface{}{}})
		return
	}

	users, err := ctrl.todoService.SearchUsers(c.Request.Context(), query, userID.(uint))
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Users list fetched successfully",
		"data":    users,
	})
}
