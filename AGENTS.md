# Guía para agentes — Promotarjetas Backend

## Propósito y alcance

Este repositorio es un BFF en Go que agrega promociones de BAC, Cuscatlán y
Agrícola, las normaliza y las sirve desde Redis mediante Gin. Mantén los cambios
pequeños, explícitos y compatibles con el contrato existente. No refactorices
áreas no relacionadas ni reviertas cambios locales que no sean tuyos.

La versión de Go de referencia es la declarada en `go.mod` (actualmente Go
1.25). Antes de incorporar una dependencia, prefiere la biblioteca estándar o
una solución ya presente en el repositorio.

## Arquitectura y contratos

- `main.go` solo compone la aplicación: configuración, Redis, servidor HTTP y
  programación del cron. La lógica de negocio no debe crecer allí.
- `controllers/` traduce HTTP a llamadas de aplicación y mapea errores a
  respuestas. No contiene scraping, reglas de categorización ni acceso directo
  a Redis.
- `services/` orquesta sincronización, normalización, orden y reglas de negocio.
- `integrations/` aísla una fuente externa por banco. Convierte sus respuestas
  al modelo unificado y no filtra detalles de su contrato al resto del sistema.
- `cache/` es la única capa que conoce Redis, sus claves, serialización y TTL.
- `models/` contiene los contratos compartidos. Conserva los nombres y etiquetas
  JSON de `PromocionUnificada` salvo que el cambio de contrato esté aprobado.
- `utils/` debe contener funciones pequeñas, puras y fáciles de probar.

Conserva, salvo autorización explícita, los endpoints y comportamientos públicos:

| Método | Ruta | Comportamiento |
| --- | --- | --- |
| `GET` | `/health` | Estado del servicio |
| `GET` | `/api/promotions` | Promociones normalizadas desde caché; sincroniza en un cache miss |
| `GET` | `/api/promotions/sync` | Fuerza una sincronización |

Los cambios en rutas, métodos HTTP, respuestas JSON, el orden de promociones,
la semántica de `CreatedAt`, la clave de caché o el TTL requieren actualización
del README, pruebas de compatibilidad y una justificación clara en el cambio.

## Convenciones de Go

- Escribe código simple y directo: nombres cortos y claros, paquetes en
  minúscula sin guiones bajos, retornos tempranos y sin `panic` como flujo de
  control. Evita retornos desnudos en funciones no triviales.
- Formatea todo Go con `gofmt`; conserva los grupos de imports estándar,
  internos y externos.
- Trata cada error. Añade contexto con `fmt.Errorf("operación recurso: %w", err)`
  cuando se propague. Usa `errors.Is` y `errors.As` para errores tipados o
  centinela. Solo ignora un error cuando sea deliberado y esté documentado.
- No uses `log.Fatal` fuera del arranque del proceso. Las capas inferiores
  devuelven errores para que el llamador decida cómo responder o registrar.
- Acepta interfaces pequeñas cuando haya una dependencia que aislar; defínelas
  en el paquete consumidor y devuelve tipos concretos. No crea interfaces solo
  por anticipación.
- Evita estado global mutable nuevo. Para código nuevo o refactorizado, inyecta
  clientes, relojes, fetchers y almacenamiento mediante constructores. Diseña
  tipos cuyo valor cero sea útil cuando resulte natural y mantén coherentes los
  receptores por valor o puntero.
- Usa `strings.Builder`, `strings.Join` o preasignación de slices cuando sean
  pertinentes; no optimices con `sync.Pool` sin medir un cuello de botella.

## HTTP, integraciones y concurrencia

- Toda operación que pueda bloquear o hacer I/O debe aceptar `context.Context`
  como primer parámetro y respetar su cancelación. Propaga el contexto de Gin a
  servicios, caché y solicitudes HTTP nuevas.
- Las integraciones externas deben crear solicitudes con contexto, usar clientes
  con timeout explícito, cerrar cuerpos de respuesta y validar códigos HTTP y
  datos antes de normalizarlos. Incluye el banco y la operación al envolver el
  error, sin registrar secretos ni cuerpos sensibles.
- Conserva la tolerancia a fallos parciales: el fallo de un banco se registra y
  no invalida los resultados correctos de los demás, salvo que el requisito del
  cambio indique lo contrario.
- Al introducir goroutines, define quién las espera y cómo terminan. No envíes a
  canales sin considerar cancelación o ausencia de receptor; protege todo estado
  compartido con el mecanismo de sincronización adecuado.
- La sincronización debe continuar siendo de una sola ejecución concurrente. No
  cambies esa garantía sin una política explícita para esperas, errores y
  respuestas de solicitudes simultáneas.
- Devuelve mensajes de error HTTP seguros y consistentes; registra internamente
  el detalle necesario. No expongas API keys, URLs privadas ni errores crudos de
  proveedores.

## Datos, configuración y seguridad

- Obtén configuración únicamente desde `config/` y variables de entorno. Nunca
  incorpores valores de `.env`, contraseñas, tokens o `CUSCATLAN_API_KEY` en
  código, pruebas, documentación o logs.
- Todo acceso a Redis pasa por `cache/`. Maneja los cache misses y los errores de
  serialización/conexión de forma explícita; no asumas que el cliente o los bytes
  de caché siempre existen.
- Mantén la normalización de categorías y el orden de salida deterministas. Las
  nuevas reglas de bancos deben producir IDs estables para preservar `CreatedAt`.
- No amplíes CORS, logging ni rutas de administración sin configurar y documentar
  sus restricciones de producción.

## Pruebas y validación

Todo código nuevo o modificado debe incluir pruebas proporcionales a su riesgo.
Prefiere pruebas de tabla y subpruebas para transformaciones; usa `httptest` y
fakes/mocks de interfaces para HTTP y Redis, sin llamar a bancos reales ni exigir
credenciales.

Como mínimo, cubre cuando aplique:

- normalización, categorías, orden e IDs de promociones;
- éxito, fallo y dato inválido de cada integración modificada;
- cache hit, cache miss, error de Redis y preservación de `CreatedAt`;
- respuestas HTTP correctas y errores seguros;
- sincronizaciones concurrentes, cancelación y prevención de carreras para
  cualquier cambio de concurrencia.

Para un cambio Go, ejecuta antes de entregar:

```bash
gofmt -w <archivos-go-modificados>
go vet ./...
go test ./...
go test -race ./...
go build ./...
```

Ejecuta `go mod tidy` y revisa `go.mod`/`go.sum` únicamente si el cambio alteró
dependencias. Si están instalados, también ejecuta `staticcheck ./...` y
`golangci-lint run`. Para cambios de contenedores, valida además:

```bash
docker compose config
```

Informa los comandos no ejecutados y el motivo. No modifiques archivos generados
o binarios de compilación, salvo que formen parte intencional del cambio.

## Entrega

- Actualiza `README.md` cuando cambien configuración, endpoints, ejecución o
  comportamiento visible.
- Resume qué cambió, los riesgos o compatibilidades relevantes y la validación
  ejecutada.
- Mantén los commits y diffs enfocados; no incluyas secretos, archivos `.env`,
  cachés locales ni artefactos accidentales.
