package domain

import "errors"

var (
	// ErrNotFound no encontrado
	ErrNotFound = errors.New("recurso no encontrado")
	// ErrAlreadyExists ya existe
	ErrAlreadyExists = errors.New("el recurso ya existe")
	// ErrUnauthorized no autorizado
	ErrUnauthorized = errors.New("no autorizado")
	// ErrForbidden acceso prohibido
	ErrForbidden = errors.New("acceso prohibido")
	// ErrInvalidInput entrada invalida
	ErrInvalidInput = errors.New("datos de entrada inválidos")
	// ErrInternal error interno
	ErrInternal = errors.New("error interno")
)
