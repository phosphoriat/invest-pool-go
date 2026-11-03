package middleware

import (
	"bytes"
	"encoding/json"
	"io"

	"github.com/gofiber/fiber/v3"
)

// DecodeJSON Decode one JSON request and reject unknown fields or trailing values
func DecodeJSON(c fiber.Ctx, target any) error {
	d := json.NewDecoder(bytes.NewReader(c.Body()))
	d.DisallowUnknownFields()

	if err := d.Decode(target); err != nil {
		return fiber.NewError(400, "Некорректный JSON или тип поля")
	}
	if err := d.Decode(new(any)); err != io.EOF {
		return fiber.NewError(400, "Ожидается один JSON объект")
	}

	return nil
}
