package main

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"time"
)

type CreateTransactionData struct {
	Amount      int    `json:"valor" validate:"required,gt=0"`
	Description string `json:"descricao" validate:"required,min=1,max=10"`
	Type        string `json:"tipo" validate:"required,oneof=c d"`
}

type CreateTransactionResult struct {
	AccountLimit   int `json:"limite"`
	AccountBalance int `json:"saldo"`
}

type BankExtract struct {
	BalanceDetails   BalanceDetails `json:"saldo"`
	LastTransactions []Transaction `json:"ultimas_transacoes"`
}

type BalanceDetails struct {
	AccountBalance int `json:"total"`
	AccountLimit   int `json:"limite"`
	BalanceDate    time.Time `json:"data_extrato"`
}

type Transaction struct {
	Amount      int `json:"valor"`
	Type        string `json:"tipo"`
	Description string `json:"descricao"`
	CreatedAt   time.Time `json:"realizada_em"`
}

func CreateTransaction(clientId int, transactionData *CreateTransactionData) (*CreateTransactionResult, error) {
	tx, err := dbConnection.Begin(context.Background())
	if err != nil {
		return nil, err
	}

	defer tx.Rollback(context.Background())

	var accountLimit int
	var accountBalance int

	err = tx.QueryRow(context.Background(), "SELECT account_limit, balance FROM clients WHERE id = $1 FOR UPDATE", clientId).Scan(&accountLimit, &accountBalance)
	if err != nil {
		return nil, err
	}

	var newAccountBalance int

	if transactionData.Type == "d" {
		newAccountBalance = accountBalance - transactionData.Amount
	} else {
		newAccountBalance = accountBalance + transactionData.Amount
	}

	if (accountLimit + newAccountBalance) < 0 {
		return nil, errors.New("LIMIT_EXCEEDED")
	}

	_, err = tx.Exec(context.Background(), "INSERT INTO transactions(client_id,amount,operation,description) values ($1, $2, $3, $4)", clientId, transactionData.Amount, transactionData.Type, transactionData.Description)
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec(context.Background(), "UPDATE clients SET balance = $1 WHERE id = $2", newAccountBalance, clientId)
	if err != nil {
		return nil, err
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return nil, err
	}

	result := CreateTransactionResult{
		AccountBalance: newAccountBalance,
		AccountLimit:   accountLimit,
	}

	return &result, nil
}

func GetBankExtract(clientId int) (*BankExtract, error) {
	rows, _ := dbConnection.Query(context.Background(), "SELECT balance, account_limit, now() FROM clients WHERE id = $1", clientId)
	balanceDetails, err := pgx.CollectOneRow(rows, pgx.RowToStructByPos[BalanceDetails])
	if err != nil {
		return nil, err
	}
	
	rows, _ = dbConnection.Query(context.Background(), "SELECT amount, operation, description, created_at FROM transactions WHERE client_id = $1 ORDER BY id DESC LIMIT 10", clientId)
	lastTransactions, err := pgx.CollectRows(rows, pgx.RowToStructByPos[Transaction])
	if err != nil {
		return nil, err
	}
	
	result := BankExtract{
		BalanceDetails: balanceDetails,
		LastTransactions: lastTransactions,
	}
	
	return &result, nil
}
