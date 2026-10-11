# Reporte de cobertura

Cobertura de los tests automatizados (`go test -cover ./...`), actualizada al cerrar cada historia.

## Cobertura actual

- **Fecha:** 11/10/2026
- **Última historia con cobertura medida:** HU-03 · Consultar estado del proyecto (rama `feature/HU-03-consultar-proyecto`, pendiente de integrar a `main` por Pull Request)
- **Comando:** `go test -cover ./...`

| Paquete | Cobertura |
|---|---|
| `internal/project` | 100.0% |

```
$ go test -cover ./...
ok  	seguimiento-medicion-gpc/internal/project	0.887s	coverage: 100.0% of statements
```

## Historial

| Fecha | Historia | Paquete | Cobertura | Observaciones |
|---|---|---|---|---|
| 10/10/2026 | HU-01 | `internal/project` | 95.7% | Sin cubrir: el error de `List()` del repositorio, que en memoria nunca falla (se cubre con TEC-05) |
| 11/10/2026 | HU-01 (ajustes de la revisión) | `internal/project` | 100.0% | Se cubrió el error del repositorio con un repositorio falso |
| 11/10/2026 | HU-03 | `internal/project` | 100.0% | Se agregó `Get`/`GetByID` con cobertura total; verificación con `-race` hecha por Peñalbé Hernán y por el CI del PR |
