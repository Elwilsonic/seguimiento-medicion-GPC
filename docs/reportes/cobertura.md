# Reporte de cobertura

Cobertura de los tests automatizados (`go test -cover ./...`), actualizada al cerrar cada historia.

## Cobertura actual

- **Fecha:** 17/10/2026
- **Última tarea con cobertura medida:** HU-07 · Modificar datos de un ítem (rama `feature/HU-07-modificar-item`, pendiente de integrar a `main` por Pull Request)
- **Comando:** `go test -cover ./...`

| Paquete | Cobertura |
|---|---|
| `cmd/server` | 28.6% |
| `internal/backlog` | 100.0% |
| `internal/httpapi` | 100.0% |
| `internal/project` | 100.0% |

`cmd/server` no llega al 100% porque `main()` arranca el servidor y bloquea en `http.ListenAndServe`, así que no se prueba con un test unitario; la elección del puerto está separada en `address`, con cobertura total (ver la [evidencia de TEC-01](../evidencias/TEC-01-tdd.md)).

```
$ go tool cover -func=server-cover.out
seguimiento-medicion-gpc/cmd/server/main.go:15:			main			0.0%
seguimiento-medicion-gpc/cmd/server/main.go:26:			address			100.0%
total:								(statements)		28.6%

$ go tool cover -func=httpapi-cover.out
seguimiento-medicion-gpc/internal/httpapi/server.go:63:		NewHandler		100.0%
seguimiento-medicion-gpc/internal/httpapi/server.go:75:		createProject		100.0%
seguimiento-medicion-gpc/internal/httpapi/server.go:90:		getProject		100.0%
seguimiento-medicion-gpc/internal/httpapi/server.go:105:	updateProject		100.0%
seguimiento-medicion-gpc/internal/httpapi/server.go:125:	createItem		100.0%
seguimiento-medicion-gpc/internal/httpapi/server.go:145:	listItems		100.0%
seguimiento-medicion-gpc/internal/httpapi/server.go:165:	pathID			100.0%
seguimiento-medicion-gpc/internal/httpapi/server.go:175:	statusFor		100.0%
seguimiento-medicion-gpc/internal/httpapi/server.go:193:	writeError		100.0%
seguimiento-medicion-gpc/internal/httpapi/server.go:202:	writeJSON		100.0%
seguimiento-medicion-gpc/internal/httpapi/server.go:208:	toProjectResponse	100.0%
seguimiento-medicion-gpc/internal/httpapi/server.go:212:	toItemResponse		100.0%
total:								(statements)		100.0%

$ go tool cover -func=backlog-cover.out
seguimiento-medicion-gpc/internal/backlog/create.go:11:			Create			100.0%
seguimiento-medicion-gpc/internal/backlog/create.go:55:			normalizePriority	100.0%
seguimiento-medicion-gpc/internal/backlog/create.go:68:			cleanCriteria		100.0%
seguimiento-medicion-gpc/internal/backlog/get.go:13:			GetItem			100.0%
seguimiento-medicion-gpc/internal/backlog/list.go:8:			ListItems		100.0%
seguimiento-medicion-gpc/internal/backlog/manager.go:21:		NewManager		100.0%
seguimiento-medicion-gpc/internal/backlog/memory_repository.go:13:	NewMemoryRepository	100.0%
seguimiento-medicion-gpc/internal/backlog/memory_repository.go:20:	Save			100.0%
seguimiento-medicion-gpc/internal/backlog/memory_repository.go:35:	GetByID			100.0%
seguimiento-medicion-gpc/internal/backlog/memory_repository.go:48:	Update			100.0%
seguimiento-medicion-gpc/internal/backlog/memory_repository.go:68:	ListByProject		100.0%
seguimiento-medicion-gpc/internal/backlog/update.go:22:			Update			100.0%
seguimiento-medicion-gpc/internal/backlog/update.go:75:			validateEstimatedHours	100.0%
total:									(statements)		100.0%

$ go tool cover -func=project-cover.out
seguimiento-medicion-gpc/internal/project/create.go:6:			Create			100.0%
seguimiento-medicion-gpc/internal/project/get.go:10:			Get			100.0%
seguimiento-medicion-gpc/internal/project/manager.go:15:		NewManager		100.0%
seguimiento-medicion-gpc/internal/project/memory_repository.go:13:	NewMemoryRepository	100.0%
seguimiento-medicion-gpc/internal/project/memory_repository.go:18:	Save			100.0%
seguimiento-medicion-gpc/internal/project/memory_repository.go:26:	List			100.0%
seguimiento-medicion-gpc/internal/project/memory_repository.go:32:	GetByID			100.0%
seguimiento-medicion-gpc/internal/project/memory_repository.go:43:	Update			100.0%
seguimiento-medicion-gpc/internal/project/update.go:19:			Update			100.0%
seguimiento-medicion-gpc/internal/project/update.go:64:			normalizeStatus		100.0%
total:									(statements)		100.0%
```

## Historial

| Fecha | Historia | Paquete | Cobertura | Observaciones |
|---|---|---|---|---|
| 10/10/2026 | HU-01 | `internal/project` | 95.7% | Sin cubrir: el error de `List()` del repositorio, que en memoria nunca falla (se cubre con TEC-05) |
| 11/10/2026 | HU-01 (ajustes de la revisión) | `internal/project` | 100.0% | Se cubrió el error del repositorio con un repositorio falso |
| 11/10/2026 | HU-03 | `internal/project` | 100.0% | Se agregó `Get`/`GetByID` con cobertura total; verificación con `-race` hecha por Peñalbé Hernán y por el CI del PR |
| 12/10/2026 | HU-04 | `internal/backlog` (nuevo) e `internal/project` | 100.0% y 100.0% | Nuevo paquete del backlog con cobertura total; sin regresiones en `internal/project`; `-race` (CA-04.7) lo verifica el CI porque el entorno local no tiene cgo |
| 12/10/2026 | HU-06 | `internal/backlog` e `internal/project` | 100.0% y 100.0% | `ListItems` y `ListByProject` entran con cobertura total; sin regresiones en `internal/project` ni en HU-04; `-race` (CA-06.5) lo verifica el CI porque el entorno local no tiene cgo |
| 13/10/2026 | HU-02 | `internal/project` e `internal/backlog` | 100.0% y 100.0% | `Update` y `normalizeStatus` entran con cobertura total (tras un ajuste: el test de contrato de `MemoryRepository.Update`, ver la desviación del ciclo TDD en la evidencia de HU-02); sin regresiones en backlog ni en HU-01/HU-03; `-race` (CA-02.7) lo verifica el CI y Hernán sobre Linux porque el entorno local no tiene cgo |
| 14/10/2026 | TEC-01 | `cmd/server` (nuevo), `internal/httpapi` (nuevo), `internal/backlog` e `internal/project` | 28.6%, 100.0%, 100.0% y 100.0% | Servidor HTTP con `net/http`; `internal/httpapi` llega al 100% tras un ajuste (test del id no numérico en las demás rutas, ver la desviación del ciclo TDD en la evidencia de TEC-01); `main()` no se prueba con tests unitarios; `-race` verificado sobre Linux |
| 15/10/2026 | HU-05 | `internal/backlog` | 100.0% | `GetItem` y `GetByID` entran con cobertura total; sin regresiones en HU-04, HU-06 ni en el resto de los paquetes; `-race` (CA-05.4) lo verifican Hernán sobre Linux y el CI porque el entorno local no tiene cgo |
| 17/10/2026 | HU-07 | `internal/backlog` | 100.0% | `Manager.Update` y `validateEstimatedHours` (`update.go`) y `MemoryRepository.Update` entran con cobertura total; el campo `EstimatedHours` entra en `Item` sin bajas; sin regresiones en HU-04, HU-05, HU-06 ni en el resto de los paquetes; `-race` (CA-07.6) lo verifica el CI del Pull Request porque el entorno local no tiene cgo |
