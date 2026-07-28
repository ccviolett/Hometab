package response

import "github.com/gofiber/fiber/v2"

func OK(c *fiber.Ctx, data interface{}) error {
	return c.JSON(data)
}

func Created(c *fiber.Ctx, data interface{}) error {
	return c.Status(fiber.StatusCreated).JSON(data)
}

func Message(c *fiber.Ctx, msg string) error {
	return c.JSON(fiber.Map{"message": msg})
}

func Error(c *fiber.Ctx, status int, detail string) error {
	return c.Status(status).JSON(fiber.Map{"detail": detail})
}

func NotFound(c *fiber.Ctx, resource string) error {
	return Error(c, fiber.StatusNotFound, resource+" not found")
}

func BadRequest(c *fiber.Ctx, detail string) error {
	return Error(c, fiber.StatusBadRequest, detail)
}

func InternalError(c *fiber.Ctx, detail string) error {
	return Error(c, fiber.StatusInternalServerError, detail)
}
