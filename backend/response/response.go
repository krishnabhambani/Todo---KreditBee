package response

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/todo-app/backend/apperrors"
	"github.com/todo-app/backend/logger"
)

// Success writes a standardized success JSON response.
func Success(c *gin.Context, statusCode int, message string, data any) {
	if data == nil {
		data = map[string]any{} // Use empty object if data is nil
	}
	c.JSON(statusCode, gin.H{
		"success": true,
		"message": message,
		"data":    data,
	})
}

// SuccessWithMeta writes a standardized success JSON response along with metadata.
func SuccessWithMeta(c *gin.Context, statusCode int, message string, data any, meta any) {
	if data == nil {
		data = map[string]any{} // Use empty object if data is nil
	}
	c.JSON(statusCode, gin.H{
		"success": true,
		"message": message,
		"data":    data,
		"meta":    meta,
	})
}

// Error writes a standardized error JSON response.
func Error(c *gin.Context, statusCode int, message string, errData any) {
	if errData == nil {
		errData = map[string]any{} // Use empty object if error is nil
	}
	// For AbortWithStatusJSON use case in middleware
	c.AbortWithStatusJSON(statusCode, gin.H{
		"success": false,
		"message": message,
		"error":   errData,
	})
}

// HandleError is the centralized error handler that parses known errors and formats the response.
func HandleError(c *gin.Context, err error) {
	// Try to get the logger from the context (injected by logger middleware)
	var log logger.Logger
	if logObj, exists := c.Get("logger"); exists {
		if l, ok := logObj.(logger.Logger); ok {
			log = l
		}
	}

	// 1. Validation Errors (from Gin ShouldBind)
	var validationErrs validator.ValidationErrors
	if errors.As(err, &validationErrs) {
		errMap := make(map[string]string)
		for _, v := range validationErrs {
			errMap[v.Field()] = v.Tag()
		}
		Error(c, http.StatusUnprocessableEntity, "Validation failed", errMap)
		return
	}

	// 2. Application Errors
	var appErr *apperrors.AppError
	if errors.As(err, &appErr) {
		if appErr.StatusCode >= 500 && log != nil {
			log.Error(c.Request.Context(), "internal server error", logger.F("details", appErr.Error()), logger.F("path", c.Request.URL.Path))
		} else if log != nil {
			log.Warn(c.Request.Context(), "client error", logger.F("details", appErr.Error()), logger.F("status", appErr.StatusCode))
		}
		Error(c, appErr.StatusCode, appErr.Message, appErr.Details)
		return
	}

	// 3. Generic/Unknown Errors
	if log != nil {
		log.Error(c.Request.Context(), "unhandled generic error", logger.F("details", err.Error()), logger.F("path", c.Request.URL.Path))
	}
	Error(c, http.StatusInternalServerError, "Internal Server Error", nil)
}
