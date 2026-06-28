package controllers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/todo-app/backend/apperrors"
	"github.com/todo-app/backend/controllers/dto"
	"github.com/todo-app/backend/response"
	"github.com/todo-app/backend/services"
)

type ShareController interface {
	ShareGroup(c *gin.Context)
	RemoveShare(c *gin.Context)
	UpdateSharePermission(c *gin.Context)
	GetGroupMembers(c *gin.Context)
	GetSharedGroups(c *gin.Context)
	SearchUsers(c *gin.Context)
}

type shareController struct {
	todoService services.TodoService
}

func NewShareController(service services.TodoService) ShareController {
	return &shareController{todoService: service}
}

func (ctrl *shareController) ShareGroup(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		response.HandleError(c, apperrors.NewUnauthorized("unauthorized"))
		return
	}

	groupIDStr := c.Param("id")
	groupID, err := strconv.ParseUint(groupIDStr, 10, 32)
	if err != nil {
		response.HandleError(c, apperrors.NewBadRequest("invalid group ID"))
		return
	}

	var req dto.ShareGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.HandleError(c, err)
		return
	}

	share, err := ctrl.todoService.ShareGroup(c.Request.Context(), uint(groupID), userID.(uint), req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusCreated, "Group shared successfully", dto.MapGroupShare(share))
}

func (ctrl *shareController) RemoveShare(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		response.HandleError(c, apperrors.NewUnauthorized("unauthorized"))
		return
	}

	groupIDStr := c.Param("id")
	groupID, err := strconv.ParseUint(groupIDStr, 10, 32)
	if err != nil {
		response.HandleError(c, apperrors.NewBadRequest("invalid group ID"))
		return
	}

	targetUserIDStr := c.Param("userId")
	targetUserID, err := strconv.ParseUint(targetUserIDStr, 10, 32)
	if err != nil {
		response.HandleError(c, apperrors.NewBadRequest("invalid target user ID"))
		return
	}

	err = ctrl.todoService.RemoveShare(c.Request.Context(), uint(groupID), userID.(uint), uint(targetUserID))
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Share removed successfully", nil)
}

func (ctrl *shareController) UpdateSharePermission(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		response.HandleError(c, apperrors.NewUnauthorized("unauthorized"))
		return
	}

	groupIDStr := c.Param("id")
	targetUserIDStr := c.Param("userId")

	groupID, err1 := strconv.ParseUint(groupIDStr, 10, 32)
	targetUserID, err2 := strconv.ParseUint(targetUserIDStr, 10, 32)

	if err1 != nil || err2 != nil {
		response.HandleError(c, apperrors.NewBadRequest("invalid group ID or user ID"))
		return
	}

	var req dto.UpdateShareRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.HandleError(c, err)
		return
	}

	if err := ctrl.todoService.UpdateSharePermission(c.Request.Context(), uint(groupID), userID.(uint), uint(targetUserID), req); err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Share permission updated successfully", nil)
}

func (ctrl *shareController) GetGroupMembers(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		response.HandleError(c, apperrors.NewUnauthorized("unauthorized"))
		return
	}

	groupIDStr := c.Param("id")
	groupID, err := strconv.ParseUint(groupIDStr, 10, 32)
	if err != nil {
		response.HandleError(c, apperrors.NewBadRequest("invalid group ID"))
		return
	}

	members, err := ctrl.todoService.GetGroupMembers(c.Request.Context(), uint(groupID), userID.(uint))
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Group members fetched successfully", dto.MapGroupShareList(members))
}

func (ctrl *shareController) GetSharedGroups(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		response.HandleError(c, apperrors.NewUnauthorized("unauthorized"))
		return
	}

	status := c.Query("status")
	sortParam := c.Query("sort")
	search := c.Query("search")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	groups, meta, err := ctrl.todoService.GetSharedGroups(c.Request.Context(), userID.(uint), search, status, sortParam, page, limit)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.SuccessWithMeta(c, http.StatusOK, "Shared groups fetched successfully", dto.MapGroupShareList(groups), meta)
}

func (ctrl *shareController) SearchUsers(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		response.HandleError(c, apperrors.NewUnauthorized("unauthorized"))
		return
	}

	query := strings.TrimSpace(c.Query("query"))
	if query == "" {
		response.Success(c, http.StatusOK, "", []interface{}{})
		return
	}

	users, err := ctrl.todoService.SearchUsers(c.Request.Context(), query, userID.(uint))
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Users list fetched successfully", users)
}
