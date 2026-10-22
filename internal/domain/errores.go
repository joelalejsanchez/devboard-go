package domain

import "errors"

var (
	ErrNotFound      = errors.New("recurso no encontrado")
	ErrAlreadyExists = errors.New("el recurso ya existe")
	ErrUnauthorized  = errors.New("no autorizado")
	ErrForbidden     = errors.New("acceso prohibido")
	ErrInvalidInput  = errors.New("datos de entrada inválidos")
	ErrInternal      = errors.New("error interno")
)
