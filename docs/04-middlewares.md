# Middlewares

## Middleware

Un middleware es una función que envuelve un `http.Handler` para añadir una tarea común antes o después de que este procese una solicitud. Por ejemplo, puede medir la duración de una petición, comprobar una credencial o recuperar un panic.

En `net/http`, la forma habitual es recibir el siguiente handler y devolver uno nuevo:

```go
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Lógica previa a la solicitud.
		next.ServeHTTP(w, r)
		// Lógica posterior a la solicitud.
	})
}
```

`next.ServeHTTP` delega el procesamiento al siguiente elemento de la cadena. Si el middleware no llama a `next`, la cadena se detiene en esa capa. Esto es útil, por ejemplo, cuando una validación falla y se responde con `401 Unauthorized` o `403 Forbidden`.

Los middlewares suelen configurarse con una función externa que recibe dependencias, como un logger, y devuelve la función que envuelve el handler:

```go
func Logger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			next.ServeHTTP(w, r)
			logger.InfoContext(r.Context(), "solicitud completada",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Duration("duration", time.Since(start)),
			)
		})
	}
}
```

Al componer capas, la más externa se ejecuta primero al entrar y última al salir. Por ejemplo, `Recovery(Logger(mux))` hace que `Recovery` envuelva a `Logger`, que a su vez envuelve al mux. El orden importa: una capa externa puede observar o manejar lo que ocurre dentro de las capas internas.

Un middleware de logging puede necesitar conocer el estado de la respuesta. Como `http.ResponseWriter` no permite consultar el estado enviado, suele envolverse en un tipo propio que registra `WriteHeader`. El primer `Write` sin una llamada previa a `WriteHeader` envía implícitamente `200 OK`. Al crear este wrapper hay que tener cuidado: `ResponseWriter` puede ofrecer otras interfaces, como `http.Flusher` para streaming; ocultarlas podría afectar a ciertos handlers.

Un middleware de recuperación puede usar `defer` y `recover` para capturar un panic que ocurra en la misma goroutine durante el procesamiento de la solicitud. Debe registrar el error y, si la respuesta aún no comenzó, devolver `500 Internal Server Error`. Si el handler ya envió headers o parte del cuerpo, no es posible reemplazar limpiamente esa respuesta. `recover` tampoco captura panics de goroutines independientes que el handler haya iniciado.

## Log/Slog

El paquete estándar `log/slog` permite escribir logs estructurados. En lugar de construir una línea de texto donde todos los datos quedan mezclados, cada evento tiene un mensaje y atributos con nombre, como `method`, `path`, `status` y `duration`. Los atributos facilitan filtrar y analizar logs, especialmente cuando se usa salida JSON.

Una instancia de `*slog.Logger` se crea con un handler. La biblioteca estándar ofrece `slog.NewTextHandler` para texto legible y `slog.NewJSONHandler` para JSON. El nivel (`Debug`, `Info`, `Warn`, `Error`) permite controlar qué eventos se emiten:

```go
logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
	Level: slog.LevelInfo,
}))

logger.Info("servidor iniciado",
	slog.String("addr", ":8000"),
)
```

También se puede usar `InfoContext`, `ErrorContext` y sus equivalentes para asociar el evento al contexto de la solicitud. Esto permite que los handlers y procesadores consideren valores contextuales cuando corresponda:

```go
logger.InfoContext(r.Context(), "solicitud recibida",
	slog.String("method", r.Method),
	slog.String("path", r.URL.Path),
)
```

En un middleware HTTP, normalmente se registra el evento después de que `next` retorne para incluir la duración y el estado final observado. Para registrar el estado, se puede envolver `ResponseWriter` y capturar la primera llamada a `WriteHeader`; si el handler escribe el cuerpo sin definir estado, el valor observado debe ser `200`. Conviene medir duración con `time.Since(start)`, que devuelve directamente `time.Duration`.

Los logs deben ser útiles para diagnosticar sin exponer información sensible. Evita registrar contraseñas, tokens, cookies o cuerpos completos que puedan contener datos personales. Además, registra errores con sus campos y contexto relevante, pero evita duplicar el mismo error en cada capa si dificulta identificar su origen.

## Context HTTP

`r.Context()` devuelve el contexto asociado a una solicitud HTTP. El contexto permite propagar cancelación, plazos límite y valores específicos de la operación a funciones que participan en su procesamiento. Es especialmente importante al acceder a bases de datos o a otros servicios, para que el trabajo pueda detenerse si el cliente cancela la petición o vence el plazo.

Por ejemplo, `QueryContext` acepta el contexto y puede cancelar la consulta cuando la solicitud deja de estar activa:

```go
rows, err := db.QueryContext(r.Context(), "SELECT id, name FROM users")
if err != nil {
	http.Error(w, "no se pudo consultar usuarios", http.StatusInternalServerError)
	return
}
defer rows.Close()
```

Para añadir un valor contextual se crea un contexto derivado. Se recomienda usar un tipo propio como clave para evitar colisiones entre paquetes:

```go
type contextKey string

const userIDKey contextKey = "userID"

ctx := context.WithValue(r.Context(), userIDKey, userID)
r = r.WithContext(ctx)
next.ServeHTTP(w, r)
```

El contexto derivado conserva cancelación y deadline del original. Los valores de contexto deben reservarse para datos que cruzan límites de API entre funciones, como una identidad autenticada o un identificador de trazabilidad. No es un reemplazo para parámetros explícitos ni un almacén genérico de configuración.

Cuando se crea un contexto con deadline o timeout, siempre debe llamarse la función de cancelación para liberar recursos:

```go
ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
defer cancel()

req, err := http.NewRequestWithContext(ctx, http.MethodGet, upstreamURL, nil)
```

No guardes `Request`, `ResponseWriter` ni el contexto de una solicitud para utilizarlos después de que la respuesta haya terminado. Tampoco pases `nil` como contexto; cuando una función no procesa una solicitud HTTP, usa un contexto válido como `context.Background()` o `context.TODO()` según el caso.

## Graceful shutdown

Un apagado ordenado (graceful shutdown) permite dejar de aceptar solicitudes nuevas y dar tiempo a las que ya están en curso para terminar antes de cerrar el proceso. Esto evita cortar respuestas a medias cuando el servidor se detiene por un despliegue, una señal del sistema o una interrupción controlada.

En Go, `http.Server.Shutdown(ctx)` inicia este proceso. El contexto establece cuánto tiempo se espera antes de forzar el cierre. Si el plazo vence, `Shutdown` devuelve un error y el servidor puede cerrarse de forma forzada con `Close`:

```go
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

if err := server.Shutdown(ctx); err != nil {
	log.Printf("apagado ordenado incompleto: %v", err)
	_ = server.Close()
}
```

Para iniciar el apagado al recibir una señal del sistema, se puede usar `signal.NotifyContext`:

```go
ctx, stop := signal.NotifyContext(
	context.Background(),
	os.Interrupt,
	syscall.SIGTERM,
)
defer stop()

<-ctx.Done() // Espera una señal de terminación.
```

En una aplicación real, el servidor suele ejecutarse en una goroutine para poder esperar la señal mientras `ListenAndServe` atiende solicitudes. Cuando llega la señal, se llama a `Shutdown` con un plazo razonable. `http.ErrServerClosed` es el error esperado al detener el servidor y normalmente no debe tratarse como una falla de inicio.

El apagado ordenado también requiere cerrar recursos de la aplicación, como conexiones a la base de datos, después de que el servidor deje de aceptar solicitudes y las solicitudes activas hayan terminado. El plazo debe ser coherente con los límites de tiempo de las operaciones para evitar que el proceso quede esperando indefinidamente.
