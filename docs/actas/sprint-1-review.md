# Acta de Review — Sprint 1 (El MVP)

> Issue del repositorio: #64

## Datos generales

- **Proyecto:** Software Metrics & Estimation
- **Sprint:** Sprint 1 — El MVP
- **Período del Sprint:** 09/10/2026 al 15/10/2026
- **Fecha de la Review:** 15/10/2026
- **Sprint Goal:** Poder crear, modificar y consultar proyectos y cargar y listar su Product Backlog desde una aplicación ejecutable, con cada cambio validado por integración continua.

## Participantes

| Rol | Integrante | Asistió |
|---|---|---|
| Product Architect | Profesores de la cátedra | Sí |
| Agile Enabler | Peñalbé Hernán | Sí |
| Product Builder | Buttini Cristobal | Sí |
| Product Builder | Cardozo Leandro | Sí |
| Product Builder | García Santiago | Sí |
| Product Builder | Sosa Ricardo | Sí |

## Resultado del Sprint Goal

**Cumplido.** Las cinco historias comprometidas y las dos tareas técnicas están integradas en `main`. La aplicación arranca con `go run ./cmd/server` y expone por HTTP la creación, modificación y consulta de proyectos y la carga y el listado de su Product Backlog. Cada Pull Request de código pasó el check de CI "gofmt y tests" antes del merge; desde el PR #72 (11/10) el check corre `go test -race ./...`.

## Demo

Se mostró la aplicación corriendo (TEC-01) desde `main`, siguiendo la sección "Cómo ejecutar" del README:

1. Crear un proyecto (`POST /proyectos`, HU-01) y verificar que un nombre repetido devuelve 409.
2. Consultar el proyecto (`GET /proyectos/{id}`, HU-03) y verificar que uno inexistente devuelve 404.
3. Modificar nombre, descripción y estado (`PUT /proyectos/{id}`, HU-02).
4. Cargar un ítem al backlog (`POST /proyectos/{id}/backlog`, HU-04) y verificar que un ítem sin criterios devuelve 400.
5. Listar el backlog del proyecto (`GET /proyectos/{id}/backlog`, HU-06).

La demo la hizo Peñalbé Hernán.

## Historias y tareas comprometidas

Las fechas de merge están en hora de Argentina.

| ID | Issue | Story Points | Responsable | Pull Request | Aprobado por | Merge | ¿Cumple la DoD? |
|---|---|---|---|---|---|---|---|
| TEC-02 | #32 | — | Peñalbé Hernán | #67 | García Santiago | 10/10 | ✅ Sí |
| HU-01 | #14 | 8 | Peñalbé Hernán | #68 (y ajustes #72, #82) | Cardozo Leandro | 11/10 | ✅ Sí, con una desviación del ciclo TDD documentada |
| HU-03 | #16 | 2 | Sosa Ricardo | #74 | García Santiago | 11/10 | ✅ Sí |
| HU-04 | #17 | 5 | García Santiago | #77 | Buttini Cristobal | 12/10 | ✅ Sí, con una desviación del ciclo TDD documentada |
| HU-06 | #21 | 3 | Buttini Cristobal | #78 | García Santiago | 12/10 | ✅ Sí |
| HU-02 | #15 | 3 | Cardozo Leandro | #80 | Buttini Cristobal | 13/10 | ✅ Sí, con una desviación del ciclo TDD documentada |
| TEC-01 | #23 | — | Peñalbé Hernán | #81 | Cardozo Leandro | 14/10 | ✅ Sí, con una desviación del ciclo TDD documentada |

Las cuatro desviaciones están documentadas en la evidencia de cada historia, con prueba de mutación, y se tratan en la Retrospectiva: en HU-01, HU-02 y TEC-01 se escribió en el GREEN una rama de código sin un test que la pidiera (el test se agregó después); en HU-04 se ajustaron en el GREEN dos pasos BDD con errores de texto.

El #73 (tests del recorte del espacio no separable en HU-01, observación de la revisión del PR #72) también se resolvió en el Sprint: PR #82, mergeado el 14/10. Además entró la historia opcional HU-05 (ver abajo).

## Velocidad

| | Story Points |
|---|---|
| Comprometidos | 21 |
| Completados | 21 |
| **Velocidad del Sprint 1** | **21** |

Las tareas técnicas (TEC-01 y TEC-02) no se estiman en Story Points (decisión 5 del acta de Planning).

## Historias opcionales

| ID | Issue | Estado | Destino |
|---|---|---|---|
| HU-05 · Consultar un ítem del backlog | #18 | ✅ Completada (Sosa Ricardo): PR #83, aprobado por Buttini Cristobal, mergeado el 15/10; sin desviaciones del ciclo TDD | — |
| HU-07 · Modificar un ítem del backlog | #19 | No se empezó | Sprint 2 |
| HU-08 · Cambiar el estado de un ítem | #20 | No se empezó | Sprint 2 |

HU-05 entró en el Sprint: se implementó el 14/10 y el 15/10 y se integró el último día. HU-07 y HU-08 pasan al Sprint 2, como preveía el acta de Planning. HU-05 no suma a la velocidad porque, como historia opcional, no se estimó en el Planning.

## Calidad al cierre

Medido sobre `main` (e1d19ea, con HU-05 integrada) con Go 1.27.1 en Linux:

```
$ go test -race -cover ./...
ok  	seguimiento-medicion-gpc/cmd/server	coverage: 28.6% of statements
ok  	seguimiento-medicion-gpc/internal/backlog	coverage: 100.0% of statements
ok  	seguimiento-medicion-gpc/internal/httpapi	coverage: 100.0% of statements
ok  	seguimiento-medicion-gpc/internal/project	coverage: 100.0% of statements
```

`cmd/server` no llega al 100% porque `main()` arranca el servidor y no se prueba con tests unitarios (ver la evidencia de TEC-01). El detalle está en `docs/reportes/cobertura.md`.

## Observaciones del Product Architect

No hubo observaciones de los profesores.

## Cambios al Product Backlog

- HU-07 y HU-08 pasan al Sprint 2.
- HU-05 no tiene endpoint HTTP: agregar `GET /items/{id}` al servidor se trata en el Planning del Sprint 2.
- Las mejoras del servidor HTTP observadas en la revisión del PR #81 (límite de tamaño del cuerpo, rechazo de texto después del JSON y timeouts de `ListenAndServe`) se tratan en el Planning del Sprint 2.

## Línea base

Al integrar esta acta se crea el tag `sprint-1` sobre `main`.

## Próximo paso

Retrospectiva del Sprint 1 (#65) y Planning del Sprint 2 (16/10/2026). Para el Sprint 2, que incluye TEC-05 (PostgreSQL en Docker), cada integrante tiene que tener Docker instalado y funcionando (riesgo registrado en el acta de Planning).
