# HU-02 · Modificar datos del proyecto

> Issue del repositorio: #15

## Historia de usuario

Como Agile Enabler, quiero modificar el nombre, la descripción y el estado de un proyecto, para mantener su información actualizada.

## Objetivo

Cambiar el nombre, la descripción y el estado de un proyecto ya guardado, identificado por su identificador, aplicando las mismas reglas del nombre que la creación (obligatorio, único y sin espacios en los extremos) y controlando que el estado sea uno de los válidos, para que el resto del sistema (la consulta de HU-03, el backlog y las métricas) trabaje siempre con información actualizada y consistente.

Es la operación de escritura pendiente de la épica E1: HU-01 crea el proyecto en `PLANIFICADO` (RN-01.6), HU-03 lo consulta y HU-02 es la única historia que lo modifica.

## Entradas

| Campo | Tipo | Obligatorio | Descripción |
|---|---|---|---|
| ID | entero | Sí | Identificador del proyecto a modificar |
| Nombre | texto | Sí | Nuevo nombre del proyecto |
| Descripción | texto | No | Nueva descripción; puede estar vacía |
| Estado | texto | Sí | Nuevo estado; se acepta en cualquier combinación de mayúsculas y minúsculas y con espacios en los extremos: se normaliza (sin espacios de extremos y en mayúsculas) y tiene que quedar en `PLANIFICADO`, `EN_CURSO` o `FINALIZADO` |

La modificación reemplaza los tres campos en una sola llamada: para cambiar un solo campo, el llamador reenvía los valores actuales de los demás (decisión 1 de "Decisiones tomadas").

## Salidas esperadas

- **Si la modificación es válida:** el proyecto actualizado, como **copia por valor**, con:
  - **ID:** el mismo del proyecto modificado (no cambia).
  - **Nombre:** el recibido, sin los espacios del principio y del final.
  - **Descripción:** la recibida, tal como llegó.
  - **Estado:** el recibido, normalizado (sin espacios de extremos y en mayúsculas), uno de los tres válidos.
- **Si la modificación no es válida:** un error que indica el motivo (ver "Condiciones de error"), y el proyecto guardado queda exactamente como estaba (nunca se aplica una modificación parcial).

## Reglas de negocio

1. **RN-02.1 Nombre obligatorio.** El nombre no puede quedar vacío ni contener solo espacios, igual que en la creación (RN-01.1). Se consideran espacios los caracteres de espacio en blanco Unicode que reconoce `strings.TrimSpace` (por ejemplo, espacios, tabs, saltos de línea y el espacio no separable).
2. **RN-02.2 Nombre sin espacios en los extremos.** El nombre nuevo se guarda sin los espacios del principio y del final (RN-01.2). Los espacios internos se conservan.
3. **RN-02.3 Nombre único, excluyendo el propio proyecto.** Aplica RN-01.3 al modificar: el nombre nuevo no puede repetir el de **otro** proyecto con ID distinto, comparando sin distinguir mayúsculas de minúsculas y sin contar los espacios de los extremos. Dejar el nombre del propio proyecto (igual o con otras mayúsculas) es válido: el proyecto se excluye de su propio control de duplicados.
4. **RN-02.4 Estado válido, sin distinguir mayúsculas ni espacios de los extremos.** Se acepta el estado escrito en cualquier combinación de mayúsculas y minúsculas y con espacios en los extremos (por ejemplo, "en_curso" o " En_Curso "): se normaliza quitando los espacios de los extremos y pasando a mayúsculas, igual que la prioridad de los ítems del backlog en HU-04. El valor normalizado tiene que ser uno de los tres estados válidos de RN-01.6 (`PLANIFICADO`, `EN_CURSO` o `FINALIZADO`); si no lo es (por ejemplo, "PAUSADO") o quedó vacío, es un error (`ErrInvalidStatus`). No hay transiciones prohibidas: cualquier estado válido puede pasar a cualquier otro.
5. **RN-02.5 Descripción opcional.** La descripción puede quedar vacía y se guarda tal como llega (RN-01.4).
6. **RN-02.6 Proyecto existente.** La modificación requiere que exista un proyecto con el ID recibido. Si no existe, se devuelve `ErrProjectNotFound` (el mismo error de HU-03, RN-03.2); no se define un error aparte.
7. **RN-02.7 Copia por valor.** La modificación devuelve una copia del proyecto actualizado. Modificar el resultado devuelto no altera el proyecto guardado (mismo criterio que CA-01.6 y CA-03.3).

## Decisión sobre entidades referenciadas

HU-02 opera sobre una entidad referenciada por ID: el proyecto a modificar. **Sí se valida su existencia** (RN-02.6): es la condición de error central de la historia (CA-02.4). Si el proyecto no existe, no hay nada que actualizar y se devuelve `ErrProjectNotFound`. No referencia ninguna otra entidad: los tres campos que cambia son propios del proyecto.

## Restricciones

- **Organización del código.** Sigue la organización de HU-01 en `internal/project`: el código de producción va en un archivo propio, `update.go`, con el método `Update` del `Manager`, el error `ErrInvalidStatus` y la normalización y verificación de validez del estado. No se tocan `project.go`, `create.go` ni `get.go`.
- **Interfaz de repositorio.** "Las historias siguientes agregan las operaciones que necesiten" (HU-01; HU-03 ya lo hizo con `GetByID`). HU-02 agrega a la interfaz `Repository` la operación `Update(p Project) (Project, error)`: reemplaza el proyecto guardado cuyo ID coincida, devuelve una copia y `ErrProjectNotFound` si no existe. La firma es coherente con `Save`, que también recibe y devuelve el proyecto completo por valor. `MemoryRepository` la implementa; TEC-05 (PostgreSQL) también deberá cumplirla. La lógica de negocio sigue sin acceder al almacenamiento directamente.
- **Reutilización de HU-03.** El `Manager` usa `Repository.GetByID` (agregado por HU-03) para verificar la existencia; no se duplica la búsqueda por ID.
- **Consecuencia de ampliar la interfaz.** El método `Update` de los repositorios falsos de test que implementan `project.Repository` — `failingRepository` (`create_test.go`, de HU-01) y `fixedStatusRepository` (`get_test.go`, de HU-03) — se agrega **en el commit RED, junto con los tests**, para que el paquete compile y los tests fallen por la funcionalidad nueva y no por errores de compilación ajenos. Esos archivos no reciben ningún otro cambio: **no se refactorizan los tests de HU-03** y `fixedStatusRepository` sigue existiendo tal como está.
- **Orden de validaciones.** Primero se valida el contenido recibido, sin consultar el repositorio: nombre vacío (RN-02.1) y estado (normalización y validez de RN-02.4), en ese orden. Después, con el mutex tomado: existencia (RN-02.6, con `GetByID`), nombre duplicado (RN-02.3, con `List`) y finalmente la escritura (`Update` del repositorio). Si falla cualquiera, no se escribe nada.
- **Copias por valor.** `Manager.Update` y `Repository.Update` devuelven `Project` por valor, nunca punteros al estado interno.
- **Errores identificables.** `ErrInvalidStatus` es nuevo, exportado y comparable con `errors.Is` (patrón de HU-01 y HU-03). Se reutilizan `ErrEmptyName`, `ErrDuplicateName` y `ErrProjectNotFound` sin duplicarlos.
- **Concurrencia.** La modificación hace varios pasos sobre el repositorio (buscar, listar, escribir), así que `Update` toma el mutex del `Manager`, igual que `Create` y `Get`. El criterio de HU-03 ("no se escribe código sin un test que lo pida") se cumple con CA-02.7, cuyo test corre con `go test -race` en el CI.
- **Pasos godog.** Los pasos de los escenarios BDD son código de test en `update_features_test.go`, con su runner `TestFeatures_HU02ModificarProyecto`. No se tocan los archivos de features tests de las otras historias.
- **Librerías.** Solo la librería estándar en el código; godog solo en los tests.
- **Exposición HTTP.** Exponer la modificación por HTTP (por ejemplo, `PUT /proyectos/{id}`) es alcance de TEC-01, no de esta historia.

## Casos límite

| Caso | Resultado esperado |
|---|---|
| Cambiar solo el nombre (se reenvían descripción y estado actuales) | Cambia el nombre; lo demás queda igual |
| Cambiar solo el estado | Cambia el estado; nombre y descripción quedan igual |
| Pasar por los tres estados (`PLANIFICADO` → `EN_CURSO` → `FINALIZADO` → `PLANIFICADO`) | Todos los cambios son válidos; no hay transiciones prohibidas |
| Estado en minúsculas o con espacios en los extremos ("en_curso", " En_Curso ") | Se guarda normalizado: `EN_CURSO` |
| Nombre nuevo con espacios en los extremos ("  Proyecto B  ") | Se guarda "Proyecto B" |
| Nombre nuevo vacío o con solo espacios | Error: `ErrEmptyName`; el proyecto queda como estaba |
| Nombre nuevo igual al de otro proyecto, con otras mayúsculas o espacios (" proyecto b " existiendo "Proyecto B") | Error: `ErrDuplicateName`; los dos proyectos quedan como estaban |
| Nombre nuevo igual al del propio proyecto (o con otras mayúsculas) | Válido: se excluye el propio proyecto del control (RN-02.3) |
| Descripción nueva vacía | Se guarda vacía |
| Estado que no existe ("PAUSADO") o estado vacío | Error: `ErrInvalidStatus`; el proyecto queda como estaba |
| Modificar un ID inexistente (por ejemplo 999), 0 o negativo | Error: `ErrProjectNotFound` |
| Modificar cuando no hay proyectos creados | Error: `ErrProjectNotFound` |
| Modificar el proyecto devuelto por `Update` | El proyecto guardado no cambia |
| Modificaciones y creaciones simultáneas | Sin accesos concurrentes sin proteger; `go test -race` (en el CI) no reporta condiciones de carrera |

## Condiciones de error

| Condición | Error | Efecto |
|---|---|---|
| Nombre vacío o con solo espacios (RN-02.1) | `ErrEmptyName` ("el nombre del proyecto es obligatorio") | No se modifica nada |
| Estado que, normalizado, no es uno de los tres válidos: no existe ("PAUSADO") o vacío (RN-02.4) | `ErrInvalidStatus` ("el estado del proyecto no es válido") | No se modifica nada |
| No existe un proyecto con ese ID (RN-02.6; incluye 0 y negativos) | `ErrProjectNotFound` ("no existe un proyecto con ese id") | No se modifica nada |
| Nombre repetido con otro proyecto (RN-02.3) | `ErrDuplicateName` ("ya existe un proyecto con ese nombre") | No se modifica nada |
| El repositorio falla al buscar, listar o actualizar | El error que devuelve el repositorio | No se modifica nada y el proyecto queda como estaba |

Los errores de contenido (`ErrEmptyName`, `ErrInvalidStatus`) se devuelven antes de consultar el repositorio, igual que hacía `Create` con el nombre. Por eso, si el nombre está vacío y además el ID no existe, se informa `ErrEmptyName`.

## Criterios de aceptación

- **CA-02.1** Se puede cambiar el nombre, la descripción y el estado de un proyecto existente en una sola operación; cada campo puede quedar igual reenviando su valor actual.
- **CA-02.2** El estado nuevo se acepta en cualquier combinación de mayúsculas y minúsculas y con espacios en los extremos (por ejemplo, "en_curso" o " En_Curso ") y se guarda normalizado a mayúsculas sin espacios de extremos (`EN_CURSO`); si el valor normalizado no es uno de los tres estados válidos (por ejemplo, "PAUSADO" o vacío), devuelve `ErrInvalidStatus` y no cambia nada. No hay transiciones prohibidas entre estados válidos.
- **CA-02.3** No se acepta un nombre vacío o con solo espacios (`ErrEmptyName`) ni un estado inválido (`ErrInvalidStatus`); en esos casos el proyecto guardado queda exactamente como estaba.
- **CA-02.4** Modificar un proyecto que no existe devuelve `ErrProjectNotFound` (incluye ID inexistente, 0, negativo y no tener proyectos creados).
- **CA-02.5** El nombre nuevo no puede repetir el de otro proyecto distinto (RN-02.3): devuelve `ErrDuplicateName` y ningún proyecto cambia. Dejar el nombre del propio proyecto es válido.
- **CA-02.6** `Update` devuelve una copia por valor: modificar el resultado devuelto no altera el proyecto guardado.
- **CA-02.7** Modificar mientras se crean o consultan proyectos en simultáneo no produce accesos concurrentes sin protección: se verifica con `go test -race` (lo ejecuta el CI de TEC-02).
- **CA-02.8** Si el repositorio falla al buscar, listar o actualizar, se devuelve ese error y el proyecto no cambia.

## Comparación con el acta de Planning

El acta del Sprint 1 (`docs/actas/sprint-1-planning.md`) describe HU-02 como "Modificar datos del proyecto (nombre, descripción y estado)" (Issue #15, 3 Story Points, responsable Cardozo Leandro, orden 3 del sprint) y la estima como "update con validaciones y cambio de estado; mismo patrón que HU-01". **El alcance coincide**: una única operación de modificación con las validaciones del nombre y del estado, sobre el `Manager` del paquete `internal/project`. No hay diferencia entre el acta y esta SDD.

## Comparación con los criterios del Issue #15 (borrador)

| CA en el Issue #15 (borrador) | En esta SDD | Ajuste |
|---|---|---|
| CA-02.1 Se pueden cambiar el nombre, la descripción y el estado de un proyecto existente. | CA-02.1 (precisado) | Se aclara que es una sola operación y que cada campo puede quedar igual |
| CA-02.2 El estado debe ser uno de los estados válidos. | CA-02.2 (precisado) | Se listan los tres estados válidos, se acepta sin distinguir mayúsculas ni espacios de extremos (igual que la prioridad en HU-04), se identifica el error (`ErrInvalidStatus`) y se aclara que no hay transiciones prohibidas |
| CA-02.3 No se acepta un nombre vacío ni un estado inválido. | CA-02.3 (precisado) | Se identifican los errores y se aclara que el proyecto queda como estaba |
| CA-02.4 Modificar un proyecto que no existe devuelve error. | CA-02.4 (precisado) | Se identifica el error (`ErrProjectNotFound`) y se incluyen ID 0, negativo y lista vacía |
| — | CA-02.5 (nuevo) | Control de nombre duplicado con otro proyecto, aplicando RN-01.3 al modificar (decidido con el Agile Enabler) |
| — | CA-02.6 (nuevo) | Copia por valor, mismo patrón que CA-01.6 y CA-03.3 |
| — | CA-02.7 (nuevo) | Concurrencia verificada con `go test -race`, mismo patrón que CA-03.4 |
| — | CA-02.8 (nuevo) | Error del repositorio al buscar, listar o actualizar: se devuelve y el proyecto no cambia |

**Texto definitivo para reemplazar la sección "Criterios de aceptación" del Issue #15** (a actualizar cuando la SDD se apruebe):

```markdown
## Criterios de aceptación

- [ ] CA-02.1 Se puede cambiar el nombre, la descripción y el estado de un proyecto existente en una sola operación; cada campo puede quedar igual reenviando su valor actual.
- [ ] CA-02.2 El estado nuevo se acepta en cualquier combinación de mayúsculas y minúsculas y con espacios en los extremos (por ejemplo, "en_curso" o " En_Curso ") y se guarda normalizado a mayúsculas sin espacios de extremos; si el valor normalizado no es `PLANIFICADO`, `EN_CURSO` o `FINALIZADO` (por ejemplo, "PAUSADO" o vacío), devuelve error (`ErrInvalidStatus`) y no cambia nada. No hay transiciones prohibidas entre estados válidos.
- [ ] CA-02.3 No se acepta un nombre vacío o con solo espacios (`ErrEmptyName`) ni un estado inválido (`ErrInvalidStatus`); el proyecto guardado queda exactamente como estaba.
- [ ] CA-02.4 Modificar un proyecto que no existe devuelve error (`ErrProjectNotFound`), incluidos ID inexistente, 0, negativo y no tener proyectos creados.
- [ ] CA-02.5 El nombre nuevo no puede repetir el de otro proyecto distinto (`ErrDuplicateName`); dejar el nombre del propio proyecto es válido.
- [ ] CA-02.6 `Update` devuelve una copia por valor: modificar el resultado no altera el proyecto guardado.
- [ ] CA-02.7 Modificar mientras se crean o consultan proyectos en simultáneo no produce accesos concurrentes sin protección (verificado con `go test -race`).
- [ ] CA-02.8 Si el repositorio falla al buscar, listar o actualizar, se devuelve ese error y el proyecto no cambia.
```

## Decisiones tomadas (dudas resueltas con el Agile Enabler)

1. Una sola operación: `Update(id int, name, description, status string) (Project, error)` reemplaza los tres campos; para cambiar uno se reenvían los valores actuales de los demás (mismo patrón que `Create`). El estado llega como texto ("en_curso", " En_Curso ") y se normaliza adentro, igual que `priority string` en el `Create` de HU-04; no se usa el tipo `Status` como parámetro.
2. Nombre duplicado: sí se valida al modificar, excluyendo el propio proyecto (RN-02.3, aplicación de RN-01.3).
3. Estado sin distinguir mayúsculas ni espacios de los extremos: "en_curso" o " En_Curso " se guardan `EN_CURSO`, normalizado igual que la prioridad en HU-04; solo un estado que no existe ("PAUSADO") o vacío devuelve `ErrInvalidStatus`.
4. Sin transiciones prohibidas: los criterios del Issue solo exigen que el estado sea uno de los válidos; cualquier estado válido puede pasar a cualquier otro.
5. Error nuevo: `ErrInvalidStatus`, exportado y declarado en `update.go` (patrón de HU-03 con `ErrProjectNotFound`, para no tocar `project.go`).
6. Orden de validaciones: contenido (nombre, estado) antes de existencia y duplicados; los errores de contenido no consultan el repositorio.
7. Mutex sí, justificado por CA-02.7 y su test con `go test -race` (no por uniformidad).
