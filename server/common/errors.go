package common

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type InternalServerError struct {
	Message string
}

func (e *InternalServerError) Error() string {
	return e.Message
}

type BadRequestError struct {
	Message string
}

func (e *BadRequestError) Error() string {
	return e.Message
}

type NotFoundError struct {
	Message string
}

func (e *NotFoundError) Error() string {
	return e.Message
}

type ForbiddenError struct {
	Message string
}

func (e *ForbiddenError) Error() string {
	return e.Message
}

func HandleError(err error, c *gin.Context) {
	switch err.(type) {
	case *InternalServerError:
		c.String(http.StatusInternalServerError, err.Error())
	case *BadRequestError:
		c.String(http.StatusBadRequest, err.Error())
	case *NotFoundError:
		c.String(http.StatusNotFound, err.Error())
	case *ForbiddenError:
		c.String(http.StatusForbidden, err.Error())
	default:
		c.String(http.StatusInternalServerError, "Internal Server Error")
	}
}
