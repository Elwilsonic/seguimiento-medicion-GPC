# Spec Delta

## Purpose

Gestión de proyectos (épica E1): crear, consultar, modificar y listar los proyectos, con sus fechas, para que el equipo pueda elegir y seguir cada uno.

## ADDED Requirements

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
