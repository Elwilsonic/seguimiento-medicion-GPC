# HU-03 · Consultar estado del proyecto

> Issue del repositorio: #16

## Historia de usuario

Como Agile Enabler, quiero consultar un proyecto y su estado actual, para saber en qué situación está.

## Objetivo

Obtener un proyecto guardado a partir de su identificador, con su nombre, descripción y estado actual, para que el Agile Enabler (y más adelante TEC-01 y el dashboard) puedan ver en qué situación está cada proyecto.

Es una operación de solo lectura: no modifica el proyecto ni sus reglas de negocio. La creación es HU-01 y la modificación (incluido el cambio de estado) es HU-02.

## Alcance

- **Incluye:** consultar un proyecto por ID a través del `Manager` del paquete `internal/project`, con protección frente a accesos concurrentes sin protección (CA-03.4).
- **No incluye:** listar proyectos, modificar el proyecto o su estado (HU-02), la exposición HTTP (TEC-01) ni ningún filtro o búsqueda por nombre.

## Contrato

### Entradas

| Campo | Tipo | Obligatorio | Descripción |
|---|---|---|---|
| ID | entero | Sí | Identificador del proyecto a consultar |

No hay más entradas: nombre, descripción y estado son datos que se devuelven, no se reciben.

### Salidas esperadas

- **Si el proyecto existe:** el proyecto con ese ID, como **copia por valor** (nunca un puntero al estado interno), con:
  - **ID:** el identificador consultado.
  - **Nombre:** el guardado.
  - **Descripción:** la guardada (puede estar vacía, según RN-01.4).
  - **Estado:** el estado actual guardado: `PLANIFICADO`, `EN_CURSO` o `FINALIZADO` (RN-01.6).
- **Si no existe:** un error que indica el motivo (ver "Condiciones de error"), y no se devuelve ningún proyecto.

### Errores

| Condición | Error | Efecto |
|---|---|---|
| No existe un proyecto con ese ID (incluye ID inexistente, 0 o negativo) | `ErrProjectNotFound` ("no existe un proyecto con ese id") | No se devuelve ningún proyecto |
| El repositorio falla al buscar | El error que devuelve el repositorio | No se devuelve ningún proyecto |

`ErrProjectNotFound` se declara en `get.go`, junto a la operación de consulta, para no tocar `project.go` (archivo de HU-01) y evitar conflictos de merge con HU-02. Es exportado y comparable con `errors.Is`, siguiendo la decisión de HU-01 sobre errores identificables.

## Reglas de negocio

1. **RN-03.1 Búsqueda por identificador.** Se devuelve el único proyecto cuyo ID coincide exactamente con el recibido. Los IDs los asignó el repositorio (RN-01.5); la consulta no los transforma ni los normaliza.
2. **RN-03.2 Proyecto inexistente.** Si ningún proyecto guardado tiene el ID recibido, se devuelve `ErrProjectNotFound`. Esto incluye IDs que nunca se asignaron (por ejemplo, uno mayor al último creado) e IDs imposibles (0 o negativos), porque en todos los casos el motivo es el mismo: no existe un proyecto con ese identificador. No se define un error aparte para IDs inválidos.
3. **RN-03.3 Copia por valor.** La consulta devuelve una copia del proyecto. Modificar el resultado devuelto no altera el proyecto guardado (CA-03.3). El repositorio devuelve copias y el `Manager` no expone referencias internas.
4. **RN-03.4 Solo lectura.** La consulta no cambia el proyecto, su estado ni ningún otro dato guardado. El cambio de estado es exclusividad de HU-02.

## Decisión sobre entidades referenciadas

HU-03 referencia una entidad por ID: el proyecto consultado. **Sí se valida su existencia**: si el proyecto no existe, la operación falla con `ErrProjectNotFound` y no devuelve un valor vacío silenciosamente. Es la condición de error central de la historia (CA-03.2) y evita que el llamador trabaje sobre un proyecto inexistente.

## Restricciones

- **Organización del código.** Sigue la organización de HU-01 en el paquete `internal/project`: el código de producción se agrega en un archivo propio, `get.go`, que contiene el método `Get` del `Manager` y el error `ErrProjectNotFound`. No se modifican los archivos de HU-01 (salvo la excepción de `repository.go`, `memory_repository.go` y `create_test.go`, ver abajo).
- **Pasos godog.** Los pasos de los escenarios BDD de esta historia son **código de test** y van en un archivo propio de test, `get_features_test.go`, con su runner `TestFeatures_HU03ConsultarProyecto`. No se toca `features_test.go` de HU-01.
- **Operación compartida con HU-02 y HU-04.** `Repository.GetByID` y `Manager.Get` son **la** búsqueda por ID del paquete: HU-02 (modificar proyecto) y HU-04 (crear ítem del backlog, que valida que el proyecto exista) van a reutilizarlas. No se redefine ni se duplica en esas historias; si alguna necesita variaciones, se resuelve reutilizando esta operación o acordándolo antes de codificar.
- **Interfaz de repositorio.** HU-01 definió solo `Save` y `List`; "las historias siguientes agregan las operaciones que necesiten" (acta del Sprint 1, decisión de HU-01). HU-03 agrega a la interfaz `Repository` la operación `GetByID(id int) (Project, error)`, que devuelve `ErrProjectNotFound` cuando no existe. La implementación en memoria la cumple y TEC-05 (PostgreSQL) también deberá cumplirla. La lógica de negocio sigue sin acceder al almacenamiento directamente.
- **Consecuencia de ampliar la interfaz.** Agregar `GetByID` a la interfaz `Repository` obliga a actualizar el repositorio falso de los tests de HU-01 (`failingRepository` en `create_test.go`) para que siga cumpliendo la interfaz. Además de `repository.go` y `memory_repository.go`, ese es el único archivo de HU-01 que se toca.
- **Estado distinto del inicial.** Como HU-02 (que cambia el estado) aún no existe, el escenario y el test que consultan un proyecto en `EN_CURSO` usan un repositorio falso de prueba (`fixedStatusRepository`, definido junto a los tests de esta historia) que guarda los proyectos con un estado fijo. Es código de test: no agrega comportamiento al código de producción.
- **Copias por valor.** Tanto el repositorio como el `Manager` devuelven valores (`Project`), nunca punteros a estructuras internas.
- **Errores identificables.** `ErrProjectNotFound` es exportado y comparable con `errors.Is`, para que TEC-01 y los tests puedan distinguirlo.
- **Concurrencia.** `Get` toma el mutex del `Manager`. El criterio es "no se escribe código sin un test que lo pida": la protección se justifica y verifica con CA-03.4, cuyo test de concurrencia corre bajo `go test -race` (el CI de TEC-02 ya lo ejecuta). Sin CA-03.4 y su test, el mutex no entraría por "uniformidad".
- **Librerías.** Solo la librería estándar en el código; godog solo en los tests.

## Casos límite

| Caso | Resultado esperado |
|---|---|
| Consultar el primer proyecto creado (ID 1) | Se devuelve con sus datos completos |
| Consultar un ID mayor al último asignado (por ejemplo, 999) | Error: `ErrProjectNotFound` |
| Consultar cuando no hay ningún proyecto creado | Error: `ErrProjectNotFound` |
| Consultar el ID 0 | Error: `ErrProjectNotFound` |
| Consultar un ID negativo (por ejemplo, -1) | Error: `ErrProjectNotFound` |
| Proyecto con descripción vacía | Se devuelve con la descripción vacía |
| Proyecto en estado distinto de `PLANIFICADO` (por ejemplo `EN_CURSO`) | Se devuelve con el estado actual guardado |
| Modificar el proyecto devuelto (nombre, descripción o estado) | El proyecto guardado no cambia; reconsultarlo devuelve los datos originales |
| Consultas simultáneas mientras se crean proyectos | Ningún acceso concurrente sin protección; `go test -race` no reporta condiciones de carrera |

## Criterios de aceptación

- **CA-03.1** Se obtiene el proyecto por su identificador, con nombre, descripción y estado actual.
- **CA-03.2** Consultar un proyecto que no existe devuelve `ErrProjectNotFound` (incluye ID inexistente, 0, negativo, y consultar sin proyectos creados).
- **CA-03.3** Modificar el resultado devuelto no altera el proyecto guardado; reconsultar el proyecto devuelve los datos originales.
- **CA-03.4** Consultar mientras se crean proyectos en simultáneo no produce accesos concurrentes sin protección: el test de concurrencia corre con `go test -race` (el CI ya lo ejecuta) y no reporta condiciones de carrera.

> Nota: la "solo lectura" quedó como regla RN-03.4 porque no es verificable como criterio por sí solo.

## Comparación con el acta de Planning

El acta del Sprint 1 (`docs/actas/sprint-1-planning.md`) describe HU-03 como "Consultar estado del proyecto" (Issue #16, 2 Story Points, historia de referencia por ser solo lectura, responsable Sosa Ricardo, orden 4 del sprint). **El alcance coincide**: consulta por ID de solo lectura, sin modificación ni listado. No hay diferencia entre el acta y esta SDD.

## Comparación con los criterios del Issue #16 (borrador)

| CA en el Issue #16 (borrador) | En esta SDD | Ajuste |
|---|---|---|
| CA-03.1 Se obtiene el proyecto por su identificador, con nombre, descripción y estado | CA-03.1 (igual) | Ninguno |
| CA-03.2 Consultar un proyecto que no existe devuelve error | CA-03.2 (precisado) | Se identifica el error (`ErrProjectNotFound`) y se incluyen ID 0 y negativo |
| CA-03.3 Modificar el resultado devuelto no altera el proyecto guardado | CA-03.3 (igual) | Ninguno |
| — | CA-03.4 (nuevo) | Concurrencia verificada con `go test -race`, para justificar el mutex con un test |

**Texto definitivo para reemplazar la sección "Criterios de aceptación" del Issue #16** (a actualizar cuando la SDD se apruebe):

```markdown
## Criterios de aceptación

- [ ] CA-03.1 Se obtiene el proyecto por su identificador, con nombre, descripción y estado.
- [ ] CA-03.2 Consultar un proyecto que no existe devuelve error (`ErrProjectNotFound`), incluidos ID inexistente, 0 y negativo.
- [ ] CA-03.3 Modificar el resultado devuelto no altera el proyecto guardado.
- [ ] CA-03.4 Consultar mientras se crean proyectos en simultáneo no produce accesos concurrentes sin protección (verificado con `go test -race`).
```

## Decisiones tomadas (dudas resueltas con el Agile Enabler)

1. IDs 0 o negativos → `ErrProjectNotFound`, sin error aparte (RN-03.2).
2. HU-03 no cambia el estado; el cambio de estado es de HU-02 (RN-03.4).
3. Firma elegida: `GetByID(id int) (Project, error)`, devolviendo `ErrProjectNotFound` (patrón de errores de HU-01).
4. Mutex sí, justificado por CA-03.4 y su test con `-race` (no por uniformidad).
5. La exposición HTTP (por ejemplo, `GET /proyectos/{id}`) es alcance de TEC-01.
