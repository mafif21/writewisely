package server

import "github.com/gofiber/fiber/v2"

func New(db Pinger) *fiber.App {
	app := fiber.New(fiber.Config{AppName: "WriteWisely API"})
	RegisterRoutes(app, NewHealthHandler(db))
	return app
}
