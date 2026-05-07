package middleware

import (
	"net/http"
	"strings"
	"user-service/internal/response"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

func JWTMiddleware(secret string) echo.MiddlewareFunc {

	return func(next echo.HandlerFunc) echo.HandlerFunc {

		return func(c echo.Context) error {

			authHeader := c.Request().Header.Get("Authorization")

			// Missing header
			if authHeader == "" {
				response.SendError(c.Response().Writer, http.StatusUnauthorized, "user unauthorized")
				return nil
			}

			// Invalid format
			if !strings.HasPrefix(authHeader, "Bearer ") {
				response.SendError(c.Response().Writer, http.StatusUnauthorized, "invalid authorization format")
				return nil
			}

			tokenString := strings.TrimPrefix(authHeader, "Bearer ")

			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				return []byte(secret), nil
			})

			// Invalid token
			if err != nil || !token.Valid {
				response.SendError(c.Response().Writer, http.StatusUnauthorized, "invalid token")
				return nil
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				response.SendError(c.Response().Writer, http.StatusUnauthorized, "invalid token claims")
				return nil
			}

			email, ok := claims["sub"].(string)
			if !ok {
				response.SendError(c.Response().Writer, http.StatusUnauthorized, "invalid token payload")
				return nil
			}

			// attach user to context
			c.Set("email", email)

			return next(c)
		}
	}
}
