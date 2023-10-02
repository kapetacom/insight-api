package logging

import "github.com/labstack/echo/v4"

func LogHandler(mode string) echo.HandlerFunc {
	return func(c echo.Context) error {

		if mode == "kubernetes-only" {
			return LogByInstanceID(c)
		}
		return GCPLogHandler(c)
	}
}
