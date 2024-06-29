package common

import "net/http"

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

func HandleError(err error, w http.ResponseWriter) {
	switch err.(type) {
	case *InternalServerError:
		http.Error(w, err.Error(), http.StatusInternalServerError)
	case *BadRequestError:
		http.Error(w, err.Error(), http.StatusBadRequest)
	case *NotFoundError:
		http.Error(w, err.Error(), http.StatusNotFound)
	case *ForbiddenError:
		http.Error(w, err.Error(), http.StatusForbidden)
	default:
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
