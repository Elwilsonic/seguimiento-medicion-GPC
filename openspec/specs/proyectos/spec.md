# proyectos Specification

## Purpose

Gestión de proyectos (épica E1): crear, consultar, modificar y listar los proyectos, con sus fechas, para que el equipo pueda elegir y seguir cada uno.

## Requirements

### Requirement: Consultar la lista de proyectos
El sistema SHALL devolver todos los proyectos guardados, cada uno con ID, nombre, descripción, estado, fecha de inicio y fecha de fin. Por la API, `GET /api/proyectos` responde 200 con un arreglo JSON donde cada proyecto tiene el mismo formato que `GET /api/proyectos/{id}`: las fechas en `AAAA-MM-DD`, o `""` si no están registradas. (HU-50, CA-50.1 y CA-50.5)

La consulta no recibe el ID de ninguna entidad, así que no hay existencia que validar.

#### Scenario: Listar los proyectos con todos sus datos (caso normal)
- **WHEN** existen los proyectos "Sistema de ventas" (descripción "Ventas online", PLANIFICADO, sin fechas) e "Inventario" (EN_CURSO, inicio 2026-03-01, fin 2026-06-30) y se consulta la lista de proyectos
- **THEN** la consulta devuelve los dos proyectos, cada uno con su ID, nombre, descripción, estado y fechas
- **AND** por la API responde 200 con `[{"id":…,"nombre":"Inventario","descripcion":"","estado":"EN_CURSO","fecha_inicio":"2026-03-01","fecha_fin":"2026-06-30"},{"id":…,"nombre":"Sistema de ventas","descripcion":"Ventas online","estado":"PLANIFICADO","fecha_inicio":"","fecha_fin":""}]`

#### Scenario: Un proyecto modificado aparece con sus datos nuevos (caso alternativo)
- **WHEN** el proyecto "Inventario" se modificó con HU-02 a nombre "Inventario central" y estado FINALIZADO, y se consulta la lista
- **THEN** el proyecto aparece con el nombre "Inventario central" y el estado FINALIZADO, y no aparece "Inventario"

#### Scenario: Un único proyecto (caso límite)
- **WHEN** existe un solo proyecto y se consulta la lista
- **THEN** la lista tiene exactamente ese proyecto

### Requirement: Lista vacía sin proyectos
Si no hay proyectos guardados, el sistema SHALL devolver una lista vacía, no `nil` ni un error. Por la API, `GET /api/proyectos` responde 200 con `[]`, no `null`. (HU-50, CA-50.2)

#### Scenario: Consultar la lista sin proyectos (caso límite)
- **WHEN** no existe ningún proyecto y se consulta la lista
- **THEN** la consulta es exitosa y devuelve una lista vacía
- **AND** por la API responde 200 con el cuerpo `[]`

### Requirement: Orden de la lista
El sistema SHALL devolver los proyectos ordenados alfabéticamente por nombre, sin distinguir mayúsculas de minúsculas, sin importar su ID ni el orden en que se crearon. La API y la interfaz web respetan ese orden. (HU-50, CA-50.6)

#### Scenario: Proyectos creados en otro orden (caso alternativo)
- **WHEN** se crean, en este orden, los proyectos "Sistema de ventas", "inventario" y "Agenda", y se consulta la lista
- **THEN** la lista los devuelve en el orden "Agenda", "inventario", "Sistema de ventas"

### Requirement: La consulta no modifica los proyectos
Consultar la lista SHALL NOT modificar ningún proyecto. La lista devuelta es una copia: modificar sus elementos o la lista misma no altera los proyectos guardados. (HU-50, CA-50.4)

#### Scenario: Modificar la lista devuelta (caso límite)
- **WHEN** se consulta la lista y se cambian el nombre y el estado del primer proyecto de la lista devuelta
- **THEN** consultar ese proyecto por su ID (HU-03) y volver a consultar la lista devuelven los datos originales

### Requirement: Error al leer los proyectos
Si el almacenamiento falla al leer los proyectos, el sistema SHALL devolver ese error y ninguna lista, ni siquiera parcial. Por la API, `GET /api/proyectos` responde 500 con `{"error":"error interno del servidor"}`, sin mostrar el error interno. (HU-50, CA-50.5)

#### Scenario: Falla el almacenamiento (caso de error)
- **WHEN** el repositorio de proyectos falla al listar y se consulta la lista
- **THEN** la consulta devuelve el error del repositorio y ninguna lista
- **AND** por la API responde 500 con `{"error":"error interno del servidor"}`

### Requirement: Consultas simultáneas
Consultar la lista mientras se crean proyectos al mismo tiempo SHALL NOT producir accesos concurrentes sin protección: cada consulta devuelve una lista coherente y `go test -race` no reporta condiciones de carrera. (HU-50, CA-50.7)

#### Scenario: Listar mientras se crean proyectos
- **WHEN** varias consultas de la lista corren al mismo tiempo que la creación de proyectos, con `go test -race`
- **THEN** no se reportan condiciones de carrera y, al terminar, la lista tiene todos los proyectos creados

### Requirement: Elegir el proyecto actual desde la lista en la interfaz web
La interfaz web SHALL mostrar, con el botón «Ver proyectos», una tabla con los proyectos en el orden de la lista (ID, nombre, descripción, estado, inicio, fin) y el mensaje "N proyecto(s).". El botón «Elegir» de cada fila SHALL dejar ese proyecto como proyecto actual y cargarlo como «Cargar proyecto», sin escribir su número. Si la API responde con error, la interfaz SHALL mostrar su mensaje. (HU-50, CA-50.3 y CA-50.5)

#### Scenario: Elegir un proyecto de la lista (caso normal)
- **WHEN** el usuario pulsa «Ver proyectos» y después «Elegir» en la fila de "Inventario"
- **THEN** «Proyecto actual» pasa a tener el ID de "Inventario", se muestran sus datos y su backlog, y el mensaje avisa que se eligió ese proyecto

#### Scenario: Sin proyectos en la interfaz (caso límite)
- **WHEN** no hay proyectos y el usuario pulsa «Ver proyectos»
- **THEN** la tabla queda vacía y el mensaje dice "0 proyecto(s)."

#### Scenario: La API falla (caso de error)
- **WHEN** el usuario pulsa «Ver proyectos» y la API responde 500
- **THEN** la interfaz muestra el mensaje de error de la API ("error interno del servidor") y no cambia el proyecto actual

### Requirement: Identificador público UUID al crear un proyecto
Al crear un proyecto, el sistema SHALL asignarle un identificador público UUID versión 4 (variante RFC 9562), en forma canónica de 36 caracteres en minúsculas. Es el único identificador del proyecto que muestran la API y la interfaz, y reemplaza hacia afuera al número correlativo de RN-01.5. `POST /api/proyectos` responde 201 con ese UUID en `id`. (HU-51, CA-51.1)

#### Scenario: Crear un proyecto (caso normal)
- **WHEN** se crea el proyecto "Sistema de ventas"
- **THEN** el proyecto creado tiene un identificador público de 36 caracteres con la forma `xxxxxxxx-xxxx-4xxx-Nxxx-xxxxxxxxxxxx`, donde N es 8, 9, a o b
- **AND** por la API responde 201 con ese identificador como texto en `id`, y no con un número

#### Scenario: El primer proyecto no es el 1 (caso límite)
- **WHEN** no hay proyectos y se crea el primero
- **THEN** su identificador público no es "1" ni ningún número, y no es el UUID nulo `00000000-0000-0000-0000-000000000000`

### Requirement: Identificadores únicos y no deducibles
Dos proyectos SHALL NOT tener el mismo identificador público, y a partir del identificador de un proyecto SHALL NOT poderse deducir el de otro ni la cantidad de proyectos: los identificadores no son correlativos. (HU-51, CA-51.2)

#### Scenario: Muchos proyectos creados seguidos (caso alternativo)
- **WHEN** se crean 1000 proyectos con nombres distintos
- **THEN** los 1000 identificadores públicos son distintos y todos son UUID versión 4
- **AND** ordenar los identificadores no da el orden de creación, es decir, no son correlativos

### Requirement: Creaciones simultáneas con identificadores únicos
Crear proyectos al mismo tiempo SHALL generar identificadores públicos únicos, sin accesos concurrentes sin protección: `go test -race` no reporta condiciones de carrera. (HU-51, CA-51.7)

#### Scenario: Crear proyectos en paralelo
- **WHEN** 50 creaciones de proyectos con nombres distintos corren al mismo tiempo, con `go test -race`
- **THEN** se crean los 50, con 50 identificadores públicos distintos, y no se reportan condiciones de carrera

### Requirement: Operar un proyecto por su UUID
Toda ruta de la API que recibe un proyecto SHALL identificarlo por su UUID: `GET` y `PUT /api/proyectos/{id}`, `PATCH /api/proyectos/{id}/fecha-inicio`, `/fecha-fin` y `/fechas`, las rutas `/integrantes`, `/backlog`, `/sprints`, `/defectos`, `/metricas`, `/metricas/sprints`, `/resumen`, `/reporte` y `/reporte/pdf`. Las mayúsculas del UUID se aceptan. La existencia del proyecto se valida, igual que hoy (RN-03.2). (HU-51, CA-51.3)

#### Scenario: Consultar un proyecto por su UUID (caso normal)
- **WHEN** existe el proyecto "Inventario" y se pide `GET /api/proyectos/{su UUID}`
- **THEN** responde 200 con sus datos, en el formato de HU-03, y con el mismo UUID en `id`

#### Scenario: UUID en mayúsculas (caso alternativo)
- **WHEN** se pide `GET /api/proyectos/{su UUID en mayúsculas}`
- **THEN** responde 200 con el proyecto, y el `id` de la respuesta está en minúsculas

#### Scenario: Operaciones que dependen del proyecto (caso normal)
- **WHEN** se crea un ítem con `POST /api/proyectos/{UUID}/backlog`, un sprint con `POST /api/proyectos/{UUID}/sprints` y un integrante con `POST /api/proyectos/{UUID}/integrantes`
- **THEN** los tres se crean en ese proyecto, y `GET /api/proyectos/{UUID}/backlog`, `/sprints`, `/integrantes`, `/metricas`, `/resumen` y `/reporte` los devuelven

### Requirement: Ninguna respuesta expone el ID entero del proyecto
Toda respuesta de la API que identifica un proyecto SHALL hacerlo con su UUID: `id` en los proyectos (también en la lista de HU-50 y en el reporte de HU-38), `proyecto_id` en ítems, sprints, integrantes, métricas y resumen, y el nombre del archivo del PDF de HU-40. (HU-51, CA-51.3)

#### Scenario: Un ítem consultado por su propio ID (caso alternativo)
- **WHEN** un ítem del proyecto "Inventario" se consulta con `GET /api/items/{id del ítem}`
- **THEN** su `proyecto_id` es el UUID de "Inventario" y no un número

#### Scenario: El PDF del reporte (caso alternativo)
- **WHEN** se pide `GET /api/proyectos/{UUID}/reporte/pdf`
- **THEN** responde 200 y el encabezado `Content-Disposition` nombra el archivo `reporte-proyecto-{UUID}.pdf`

### Requirement: Identificador con formato inválido
Si el `{id}` de una ruta de proyecto no tiene la forma canónica `xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx` (36 caracteres hexadecimales con guiones), la API SHALL responder 400 con `{"error":"el id del proyecto debe ser un UUID"}`, sin consultar el proyecto. Se rechazan los números, las llaves, el prefijo `urn:uuid:` y la forma sin guiones. (HU-51, CA-51.4)

#### Scenario: Un ID numérico (caso de error)
- **WHEN** se pide `GET /api/proyectos/1`
- **THEN** responde 400 con `{"error":"el id del proyecto debe ser un UUID"}`, y no 404

#### Scenario: Otras formas de UUID (caso límite)
- **WHEN** se pide `GET /api/proyectos/{UUID}` con el UUID de un proyecto existente escrito entre llaves, con `urn:uuid:` adelante o sin guiones
- **THEN** responde 400 con `{"error":"el id del proyecto debe ser un UUID"}`

#### Scenario: ID inválido y cuerpo inválido (caso límite)
- **WHEN** se pide `PATCH /api/proyectos/abc/fecha-inicio` con un cuerpo que no es JSON
- **THEN** responde 400 con el mensaje de ID de proyecto inválido, porque el ID de la ruta se valida primero (orden de HU-18)

#### Scenario: Las demás entidades conservan su ID entero (caso límite)
- **WHEN** se pide `GET /api/items/abc`
- **THEN** responde 400 con `{"error":"el id debe ser un número entero"}`, igual que hoy, porque los ítems siguen con IDs enteros (parte pendiente de HU-51)

### Requirement: Proyecto inexistente
Si el `{id}` tiene la forma de un UUID pero no corresponde a ningún proyecto, el sistema SHALL devolver `ErrProjectNotFound` y la API SHALL responder 404. Las validaciones de contenido de cada ruta conservan su orden: siguen antes que la existencia del proyecto. (HU-51, CA-51.4)

#### Scenario: UUID que no existe (caso de error)
- **WHEN** se pide `GET /api/proyectos/{UUID versión 4 que nunca se asignó}`
- **THEN** responde 404 con `{"error":"no existe un proyecto con ese id"}`

#### Scenario: El UUID nulo (caso límite)
- **WHEN** se pide `GET /api/proyectos/00000000-0000-0000-0000-000000000000`
- **THEN** responde 404, porque tiene la forma de un UUID pero nunca se asigna

#### Scenario: Se conserva el orden de errores (caso límite)
- **WHEN** se pide `PATCH /api/proyectos/{UUID inexistente}/fecha-inicio` con `{"fecha_inicio":""}`
- **THEN** responde 400 con el mensaje de `ErrMissingStartDate`, igual que hoy con un ID entero inexistente (HU-18)

### Requirement: Mismo comportamiento en memoria y en PostgreSQL
El identificador público SHALL guardarse y buscarse igual en memoria y en PostgreSQL. En PostgreSQL se guarda en una columna de tipo UUID, única, de la tabla de proyectos. Las relaciones de ítems, sprints e integrantes con su proyecto siguen funcionando. (HU-51, CA-51.5)

#### Scenario: Guardar y buscar por UUID en los dos repositorios
- **WHEN** se guarda un proyecto y después se lo busca por su identificador público, en memoria y en PostgreSQL
- **THEN** los dos repositorios devuelven el mismo proyecto con el mismo UUID, y un UUID que no existe devuelve `ErrProjectNotFound`

#### Scenario: Dos proyectos con el mismo UUID en PostgreSQL (caso de error)
- **WHEN** se intenta guardar en PostgreSQL un proyecto con un identificador público que ya usa otro
- **THEN** la base lo rechaza y el guardado devuelve un error, sin crear el proyecto

### Requirement: La interfaz web usa el UUID
En la interfaz web, el campo «Proyecto actual» SHALL aceptar el UUID como texto. Crear un proyecto, «Ver proyectos» y «Elegir» (HU-50) SHALL dejar el UUID en ese campo, y todas las secciones que usan el proyecto actual lo envían en la ruta. La lista de HU-50 conserva su orden por nombre. (HU-51, CA-51.6)

#### Scenario: Elegir un proyecto de la lista (caso normal)
- **WHEN** el usuario pulsa «Ver proyectos» y «Elegir» en la fila de "Inventario"
- **THEN** «Proyecto actual» muestra el UUID de "Inventario" y se cargan sus datos y su backlog

#### Scenario: Escribir un ID que no es UUID (caso de error)
- **WHEN** el usuario escribe "1" en «Proyecto actual» y pulsa «Cargar proyecto»
- **THEN** la interfaz muestra el mensaje "el id del proyecto debe ser un UUID" que devuelve la API y no carga ningún proyecto

#### Scenario: Descargar el PDF
- **WHEN** el usuario descarga el reporte en PDF del proyecto actual
- **THEN** el archivo se llama `reporte-proyecto-{UUID}.pdf`
