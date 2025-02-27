package controller

import (
	"github.com/labstack/echo/v4"
	"github.com/roundfeather/configuration-server/internal/config"
	"github.com/roundfeather/configuration-server/internal/utils"
	"io/fs"
	"net/http"
	"os"
	"path"

	log "github.com/sirupsen/logrus"
)

func GetV1Configuration(c echo.Context) error {
	lf := utils.GetLogFields(c)
	serviceName := c.QueryParam("service")
	log.WithFields(lf).Info(serviceName)

	dir := "config-repo"
	env := c.Get("cfg").(config.Cfg).Profile

	configSources := []string{
		"application.yml",
		"application-" + env + ".yml",
		serviceName + ".yml",
		serviceName + "-" + env + ".yml",
	}

	properties := map[string]string{}

	for _, fileName := range configSources {
		configFile, e := fs.Glob(os.DirFS(dir), fileName)
		if e == nil && len(configFile) > 0 {
			cfg := utils.Configuration{Path: path.Join(dir, configFile[0])}
			cfg.GetProperties(properties)
		}
	}

	return c.JSON(http.StatusOK, properties)
}
