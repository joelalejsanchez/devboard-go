# Patrones de Diseño

Un patrón de diseño es una solución conocida para un problema recurrente de organización del código. No es una regla que haya que aplicar siempre ni una función incluida necesariamente en el lenguaje. En Go, muchos patrones se expresan con funciones, interfaces pequeñas y composición de tipos, procurando que la solución sea fácil de leer y mantener.

## Patrones Utiles en GO

- **Patrón Repository:** separa el acceso a datos de la lógica de aplicación. Un caso de uso depende de operaciones como buscar o guardar, en vez de depender directamente de SQL o de un cliente de base de datos.

  ```go
  type UserRepository interface {
  	FindByID(ctx context.Context, id int64) (User, error)
  	Save(ctx context.Context, user User) error
  }
  ```

  Una implementación puede usar PostgreSQL y otra una estructura en memoria para pruebas. La interfaz conviene declararla donde se consume y con los métodos realmente necesarios. No es obligatorio crear una interfaz para cada struct: si solo hay una implementación y no se necesita sustituirla, una abstracción adicional puede complicar el código.

- **Patron Strategy:** permite elegir entre distintos algoritmos que cumplen el mismo contrato. Por ejemplo, un sistema puede admitir varias formas de calcular un costo:

  ```go
  type PricingStrategy interface {
  	Calculate(order Order) Money
  }

  type Checkout struct {
  	pricing PricingStrategy
  }
  ```

  `Checkout` usa la estrategia recibida sin conocer sus detalles. Así se puede cambiar el cálculo mediante otra implementación, sin llenar el flujo principal de condicionales. En Go la estrategia puede ser una interfaz o, para operaciones simples, una función con una firma apropiada.

- **Patron Adapter:** convierte la interfaz de un componente en la interfaz que espera otro. Es útil para integrar una biblioteca externa, un servicio legado o una API de terceros sin propagar sus tipos por toda la aplicación.

  ```go
  type PaymentGateway interface {
  	Charge(ctx context.Context, amount Money) error
  }

  type ExternalPaymentClient interface {
  	CreatePayment(ctx context.Context, cents int64) error
  }

  type paymentAdapter struct {
  	client ExternalPaymentClient
  }

  func (a paymentAdapter) Charge(ctx context.Context, amount Money) error {
  	return a.client.CreatePayment(ctx, amount.Cents())
  }
  ```

  El adaptador traduce la llamada y los tipos; la lógica de aplicación conserva su contrato propio.

- **Patron Middleware:** envuelve un handler para ejecutar comportamiento transversal antes o después del procesamiento, por ejemplo logging, autenticación o recuperación de panics. Es una forma de composición especialmente común en servidores HTTP. El documento [04-middlewares.md](/C:/joelsanchz/dev/learning/go-devtalles/devboard/docs/04-middlewares.md) desarrolla este patrón aplicado a `net/http`.

Estos patrones pueden combinarse. Por ejemplo, un caso de uso puede depender de un repository y una estrategia; un adaptador puede implementar la interfaz del repository y un middleware puede rodear el handler que invoca el caso de uso. La intención es separar responsabilidades, no añadir capas por costumbre.

## Functional Options

Functional Options es una técnica para configurar un valor con opciones opcionales, evitando constructores con muchos parámetros posicionales o variantes de constructor. Cada opción es una función que modifica la configuración:

```go
type Server struct {
	addr         string
	readTimeout  time.Duration
	writeTimeout time.Duration
}

type ServerOption func(*Server)

func WithReadTimeout(timeout time.Duration) ServerOption {
	return func(s *Server) {
		s.readTimeout = timeout
	}
}

func WithWriteTimeout(timeout time.Duration) ServerOption {
	return func(s *Server) {
		s.writeTimeout = timeout
	}
}

func NewServer(addr string, options ...ServerOption) *Server {
	s := &Server{
		addr:         addr,
		readTimeout:  5 * time.Second,
		writeTimeout: 10 * time.Second,
	}
	for _, option := range options {
		option(s)
	}
	return s
}
```

El consumidor puede elegir solo las opciones que necesita:

```go
server := NewServer(":8000", WithReadTimeout(2*time.Second))
```

Los valores por defecto viven en el constructor, por lo que la configuración sigue siendo válida aunque no se pase ninguna opción. Las funciones `With...` comunican la intención en la llamada y permiten añadir opciones nuevas sin cambiar la lista de argumentos posicionales existente.

Este patrón resulta útil cuando hay varias opciones opcionales y una configuración por defecto razonable. Para dos o tres valores obligatorios, parámetros explícitos o una estructura de configuración suelen ser más sencillos. Si una opción puede ser inválida, define cómo se reporta: una función de opción que devuelve `error` permite validar durante la construcción, por ejemplo `type ServerOption func(*Server) error`.

## Decorator Pattern - Middleware

El patrón Decorator añade comportamiento a un objeto envolviéndolo con otro que satisface la misma interfaz. El objeto decorador delega la operación principal al componente interno e intercala su propia lógica. En Go se logra mediante composición; no se requiere herencia.

Por ejemplo, una interfaz de almacenamiento puede envolverse para registrar cada operación:

```go
type Store interface {
	Get(ctx context.Context, key string) (string, error)
}

type loggingStore struct {
	next   Store
	logger *slog.Logger
}

func (s loggingStore) Get(ctx context.Context, key string) (string, error) {
	start := time.Now()
	value, err := s.next.Get(ctx, key)
	s.logger.InfoContext(ctx, "store get",
		slog.Duration("duration", time.Since(start)),
		slog.Bool("error", err != nil),
	)
	return value, err
}
```

`loggingStore` implementa `Store`, así que puede sustituir al almacenamiento original. Se puede envolver una capa con otra —por ejemplo, logging y luego caché— y decidir el orden según el comportamiento deseado.

El middleware HTTP es una aplicación del mismo principio: recibe un `http.Handler`, devuelve otro `http.Handler` y delega mediante `ServeHTTP`. La firma de middleware facilita encadenar decoradores alrededor de rutas o del mux completo. No obstante, al envolver `http.ResponseWriter`, hay que preservar las capacidades adicionales que use el handler, como `http.Flusher` para streaming.

## Strategy Pattern

Strategy define una familia de algoritmos intercambiables detrás de un contrato común. El código que utiliza la estrategia depende del contrato, no de una implementación específica. Es útil cuando una decisión algorítmica debe variar según configuración, tipo de operación o contexto.

```go
type Notifier interface {
	Send(ctx context.Context, recipient, message string) error
}

type AlertService struct {
	notifier Notifier
}

func (s AlertService) Alert(ctx context.Context, recipient, message string) error {
	return s.notifier.Send(ctx, recipient, message)
}
```

`Notifier` podría tener implementaciones para correo, SMS o una salida de consola. `AlertService` puede usarlas sin cambiar su flujo. También se puede definir la estrategia con una función:

```go
type DiscountFunc func(Order) Money

func ApplyDiscount(order Order, discount DiscountFunc) Money {
	return discount(order)
}
```

No hace falta Strategy si solo hay un algoritmo y no se prevé variación: una función directa y clara puede bastar. Si aparecen varios `if` o `switch` que seleccionan comportamientos alternativos, extraer una estrategia puede hacer que cada variante sea más fácil de entender y probar.

## Anti-Patrones

Un anti-patrón es una solución que parece conveniente en el momento, pero tiende a producir problemas de mantenimiento, acoplamiento o complejidad. Algunos errores comunes al aplicar patrones en Go son:

- **Abstraer antes de tener una necesidad:** crear interfaces para cada tipo o anticipar muchas implementaciones futuras puede ocultar el flujo real. Introduce una interfaz cuando ayude a separar una dependencia, habilite una sustitución útil o simplifique una prueba.
- **Crear demasiadas capas:** un handler que llama a un servicio que solo reenvía a otro servicio, que finalmente delega sin aportar una regla o frontera clara, hace más difícil seguir una operación. Cada capa debería tener una responsabilidad comprensible.
- **Usar patrones como objetivo:** un patrón no mejora por sí solo el diseño. Elige la solución más pequeña que represente el problema actual; refactoriza cuando el cambio de comportamiento o la duplicación lo justifiquen.
- **Usar estado global mutable:** dependencias globales dificultan pruebas aisladas y pueden causar interferencias entre solicitudes concurrentes. Es preferible construir componentes y pasarles sus dependencias explícitamente.
- **Ocultar errores o convertirlos en texto:** descartar errores, hacer `panic` para fallos esperables o comparar `err.Error()` rompe la propagación y el tratamiento consistente. Devuelve errores, envuélvelos con contexto cuando aporte valor y clasifícalos con `errors.Is` o `errors.As`.
- **Crear interfaces demasiado grandes:** contratos con muchos métodos obligan a los consumidores a depender de operaciones que no usan. Mantén interfaces pequeñas y orientadas a quien las consume.

Una buena guía es empezar con código directo, observar dónde aparece el acoplamiento o la repetición y aplicar un patrón que resuelva ese problema concreto. El resultado debería ser más sencillo de leer, cambiar y probar que la versión anterior.
