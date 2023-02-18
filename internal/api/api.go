// Package api provides api handlers for the Lambda
package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

// Handler the API Handler.
type Handler struct{}

// NewHandler creates a new API Handler.
func NewHandler() *Handler {
	return &Handler{}
}

// HelloWorld handler for Hello World request.
func (h *Handler) HelloWorld(c echo.Context) error {
	c.Logger().Info("Hello World!")
	return c.JSON(http.StatusOK, h.createResponse("Hello World!"))
}

// YourNameRequest request for the YourName handler.
type YourNameRequest struct {
	Name string `json:"name"`
}

// YourName handlers for your name request.
func (h *Handler) YourName(c echo.Context) error {
	r := c.Request()
	payload := YourNameRequest{}
	defer func() {
		err := r.Body.Close()
		if err != nil {
			c.Logger().Error("unable to close response body: %s", err)
		}
	}()
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		return c.JSON(http.StatusBadRequest, h.createResponse("invalid request body provided"))
	}
	if payload.Name == "" {
		return c.JSON(http.StatusBadRequest, h.createResponse("if you don't tell me I don't know your name"))
	}

	return c.JSON(http.StatusOK, h.createResponse(fmt.Sprintf("your name is: %s", payload.Name)))
}

func (h *Handler) createResponse(msg string) any {
	return struct {
		Message string `json:"message"`
	}{
		Message: msg,
	}
}
