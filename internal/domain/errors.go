package domain

import "errors"

var (
	// ErrNotFound indicates that a requested resource does not exist.
	ErrNotFound = errors.New("recurso no encontrado")
	// ErrAlreadyExists indicates that a resource with the same identity already exists.
	ErrAlreadyExists = errors.New("el recurso ya existe")
	// ErrUnauthorized indicates that the request is not authenticated.
	ErrUnauthorized = errors.New("no autorizado")
	// ErrForbidden indicates that the authenticated caller lacks permission.
	ErrForbidden = errors.New("acceso prohibido")
	// ErrInvalidInput indicates that the input is invalid.
	ErrInvalidInput = errors.New("datos de entrada inválidos")
	// ErrInternal indicates that an internal error occurred.
	ErrInternal = errors.New("error interno")
)
