# integrantes Specification

## Purpose

Integrantes de un proyecto (épica E1): registrar, modificar y consultar los integrantes con su rol. El comportamiento anterior a HU-52 está en las SDD de HU-15 a HU-17 (`docs/especificaciones/`).

## Requirements

### Requirement: Identificador público UUID al registrar un integrante
Al registrar un integrante, el sistema SHALL asignarle un identificador público UUID versión 4, en forma canónica de 36 caracteres en minúsculas. Es el único identificador del integrante que muestran la API y la interfaz, y reemplaza hacia afuera al entero correlativo de RN-15.7, que sigue como clave interna. `POST /api/proyectos/{id}/integrantes` responde 201 con ese UUID en `id`. (HU-52, CA-52.1)

#### Scenario: Registrar un integrante (caso normal)
- **WHEN** se registra a "Ana Pérez" como PRODUCT_BUILDER en un proyecto
- **THEN** responde 201 y el `id` del integrante es un texto con la forma de un UUID versión 4

#### Scenario: El primer integrante no es el 1 (caso límite)
- **WHEN** no hay integrantes y se registra el primero
- **THEN** su `id` no es "1" ni ningún número, y no es el UUID nulo

### Requirement: Identificadores de integrantes únicos y no deducibles
Dos integrantes SHALL NOT tener el mismo identificador público, y a partir del de un integrante SHALL NOT poderse deducir el de otro ni la cantidad de integrantes. (HU-52, CA-52.2)

#### Scenario: Muchos integrantes registrados seguidos (caso alternativo)
- **WHEN** se registran 1000 integrantes con nombres distintos
- **THEN** los 1000 identificadores públicos son distintos, todos son UUID versión 4, y ordenarlos no da el orden de registro

### Requirement: Registros simultáneos de integrantes
Registrar integrantes al mismo tiempo SHALL generar identificadores públicos únicos, sin condiciones de carrera (`go test -race` en la CI). (HU-52, CA-52.7)

#### Scenario: Registrar integrantes en paralelo
- **WHEN** 50 registros de integrantes con nombres distintos corren al mismo tiempo
- **THEN** se registran los 50, con 50 identificadores públicos distintos

### Requirement: Modificar un integrante por su UUID
`PUT /api/integrantes/{id}` SHALL identificar al integrante por su UUID, con mayúsculas aceptadas. La existencia del integrante se valida igual que hoy (HU-16, `ErrMemberNotFound`), porque la modificación necesita el integrante guardado. `GET /api/integrantes/{id}` sigue respondiendo 405. (HU-52, CA-52.3)

#### Scenario: Modificar un integrante por su UUID (caso normal)
- **WHEN** se pide `PUT /api/integrantes/{UUID de "Ana Pérez"}` con el nombre "Ana María" y el rol AGILE_ENABLER
- **THEN** responde 200 con el integrante modificado y con el mismo UUID en `id`

#### Scenario: UUID del integrante en mayúsculas (caso alternativo)
- **WHEN** se pide `PUT /api/integrantes/{su UUID en mayúsculas}` con datos válidos
- **THEN** responde 200 y el `id` de la respuesta está en minúsculas

### Requirement: Ninguna respuesta expone el ID entero de un integrante
Toda respuesta de la API que identifica un integrante SHALL hacerlo con su UUID en `id`: el registro, la lista de HU-17 y la modificación. (HU-52, CA-52.3)

#### Scenario: La lista de integrantes (caso alternativo)
- **WHEN** se pide `GET /api/proyectos/{id}/integrantes` con dos integrantes registrados
- **THEN** cada integrante de la lista tiene su UUID en `id`, y ninguno tiene un número

### Requirement: Identificador de integrante con formato inválido
Si el `{id}` de `PUT /api/integrantes/{id}` no tiene la forma canónica de un UUID, la API SHALL responder 400 con `{"error":"el id del integrante debe ser un UUID"}`, sin consultar el integrante. Se rechazan los números, las llaves, el prefijo `urn:uuid:` y la forma sin guiones. (HU-52, CA-52.4)

#### Scenario: Un ID numérico de integrante (caso de error)
- **WHEN** se pide `PUT /api/integrantes/1` con datos válidos
- **THEN** responde 400 con `{"error":"el id del integrante debe ser un UUID"}`, y no 404

#### Scenario: El ID del integrante se valida antes que el cuerpo (caso límite)
- **WHEN** se pide `PUT /api/integrantes/abc` con un cuerpo que no es JSON
- **THEN** responde 400 con el mensaje del ID del integrante

### Requirement: Integrante inexistente
Si el `{id}` tiene la forma de un UUID pero no corresponde a ningún integrante, la API SHALL responder 404 con `{"error":"no existe un integrante con ese id"}`. Las validaciones de contenido conservan su orden: siguen antes que la existencia. (HU-52, CA-52.4)

#### Scenario: UUID de integrante que no existe (caso de error)
- **WHEN** se pide `PUT /api/integrantes/{UUID versión 4 que nunca se asignó}` con datos válidos
- **THEN** responde 404 con `{"error":"no existe un integrante con ese id"}`

#### Scenario: Se conserva el orden de errores (caso límite)
- **WHEN** se pide `PUT /api/integrantes/{UUID inexistente}` con el nombre vacío
- **THEN** responde 400 con `{"error":"el nombre del integrante es obligatorio"}`, igual que hoy con un ID entero inexistente (HU-16)

### Requirement: Integrantes en memoria y en PostgreSQL
El identificador público del integrante SHALL guardarse y buscarse igual en memoria y en PostgreSQL, en una columna UUID única de la tabla de integrantes. La relación con su proyecto sigue con el entero. (HU-52, CA-52.5)

#### Scenario: Guardar y buscar un integrante por su UUID
- **WHEN** se guarda un integrante y después se lo busca por su identificador público, en memoria y en PostgreSQL
- **THEN** los dos repositorios devuelven el mismo integrante, y un UUID que no existe da `ErrMemberNotFound`

#### Scenario: Dos integrantes con el mismo UUID en PostgreSQL (caso de error)
- **WHEN** se intenta guardar en PostgreSQL un integrante con un identificador público que ya usa otro
- **THEN** la base lo rechaza y el guardado devuelve un error, sin crear el integrante

### Requirement: La interfaz web usa el UUID del integrante
En la interfaz web, el campo «ID del integrante» de "Modificar integrante" SHALL aceptar el UUID como texto. (HU-52, CA-52.6)

#### Scenario: Escribir un ID de integrante que no es UUID (caso de error)
- **WHEN** el usuario escribe "1" en «ID del integrante» y guarda
- **THEN** la interfaz muestra "el id del integrante debe ser un UUID" y no modifica nada
