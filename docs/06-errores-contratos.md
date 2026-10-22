# Errores y Contratos

Una API define un contrato entre servidor y cliente: qué solicitudes acepta, qué respuestas devuelve y cómo representa tanto los resultados como los errores. En Go, el manejo de errores tiene dos partes relacionadas, pero distintas: representar fallos dentro del programa mediante valores `error`, y traducir esos fallos a respuestas HTTP claras y estables.

## Sentinel errores

Un sentinel error es un valor de error definido para representar una condición conocida que otros componentes pueden reconocer. Normalmente se declara como variable a nivel de paquete con `errors.New`:

```go
var ErrNotFound = errors.New("recurso no encontrado")
```

Una función puede devolver ese valor cuando no encuentra el recurso:

```go
func (r *UserRepository) Get(ctx context.Context, id int64) (User, error) {
	user, err := r.find(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return User{}, ErrNotFound
	}
	if err != nil {
		return User{}, fmt.Errorf("buscar usuario %d: %w", id, err)
	}
	return user, nil
}
```

El operador `%w` envuelve un error y conserva la cadena causal. En las capas superiores se usa `errors.Is` para saber si un error envuelto corresponde al sentinel, sin comparar mensajes de texto:

```go
user, err := repository.Get(ctx, id)
if errors.Is(err, ErrNotFound) {
	// Convertir esta condición a una respuesta HTTP 404.
}
```

`errors.As` sirve para encontrar un error de un tipo concreto dentro de la cadena de errores. Se usa cuando el error proporciona datos estructurados, por ejemplo, un error de validación con información de los campos.

Los sentinel errors son útiles cuando una condición forma parte del contrato entre paquetes. Conviene exportarlos solo si los consumidores necesitan tomar decisiones basadas en ellos; si el error es un detalle interno, puede mantenerse privado. Evita comparar `err.Error()` con una cadena, porque los mensajes pueden cambiar y no son identificadores estables. También evita envolver con `%w` un error interno que no quieras exponer como parte del contrato público de tu paquete.

Una capa HTTP suele traducir errores del dominio a códigos de estado: no encontrado a `404`, entrada inválida a `400` o `422` según el contrato, falta de autenticación a `401`, falta de autorización a `403` y fallos inesperados a `500`. El cliente recibe una respuesta segura y consistente; los detalles técnicos completos se registran en el servidor.

## RFC 7807

RFC 7807 definió el formato **Problem Details for HTTP APIs**: una estructura común para comunicar detalles legibles por máquinas en el cuerpo de una respuesta de error HTTP. Su objetivo es complementar el código de estado con información específica, sin inventar un formato de error distinto para cada endpoint. El tipo de contenido JSON es `application/problem+json`.

El objeto define cinco miembros estándar:

- `type`: URI que identifica la clase de problema. Si se omite, equivale a `about:blank`.
- `title`: resumen breve y estable del tipo de problema.
- `status`: código HTTP asociado. Es informativo; el código real de la respuesta debe coincidir.
- `detail`: explicación de esta ocurrencia, útil para que el consumidor entienda o corrija el problema.
- `instance`: URI que identifica esta ocurrencia concreta.

Se pueden añadir miembros de extensión específicos del tipo, como una lista de campos inválidos. Los clientes deben ignorar extensiones que no reconozcan, lo que permite evolucionar el contrato.

Ejemplo de respuesta:

```http
HTTP/1.1 400 Bad Request
Content-Type: application/problem+json
```

```json
{
  "type": "https://api.example.com/problems/validation-error",
  "title": "La solicitud no es válida",
  "status": 400,
  "detail": "Uno o más campos no cumplen las reglas requeridas.",
  "instance": "/requests/req-7f31",
  "errors": [
    { "field": "email", "reason": "debe tener un formato válido" }
  ]
}
```

El `title` describe la clase de problema; `detail` puede variar entre ocurrencias. No incluyas en `detail` trazas, consultas, nombres de tablas, rutas internas ni secretos. El formato describe el problema que la API presenta al consumidor, no es un mecanismo para filtrar detalles de depuración.

RFC 7807 fue reemplazado por RFC 9457. Este apartado lo conserva porque aparece en documentación y APIs existentes; para diseños nuevos, consulta el apartado RFC 9457. [RFC 7807 en el RFC Editor](https://www.rfc-editor.org/rfc/rfc7807.html) indica su obsolescencia y referencia el estándar sucesor.

## RFC 9457

RFC 9457 es la versión vigente del estándar Problem Details y reemplaza RFC 7807. Mantiene el mismo modelo básico de cinco miembros (`type`, `title`, `status`, `detail` e `instance`) y el tipo de contenido `application/problem+json`, con aclaraciones y ajustes en la especificación. La intención de uso sigue siendo la misma: asociar un formato común a un código HTTP apropiado.

Una estructura Go sencilla puede representar el formato y añadir extensiones cuando una API lo necesite:

```go
type Problem struct {
	Type     string `json:"type"`
	Title    string `json:"title"`
	Status   int    `json:"status,omitempty"`
	Detail   string `json:"detail,omitempty"`
	Instance string `json:"instance,omitempty"`
}

func writeProblem(w http.ResponseWriter, status int, problem Problem) {
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(status)
	problem.Status = status
	_ = json.NewEncoder(w).Encode(problem)
}
```

Es importante que `status` en el cuerpo coincida con el estado HTTP. El estado HTTP es la señal primaria para clientes, proxies y herramientas; el miembro del JSON es una copia útil para consumidores que procesan el cuerpo por separado.

Usa un `type` estable y específico para cada clase de problema, preferiblemente una URI bajo el control de tu API que documente el problema. No construyas clientes que dependan de analizar `title` o `detail`: esos campos son texto para personas y pueden cambiar o localizarse. Para información que el cliente deba procesar, usa campos de extensión con nombres y tipos documentados.

Problem Details no reemplaza todos los cuerpos de error. Si la respuesta sigue siendo una representación normal del recurso, puede ser más claro conservar el formato propio de ese recurso. Para errores HTTP de una API, RFC 9457 ofrece un contrato común y extensible. Consulta [RFC 9457 en el RFC Editor](https://www.rfc-editor.org/rfc/rfc9457.html) como referencia normativa.

## Paginación Cursor Based

La paginación limita el volumen de resultados por respuesta. En la paginación por cursor, el cliente recibe un token que representa dónde continuar y lo envía en la siguiente solicitud. Por ejemplo:

```http
GET /orders?limit=20
```

La respuesta puede contener una página y un cursor para continuar:

```json
{
  "data": [
    { "id": "ord_105", "created_at": "2026-10-21T10:00:00Z" }
  ],
  "next_cursor": "eyJjcmVhdGVkX2F0Ijoi..."
}
```

El cliente usa el cursor opaco en la próxima llamada:

```http
GET /orders?limit=20&after=eyJjcmVhdGVkX2F0Ijoi...
```

Un cursor no debería ser simplemente un número de página si el conjunto de datos cambia a menudo. Normalmente codifica la posición del último elemento según un orden estable, por ejemplo `(created_at, id)`. La consulta siguiente pide los elementos posteriores a esa tupla:

```sql
SELECT id, created_at, total
FROM orders
WHERE (created_at, id) > ($1, $2)
ORDER BY created_at, id
LIMIT $3;
```

El criterio de orden debe ser determinista y único en conjunto; añadir el ID como segundo criterio evita ambigüedades cuando varias filas comparten la misma fecha. Según el orden elegido, la comparación puede ser `<` en lugar de `>`.

El cursor suele tratarse como opaco: el cliente lo devuelve sin interpretarlo. Puede codificar los valores de posición en Base64 URL-safe, pero Base64 no cifra ni autentica los datos. Si el contenido no debe ser modificable, firma el cursor o guarda el estado en el servidor. Valida también límites de `limit` y rechaza cursores malformados o incompatibles.

Una paginación basada en cursor suele ser más estable y eficiente que `OFFSET` para recorrer conjuntos grandes, aunque no crea una instantánea automática: inserciones y eliminaciones entre solicitudes pueden afectar qué observa el cliente. Define claramente el orden, los filtros y la semántica del cursor. Si no hay más resultados, `next_cursor` puede omitirse, ser `null` o acompañarse de un indicador `has_more`; elige una convención y manténla consistente.

## Swaggo

Swaggo es un conjunto de herramientas de Go que lee comentarios declarativos en el código y genera documentación Swagger/OpenAPI. El proyecto `swaggo/swag` genera una especificación Swagger 2.0 (OpenAPI 2.0); otros componentes del ecosistema, como `http-swagger`, pueden servir una interfaz Swagger UI. Conviene confirmar qué versión de especificación genera la herramienta elegida antes de documentar una API.

Las anotaciones se escriben en comentarios Go. Las generales suelen estar junto a la función `main` y las anotaciones de endpoint encima de cada handler:

```go
// GetUser godoc
// @Summary      Obtiene un usuario
// @Description  Busca un usuario usando su identificador.
// @Tags         users
// @Produce      json
// @Param        id path int true "Identificador del usuario"
// @Success      200 {object} UserResponse
// @Failure      404 {object} Problem
// @Router       /users/{id} [get]
func GetUser(w http.ResponseWriter, r *http.Request) {
	// Implementación del endpoint.
}
```

Los comentarios describen el contrato: método, ruta, parámetros, cuerpos, respuestas y códigos posibles. Después se ejecuta el comando `swag init` para generar la documentación a partir de las anotaciones; los archivos generados se sirven o se integran con la interfaz correspondiente según la configuración del proyecto. La guía oficial de [swaggo/swag](https://github.com/swaggo/swag) documenta su instalación, formato de anotaciones y opciones del generador.

La documentación generada no sustituye la validación del handler ni garantiza por sí sola que el comportamiento real coincida con lo descrito. Mantén sincronizados los esquemas de solicitud y respuesta, documenta también los errores Problem Details y revisa los artefactos generados como parte del flujo de trabajo. No edites manualmente archivos generados si el siguiente `swag init` los sobrescribirá; cambia las anotaciones o la configuración fuente.
