# Reporte de cobertura

Cobertura de los tests automatizados (`go test -cover ./...`), actualizada al cerrar cada historia.

## Cobertura actual

- **Fecha:** 12/10/2026
- **Última historia con cobertura medida:** HU-04 · Crear ítem del backlog (rama `feature/HU-04-crear-item-backlog`, pendiente de integrar a `main` por Pull Request)
- **Comando:** `go test -cover ./...` (en el entorno local se ejecutó con binarios `-cover` y `-test.coverprofile` por una política de Control de Aplicaciones de Windows; ver la evidencia de HU-04)

| Paquete | Cobertura |
|---|---|
| `internal/backlog` | 100.0% |
| `internal/project` | 100.0% |

```
$ go tool cover -func=backlog-cover.out
seguimiento-medicion-gpc/internal/backlog/create.go:11:			Create			100.0%
seguimiento-medicion-gpc/internal/backlog/create.go:55:			normalizePriority	100.0%
seguimiento-medicion-gpc/internal/backlog/create.go:68:			cleanCriteria		100.0%
seguimiento-medicion-gpc/internal/backlog/manager.go:21:		NewManager		100.0%
seguimiento-medicion-gpc/internal/backlog/memory_repository.go:13:	NewMemoryRepository	100.0%
seguimiento-medicion-gpc/internal/backlog/memory_repository.go:20:	Save			100.0%
total:									(statements)		100.0%

$ go tool cover -func=project-cover.out
seguimiento-medicion-gpc/internal/project/create.go:6:			Create			100.0%
seguimiento-medicion-gpc/internal/project/get.go:10:			Get			100.0%
seguimiento-medicion-gpc/internal/project/manager.go:15:		NewManager		100.0%
seguimiento-medicion-gpc/internal/project/memory_repository.go:13:	NewMemoryRepository	100.0%
seguimiento-medicion-gpc/internal/project/memory_repository.go:18:	Save			100.0%
seguimiento-medicion-gpc/internal/project/memory_repository.go:26:	List			100.0%
seguimiento-medicion-gpc/internal/project/memory_repository.go:32:	GetByID			100.0%
total:									(statements)		100.0%
```

## Historial

| Fecha | Historia | Paquete | Cobertura | Observaciones |
|---|---|---|---|---|
| 10/10/2026 | HU-01 | `internal/project` | 95.7% | Sin cubrir: el error de `List()` del repositorio, que en memoria nunca falla (se cubre con TEC-05) |
| 11/10/2026 | HU-01 (ajustes de la revisión) | `internal/project` | 100.0% | Se cubrió el error del repositorio con un repositorio falso |
| 11/10/2026 | HU-03 | `internal/project` | 100.0% | Se agregó `Get`/`GetByID` con cobertura total; verificación con `-race` hecha por Peñalbé Hernán y por el CI del PR |
| 12/10/2026 | HU-04 | `internal/backlog` (nuevo) e `internal/project` | 100.0% y 100.0% | Nuevo paquete del backlog con cobertura total; sin regresiones en `internal/project`; `-race` (CA-04.7) lo verifica el CI porque el entorno local no tiene cgo |
