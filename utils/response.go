package utils

import (
    "net/http"
    "github.com/gin-gonic/gin"
)

// Respon struktur
type Response struct {
    Success bool        `json:"success"`
    Message string      `json:"message"`
    Data    interface{} `json:"data,omitempty"`
    Errors  interface{} `json:"errors,omitempty"`
}

// Response sukses dengan data
func SuccessResponse(c *gin.Context, statusCode int, message string, data interface{}) {
    c.JSON(statusCode, Response{
        Success: true,
        Message: message,
        Data:    data,
    })
}

// Response sukses dengan status 200
func SuccessResponseOK(c *gin.Context, message string, data interface{}) {
    SuccessResponse(c, http.StatusOK, message, data)
}

// Response sukses dengan status 201
func SuccessResponseCreated(c *gin.Context, message string, data interface{}) {
    SuccessResponse(c, http.StatusCreated, message, data)
}

// ErrorResponse untuk response error manual
func ErrorResponse(c *gin.Context, statusCode int, message string, errorDetails interface{}) {
    c.JSON(statusCode, Response{
        Success: false,
        Message: message,
        Errors:  errorDetails,
    })
}

// Response error 400 ("ex:Validasi gagal")
func ErrorResponseBadRequest(c *gin.Context, message string, errorDetails interface{}) {
    ErrorResponse(c, http.StatusBadRequest, message, errorDetails)
}

// Response error 401 ("ex:Token tidak valid")
func ErrorResponseUnauthorized(c *gin.Context, message string) {
    ErrorResponse(c, http.StatusUnauthorized, message, nil)
}

// Response error 404 ("ex:Data tidak ditemukan")
func ErrorResponseNotFound(c *gin.Context, message string) {
    ErrorResponse(c, http.StatusNotFound, message, nil)
}

// Response error 409 ("ex:Email sudah terdaftar")
func ErrorResponseConflict(c *gin.Context, message string) {
    ErrorResponse(c, http.StatusConflict, message, nil)
}

// Response error 500 ("ex:Terjadi kesalahan pada server")
func ErrorResponseInternal(c *gin.Context, message string) {
    ErrorResponse(c, http.StatusInternalServerError, message, nil)
}