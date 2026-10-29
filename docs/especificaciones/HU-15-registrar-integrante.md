# HU-15 · Registrar integrante del proyecto

> Issue del repositorio: #29

Aprobada por: Peñalbé Hernán (Agile Enabler), 29/10/2026.

## Historia de usuario

Como Agile Enabler, quiero registrar un integrante con nombre y rol en un proyecto, para saber quién participa.

**Épica:** E4 · Registro de integrantes

**Prioridad:** MEDIA (la del Issue #29)

**Story Points:** 5 (acta de Planning del Sprint 2: "Paquete nuevo, como HU-10")

## Objetivo

Registrar los integrantes de un proyecto existente, cada uno con su nombre y uno de los tres roles de la metodología: Product Architect, Agile Enabler o Product Builder. Es la primera historia de E4: crea el paquete `internal/member`, con su `Manager`, su interfaz `Repository` y sus dos implementaciones (memoria y PostgreSQL), la tabla `members`, el endpoint `POST /api/proyectos/{id}/integrantes` y el formulario de la interfaz web. HU-16 (modificar) y HU-17 (consultar) se apoyan en este paquete.

## Alcance

- **Incluye:**
  - el tipo `Member` y los roles, en `internal/member`;
  - la operación `Register(projectID int, name, role string) (Member, error)` del `Manager`;
  - la interfaz `Repository` con `Save` y `ListByProject`, y sus implementaciones en memoria y en PostgreSQL, con tests de contrato;
  - la tabla `members` en `db/init/01-schema.sql` y su vaciado en `internal/postgrestest`;
  - el armado en `cmd/server` (memoria o PostgreSQL) y en `httpapi.NewHandler`;
  - el endpoint `POST /api/proyectos/{id}/integrantes`;
  - una sección "Integrantes" en la interfaz web, con el formulario "Registrar integrante".
- **No incluye:**
  - modificar un integrante o su rol (HU-16 y HU-48);
  - listar los integrantes de un proyecto (HU-17). `ListByProject` se agrega al repositorio porque lo necesita la regla de duplicados (RN-15.4), pero no se expone en el `Manager` ni en la API;
  - dar de baja un integrante;
  - relacionar los integrantes con los esfuerzos: el integrante de un esfuerzo (HU-26) sigue siendo un texto libre, y el reporte de HU-38 sigue listando los que registraron esfuerzo (tema 5 del Planning del Sprint 4);
  - validar que un esfuerzo o un voto de Planning Poker use un integrante registrado.

## Contrato

### Entradas

| Campo | Tipo | Obligatorio | Descripción |
|---|---|---|---|
| Proyecto | entero | Sí | ID del proyecto (en la ruta) |
| Nombre | texto | Sí | Nombre del integrante |
| Rol | texto | Sí | `PRODUCT_ARCHITECT`, `AGILE_ENABLER` o `PRODUCT_BUILDER` |

Firma: `Register(projectID int, name, role string) (Member, error)`, en `internal/member/register.go`. El rol llega como texto y se normaliza adentro, como el estado de un proyecto (HU-02) o la severidad de un defecto (HU-29).

### Salidas esperadas

- **Si el registro es válido:** el integrante guardado, como copia por valor:

  ```go
  type Member struct {
      ID        int
      ProjectID int
      Name      string
      Role      Role
  }
  ```

  con un ID correlativo y global, el nombre sin espacios en los extremos y el rol normalizado. `Member` no tiene slices ni mapas, así que la copia del struct no comparte estado.
- **Si no es válido:** el error que indica el motivo (ver "Condiciones de error"); no se guarda nada.

### HTTP

`POST /api/proyectos/{id}/integrantes`

- **Cuerpo:**

  ```json
  {"nombre": "Ana Pérez", "rol": "PRODUCT_BUILDER"}
  ```

- **201 Created:**

  ```json
  {"id": 1, "proyecto_id": 1, "nombre": "Ana Pérez", "rol": "PRODUCT_BUILDER"}
  ```

- **Errores:**

  | Caso | Código | Mensaje |
  |---|---|---|
  | Nombre vacío o solo espacios | 400 | El de `ErrEmptyName` ("el nombre del integrante es obligatorio") |
  | Rol vacío | 400 | El de `ErrMissingRole` ("el rol del integrante es obligatorio") |
  | Rol que no es uno de los tres | 400 | El de `ErrInvalidRole` ("el rol debe ser PRODUCT_ARCHITECT, AGILE_ENABLER o PRODUCT_BUILDER") |
  | Cuerpo que no es JSON, campo que no es texto o texto después del JSON | 400 | El de `errInvalidJSON` |
  | `{id}` que no es un número entero | 400 | El de `errInvalidID` |
  | Proyecto inexistente (incluidos 0 y negativos) | 404 | El de `project.ErrProjectNotFound` |
  | Ya hay un integrante con ese nombre en el proyecto | 409 | El de `ErrDuplicateMember` ("ya hay un integrante con ese nombre en el proyecto") |
  | Cuerpo de más de 1 MiB | 413 | El de `errBodyTooLarge` |
  | Falla del almacenamiento | 500 | El de `errInternal`, sin detalle |

  El orden de los errores es: ID de la ruta, JSON y `Manager`.

### Interfaz web

Una sección nueva, **"Integrantes"** (`id="integrantes"`), después de "Proyectos":

- Formulario **"Registrar integrante"** (`form-registrar-integrante`) para el proyecto actual, con el campo "Nombre" (`name="nombre"`, obligatorio), la lista "Rol" (`name="rol"`) con las opciones "Product Architect" (`PRODUCT_ARCHITECT`), "Agile Enabler" (`AGILE_ENABLER`) y "Product Builder" (`PRODUCT_BUILDER`), y el botón "Registrar integrante".
- Al registrar, muestra en `msg-integrantes` "Integrante N registrado: Nombre (Rol)." con el nombre del rol legible, y limpia el campo del nombre.
- Sin proyecto actual, el mensaje de siempre de `proyectoActual()`; con un error de la API, su mensaje. Los dos en `msg-integrantes`.
- Los textos se escriben con `textContent`, nunca con `innerHTML`.

## Reglas de negocio

1. **RN-15.1 Nombre obligatorio.** El nombre se guarda sin espacios en blanco en los extremos (`strings.TrimSpace`) y no puede quedar vacío (`ErrEmptyName`), como el nombre de un proyecto (RN-01.1).
2. **RN-15.2 Rol obligatorio y válido.** El rol es uno de `PRODUCT_ARCHITECT`, `AGILE_ENABLER` o `PRODUCT_BUILDER`, los tres roles de la metodología (AGENTS.md, "Terminología del proyecto"). Se acepta sin distinguir mayúsculas y sin contar los espacios de los extremos, y se guarda normalizado. Vacío es `ErrMissingRole`; otro valor, `ErrInvalidRole` (decisión D2).
3. **RN-15.3 Proyecto existente.** El proyecto tiene que existir; se valida con `project.Manager.Get` (HU-03) y su error se propaga tal cual: `ErrProjectNotFound`, incluidos los IDs 0 y negativos.
4. **RN-15.4 Sin duplicados en el proyecto.** No puede haber dos integrantes con el mismo nombre en el mismo proyecto. Los nombres se comparan sin distinguir mayúsculas (`strings.EqualFold`), como los de los proyectos (RN-01.3), y sin importar el rol: el mismo nombre con otro rol también es duplicado (`ErrDuplicateMember`, decisión D1). El mismo nombre en otro proyecto sí se acepta.
5. **RN-15.5 Cualquier rol, sin límites.** Un proyecto puede tener cualquier cantidad de integrantes de cada rol, incluido más de un Agile Enabler: el Issue no pone límites y no se agregan reglas que no se pidieron (decisión D3).
6. **RN-15.6 Cualquier estado del proyecto.** Se puede registrar un integrante en un proyecto `PLANIFICADO`, `EN_CURSO` o `FINALIZADO`.
7. **RN-15.7 ID correlativo y global.** El ID lo asigna el repositorio: en memoria, correlativo desde 1; en PostgreSQL, `GENERATED ALWAYS AS IDENTITY`. No se reinicia por proyecto, como los demás IDs.
8. **RN-15.8 Orden de validación.** Nombre, rol, existencia del proyecto, duplicado y guardado. Se informa el primer error encontrado; si algo falla, no se guarda nada.
9. **RN-15.9 Concurrencia.** La lectura de los integrantes del proyecto, la comparación de nombres y el guardado se hacen bajo el `sync.Mutex` del `Manager` de integrantes, así dos registros simultáneos del mismo nombre no dejan un duplicado. El proyecto se valida antes de tomar el mutex, como en `sprint.Manager.Create` (RN-10.4): los proyectos no se borran, y así nunca se toman dos mutex a la vez.
10. **RN-15.10 Copias por valor.** `Register` y el repositorio devuelven copias; modificar lo devuelto no cambia lo guardado.
11. **RN-15.11 Persistencia.** Los integrantes se guardan en memoria o en PostgreSQL detrás de la misma interfaz `Repository`, con el mismo comportamiento, verificado con los tests de contrato.

## Decisión sobre entidades referenciadas

- **Proyecto:** **sí se valida que exista** (RN-15.3). Sin esa validación, un integrante podría quedar en un proyecto inexistente, y HU-17 listaría integrantes de un proyecto que no se puede consultar. Es la misma validación que hacen el backlog (HU-04) y los sprints (HU-10). En PostgreSQL, además, `project_id` es una clave foránea a `projects`.
- **Esfuerzos y votos:** no se relacionan con los integrantes (ver "No incluye").

## Decisiones de diseño (a confirmar por el Agile Enabler)

1. **D1 "Mismo integrante" es el mismo nombre en el proyecto, sin distinguir mayúsculas ni el rol.** El borrador dice "no se puede registrar dos veces al mismo integrante", y la aplicación no tiene otro dato para identificar a una persona. *Alternativas:* comparar nombre y rol (permitiría a "Ana" como Product Builder y como Agile Enabler, que es la misma persona dos veces), o no controlar duplicados (contradice CA-15.3).
2. **D2 Roles como códigos en mayúsculas.** `PRODUCT_ARCHITECT`, `AGILE_ENABLER` y `PRODUCT_BUILDER`, como los demás valores fijos de la API (`EN_CURSO`, `RESUELTO`, `ALTA`). La interfaz muestra los nombres legibles. *Alternativa:* aceptar los nombres con espacios ("Product Builder"), que obliga a normalizar espacios intermedios y no sigue el estilo del resto de la API.
3. **D3 Sin límite por rol.** Scrum suele tener un solo Agile Enabler, pero el Issue no lo pide y un proyecto real puede rotar el rol. Agregar el límite sería una regla de negocio inventada.
4. **D4 Paquete nuevo `internal/member`.** Como `effort` o `defect`: un dominio con su `Manager`, `Repository` y dos implementaciones. Los nombres del código van en inglés (`Member`, `Role`); la API y la interfaz en español ("integrante", "rol").
5. **D5 `ListByProject` en el repositorio desde HU-15.** La regla de duplicados (RN-15.4) necesita los integrantes del proyecto. HU-17 la va a reutilizar para listar; HU-15 no la expone en el `Manager` ni en la API. HU-16 agregará `GetByID` y `Update`.
6. **D6 El control de duplicados está en el `Manager`, no en la base.** Como los nombres de proyecto (HU-01): la tabla `members` no tiene un índice único, porque la comparación sin distinguir mayúsculas ya la hace el `Manager` bajo su mutex. *Alternativa:* un índice único sobre `(project_id, lower(name))`, que repetiría la regla en dos lugares.
7. **D7 Sin cambios en esfuerzos, Planning Poker ni reportes** (tema 5 del Planning del Sprint 4): el reporte de HU-38 sigue listando los integrantes que registraron esfuerzo. Unir los dos conceptos es otra historia.

## Restricciones

- **Organización del código.**
  - `internal/member` (nuevo): `doc.go`, `member.go` (`Member`, `Role`, los tres roles y los errores), `manager.go` (`Manager`, `NewManager(repo Repository, projects *project.Manager)`), `register.go` (`Register` y la normalización del rol), `repository.go`, `memory_repository.go` y `postgres_repository.go`.
  - `db/init/01-schema.sql`: tabla `members (id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY, project_id INTEGER NOT NULL REFERENCES projects (id), name TEXT NOT NULL, role TEXT NOT NULL)`.
  - `internal/postgrestest/postgrestest.go`: `members` en el `TRUNCATE`.
  - `cmd/server/repositories.go` y `main.go`: el repositorio de integrantes y su `Manager`.
  - `internal/httpapi`: `members.go` (nuevo, handler `registerMember`, `memberRequest` y `memberResponse`), `server.go` (la ruta, el parámetro nuevo de `NewHandler` y los errores en `statusFor`).
  - `internal/web`: `index.html` (sección "Integrantes") y `app.js` (el formulario).
- **`NewHandler` suma un parámetro** (`members *member.Manager`). Se actualizan sus 24 llamadas, en `cmd/server` y en los tests de `internal/httpapi` (`newHandler` y los que arman el handler a mano). Es un cambio mecánico que entra en el RED, para que el GREEN no toque tests. *Alternativa descartada:* crear el `Manager` de integrantes dentro de `NewHandler` con un repositorio en memoria, que dejaría a los integrantes sin PostgreSQL.
- **Cambia el esquema:** una base creada antes no tiene la tabla `members`; se recrea con `docker compose down -v` y `docker compose up -d --wait`, y se avisa en el README.
- **RED.** Firmas vacías de `Member`, `Role`, los roles, los errores, `Manager`, `NewManager`, `Register` (que devuelve un error), `Repository` y los dos repositorios (que devuelven un error), para que todo compile y los tests fallen por aserciones. La tabla, la ruta, el handler y la interfaz van en el GREEN. `go vet ./...` antes del commit RED.
- **Librerías.** Solo la librería estándar y `pgx/v5`, que ya está aprobada; godog en los tests.

## Casos límite

| Caso | Resultado esperado |
|---|---|
| Nombre y rol válidos | Se guarda con ID 1, el nombre sin espacios en los extremos y el rol normalizado |
| Rol `product_builder` o `  Agile_Enabler  ` | Se acepta y se guarda como `PRODUCT_BUILDER` o `AGILE_ENABLER` |
| Rol `Product Builder` (con espacio), `SCRUM_MASTER` o `DEV` | `ErrInvalidRole` |
| Rol vacío o solo espacios | `ErrMissingRole` |
| Nombre vacío o solo espacios | `ErrEmptyName` |
| Nombre y rol vacíos | `ErrEmptyName` (el nombre se valida primero) |
| "Ana" ya registrada y se registra "ana" o " Ana " | `ErrDuplicateMember` |
| "Ana" ya registrada como Product Builder y se registra como Agile Enabler | `ErrDuplicateMember` (D1) |
| "Ana" registrada en el proyecto 1 y se registra en el proyecto 2 | Se acepta |
| Dos Agile Enabler en el mismo proyecto | Se aceptan (D3) |
| Proyecto `FINALIZADO` | Se acepta (RN-15.6) |
| Proyecto inexistente, 0 o negativo | `ErrProjectNotFound` |
| Nombre vacío en un proyecto inexistente | `ErrEmptyName` (los datos se validan primero) |
| Dos registros simultáneos del mismo nombre | Uno se guarda y el otro devuelve `ErrDuplicateMember`; `go test -race` sin carreras |
| Modificar el integrante devuelto | Lo guardado no cambia |
| Falla el repositorio al leer o al guardar | Se devuelve ese error; no se guarda nada |
| PostgreSQL: registrar y leer | Mismo resultado que en memoria (contrato) |
| HTTP: cuerpo que no es JSON, `{"nombre": 5}` o texto después del JSON | 400, JSON inválido |
| HTTP: cuerpo de más de 1 MiB | 413 |
| Interfaz sin proyecto actual | Mensaje de `proyectoActual()` en `msg-integrantes` |

## Condiciones de error

| Condición | Error | HTTP |
|---|---|---|
| Nombre vacío o solo espacios (RN-15.1) | `member.ErrEmptyName` | 400 |
| Rol vacío (RN-15.2) | `member.ErrMissingRole` | 400 |
| Rol inválido (RN-15.2) | `member.ErrInvalidRole` | 400 |
| Proyecto inexistente (RN-15.3) | `project.ErrProjectNotFound` | 404 |
| Nombre repetido en el proyecto (RN-15.4) | `member.ErrDuplicateMember` | 409 |
| Falla del repositorio | El error del repositorio | 500, sin detalle |

## Criterios de aceptación

- **CA-15.1** Se registra un integrante en un proyecto existente con su nombre y un rol válido (Product Architect, Agile Enabler o Product Builder); queda con un ID, el nombre sin espacios en los extremos y el rol normalizado.
- **CA-15.2** El nombre y el rol son obligatorios: un nombre vacío devuelve `ErrEmptyName`, un rol vacío `ErrMissingRole` y un rol que no es uno de los tres `ErrInvalidRole` (400 en la API), y no se guarda nada.
- **CA-15.3** No se puede registrar dos veces al mismo integrante en el mismo proyecto: el mismo nombre, sin distinguir mayúsculas y con cualquier rol, devuelve `ErrDuplicateMember` (409); el mismo nombre en otro proyecto se acepta.
- **CA-15.4** Registrar un integrante en un proyecto inexistente (incluidos 0 y negativos) devuelve `ErrProjectNotFound` (404).
- **CA-15.5** `Register` devuelve una copia: modificarla no cambia lo guardado. Dos registros simultáneos del mismo nombre no dejan un duplicado (`go test -race`). Si falla el repositorio, se devuelve el error y no se guarda nada.
- **CA-15.6** Los integrantes se guardan en memoria y en PostgreSQL con el mismo comportamiento (tests de contrato; los de PostgreSQL se omiten sin `TEST_DATABASE_URL`).
- **CA-15.7** Se registra desde la API (`POST /api/proyectos/{id}/integrantes`: 201, y 400, 404, 409, 413 o 500 según la tabla) y desde el formulario "Registrar integrante" de la sección "Integrantes" de la interfaz web.

## Comparación con el acta de Planning

- **Sprint 2:** HU-15 con **5 SP**, "Paquete nuevo, como HU-10", con dependencias HU-01 y HU-03, en "si hay tiempo". Coincide con esta SDD.
- **Sprint 3 y Sprint 4:** E4 (HU-15 a HU-17) queda **fuera de la entrega, salvo capacidad sobrante**. Se toma el 29/10/2026 con capacidad sobrante, asignada por el Agile Enabler.
- **Tema 5 del Planning del Sprint 4:** el reporte de HU-38 lista los integrantes que registraron esfuerzo porque E4 no estaba implementada. Esta SDD no lo cambia (D7).
- **División de HU-16 (acta del Sprint 3):** HU-16 quedó como "Modificar datos" y el cambio de rol pasó a HU-48. No afecta a HU-15.

No hay diferencias de alcance con las actas: la única novedad es que una historia "fuera de la entrega" se toma con capacidad sobrante, como prevén las actas.

## Comparación con los criterios del Issue #29 (borrador)

| Borrador | En esta SDD | Ajuste |
|---|---|---|
| CA-15.1 Se registra un integrante con nombre y un rol válido. | CA-15.1 | Se agrega el proyecto existente y la normalización |
| CA-15.2 Nombre y rol son obligatorios. | CA-15.2 | Se separan rol vacío y rol inválido |
| CA-15.3 No se puede registrar dos veces al mismo integrante en el mismo proyecto. | CA-15.3 | Se precisa qué es "el mismo integrante" (D1) |
| — | CA-15.4 a CA-15.7 (nuevos) | Proyecto inexistente, copias y concurrencia, persistencia, API e interfaz |

Con la aprobación de la SDD, los criterios definitivos reemplazan al borrador en el Issue #29, con autorización.

## Pruebas previstas

- **Unitarios** (`internal/member/register_test.go`): registro válido; normalización del nombre y del rol; nombre vacío; rol vacío; rol inválido (con espacio, `SCRUM_MASTER`); orden de validación; duplicado con otra capitalización y con otro rol; mismo nombre en otro proyecto; dos Agile Enabler; proyecto `FINALIZADO`; proyecto inexistente, 0 y negativo; copia por valor; error del repositorio al leer y al guardar; registros simultáneos con `-race`.
- **BDD** (`features/HU-15-registrar-integrante.feature`, runner `TestFeatures_HU15RegistrarIntegrante` en `internal/member/register_features_test.go`): un escenario por criterio CA-15.1 a CA-15.5, con el CA en el nombre.
- **Contrato** (`internal/member/repository_contract_test.go`): `Save` asigna IDs correlativos y devuelve copias; `ListByProject` devuelve los integrantes del proyecto, ordenados por ID, y `[]` sin integrantes; contra `MemoryRepository` y `PostgresRepository` (este último se omite sin `TEST_DATABASE_URL`).
- **API** (`internal/httpapi/members_test.go`): 201 con el cuerpo exacto; un test por código de error (400 de nombre, rol vacío, rol inválido, JSON e ID; 404; 409; 500); la ruta en `bodyEndpoints` (413 y texto después del JSON) y en `TestRutasConID_IDNoNumericoDevuelve400`.
- **Interfaz** (`internal/httpapi/web_test.go`): la sección, el formulario con las tres opciones de rol, y que `app.js` llama a la ruta nueva.

## Documentación a actualizar

- `README.md`: el endpoint, un `curl`, la sección "Integrantes" de la interfaz, el 409 nuevo, la tabla `members` en el aviso de recrear la base y la fila de HU-15.
- `docs/evidencias/HU-15-tdd.md`, `docs/reportes/cobertura.md` y `docs/ia/registro.md`.
- `docs/documentacion-tecnica.md` (TEC-06): el paquete, el repositorio, la tabla, el endpoint y los errores, después del merge.

HU-15: Historia → SDD → Criterios de Aceptación → BDD → Tests → Código Go
