package controllers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/todo-app/backend/services"
)

type SubtaskController struct {
	todoService services.TodoService
}

func NewSubtaskController(service services.TodoService) *SubtaskController {
	return &SubtaskController{todoService: service}
}

type SubtaskInput struct {
	Title       string     `json:"title" binding:"required"`
	Description string     `json:"description"`
	DueDate     *time.Time `json:"due_date"`
}

func (ctrl *SubtaskController) GetSubtasks(c *gin.Context) {
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

	subtasks, err := ctrl.todoService.GetSubtasks(c.Request.Context(), uint(groupID), userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to fetch subtasks",
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Subtasks fetched successfully",
		"data":    subtasks,
	})
}

func (ctrl *SubtaskController) CreateSubtask(c *gin.Context) {
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

	var input SubtaskInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Title is required",
			"data":    nil,
		})
		return
	}

	subtask, err := ctrl.todoService.CreateSubtask(c.Request.Context(), input.Title, input.Description, input.DueDate, uint(groupID), userID.(uint))
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
		"message": "Subtask created successfully",
		"data":    subtask,
	})
}

func (ctrl *SubtaskController) UpdateSubtask(c *gin.Context) {
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
			"message": "Invalid Subtask ID",
			"data":    nil,
		})
		return
	}

	var input SubtaskInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Title is required",
			"data":    nil,
		})
		return
	}

	subtask, err := ctrl.todoService.UpdateSubtask(c.Request.Context(), uint(id), input.Title, input.Description, input.DueDate, userID.(uint))
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
		"message": "Subtask updated successfully",
		"data":    subtask,
	})
}

func (ctrl *SubtaskController) DeleteSubtask(c *gin.Context) {
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
			"message": "Invalid Subtask ID",
			"data":    nil,
		})
		return
	}

	err = ctrl.todoService.DeleteSubtask(c.Request.Context(), uint(id), userID.(uint))
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
		"message": "Subtask deleted successfully",
		"data":    nil,
	})
}

func (ctrl *SubtaskController) ToggleCompleteSubtask(c *gin.Context) {
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
			"message": "Invalid Subtask ID",
			"data":    nil,
		})
		return
	}

	subtask, err := ctrl.todoService.ToggleCompleteSubtask(c.Request.Context(), uint(id), userID.(uint))
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
		"message": "Subtask completion status updated",
		"data":    subtask,
	})
}
