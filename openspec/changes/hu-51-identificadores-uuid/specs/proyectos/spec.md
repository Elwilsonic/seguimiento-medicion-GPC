# Spec Delta

## Purpose

Gestión de proyectos (épica E1): crear, consultar, modificar y listar los proyectos, con sus fechas, para que el equipo pueda elegir y seguir cada uno.

## ADDED Requirements

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
