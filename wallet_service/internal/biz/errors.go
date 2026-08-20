package biz

import "errors"

var (
	ErrInvalidAmount       = errors.New("amount must be greater than zero")
	ErrInsufficientBalance = errors.New("insufficient balance")
	ErrSelfTransfer        = errors.New("cannot transfer to self")

	ErrWalletNotFound       = errors.New("wallet not found")
	ErrSystemWalletNotFound = errors.New("system escrow wallet not found")

	ErrDatabaseTxFailed = errors.New("database transaction failed")
)