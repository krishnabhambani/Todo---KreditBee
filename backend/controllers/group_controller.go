package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/todo-app/backend/apperrors"
	"github.com/todo-app/backend/controllers/dto"
	"github.com/todo-app/backend/services"
)

type GroupController interface {
	CreateGroup(c *gin.Context)
	GetGroups(c *gin.Context)
	GetGroupByID(c *gin.Context)
	UpdateGroup(c *gin.Context)
	DeleteGroup(c *gin.Context)
}

type groupController struct {
	todoService services.TodoService
}

func NewGroupController(service services.TodoService) GroupController {
	return &groupController{todoService: service}
}

func (ctrl *groupController) CreateGroup(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.Error(apperrors.NewUnauthorized("unauthorized"))
		return
	}

	var req dto.CreateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(apperrors.NewBadRequest("title is required"))
		return
	}

	group, err := ctrl.todoService.CreateGroup(c.Request.Context(), req, userID.(uint))
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"success": true, "message": "Group created successfully", "data": dto.MapTodo(group)})
}

func (ctrl *groupController) GetGroups(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.Error(apperrors.NewUnauthorized("unauthorized"))
		return
	}

	search := c.Query("search")
	status := c.Query("status")
	sortParam := c.Query("sort")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	groups, meta, err := ctrl.todoService.GetGroups(c.Request.Context(), userID.(uint), search, status, sortParam, page, limit)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Groups retrieved",
		"data":    dto.MapTodos(groups),
		"meta":    meta,
	})
}

func (ctrl *groupController) GetGroupByID(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.Error(apperrors.NewUnauthorized("unauthorized"))
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.Error(apperrors.NewBadRequest("invalid group ID"))
		return
	}

	group, err := ctrl.todoService.GetGroupByID(c.Request.Context(), uint(id), userID.(uint))
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Group retrieved", "data": dto.MapTodo(group)})
}

func (ctrl *groupController) UpdateGroup(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.Error(apperrors.NewUnauthorized("unauthorized"))
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.Error(apperrors.NewBadRequest("invalid group ID"))
		return
	}

	var req dto.UpdateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(apperrors.NewBadRequest("title is required"))
		return
	}

	group, err := ctrl.todoService.UpdateGroup(c.Request.Context(), uint(id), req, userID.(uint))
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Group updated successfully", "data": dto.MapTodo(group)})
}

func (ctrl *groupController) DeleteGroup(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.Error(apperrors.NewUnauthorized("unauthorized"))
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.Error(apperrors.NewBadRequest("invalid group ID"))
		return
	}

	err = ctrl.todoService.DeleteGroup(c.Request.Context(), uint(id), userID.(uint))
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Group deleted successfully", "data": nil})
}
