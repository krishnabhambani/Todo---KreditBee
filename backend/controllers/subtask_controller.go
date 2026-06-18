package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/todo-app/backend/apperrors"
	"github.com/todo-app/backend/controllers/dto"
	"github.com/todo-app/backend/services"
)

type SubtaskController struct {
	todoService services.TodoService
}

func NewSubtaskController(service services.TodoService) *SubtaskController {
	return &SubtaskController{todoService: service}
}

func (ctrl *SubtaskController) GetSubtasks(c *gin.Context) {
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

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	subtasks, meta, err := ctrl.todoService.GetSubtasks(c.Request.Context(), uint(groupID), userID.(uint), page, limit)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Subtasks fetched successfully",
		"data":    dto.MapTodos(subtasks),
		"meta":    meta,
	})
}

func (ctrl *SubtaskController) CreateSubtask(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.Error(apperrors.NewUnauthorized("unauthorized"))
		return
	}

	var req dto.CreateSubtaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(apperrors.NewBadRequest("title and group_id are required"))
		return
	}

	subtask, err := ctrl.todoService.CreateSubtask(c.Request.Context(), req, userID.(uint))
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"success": true, "message": "Subtask created successfully", "data": dto.MapTodo(subtask)})
}

func (ctrl *SubtaskController) UpdateSubtask(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.Error(apperrors.NewUnauthorized("unauthorized"))
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.Error(apperrors.NewBadRequest("invalid subtask ID"))
		return
	}

	var req dto.UpdateSubtaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(apperrors.NewBadRequest("title is required"))
		return
	}

	subtask, err := ctrl.todoService.UpdateSubtask(c.Request.Context(), uint(id), req, userID.(uint))
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Subtask updated successfully", "data": dto.MapTodo(subtask)})
}

func (ctrl *SubtaskController) DeleteSubtask(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.Error(apperrors.NewUnauthorized("unauthorized"))
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.Error(apperrors.NewBadRequest("invalid subtask ID"))
		return
	}

	err = ctrl.todoService.DeleteSubtask(c.Request.Context(), uint(id), userID.(uint))
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Subtask deleted successfully", "data": nil})
}

func (ctrl *SubtaskController) ToggleCompleteSubtask(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.Error(apperrors.NewUnauthorized("unauthorized"))
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.Error(apperrors.NewBadRequest("invalid subtask ID"))
		return
	}

	subtask, err := ctrl.todoService.ToggleCompleteSubtask(c.Request.Context(), uint(id), userID.(uint))
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Subtask status updated", "data": dto.MapTodo(subtask)})
}
