package biz

import "errors"

var (
	// Lỗi Logic Nghiệp vụ (Client lỗi)
	ErrInvalidAmount       = errors.New("amount must be greater than zero")
	ErrInsufficientBalance = errors.New("insufficient balance")
	ErrSelfTransfer        = errors.New("cannot transfer to self")

	// Lỗi Tài nguyên (Not Found)
	ErrWalletNotFound       = errors.New("wallet not found")
	ErrSystemWalletNotFound = errors.New("system escrow wallet not found")

	// Lỗi Hệ thống (Internal Server Error)
	ErrDatabaseTxFailed = errors.New("database transaction failed")
)
