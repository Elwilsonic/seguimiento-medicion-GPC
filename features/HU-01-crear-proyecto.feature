# language: es
# HU-01 · Crear proyecto (Issue #14)
# Especificación: docs/especificaciones/HU-01-crear-proyecto.md

Característica: HU-01 Crear proyecto
  Como Agile Enabler,
  quiero crear un proyecto con nombre, descripción y estado inicial,
  para poder empezar a planificar su trabajo.

  # Caso normal (happy path) - CA-01.1, CA-01.4
  Escenario: Crear un proyecto con nombre y descripción
    Dado que no hay proyectos creados
    Cuando creo un proyecto con nombre "Sistema de ventas" y descripción "Gestión de ventas de la empresa"
    Entonces el proyecto se crea con el ID 1
    Y el proyecto tiene el nombre "Sistema de ventas"
    Y el proyecto tiene la descripción "Gestión de ventas de la empresa"
    Y el proyecto queda en el estado "PLANIFICADO"

  # Caso alternativo - CA-01.1
  Escenario: Los proyectos reciben IDs correlativos
    Dado que existe un proyecto con nombre "Proyecto A"
    Cuando creo un proyecto con nombre "Proyecto B" y descripción "Segundo proyecto"
    Entonces el proyecto se crea con el ID 2

  # Caso alternativo - CA-01.5
  Escenario: Crear un proyecto sin descripción
    Dado que no hay proyectos creados
    Cuando creo un proyecto con nombre "Proyecto sin descripción" y descripción ""
    Entonces el proyecto se crea con el ID 1
    Y el proyecto tiene la descripción ""

  # Caso límite - CA-01.5
  Escenario: El nombre se guarda sin los espacios de los extremos
    Dado que no hay proyectos creados
    Cuando creo un proyecto con nombre "   Proyecto A   " y descripción "Con espacios"
    Entonces el proyecto tiene el nombre "Proyecto A"
    Y el proyecto guardado tiene el nombre "Proyecto A"

  # Caso límite - CA-01.3
  Escenario: Los nombres que solo difieren en los espacios internos son distintos
    Dado que existe un proyecto con nombre "Proyecto A"
    Cuando creo un proyecto con nombre "Proyecto  A" y descripción "Dos espacios internos"
    Entonces el proyecto se crea con el ID 2
    Y hay 2 proyectos guardados

  # Caso de error - CA-01.2
  Escenario: No se puede crear un proyecto con el nombre vacío
    Dado que no hay proyectos creados
    Cuando creo un proyecto con nombre "" y descripción "Sin nombre"
    Entonces la creación falla porque el nombre es obligatorio
    Y hay 0 proyectos guardados

  # Caso de error y límite - CA-01.2
  Escenario: No se puede crear un proyecto con un nombre de solo espacios
    Dado que no hay proyectos creados
    Cuando creo un proyecto con nombre "   " y descripción "Solo espacios"
    Entonces la creación falla porque el nombre es obligatorio
    Y hay 0 proyectos guardados

  # Caso de error - CA-01.3
  Esquema del escenario: No se puede crear un proyecto con un nombre repetido
    Dado que existe un proyecto con nombre "Proyecto A"
    Cuando creo un proyecto con nombre "<nombre>" y descripción "Repetido"
    Entonces la creación falla porque el nombre está repetido
    Y hay 1 proyectos guardados

    Ejemplos:
      | nombre     |
      | Proyecto A |
      | PROYECTO A |
      | proyecto a |

  # Caso de error y límite - CA-01.3
  Escenario: Un nombre repetido con espacios en los extremos también se rechaza
    Dado que existe un proyecto con nombre "Proyecto A"
    Cuando creo un proyecto con nombre "  Proyecto A  " y descripción "Repetido con espacios"
    Entonces la creación falla porque el nombre está repetido
    Y hay 1 proyectos guardados

  # Caso límite - CA-01.6
  Escenario: Modificar el proyecto devuelto no altera el proyecto guardado
    Dado que no hay proyectos creados
    Cuando creo un proyecto con nombre "Proyecto A" y descripción "Original"
    Y cambio el nombre del proyecto devuelto a "Otro nombre"
    Entonces el proyecto guardado tiene el nombre "Proyecto A"

  # Caso límite - CA-01.7
  Escenario: Creaciones simultáneas con el mismo nombre crean un solo proyecto
    Dado que no hay proyectos creados
    Cuando se crean 50 proyectos en simultáneo con el nombre "Proyecto concurrente"
    Entonces hay 1 proyectos guardados
    Y 49 creaciones fallan porque el nombre está repetido

  # Caso límite - CA-01.7
  Escenario: Creaciones simultáneas con nombres distintos reciben IDs distintos
    Dado que no hay proyectos creados
    Cuando se crean 50 proyectos en simultáneo con nombres distintos
    Entonces hay 50 proyectos guardados
    Y los IDs de los proyectos guardados son todos distintos
