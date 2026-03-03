package middleware

import (
	"net/http"

	"github.com/DevKayoS/fintech-kodify/internal/errors"
	"github.com/gin-gonic/gin"
)

func ErrorHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Next()

		if len(ctx.Errors) == 0 {
			return
		}

		err := ctx.Errors.Last().Err

		if apiErr, ok := err.(errors.ApiError); ok {
			ctx.JSON(apiErr.StatusCode, gin.H{
				"error": gin.H{
					"code": apiErr.Code,
					"msg":  apiErr.Message,
				},
			})
			return
		}

		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code": "INTERNAL_ERROR",
				"msg":  err.Error(),
			},
		})
	}
}
