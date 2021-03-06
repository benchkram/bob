// Package generated provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/deepmap/oapi-codegen version v1.9.0 DO NOT EDIT.
package generated

// Item defines model for Item.
type Item struct {
	CreatedAt int    `json:"createdAt"`
	Id        string `json:"id"`
	Name      string `json:"name"`
}

// CreateItemHandlerJSONBody defines parameters for CreateItemHandler.
type CreateItemHandlerJSONBody string

// CreateItemHandlerJSONRequestBody defines body for CreateItemHandler for application/json ContentType.
type CreateItemHandlerJSONRequestBody CreateItemHandlerJSONBody
