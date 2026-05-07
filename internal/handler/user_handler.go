package handler

import (
	"net/http"
	"time"

	"user-service/internal/model"
	"user-service/internal/service"

	"github.com/labstack/echo/v4"
)

type UserHandler struct {
	service *service.UserService
}

func NewUserHandler(s *service.UserService) *UserHandler {
	return &UserHandler{
		service: s,
	}
}

func (h *UserHandler) GetProfile(c echo.Context) error {

	//GET email from jwt middleware
	//email := c.Param("email")
	email, ok := c.Get("email").(string)
	if !ok {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Unauthorized"})
	}

	//call service to get user profile
	user, err := h.service.GetProfile(email)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "User not found"})
	}

	return c.JSON(http.StatusOK, user)
}

func (h *UserHandler) CreateUser(c echo.Context) error {

	var user model.User

	// 1. Bind request body
	if err := c.Bind(&user); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request",
		})
	}

	// 2. Basic validation
	if user.Email == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Email is required",
		})
	}

	// 3. Set timestamps
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	// 4. Save user
	err := h.service.CreateUser(&user)

	if err != nil {

		// Handle duplicate user
		if err.Error() == "user already exists" {
			return c.JSON(http.StatusConflict, map[string]string{
				"error": err.Error(),
			})
		}

		// Other errors
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to create user",
		})
	}

	return c.JSON(http.StatusCreated, user)
}

func (h *UserHandler) UpdateProfile(c echo.Context) error {

	email, ok := c.Get("email").(string)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Unauthorized",
		})
	}

	var body map[string]interface{}

	// 1. Bind request
	if err := c.Bind(&body); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request",
		})
	}

	// 2. Prevent updating restricted fields
	delete(body, "email")
	delete(body, "id")

	// 3. Add updatedAt
	body["updatedAt"] = time.Now()

	// 4. Update in DB
	err := h.service.UpdateProfile(email, body)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Update failed",
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Profile updated successfully",
	})
}
