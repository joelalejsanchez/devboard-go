# JSON y Validadores

JSON (JavaScript Object Notation) es un formato de texto común para intercambiar datos entre clientes y servidores. Una API HTTP suele recibir JSON en el cuerpo de una solicitud y devolver JSON en la respuesta. En Go, la biblioteca estándar `encoding/json` convierte entre JSON y valores de Go; la validación comprueba después que los valores recibidos tengan sentido para la aplicación.

## encoding/json

El paquete `encoding/json` incluye funciones y tipos para codificar (Go a JSON) y decodificar (JSON a Go). `json.Marshal` genera un arreglo de bytes con el documento JSON, mientras `json.Unmarshal` lee bytes existentes y los guarda en un valor destino:

```go
type User struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

user := User{Name: "Ana", Email: "ana@example.com"}
data, err := json.Marshal(user)
if err != nil {
	return err
}
fmt.Println(string(data)) // {"name":"Ana","email":"ana@example.com"}

var decoded User
if err := json.Unmarshal(data, &decoded); err != nil {
	return err
}
```

Al decodificar se pasa un puntero para que `Unmarshal` pueda modificar el valor. Los campos exportados (que comienzan con mayúscula) son los que puede codificar o llenar por defecto. Los campos no exportados se ignoran.

Para manejar cuerpos HTTP sin cargar manualmente todo el contenido, se puede usar `json.Decoder` directamente sobre `r.Body` y `json.Encoder` directamente sobre `http.ResponseWriter`:

```go
var input CreateUserRequest
if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
	http.Error(w, "JSON inválido", http.StatusBadRequest)
	return
}

w.Header().Set("Content-Type", "application/json")
if err := json.NewEncoder(w).Encode(input); err != nil {
	// Registrar el error: quizá la respuesta ya haya comenzado a escribirse.
	return
}
```

`Decode` normalmente decodifica un valor JSON; si la API requiere que el cuerpo contenga exactamente un documento y ningún dato adicional, se puede hacer una segunda decodificación y comprobar que devuelva `io.EOF`. También conviene limitar el tamaño del cuerpo de la solicitud para evitar que una petición envíe datos ilimitados; en HTTP se puede envolver el cuerpo con `http.MaxBytesReader`.

Por defecto, `Unmarshal` ignora propiedades JSON que no correspondan con campos de la estructura. Esto facilita compatibilidad evolutiva, pero quizá no sea deseable para una API que debe rechazar campos desconocidos. Para eso, `json.Decoder` ofrece `DisallowUnknownFields`. Los errores de decodificación pueden indicar sintaxis inválida o incompatibilidad de tipos, como recibir texto donde se espera un número.

JSON distingue entre `null`, valores concretos y propiedades ausentes. Un campo `string` no permite diferenciar entre una propiedad omitida y una cadena vacía una vez decodificado: en ambos casos su valor puede ser `""`. Si es necesario distinguirlos, se pueden usar punteros, tipos opcionales propios o un tipo como `json.RawMessage` para procesar el valor explícitamente.

## JSON tags

Un tag de estructura indica a `encoding/json` cómo representar un campo. Por ejemplo, `json:"email"` fija el nombre de la propiedad JSON. Sin tag, los nombres exportados se convierten en nombres JSON basados en el nombre del campo, pero es recomendable declarar los nombres explícitamente en los tipos que forman parte de una API:

```go
type CreateUserRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}
```

Opciones comunes:

- `json:"name,omitempty"`: omite el campo cuando su valor es el valor cero del tipo (por ejemplo `""`, `0`, `false`, un puntero nil o una colección vacía/nil).
- `json:"-"`: excluye el campo completamente tanto al codificar como al decodificar.

La forma general es `json:"nombre,opcion1,opcion2"`, sin espacios alrededor de las comas. Por ejemplo:

```go
type UserResponse struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email,omitempty"`
	Secret    string `json:"-"`
}
```

`omitempty` puede ocultar campos legítimos. En particular, no permite distinguir automáticamente el cero de un valor intencional del valor ausente. Si el cliente debe poder enviar `false`, `0` o `""` como un valor explícito, considera usar un puntero o una representación opcional.

Los tags controlan el formato, no la validez ni la seguridad de los datos. No conviene serializar directamente modelos internos que contengan contraseñas, secretos o campos que el cliente no debería ver. Define estructuras específicas para solicitudes y respuestas públicas.

## Enums en strings

JSON no tiene un tipo enum propio. Una técnica habitual en Go es definir un tipo basado en `string` y constantes con los valores permitidos:

```go
type Status string

const (
	StatusPending   Status = "pending"
	StatusActive    Status = "active"
	StatusSuspended Status = "suspended"
)
```

El campo se puede codificar como una cadena JSON normal:

```go
type Account struct {
	Status Status `json:"status"`
}
```

Al decodificar, el valor se convierte al tipo `Status`, pero el tipo string por sí solo no restringe los valores: JSON con `"status":"unknown"` también se puede decodificar. Se debe validar después de decodificar:

```go
func (s Status) IsValid() bool {
	switch s {
	case StatusPending, StatusActive, StatusSuspended:
		return true
	default:
		return false
	}
}
```

Y al procesar la solicitud:

```go
if !account.Status.IsValid() {
	http.Error(w, "status inválido", http.StatusBadRequest)
	return
}
```

Esta separación mantiene sencillo el formato JSON y deja explícita la regla de dominio. Si el valor no reconocido debe producir un error durante la propia decodificación, se puede implementar `UnmarshalJSON` en el tipo para validar al leerlo.

## UnmarshalJSON

Un tipo puede implementar la interfaz `json.Unmarshaler` para controlar cómo se decodifica su representación:

```go
type Unmarshaler interface {
	UnmarshalJSON(data []byte) error
}
```

Es útil cuando un valor necesita reglas de lectura propias, como aceptar solo ciertos strings para un enum, validar un formato específico o admitir varias representaciones JSON. Este ejemplo rechaza estados desconocidos:

```go
func (s *Status) UnmarshalJSON(data []byte) error {
	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return fmt.Errorf("status debe ser un string: %w", err)
	}

	status := Status(value)
	if !status.IsValid() {
		return fmt.Errorf("status desconocido: %q", value)
	}

	*s = status
	return nil
}
```

Se recibe `data` como JSON sin procesar, por eso primero se decodifica al tipo básico deseado. Como se está personalizando `Status`, llamar a `json.Unmarshal(data, s)` desde este mismo método volvería a invocar `Status.UnmarshalJSON` y causaría recursión. Decodificar a un `string` u otro tipo auxiliar evita ese problema.

La función debe asignar el valor al receptor solo cuando la lectura y la validación terminan correctamente. Así se evita dejar un valor parcialmente actualizado ante un error. Los mensajes devueltos se incorporan al error de decodificación de la estructura que contiene el campo.

Para tipos que usan la representación JSON normal salvo por algún detalle, puede definirse un tipo auxiliar nuevo con la misma estructura subyacente y convertir entre tipos. Ese patrón ayuda a reutilizar la codificación estándar sin llamar recursivamente al método personalizado.

## Validator v10

`github.com/go-playground/validator/v10` es una biblioteca externa que valida estructuras mediante tags `validate`. Se usa después de decodificar JSON: `encoding/json` comprueba que el documento pueda convertirse al tipo de Go, y el validador comprueba reglas adicionales, como que un campo sea requerido o que un email tenga un formato aceptable.

Primero se crea un validador y se anotan los campos:

```go
type CreateUserRequest struct {
	Name  string `json:"name" validate:"required,min=2,max=80"`
	Email string `json:"email" validate:"required,email"`
}

validate := validator.New()
```

Después de decodificar, se valida el valor:

```go
var input CreateUserRequest
if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
	http.Error(w, "JSON inválido", http.StatusBadRequest)
	return
}

if err := validate.Struct(input); err != nil {
	// Devolver un error de validación apropiado.
	return
}
```

Los tags se evalúan en el orden escrito. Algunas reglas frecuentes son `required`, `min`, `max`, `oneof=a b c`, `email` y `omitempty`. La interpretación de `min` y `max` depende del tipo del campo: para una cadena se refiere a su longitud; para números, al valor; y para colecciones, al tamaño.

Si `validate.Struct` devuelve un error de tipo `validator.ValidationErrors`, se pueden inspeccionar los errores campo por campo. Para una API pública, conviene transformar esos errores a mensajes estables y entendibles por el cliente, en vez de devolver directamente mensajes internos de la biblioteca:

```go
if err := validate.Struct(input); err != nil {
	var validationErrors validator.ValidationErrors
	if errors.As(err, &validationErrors) {
		for _, fieldErr := range validationErrors {
			log.Printf("campo %s no pasó la regla %s", fieldErr.Field(), fieldErr.Tag())
		}
	}
	http.Error(w, "datos inválidos", http.StatusBadRequest)
	return
}
```

Las validaciones de formato no sustituyen las reglas de negocio. Que un email tenga forma válida no demuestra que exista; que un ID sea positivo no demuestra que el recurso esté disponible. También valida en la capa apropiada las relaciones entre campos y las restricciones que dependen del estado de la aplicación o de la base de datos.

Como `validator/v10` no forma parte de la biblioteca estándar, un programa que use el ejemplo debe añadir el módulo como dependencia, por ejemplo con `go get github.com/go-playground/validator/v10`. Los tipos de solicitud pueden llevar tags tanto `json` como `validate`; cada biblioteca interpreta únicamente sus propios tags.
