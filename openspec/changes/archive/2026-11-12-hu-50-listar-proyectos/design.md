# Design: HU-50 · Listar los proyectos

## Context

Motivación y alcance: ver `proposal.md`. Lo que ya existe y condiciona el diseño:

- `project.Repository` ya tiene `List() ([]Project, error)`, usado por `Manager.Create` (HU-01) para validar el nombre único. Sus dos implementaciones devuelven copias en orden de ID:
  - `MemoryRepository.List` devuelve `slices.Clone(r.projects)`. **Sin proyectos devuelve `nil`**, porque `slices.Clone(nil)` es `nil`.
  - `PostgresRepository.List` hace `SELECT … FROM projects ORDER BY id` y devuelve una lista nueva.
  - El test de contrato (`repository_contract_test.go`) fija ese orden por ID, que `Create` no necesita pero otros tests sí.
- `Manager.Get` (HU-03) toma `m.mu` porque `MemoryRepository` no es seguro para uso concurrente (CA-03.4).
- `httpapi` registra `POST /api/proyectos` y `GET /api/proyectos/{id}`; `toProjectResponse` arma el JSON de un proyecto y `writeError` responde 500 con "error interno del servidor" ante cualquier error no clasificado.
- La interfaz (`internal/web`) elige el proyecto actual escribiendo su ID en `#proyecto-actual`; `cargarProyecto()` lo consulta con `GET /api/proyectos/{id}`, lo muestra con `mostrarProyecto(p)` y carga el backlog. HU-14 y HU-17 ya listan con un botón «Ver …», una tabla y un botón por fila (`boton`, `celda`, `accion`, `mostrar`).
- HU-51 (UUID) va a cambiar el tipo del ID en todas las capas, en paralelo con este cambio.

## Goals / Non-Goals

**Goals:**

- Una sola regla de orden para memoria y PostgreSQL, que siga valiendo cuando los IDs sean UUID.
- Reutilizar lo que existe (`Repository.List`, `toProjectResponse`, `cargarProyecto`, los helpers de la interfaz) sin cambiar su comportamiento.
- Archivos nuevos donde se pueda, para no chocar con HU-51 ni con TEC-11.

**Non-Goals:**

- Cambiar la interfaz `Repository`, sus implementaciones o el test de contrato.
- Ordenar con reglas de idioma (acentos, `ñ`): necesitaría una librería no aprobada (ver Riesgos).

## Decisions

### D1. `Manager.List()` en un archivo nuevo, sobre `Repository.List()`

`internal/project/list.go` agrega `func (m *Manager) List() ([]Project, error)`. Toma `m.mu` (como `Get`), llama a `m.repo.List()`, convierte un resultado `nil` en `[]Project{}`, ordena según D2 y devuelve la lista. Si el repositorio falla, devuelve `nil` y ese error, sin lista parcial.

La lista que devuelve el repositorio ya es una copia nueva en las dos implementaciones, así que ordenarla no toca el estado interno, y el llamador puede modificarla sin efecto sobre los proyectos guardados (CA-50.4).

El mutex se justifica con CA-50.7 y su test bajo `-race`, igual que en HU-03, no "por uniformidad".

*Alternativa descartada:* que `httpapi` llame directo al repositorio. Saltearía el `Manager`, que es quien aplica las reglas de negocio (orden, lista vacía y mutex).

*Alternativa descartada:* agregar a `Repository` una operación nueva ordenada por nombre. Obligaría a tocar las dos implementaciones, el test de contrato y los repositorios falsos de los tests, en los mismos archivos que va a tocar HU-51, sin ganar nada: la regla de orden queda en un solo lugar si la aplica el `Manager`.

### D2. Orden alfabético por nombre, sin distinguir mayúsculas

El `Manager` ordena con `slices.SortStableFunc` y la comparación `strings.Compare(strings.ToLower(a.Name), strings.ToLower(b.Name))`.

Razones:

- **No depende del ID.** Con HU-51 los IDs pasan a ser UUID aleatorios; ordenar por ID daría un orden sin sentido que cambiaría de una base a otra. El orden por nombre sigue igual antes y después de HU-51.
- **Es como se busca.** El usuario elige el proyecto por su nombre (es lo que reemplaza a "saber su número"), y una lista alfabética se recorre rápido.
- **Es total.** RN-01.3 impide dos nombres iguales sin distinguir mayúsculas, así que no hay empates. El orden estable es solo una defensa: si un repositorio falso de test devolviera dos nombres iguales, quedan en el orden del repositorio.
- **Es el mismo en memoria y en PostgreSQL**, porque lo aplica Go y no la intercalación (*collation*) de la base.

*Alternativa descartada: orden por ID (orden de creación), como HU-14 y HU-17.* Hoy coincide con el orden de creación, pero deja de hacerlo con HU-51: un UUID no dice cuándo se creó el proyecto. Mantener el orden de creación exigiría agregar al esquema una columna con la fecha de creación, un cambio de la base que el Issue no pide y que se cruzaría con la migración de HU-51.

*Alternativa descartada: `ORDER BY lower(name)` en SQL.* Dejaría la regla repetida en dos implementaciones, con resultados que pueden diferir según la intercalación de la base.

### D3. `GET /api/proyectos` con el formato de `GET /api/proyectos/{id}`

`internal/httpapi/projects.go` agrega `listProjects`, registrado en `server.go` como `mux.HandleFunc("GET /api/proyectos", s.listProjects)`. Arma la respuesta con `make([]projectResponse, 0, len(projects))` y `toProjectResponse` para cada proyecto, así que el formato es idéntico al de la consulta por ID y una lista vacía sale como `[]`, no `null`. Los errores van por `writeError`, que responde 500 y no muestra el error interno.

Códigos HTTP:

| Condición | Código | Cuerpo |
|---|---|---|
| Hay proyectos | 200 | arreglo con un objeto por proyecto, en el orden de D2 |
| No hay proyectos | 200 | `[]` |
| Falla la lectura del repositorio | 500 | `{"error":"error interno del servidor"}` |

*Alternativa descartada:* un objeto `{"proyectos": [...], "cantidad": N}`. Rompe la forma de las otras listas de la API (HU-14, HU-17, defectos), que son arreglos.

### D4. Interfaz: «Ver proyectos» en la sección de contexto

En `index.html`, dentro de la sección «Proyecto y sprint actuales», van el botón `<button type="button" id="ver-proyectos" class="secundario">Ver proyectos</button>` y una tabla con el encabezado `ID · Nombre · Descripción · Estado · Inicio · Fin · Acciones` y el cuerpo `<tbody id="lista-proyectos"></tbody>`. El comentario "No hay listado de proyectos (ninguna historia lo pide)" se reemplaza por uno que cite HU-50.

En `app.js`, el handler de `ver-proyectos`:

1. llama a `api("GET", "/proyectos")` dentro de `accion("msg-contexto", …)`;
2. llena `#lista-proyectos` con `replaceChildren`, `celda` y `boton`, nunca con `innerHTML`;
3. muestra "N proyecto(s).".

El botón «Elegir» de cada fila escribe el ID en `#proyecto-actual`, llama a `cargarProyecto()` (que vuelve a consultar el proyecto y carga el backlog) y muestra "Proyecto N elegido: Nombre.".

Si la API falla, `accion` muestra el mensaje de error y el campo `#proyecto-actual` no se toca.

*Alternativa descartada:* cargar la lista sola al abrir la página. Haría una consulta en cada carga aunque el usuario ya sepa el ID, y se aparta de cómo funcionan HU-14 y HU-17.

*Alternativa descartada:* reemplazar el campo de ID por un `<select>`. Cambia un elemento que usan muchas historias y que HU-51 va a tocar, y no deja ver la descripción, el estado ni las fechas.

### D5. Tests y escenarios

- `features/HU-50-listar-proyectos.feature`: los escenarios de la spec a nivel del `Manager` (normal, modificado, orden, único, sin proyectos, copia y error del repositorio). Los pasos van en `internal/project/list_features_test.go`, con el runner `TestFeatures_HU50ListarProyectos`; el error usa un repositorio falso propio de ese archivo.
- `internal/project/list_test.go`: los tests unitarios del `Manager`, incluidos:
  - la lista vacía no `nil`, con `MemoryRepository`, que es el que devuelve `nil`;
  - la concurrencia bajo `-race` (CA-50.7).
- `internal/httpapi/projects_list_test.go`: 200 con formato exacto y orden, 200 con `[]`, 500 con un repositorio que falla.
- `internal/httpapi/projects_list_web_test.go`: los tests de la interfaz, con los helpers que ya existen (`get`, `handlerBody`). Van en un archivo nuevo para no chocar con TEC-11, que está comentando `web_test.go`.
- Cada test lleva arriba el comentario `// HU-50 · CA-50.N · Verifica que …`.

La interfaz y la ruta HTTP se cubren con tests de Go, no con escenarios godog, como en HU-14 y HU-17.

### Qué implementa cada requisito

| Requisito (spec) | Criterio | Código |
|---|---|---|
| Consultar la lista de proyectos | CA-50.1, CA-50.5 | `Manager.List` (`internal/project/list.go`); `server.listProjects` (`internal/httpapi/projects.go`) y su ruta en `internal/httpapi/server.go` |
| Lista vacía sin proyectos | CA-50.2 | `Manager.List` (`nil` → `[]Project{}`); `listProjects` (`make(…, 0, …)`) |
| Orden de la lista | CA-50.6 | `Manager.List` (`slices.SortStableFunc`) |
| La consulta no modifica los proyectos | CA-50.4 | `Manager.List` (lista nueva sobre la copia del repositorio) |
| Error al leer los proyectos | CA-50.5 | `Manager.List`; `listProjects` → `writeError` (500) |
| Consultas simultáneas | CA-50.7 | `Manager.List` (`m.mu`) |
| Elegir el proyecto actual en la interfaz | CA-50.3, CA-50.5 | `index.html` (`#ver-proyectos`, `#lista-proyectos`); handler de `ver-proyectos` en `internal/web/app.js` |

## Risks / Trade-offs

- **[Riesgo] El orden por bytes deja los nombres que empiezan con acento o con `ñ` después de la "z"** (por ejemplo, "Ábaco" queda después de "Zeta"). → Se acepta y se documenta en el README: corregirlo exige reglas de idioma (`golang.org/x/text/collate`), una librería que no está aprobada. Si el Agile Enabler la aprueba, se cambia solo la comparación de D2.
- **[Riesgo] HU-51 toca las mismas capas** (`projectResponse.ID`, `#proyecto-actual` de tipo `number`, `cargarProyecto`). → HU-50 va primero (Planning del Sprint 5) y agrega casi todo en archivos nuevos. El orden de D2 no depende del ID. HU-51 solo cambia el tipo de `id`, y el botón «Elegir» funciona igual con un UUID si el campo pasa a `text`.
- **[Trade-off] Sin paginación**, la lista trae todos los proyectos en cada consulta. → Para el volumen del proyecto (decenas de proyectos) no importa; la paginación quedó fuera de alcance.
- **[Trade-off] La lista no se recarga sola** al crear o modificar un proyecto. → El usuario vuelve a pulsar «Ver proyectos», igual que en HU-14 y HU-17.

## Migration Plan

No hay migración: no cambia el esquema ni los datos, y las rutas existentes quedan iguales. El despliegue es el habitual (merge a `main` y CI). Para volver atrás, alcanza con revertir el merge.
