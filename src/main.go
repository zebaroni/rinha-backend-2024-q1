package main

import (
	"context"
	"github.com/go-playground/validator/v10"
	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"os"
)

var validate *validator.Validate
var dbConnection *pgxpool.Pool

func CreateTransactionHandler(c *fiber.Ctx) error {
	transactionData := new(CreateTransactionData)

	clientId, err := c.ParamsInt("id")
	if err != nil {
		return c.SendStatus(422)
	}

	err = c.BodyParser(&transactionData)
	if err != nil {
		return c.SendStatus(422)
	}

	err = validate.Struct(transactionData)
	if err != nil {
		return c.SendStatus(422)
	}

	client, err := GetClient(clientId)
	if err != nil {
		return c.SendStatus(500)
	}
	if client == nil {
		return c.SendStatus(404)
	}

	transactionResult, err := CreateTransaction(clientId, transactionData)
	if err != nil {
		if err.Error() == "LIMIT_EXCEEDED" {
			return c.SendStatus(422)
		} else {
			return c.SendStatus(500)
		}
	}

	return c.JSON(transactionResult)
}

func GetExtractHandler(c *fiber.Ctx) error {
	clientId, err := c.ParamsInt("id")
	if err != nil {
		return c.SendStatus(422)
	}
	
	client, err := GetClient(clientId)
	if err != nil {
		return c.SendStatus(500)
	}
	if client == nil {
		return c.SendStatus(404)
	}
	
	bankExtract, err := GetBankExtract(clientId)
	
	return c.JSON(bankExtract)
}

func main() {
	con, err := pgxpool.New(context.Background(), "postgres://rinha:rinha@localhost:5432/rinha")
	dbConnection = con

	if err != nil {
		os.Exit(1)
	}

	defer dbConnection.Close()

	validate = validator.New(validator.WithRequiredStructEnabled())

	app := fiber.New(fiber.Config{
		JSONEncoder: json.Marshal,
		JSONDecoder: json.Unmarshal,
	})

	app.Post("/clientes/:id/transacoes", CreateTransactionHandler)
	app.Get("/clientes/:id/extrato", GetExtractHandler)

	log.Fatal(app.Listen(":9999"))
}
