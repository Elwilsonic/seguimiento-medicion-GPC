# language: es
# HU-03 · Consultar estado del proyecto (Issue #16)
# Especificación: docs/especificaciones/HU-03-consultar-proyecto.md

Característica: HU-03 Consultar estado del proyecto
  Como Agile Enabler,
  quiero consultar un proyecto y su estado actual,
  para saber en qué situación está.

  # Caso normal (happy path) - CA-03.1
  Escenario: Consultar un proyecto existente por su ID
    Dado que existe un proyecto con nombre "Sistema de ventas" y descripción "Gestión de ventas"
    Cuando consulto el proyecto con el ID 1
    Entonces obtengo el proyecto con el nombre "Sistema de ventas"
    Y el proyecto tiene la descripción "Gestión de ventas"
    Y el proyecto está en el estado "PLANIFICADO"

  # Caso alternativo - CA-03.1
  Escenario: Consultar un proyecto que no es el primero
    Dado que existe un proyecto con nombre "Proyecto A"
    Y que existe un proyecto con nombre "Proyecto B"
    Cuando consulto el proyecto con el ID 2
    Entonces obtengo el proyecto con el nombre "Proyecto B"
    Y el proyecto está en el estado "PLANIFICADO"

  # Caso alternativo - CA-03.1
  Escenario: Consultar un proyecto con descripción vacía
    Dado que existe un proyecto con nombre "Proyecto A" y descripción ""
    Cuando consulto el proyecto con el ID 1
    Entonces obtengo el proyecto con el nombre "Proyecto A"
    Y el proyecto tiene la descripción ""

  # Caso alternativo - CA-03.1
  Escenario: Consultar un proyecto en un estado distinto del inicial
    Dado que existe un proyecto con nombre "Proyecto A" en el estado "EN_CURSO"
    Cuando consulto el proyecto con el ID 1
    Entonces obtengo el proyecto con el nombre "Proyecto A"
    Y el proyecto está en el estado "EN_CURSO"

  # Caso límite - CA-03.2
  Escenario: Consultar un ID mayor al último asignado
    Dado que existe un proyecto con nombre "Proyecto A"
    Cuando consulto el proyecto con el ID 999
    Entonces la consulta falla porque el proyecto no existe

  # Caso límite - CA-03.2
  Escenario: Consultar cuando no hay proyectos creados
    Dado que no hay proyectos creados
    Cuando consulto el proyecto con el ID 1
    Entonces la consulta falla porque el proyecto no existe

  # Caso límite - CA-03.2
  Esquema del escenario: Consultar un ID imposible
    Dado que existe un proyecto con nombre "Proyecto A"
    Cuando consulto el proyecto con el ID <id>
    Entonces la consulta falla porque el proyecto no existe

    Ejemplos:
      | id |
      | 0  |
      | -1 |

  # Caso límite - CA-03.3
  Escenario: Modificar el proyecto devuelto no altera el proyecto guardado
    Dado que existe un proyecto con nombre "Proyecto A" y descripción "Original"
    Cuando consulto el proyecto con el ID 1
    Y cambio el nombre del proyecto obtenido a "Otro nombre"
    Entonces el proyecto guardado con el ID 1 tiene el nombre "Proyecto A"
    Y al reconsultar el proyecto con el ID 1 obtengo el nombre "Proyecto A"

  # Caso límite - CA-03.4
  Escenario: Consultar mientras se crean proyectos en simultáneo
    Dado que no hay proyectos creados
    Cuando se crean 50 proyectos en simultáneo con nombres distintos mientras se consultan en simultáneo los IDs del 1 al 50
    Entonces hay 50 proyectos guardados
    Y ninguna consulta simultánea devuelve un error inesperado
    Y al reconsultar los IDs del 1 al 50 se obtienen los 50 proyectos
