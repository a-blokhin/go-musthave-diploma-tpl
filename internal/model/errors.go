package model

import "errors"

var ErrUserAlreadyExists = errors.New("user already exists")

type UserAlreadyExistsError struct {
	Login string
}

func (e *UserAlreadyExistsError) Error() string {
	return ErrUserAlreadyExists.Error()
}

func (e *UserAlreadyExistsError) Unwrap() error {
	return ErrUserAlreadyExists
}

var ErrInvalidCredentials = errors.New("invalid credentials")

type InvalidCredentialsError struct {
}

func (e *InvalidCredentialsError) Error() string {
	return ErrInvalidCredentials.Error()
}

func (e *InvalidCredentialsError) Unwrap() error {
	return ErrInvalidCredentials
}

var ErrOrderNotFound = errors.New("order not found")

type OrderNotFoundError struct {
	Number string
}

func (e *OrderNotFoundError) Error() string {
	return ErrOrderNotFound.Error()
}

func (e *OrderNotFoundError) Unwrap() error {
	return ErrOrderNotFound
}

var ErrOrderAlreadyExists = errors.New("order already exists")

type OrderAlreadyExistsError struct {
	Number string
}

func (e *OrderAlreadyExistsError) Error() string {
	return ErrOrderAlreadyExists.Error()
}

func (e *OrderAlreadyExistsError) Unwrap() error {
	return ErrOrderAlreadyExists
}

var ErrInvalidOrderNumber = errors.New("invalid order number format")

type InvalidOrderNumberError struct {
	Number string
}

func (e *InvalidOrderNumberError) Error() string {
	return ErrInvalidOrderNumber.Error()
}

func (e *InvalidOrderNumberError) Unwrap() error {
	return ErrInvalidOrderNumber
}

var ErrInsufficientBalance = errors.New("insufficient balance")

type InsufficientBalanceError struct {
}

func (e *InsufficientBalanceError) Error() string {
	return ErrInsufficientBalance.Error()
}

func (e *InsufficientBalanceError) Unwrap() error {
	return ErrInsufficientBalance
}

var ErrInvalidWithdrawalOrder = errors.New("invalid withdrawal order number")

type InvalidWithdrawalOrderError struct {
	Order string
}

func (e *InvalidWithdrawalOrderError) Error() string {
	return ErrInvalidWithdrawalOrder.Error()
}

func (e *InvalidWithdrawalOrderError) Unwrap() error {
	return ErrInvalidWithdrawalOrder
}
