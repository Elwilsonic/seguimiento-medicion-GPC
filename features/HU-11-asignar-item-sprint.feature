# language: es
# HU-11 · Asignar historias del backlog a un sprint (Issue #25)
# Especificación: docs/especificaciones/HU-11-asignar-item-sprint.md

Característica: HU-11 Asignar historias del backlog a un sprint
  Como Agile Enabler,
  quiero asignar ítems del backlog a un sprint,
  para definir el alcance de la iteración.

  # Caso normal (happy path) - CA-11.1, RN-11.1, RN-11.2 y RN-11.4
  Escenario: Asignar un ítem a un sprint abierto del mismo proyecto
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Y que existe un ítem con título "Login de usuarios", descripción "Formulario de acceso", prioridad "ALTA" y criterios de aceptación "Login OK" en el proyecto "Sistema de ventas"
    Y que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Cuando asigno el ítem con el ID 1 al sprint con el ID 1
    Entonces la asignación es exitosa
    Y el ítem devuelto queda asignado al sprint con el ID 1
    Y el ítem guardado queda asignado al sprint con el ID 1

  # Caso normal - CA-11.1
  Escenario: Asignar dos ítems distintos al mismo sprint
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Y que existe un ítem con título "Login de usuarios", descripción "", prioridad "ALTA" y criterios de aceptación "Login OK" en el proyecto "Sistema de ventas"
    Y que existe un ítem con título "Listar proyectos", descripción "", prioridad "MEDIA" y criterios de aceptación "Lista OK" en el proyecto "Sistema de ventas"
    Y que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Cuando asigno el ítem con el ID 1 al sprint con el ID 1
    Y asigno el ítem con el ID 2 al sprint con el ID 1
    Entonces los dos ítems quedan asignados al sprint con el ID 1

  # Caso alternativo - RN-11.6 (el estado del ítem es indistinto)
  Escenario: Un ítem terminado se asigna igual
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Y que existe un ítem con título "Login de usuarios", descripción "", prioridad "ALTA" y criterios de aceptación "Login OK" en el estado "TERMINADO" en el proyecto "Sistema de ventas"
    Y que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Cuando asigno el ítem con el ID 1 al sprint con el ID 1
    Entonces la asignación es exitosa
    Y el ítem guardado queda asignado al sprint con el ID 1

  # Caso límite - CA-11.5 (la asignación no toca el resto del ítem)
  Escenario: Asignar no cambia ningún otro campo del ítem
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Y que existe un ítem con título "Login de usuarios", descripción "Formulario de acceso", prioridad "ALTA" y criterios de aceptación "Login OK" y "Carga correcta" en el proyecto "Sistema de ventas"
    Y que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Cuando asigno el ítem con el ID 1 al sprint con el ID 1
    Entonces el ítem guardado conserva el título "Login de usuarios", la descripción "Formulario de acceso", la prioridad "ALTA", los criterios de aceptación "Login OK" y "Carga correcta", el estado "PENDIENTE", los story points 0 y las horas estimadas 0

  # Caso límite - CA-11.5 (el Update de HU-07 conserva la asignación)
  Escenario: Modificar el ítem con el Update de HU-07 conserva la asignación
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Y que existe un ítem con título "Login de usuarios", descripción "", prioridad "ALTA" y criterios de aceptación "Login OK" en el proyecto "Sistema de ventas"
    Y que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Cuando asigno el ítem con el ID 1 al sprint con el ID 1
    Y modifico el ítem con el ID 1 con título "Login con email", descripción "Acceso por email", prioridad "MEDIA", criterios de aceptación "Login por email OK" y horas estimadas 4 con el Update de HU-07
    Entonces la modificación es exitosa
    Y el ítem guardado sigue asignado al sprint con el ID 1

  # Caso límite - CA-11.7
  Escenario: Modificar el ítem devuelto no altera el ítem guardado
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Y que existe un ítem con título "Login de usuarios", descripción "", prioridad "ALTA" y criterios de aceptación "Login OK" en el proyecto "Sistema de ventas"
    Y que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Cuando asigno el ítem con el ID 1 al sprint con el ID 1
    Y cambio el título, los criterios y el sprint del ítem devuelto
    Entonces al reconsultar el ítem con el ID 1 obtengo el título "Login de usuarios" y sigue asignado al sprint con el ID 1

  # Caso límite - CA-11.6
  Escenario: Asignaciones simultáneas del mismo ítem a dos sprints
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Y que existe un ítem con título "Login de usuarios", descripción "", prioridad "ALTA" y criterios de aceptación "Login OK" en el proyecto "Sistema de ventas"
    Y que existen dos sprints abiertos "Sprint 1" y "Sprint 2" en el proyecto "Sistema de ventas"
    Cuando se asignan en simultáneo 10 veces el ítem con el ID 1 a los sprints con los IDs 1 y 2
    Entonces exactamente una asignación tiene éxito
    Y las demás fallan porque el ítem ya está asignado a un sprint
    Y el ítem guardado queda asignado a uno de los sprints

  # Caso límite - CA-11.6
  Escenario: Asignaciones simultáneas de ítems distintos al mismo sprint
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Y que existen 5 ítems en el proyecto "Sistema de ventas"
    Y que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Cuando se asignan en simultáneo los 5 ítems al sprint con el ID 1
    Entonces las 5 asignaciones tienen éxito
    Y los 5 ítems quedan asignados al sprint con el ID 1

  # Caso de error - CA-11.2 y RN-11.3
  Esquema del escenario: Asignar un ítem ya asignado
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Y que existe un ítem con título "Login de usuarios", descripción "", prioridad "ALTA" y criterios de aceptación "Login OK" en el proyecto "Sistema de ventas"
    Y que existen dos sprints abiertos "Sprint 1" y "Sprint 2" en el proyecto "Sistema de ventas"
    Cuando asigno el ítem con el ID 1 al sprint con el ID 1
    Y asigno el ítem con el ID 1 al sprint con el ID <sprint>
    Entonces la asignación falla porque el ítem ya está asignado a un sprint
    Y el ítem guardado sigue asignado al sprint con el ID 1

    Ejemplos:
      | caso          | sprint |
      | mismo sprint | 1      |
      | otro sprint  | 2      |

  # Caso de error - CA-11.3 y RN-11.1
  Escenario: Asignar a un sprint cerrado
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Y que existe un ítem con título "Login de usuarios", descripción "", prioridad "ALTA" y criterios de aceptación "Login OK" en el proyecto "Sistema de ventas"
    Y que existe un sprint cerrado "Sprint 1" en el proyecto "Sistema de ventas"
    Cuando asigno el ítem con el ID 1 al sprint con el ID 1
    Entonces la asignación falla porque el sprint está cerrado
    Y el ítem guardado sigue sin sprint asignado

  # Caso de error - CA-11.3 y RN-11.2
  Escenario: Asignar un ítem de otro proyecto
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Y que existe un proyecto con nombre "Portal de clientes"
    Y que existe un ítem con título "Login de usuarios", descripción "", prioridad "ALTA" y criterios de aceptación "Login OK" en el proyecto "Portal de clientes"
    Y que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Cuando asigno el ítem con el ID 1 al sprint con el ID 1
    Entonces la asignación falla porque el ítem pertenece a otro proyecto
    Y el ítem guardado sigue sin sprint asignado

  # Caso de error - CA-11.4 y RN-11.1
  Esquema del escenario: Asignar a un sprint inexistente
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Y que existe un ítem con título "Login de usuarios", descripción "", prioridad "ALTA" y criterios de aceptación "Login OK" en el proyecto "Sistema de ventas"
    Y que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Cuando asigno el ítem con el ID 1 al sprint con el ID <sprint>
    Entonces la asignación falla porque no existe un sprint con ese id
    Y el ítem guardado sigue sin sprint asignado

    Ejemplos:
      | sprint |
      | 999    |
      | 0      |
      | -1     |

  # Caso de error - CA-11.4
  Escenario: Asignar cuando no hay sprints creados
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Y que existe un ítem con título "Login de usuarios", descripción "", prioridad "ALTA" y criterios de aceptación "Login OK" en el proyecto "Sistema de ventas"
    Cuando asigno el ítem con el ID 1 al sprint con el ID 1
    Entonces la asignación falla porque no existe un sprint con ese id

  # Caso de error - CA-11.4
  Esquema del escenario: Asignar un ítem inexistente
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Y que existe un ítem con título "Login de usuarios", descripción "", prioridad "ALTA" y criterios de aceptación "Login OK" en el proyecto "Sistema de ventas"
    Y que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Cuando asigno el ítem con el ID <item> al sprint con el ID 1
    Entonces la asignación falla porque no existe un ítem del backlog con ese id

    Ejemplos:
      | item |
      | 999  |
      | 0    |
      | -1   |

  # Caso de error - CA-11.4
  Escenario: Asignar cuando no hay ítems creados
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Y que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Cuando asigno el ítem con el ID 1 al sprint con el ID 1
    Entonces la asignación falla porque no existe un ítem del backlog con ese id
