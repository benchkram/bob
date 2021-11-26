package restapi

import (
	"server-db/server"

	"github.com/labstack/echo/v4"

	"server-db/server/item"
	"server-db/server/rest-api/generated"
)

func (api *API) CreateItemHandler(c echo.Context) error {
	var name string

	err := c.Bind(&name)
	if err != nil {
		return c.JSON(400, "invalid name")
	}

	i, err := api.CreateItem(item.WithName(name))
	if err != nil {
		return c.JSON(500, "item creation failed: "+err.Error())
	}

	return c.JSON(200, generated.Item{
		Id:        i.Id(),
		Name:      i.Name(),
		CreatedAt: int(i.CreatedAt().Unix()),
	})
}

func (api *API) GetItemHandler(c echo.Context, itemId string) error {
	i, err := api.Application.Item(itemId)
	if err == server.ErrItemNotFound {
		return c.JSON(404, "item not found")
	} else if err != nil {
		return c.JSON(500, "item lookup failed: "+err.Error())
	}

	return c.JSON(200, generated.Item{
		Id:        i.Id(),
		Name:      i.Name(),
		CreatedAt: int(i.CreatedAt().Unix()),
	})
}

func (api *API) Ping(c echo.Context) error {
	return c.JSON(200, "pong!")
}
