package restapi

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"server-db/server"
	"server-db/server/rest-api/generated"
)

type API struct {
	server.Application
	router  *echo.Echo
	address string
}

func New(app server.Application) *API {
	r := echo.New()

	r.Use(middleware.Logger())

	api := &API{router: r, Application: app}

	generated.RegisterHandlers(r, api)

	return api
}

func (api *API) Run(addr string) error {
	go func() {
		err := api.router.Start(addr)
		if err != nil && err != http.ErrServerClosed {
			log.Print(err)
		}
	}()

	return nil
}

func (api *API) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := api.router.Shutdown(ctx)
	if err != nil {
		return err
	}

	return nil
}
