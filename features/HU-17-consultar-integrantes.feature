# language: es
# HU-17 · Consultar los integrantes de un proyecto (Issue #31)
# Especificación: docs/especificaciones/HU-17-consultar-integrantes.md
# CA-17.6 (interfaz) y la ruta GET /api/proyectos/{id}/integrantes se cubren
# con tests de Go, no con escenarios.

Característica: HU-17 Consultar los integrantes de un proyecto
  Como Agile Enabler,
  quiero ver los integrantes de un proyecto con sus roles,
  para conocer la composición del equipo.

  Antecedentes:
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Y que existe un proyecto con nombre "Inventario"

  # Caso normal
  Escenario: Listar los integrantes de un proyecto con su nombre y su rol (CA-17.1)
    Dado que el integrante "Ana Pérez" está registrado con el rol "AGILE_ENABLER" en el proyecto "Sistema de ventas"
    Y que el integrante "Bruno Díaz" está registrado con el rol "PRODUCT_BUILDER" en el proyecto "Sistema de ventas"
    Y que el integrante "Carla Ruiz" está registrado con el rol "PRODUCT_ARCHITECT" en el proyecto "Sistema de ventas"
    Cuando consulto los integrantes del proyecto "Sistema de ventas"
    Entonces la consulta es exitosa
    Y obtengo estos integrantes:
      | id | nombre     | rol               |
      | 1  | Ana Pérez  | AGILE_ENABLER     |
      | 2  | Bruno Díaz | PRODUCT_BUILDER   |
      | 3  | Carla Ruiz | PRODUCT_ARCHITECT |

  # Caso alternativo - integrante modificado con HU-16
  Escenario: Un integrante modificado aparece con sus datos nuevos (CA-17.1)
    Dado que el integrante "Ana" está registrado con el rol "PRODUCT_BUILDER" en el proyecto "Sistema de ventas"
    Y que el integrante "Bruno" está registrado con el rol "PRODUCT_BUILDER" en el proyecto "Sistema de ventas"
    Y que el integrante "Ana" se modificó con el nombre "Ana Pérez" y el rol "AGILE_ENABLER"
    Cuando consulto los integrantes del proyecto "Sistema de ventas"
    Entonces la consulta es exitosa
    Y obtengo estos integrantes:
      | id | nombre    | rol             |
      | 1  | Ana Pérez | AGILE_ENABLER   |
      | 2  | Bruno     | PRODUCT_BUILDER |

  # Caso alternativo - otro proyecto con el mismo nombre
  Escenario: No aparecen los integrantes de otros proyectos, aunque tengan el mismo nombre (CA-17.3)
    Dado que el integrante "Ana" está registrado con el rol "PRODUCT_BUILDER" en el proyecto "Inventario"
    Y que el integrante "Ana" está registrado con el rol "AGILE_ENABLER" en el proyecto "Sistema de ventas"
    Y que el integrante "Bruno" está registrado con el rol "PRODUCT_BUILDER" en el proyecto "Inventario"
    Cuando consulto los integrantes del proyecto "Sistema de ventas"
    Entonces la consulta es exitosa
    Y obtengo estos integrantes:
      | id | nombre | rol           |
      | 2  | Ana    | AGILE_ENABLER |

  # Caso límite - proyecto sin integrantes
  Escenario: Un proyecto sin integrantes devuelve una lista vacía (CA-17.2)
    Dado que el integrante "Ana" está registrado con el rol "PRODUCT_BUILDER" en el proyecto "Inventario"
    Cuando consulto los integrantes del proyecto "Sistema de ventas"
    Entonces la consulta es exitosa
    Y obtengo una lista vacía de integrantes

  # Caso límite - proyecto finalizado
  Escenario: Se listan los integrantes de un proyecto finalizado (CA-17.1)
    Dado que el integrante "Ana" está registrado con el rol "PRODUCT_BUILDER" en el proyecto "Sistema de ventas"
    Y que el proyecto "Sistema de ventas" está finalizado
    Cuando consulto los integrantes del proyecto "Sistema de ventas"
    Entonces la consulta es exitosa
    Y obtengo estos integrantes:
      | id | nombre | rol             |
      | 1  | Ana    | PRODUCT_BUILDER |

  # Caso límite - la lista devuelta es una copia
  Escenario: Modificar la lista devuelta no cambia los integrantes guardados (CA-17.5)
    Dado que el integrante "Ana" está registrado con el rol "PRODUCT_BUILDER" en el proyecto "Sistema de ventas"
    Y consulto los integrantes del proyecto "Sistema de ventas"
    Cuando cambio el nombre y el rol del primer integrante devuelto
    Y consulto los integrantes del proyecto "Sistema de ventas"
    Entonces la consulta es exitosa
    Y obtengo estos integrantes:
      | id | nombre | rol             |
      | 1  | Ana    | PRODUCT_BUILDER |

  # Caso de error - proyecto inexistente
  Esquema del escenario: Consultar los integrantes de un proyecto inexistente falla (CA-17.4)
    Cuando consulto los integrantes del proyecto con el ID <id>
    Entonces la consulta falla porque no existe un proyecto con ese id

    Ejemplos:
      | id |
      | 99 |
      | 0  |
      | -1 |

  # Caso de error - falla del almacenamiento
  Escenario: Una falla del repositorio de integrantes se devuelve sin lista parcial (RN-17.8)
    Dado que el integrante "Ana" está registrado con el rol "PRODUCT_BUILDER" en el proyecto "Sistema de ventas"
    Y que el repositorio de integrantes falla al listar
    Cuando consulto los integrantes del proyecto "Sistema de ventas"
    Entonces la consulta falla con el error del repositorio
