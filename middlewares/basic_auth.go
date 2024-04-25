package middlewares

import (
	"crypto/subtle"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func BasicAuthMiddleware() echo.MiddlewareFunc {
	return middleware.BasicAuth(func(user, pass string, ctx echo.Context) (bool, error) {
		adminUser := os.Getenv("ADMIN_USERNAME")
		adminPass := os.Getenv("ADMIN_PASSWORD")
		if subtle.ConstantTimeCompare([]byte(user), []byte(adminUser)) == 1 &&
			subtle.ConstantTimeCompare([]byte(pass), []byte(adminPass)) == 1 {
			return true, nil
		}
		return false, nil
	})
}
