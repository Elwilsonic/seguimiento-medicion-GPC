  # HU-06 · Listar el backlog de un proyecto

  > Issue del repositorio: #21

  ## Historia de usuario

  Como Agile Enabler, quiero listar los ítems del backlog de un proyecto, para ver todo el trabajo pendiente en un solo lugar.

  ## Objetivo

  Obtener todos los ítems del Product Backlog de un proyecto existente, a partir de su ID, para que el Agile Enabler (y más adelante TEC-01 y el dashboard) vea el trabajo pendiente de cada proyecto en un solo lugar.

  Es una operación de solo lectura: no crea, modifica ni borra ítems.

  ## Alcance

  - **Incluye:** listar los ítems de un proyecto a través del `Manager` del paquete `internal/backlog` (creado en HU-04), validando que el proyecto exista (reutilizando HU-03) y devolviendo una lista vacía cuando el proyecto no tiene ítems.
  - **No incluye:** crear ítems (HU-04), consultar un ítem (HU-05), modificarlo (HU-07), cambiar su estado (HU-08), ordenarlo por prioridad (HU-09), filtros ni paginación, ni la exposición HTTP (TEC-01).

  ## Contrato

  ### Entradas

  | Campo | Tipo | Obligatorio | Descripción |
  |---|---|---|---|
  | ProjectID | entero | Sí | ID del proyecto cuyo backlog se lista |

  Firma: `ListItems(projectID int) ([]Item, error)`.

  ### Salidas esperadas

  - **Si el proyecto existe:** una lista nueva con copias de todos los ítems cuyo `ProjectID` coincide, ordenada por ID ascendente. Cada ítem tiene todos los campos que define HU-04 (ID, ProjectID, título, descripción, prioridad, criterios de aceptación, estado y story points).
  - **Si el proyecto existe y no tiene ítems:** una lista vacía (longitud 0, no `nil`), sin error.
  - **Si hay un error:** el error (ver "Errores") y ninguna lista (`nil`).

  ### Errores

  | Condición | Error | Efecto |
  |---|---|---|
  | No existe un proyecto con ese ProjectID (incluye 0, negativos y listar sin proyectos creados) | `ErrProjectNotFound` de `internal/project`, propagado tal cual | No se devuelve ninguna lista |
  | El repositorio falla al listar | El error que devuelve el repositorio | No se devuelve ninguna lista |

  No se define ningún error nuevo: un backlog vacío no es un error (RN-06.2), y el proyecto inexistente reutiliza el error de HU-03.

  **Orden de validación:** primero se valida el proyecto y después se lista. Si el proyecto no existe, no se consulta el repositorio de ítems.

  ## Reglas de negocio

  1. **RN-06.1 Listado completo.** Se devuelven todos los ítems del proyecto, sin filtrar por el estado del ítem ni por el estado del proyecto: un proyecto `FINALIZADO` lista su backlog igual.
  2. **RN-06.2 Backlog vacío.** Si el proyecto existe y no tiene ítems, se devuelve una lista vacía de longitud 0 (no `nil`), sin error. Un backlog vacío es un estado válido.
  3. **RN-06.3 Aislamiento por proyecto.** Solo aparecen los ítems del proyecto indicado, aunque otros proyectos tengan ítems con el mismo título.
  4. **RN-06.4 Proyecto inexistente.** Si el proyecto no existe, se devuelve `ErrProjectNotFound` y no una lista vacía. Incluye IDs nunca asignados, 0 y negativos, con el mismo criterio de RN-03.2.
  5. **RN-06.5 Copias por valor.** La lista devuelta es nueva, y cada ítem tiene su propia copia de la lista de criterios de aceptación. Modificar la lista, un ítem o sus criterios no altera los ítems guardados.
  6. **RN-06.6 Orden por ID.** La lista se ordena por ID ascendente, que es el orden de creación. Ordena el repositorio: el repositorio en memoria guarda los ítems en orden de ID, y TEC-05 usará `ORDER BY id`.

  ## Decisión sobre entidades referenciadas

  HU-06 referencia una entidad por ID: el proyecto cuyo backlog se lista. **Sí se valida su existencia.** El acta del Sprint 1 le deja esta decisión a esta SDD ("decide si valida el proyecto"), y los motivos son:

  1. Si un proyecto inexistente devolviera una lista vacía, no se podría distinguir "el proyecto no tiene ítems" (válido) de "el proyecto no existe" (un error del llamador, por ejemplo un ID mal escrito).
  2. Ya existe la búsqueda por ID (HU-03), y HU-04 ya la usa: el `Manager` de backlog llama a `project.Manager.Get`. No se duplica ni la regla ni el error.

  ## Restricciones

  - **Organización del código.** Se agrega el archivo `internal/backlog/list.go`, con el método `ListItems` del `Manager`. No se toca el paquete `internal/project` ni el código de producción de HU-04, salvo lo indicado en "Interfaz de repositorio".
  - **Interfaz de repositorio.** Se agrega a la interfaz `Repository` la operación `ListByProject(projectID int) ([]Item, error)`, que devuelve una lista nueva, ordenada por ID, con copias de los ítems (incluida la lista de criterios), y una lista vacía no `nil` cuando no hay ítems. La implementa el repositorio en memoria (`memory_repository.go`), y TEC-05 también deberá implementarla.
  - **Consecuencia de ampliar la interfaz.** Agregar `ListByProject` obliga a actualizar el repositorio falso de los tests de HU-04 (`failingRepository` en `create_test.go`) para que siga cumpliendo la interfaz, como pasó en HU-03. Además de `repository.go` y `memory_repository.go`, ese es el único archivo de HU-04 que se toca.
  - **Concurrencia.** El repositorio en memoria no es seguro para uso concurrente. `ListItems` valida el proyecto antes de tomar el mutex del `Manager` (igual que `Create` en HU-04, así no se toman dos mutex a la vez) y lista con el mutex tomado. La protección se justifica con CA-06.5 y su test bajo `go test -race`.
  - **Tests.** Igual que en HU-04, los tests son internos al paquete (`package backlog`): tests unitarios en `list_test.go`, y pasos de godog en `list_features_test.go`, con el runner `TestFeatures_HU06ListarBacklog`. El proyecto `FINALIZADO` se crea guardándolo directo con `projectRepo.Save`, como en `TestCreate_ProyectoFinalizadoAceptaItems` de HU-04.
  - **Librerías.** Solo la librería estándar en el código, y godog solo en los tests.
  - **Exposición HTTP.** Es alcance de TEC-01.

  ## Casos límite

  | Caso | Resultado esperado |
  |---|---|
  | Proyecto sin ítems | Lista vacía (longitud 0, no `nil`), sin error |
  | Proyecto con varios ítems | Aparecen todos, ordenados por ID ascendente |
  | Otro proyecto con ítems (incluso con el mismo título) | Esos ítems no aparecen |
  | Proyecto inexistente (999), 0, negativo (-1) o sin proyectos creados | `ErrProjectNotFound` |
  | Proyecto `FINALIZADO` con ítems | Se listan igual |
  | Modificar la lista devuelta, un ítem o sus criterios | Los ítems guardados no cambian; volver a listar devuelve los datos originales |
  | Listados simultáneos mientras se crean ítems | `go test -race` no reporta condiciones de carrera |
  | El repositorio falla al listar | Se devuelve ese error y la lista es `nil` |

  ## Criterios de aceptación

  - **CA-06.1** Se listan todos los ítems del proyecto indicado, con todos sus datos, ordenados por ID ascendente.
  - **CA-06.2** Un proyecto sin ítems devuelve una lista vacía (longitud 0, no `nil`), no un error.
  - **CA-06.3** No aparecen ítems de otros proyectos.
  - **CA-06.4** Listar el backlog de un proyecto que no existe devuelve `ErrProjectNotFound` (incluidos ID inexistente, 0, negativo y listar sin proyectos creados).
  - **CA-06.5** Listar mientras se crean ítems en simultáneo no produce accesos concurrentes sin protección: el test corre con `go test -race` (el CI ya lo ejecuta) sin reportar condiciones de carrera.
  - **CA-06.6** Modificar la lista devuelta, un ítem o sus criterios de aceptación no altera los ítems guardados.
  - **CA-06.7** Si el repositorio falla al listar, se devuelve ese error y ninguna lista.

  ## Comparación con el acta de Planning

  El acta del Sprint 1 describe HU-06 como "Listar el backlog de un proyecto" (Issue #21, 3 Story Points, responsable Buttini Cristobal, orden 6, después de HU-04 y HU-01), con la nota "Lista simple; decide si valida el proyecto y el caso de backlog vacío". **El alcance coincide:** es una lista simple, sin filtros, y esta SDD toma las dos decisiones que el acta le deja (valida el proyecto: RN-06.4 y CA-06.4; backlog vacío como lista sin error: RN-06.2 y CA-06.2). No hay diferencias entre el acta y esta SDD.

  ## Comparación con los criterios del Issue #21 (borrador)

  | CA en el Issue #21 (borrador) | En esta SDD | Ajuste |
  |---|---|---|
  | CA-06.1 Se listan todos los ítems del proyecto indicado | CA-06.1 (precisado) | Con todos sus datos y ordenados por ID |
  | CA-06.2 Un proyecto sin ítems devuelve una lista vacía, no un error | CA-06.2 (precisado) | Longitud 0, no `nil` |
  | CA-06.3 No aparecen ítems de otros proyectos | CA-06.3 (igual) | Ninguno |
  | — | CA-06.4 (nuevo) | Validación del proyecto, que el acta le deja a esta SDD |
  | — | CA-06.5 (nuevo) | Concurrencia verificada con `go test -race` |
  | — | CA-06.6 (nuevo) | Copias por valor, incluidos los criterios |
  | — | CA-06.7 (nuevo) | Error del repositorio propagado |

  **Texto definitivo para reemplazar la sección "Criterios de aceptación" del Issue #21** (cuando se apruebe la SDD):

  - [ ] CA-06.1 Se listan todos los ítems del proyecto indicado, con todos sus datos, ordenados por ID ascendente.
  - [ ] CA-06.2 Un proyecto sin ítems devuelve una lista vacía (longitud 0, no `nil`), no un error.
  - [ ] CA-06.3 No aparecen ítems de otros proyectos.
  - [ ] CA-06.4 Listar el backlog de un proyecto inexistente devuelve `ErrProjectNotFound`, incluidos ID 0 y negativo.
  - [ ] CA-06.5 Listar mientras se crean ítems en simultáneo no produce condiciones de carrera (verificado con `go test -race`).
  - [ ] CA-06.6 Modificar la lista devuelta, un ítem o sus criterios no altera los ítems guardados.
  - [ ] CA-06.7 Si el repositorio falla al listar, se devuelve ese error.

  ## Decisiones tomadas (dudas resueltas con el Agile Enabler)

  1. Se valida que el proyecto exista, con `ErrProjectNotFound` de HU-03 (RN-06.4).
  2. Backlog vacío: lista de longitud 0, no `nil`, sin error (RN-06.2).
  3. Firmas: `ListItems(projectID int) ([]Item, error)` en el `Manager` y `ListByProject(projectID int) ([]Item, error)` en la interfaz `Repository`.
  4. No se define ningún error nuevo: se propaga `ErrProjectNotFound` de `internal/project`.
  5. Orden por ID ascendente, y ordena el repositorio (RN-06.6).
  6. Se devuelven los ítems completos, con todos los campos que define HU-04.
  7. Mutex sí, justificado por CA-06.5 y su test con `-race`.
