# language: es
# HU-52 · Identificadores UUID del resto de las entidades (Issue #197), base compartida
# Cambio (OpenSpec): openspec/changes/hu-52-identificadores-uuid-entidades/
# Los escenarios son de dominio (D9 de design.md): verifican que los Manager
# de las 6 entidades asignen el UUID al crear, lo busquen y lo conserven. Las
# rutas, los 400 y los 404 de la API van en el .feature de cada entidad. Los
# 1000 identificadores (CA-52.2), PostgreSQL (CA-52.5) y la concurrencia
# (CA-52.7) se cubren con tests de Go.

Característica: HU-52 Base de los identificadores UUID de las entidades
  Como integrante del equipo,
  quiero que los ítems del backlog, los sprints, los integrantes, los defectos, los registros de esfuerzo y las sesiones de Planning Poker se identifiquen con un UUID en lugar de un número correlativo,
  para que no se pueda adivinar ni recorrer ninguna entidad, ni deducir cuántas hay.

  Antecedentes:
    Dado que existe el proyecto "Sistema de ventas" con un ítem en curso asignado a un sprint abierto

  # Caso normal
  Esquema del escenario: Crear una entidad le asigna un UUID versión 4 (CA-52.1)
    Cuando creo <entidad>
    Entonces <entidad> creado tiene un identificador público UUID versión 4 en minúsculas
    Y no es el UUID nulo

    Ejemplos:
      | entidad        |
      | un ítem        |
      | un sprint      |
      | un integrante  |
      | un defecto     |
      | un esfuerzo    |
      | una sesión     |

  # Caso alternativo
  Esquema del escenario: Varias creaciones dan identificadores distintos (CA-52.2)
    Cuando creo 20 veces <entidad>
    Entonces los 20 identificadores públicos son distintos

    Ejemplos:
      | entidad        |
      | un ítem        |
      | un sprint      |
      | un integrante  |
      | un defecto     |
      | un esfuerzo    |
      | una sesión     |

  # Caso normal
  Esquema del escenario: Buscar una entidad por su UUID (CA-52.3)
    Dado que creé <entidad>
    Cuando busco <entidad> por su identificador público
    Entonces obtengo <entidad> creado

    Ejemplos:
      | entidad        |
      | un ítem        |
      | un sprint      |
      | un integrante  |
      | un defecto     |
      | un esfuerzo    |
      | una sesión     |

  # Caso alternativo
  Esquema del escenario: Modificar una entidad conserva su UUID (CA-52.5)
    Dado que creé <entidad>
    Cuando modifico <entidad>
    Entonces <entidad> modificado conserva su identificador público

    Ejemplos:
      | entidad        |
      | un ítem        |
      | un sprint      |
      | un integrante  |
      | un defecto     |
      | un esfuerzo    |
      | una sesión     |

  # Caso límite
  Esquema del escenario: El UUID nulo no corresponde a ninguna entidad (CA-52.4)
    Dado que creé <entidad>
    Cuando busco <entidad> por el UUID nulo
    Entonces la búsqueda falla con "<error>"

    Ejemplos:
      | entidad        | error                                 |
      | un ítem        | no existe un ítem del backlog con ese id |
      | un sprint      | no existe un sprint con ese id        |
      | un integrante  | no existe un integrante con ese id    |
      | un defecto     | no existe un defecto con ese id       |
      | un esfuerzo    | no existe un esfuerzo con ese id      |
      | una sesión     | no existe la sesión de Planning Poker |

  # Caso de error
  Esquema del escenario: Un UUID que nunca se asignó da el error de su entidad (CA-52.4)
    Cuando busco <entidad> por un UUID versión 4 que nunca se asignó
    Entonces la búsqueda falla con "<error>"

    Ejemplos:
      | entidad        | error                                 |
      | un ítem        | no existe un ítem del backlog con ese id |
      | un sprint      | no existe un sprint con ese id        |
      | un integrante  | no existe un integrante con ese id    |
      | un defecto     | no existe un defecto con ese id       |
      | un esfuerzo    | no existe un esfuerzo con ese id      |
      | una sesión     | no existe la sesión de Planning Poker |
