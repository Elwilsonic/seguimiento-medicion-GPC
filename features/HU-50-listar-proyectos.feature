# language: es
# HU-50 · Listar los proyectos (Issue #183), vinculada a HU-03 (#16)
# Especificación (OpenSpec): openspec/specs/proyectos/spec.md
# CA-50.3 (interfaz), la ruta GET /api/proyectos (CA-50.5) y la concurrencia
# (CA-50.7) se cubren con tests de Go, no con escenarios.

Característica: HU-50 Listar los proyectos
  Como integrante del equipo,
  quiero ver la lista de todos los proyectos,
  para elegir uno sin tener que saber su número.

  # Caso normal
  Escenario: Listar los proyectos con todos sus datos (CA-50.1)
    Dado que existe un proyecto con nombre "Sistema de ventas" y descripción "Ventas online"
    Y que existe un proyecto con nombre "Inventario" y descripción ""
    Y que el proyecto "Inventario" está en el estado "EN_CURSO"
    Y que el proyecto "Inventario" tiene las fechas "2026-03-01" a "2026-06-30"
    Cuando consulto la lista de proyectos
    Entonces la consulta de la lista es exitosa
    Y obtengo estos proyectos, en este orden:
      | nombre            | descripcion   | estado      | inicio     | fin        |
      | Inventario        |               | EN_CURSO    | 2026-03-01 | 2026-06-30 |
      | Sistema de ventas | Ventas online | PLANIFICADO |            |            |
    Y cada proyecto de la lista tiene el ID que se le asignó al crearlo

  # Caso alternativo
  Escenario: Un proyecto modificado aparece con sus datos nuevos (CA-50.1)
    Dado que existe un proyecto con nombre "Inventario" y descripción "Depósito"
    Y que el proyecto "Inventario" se modificó con el nombre "Inventario central" y el estado "FINALIZADO"
    Cuando consulto la lista de proyectos
    Entonces la consulta de la lista es exitosa
    Y obtengo estos proyectos, en este orden:
      | nombre             | descripcion | estado     | inicio | fin |
      | Inventario central | Depósito    | FINALIZADO |        |     |

  # Caso alternativo
  Escenario: Los proyectos se ordenan por nombre sin distinguir mayúsculas (CA-50.6)
    Dado que existe un proyecto con nombre "Sistema de ventas" y descripción ""
    Y que existe un proyecto con nombre "inventario" y descripción ""
    Y que existe un proyecto con nombre "Agenda" y descripción ""
    Cuando consulto la lista de proyectos
    Entonces la consulta de la lista es exitosa
    Y los nombres de la lista son, en este orden: "Agenda", "inventario", "Sistema de ventas"

  # Caso límite
  Escenario: Un único proyecto (CA-50.1)
    Dado que existe un proyecto con nombre "Agenda" y descripción "Turnos"
    Cuando consulto la lista de proyectos
    Entonces la consulta de la lista es exitosa
    Y los nombres de la lista son, en este orden: "Agenda"

  # Caso límite
  Escenario: Sin proyectos, la lista está vacía y no es un error (CA-50.2)
    Dado que no hay proyectos creados
    Cuando consulto la lista de proyectos
    Entonces la consulta de la lista es exitosa
    Y la lista está vacía y no es nula

  # Caso límite
  Escenario: Modificar la lista devuelta no altera los proyectos guardados (CA-50.4)
    Dado que existe un proyecto con nombre "Agenda" y descripción "Turnos"
    Cuando consulto la lista de proyectos
    Y cambio el nombre del primer proyecto de la lista a "Otro" y su estado a "FINALIZADO"
    Entonces al consultar el proyecto "Agenda" por su ID tiene el nombre "Agenda" y el estado "PLANIFICADO"
    Y al volver a consultar la lista los nombres son, en este orden: "Agenda"

  # Caso de error
  Escenario: Falla el almacenamiento al listar (CA-50.5)
    Dado que el almacenamiento de proyectos falla al listar
    Cuando consulto la lista de proyectos
    Entonces la consulta de la lista falla con el error del almacenamiento
    Y no se devuelve ninguna lista
