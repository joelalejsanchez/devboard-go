# HTTP en Go

HTTP es el protocolo que permite que un cliente (por ejemplo, un navegador o una aplicación móvil) haga solicitudes a un servidor y reciba respuestas. En un backend, normalmente recibimos una solicitud, validamos sus datos, ejecutamos una operación y devolvemos un estado y un cuerpo.

Una solicitud contiene, entre otras cosas:

- Un **método** que expresa la acción (`GET`, `POST`, `PUT`, `PATCH`, `DELETE`).
- Una **ruta** que identifica el recurso (`/health`, `/users/42`).
- **Headers** con metadatos, como `Content-Type` o credenciales.
- Opcionalmente, un **cuerpo** con datos, frecuente en `POST` y `PUT`.

La respuesta contiene un código de estado, headers y, a menudo, un cuerpo. Por ejemplo, `200 OK` indica éxito, `201 Created` que se creó un recurso, `400 Bad Request` que la solicitud no es válida y `500 Internal Server Error` que ocurrió un error inesperado en el servidor.

## El paquete `net/http`

La biblioteca estándar de Go incluye `net/http`; con ella podemos crear servidores HTTP sin instalar dependencias externas. El flujo básico conecta un servidor con un multiplexor de rutas y uno o más handlers:

```go
mux := http.NewServeMux()
mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte("ok"))
})

server := &http.Server{
	Addr:    ":8000",
	Handler: mux,
}

log.Fatal(server.ListenAndServe())
```

`http.Server` escucha en una dirección y entrega cada solicitud a su `Handler`. `http.NewServeMux` crea un enrutador (multiplexor): decide qué handler debe atender cada solicitud según su método y ruta. En una aplicación más grande conviene configurar también límites de tiempo, como hace este proyecto en `internal/server/server.go`.

`ListenAndServe` se queda esperando solicitudes. Si el servidor no puede iniciar o deja de funcionar, devuelve un error. En una aplicación real suele ser mejor registrarlo con contexto y cerrar el servidor ordenadamente ante una señal de apagado, en vez de terminar directamente con `log.Fatal`.

## `Handler` y `HandlerFunc`

La abstracción central de `net/http` es la interfaz `http.Handler`. Cualquier tipo que implemente este método puede atender solicitudes:

```go
type Handler interface {
	ServeHTTP(http.ResponseWriter, *http.Request)
}
```

El `ResponseWriter` sirve para construir la respuesta; el `Request` contiene la solicitud entrante. Un tipo propio puede implementar la interfaz para agrupar lógica y dependencias:

```go
type HealthHandler struct{}

func (h *HealthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
```

`http.HandlerFunc` es un adaptador que convierte una función con la firma correcta en un `http.Handler`. Por eso podemos registrar una función con `HandleFunc`, sin crear un tipo:

```go
http.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
})
```

En cambio, `mux.Handle` recibe un valor `http.Handler`. El `HealthHandler` de este proyecto implementa esa interfaz y se registra con `RegisterRoute`, que por dentro llama a `ServeMux.Handle`.

### Escribir la respuesta

Los headers deben definirse antes de escribir el estado o el cuerpo. La primera llamada a `WriteHeader` envía el código de estado; si se escribe el cuerpo antes, Go envía automáticamente `200 OK`.

```go
w.Header().Set("Content-Type", "application/json")
w.WriteHeader(http.StatusCreated)
_, err := w.Write([]byte(`{"id":42}`))
```

No es necesario llamar a `WriteHeader` cuando se responde con `200 OK`. Para JSON, es más seguro usar `json.NewEncoder(w).Encode(valor)`, que codifica el valor y lo escribe en la respuesta:

```go
w.Header().Set("Content-Type", "application/json")
if err := json.NewEncoder(w).Encode(response); err != nil {
	// La respuesta puede haber comenzado a escribirse; normalmente se registra el error.
	return
}
```

Una vez enviados los headers y el cuerpo, ya no se puede cambiar el estado de la respuesta. Por ello, valida la solicitud y decide el estado antes de empezar a escribir. `http.Error(w, mensaje, codigo)` es útil para respuestas de error de texto; si tu API necesita un formato JSON uniforme, escribe el JSON de error explícitamente.

## Routing con `http.ServeMux`

Desde Go 1.22, los patrones del `ServeMux` estándar pueden incluir método y ruta. Por ejemplo, `"GET /health"` solo coincide con solicitudes `GET` a `/health`. Los patrones más específicos tienen prioridad sobre los más generales.

Los segmentos entre llaves capturan partes de la ruta y se leen con `PathValue`:

```go
mux.HandleFunc("GET /users/{id}", func(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	_, _ = fmt.Fprintf(w, "usuario: %s", id)
})
```

Para datos de la query string se usa `r.URL.Query()`; por ejemplo, `/users?limit=10`:

```go
limit := r.URL.Query().Get("limit")
```

Si una ruta existe pero no admite el método solicitado, el `ServeMux` responde `405 Method Not Allowed`. Si no hay una ruta coincidente, responde `404 Not Found`. Una API también puede registrar una ruta raíz como `"/"`, que coincide con rutas más amplias; hay que considerar su alcance al definir el resto de patrones.
