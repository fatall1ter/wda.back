package infra

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// Settings layout view parameters
type Settings struct {
	Proxy             string `json:"proxy"`
	VisibleOnline     string `json:"online_visible"`
	VisibleQueue      string `json:"queue_visible"`
	VisibleReport     string `json:"report_visible"`
	VisibleMonitoring string `json:"monitoring_visible"`
}

// apiSettings docs
// @Summary Get settings
// @Description get settings for gui configure
// @Produce  json
// @Tags settings
// @Success 200 {object} infra.Settings
// @Failure 404 {object} infra.ErrResponse
// @Failure 500 {object} infra.ErrResponse
// @Router /v1/layout/settings [get]
func (s *Server) apiSettings(c echo.Context) error {
	s.log.Debugf("request settings: %v", s.guiSettings)
	return c.JSON(http.StatusOK, s.guiSettings)
}
