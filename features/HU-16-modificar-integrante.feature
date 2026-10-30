# language: es
# HU-16 · Modificar los datos o el rol de un integrante (Issue #30)
# Especificación: docs/especificaciones/HU-16-modificar-integrante.md
# CA-16.5 (concurrencia y fallas del repositorio), CA-16.6 (PostgreSQL) y
# CA-16.7 (API e interfaz) se cubren con tests de Go, no con escenarios.

Característica: HU-16 Modificar los datos o el rol de un integrante
  Como Agile Enabler,
  quiero modificar el nombre o el rol de un integrante ya registrado,
  para mantener la información del equipo al día.

  Antecedentes:
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Y que el integrante "Ana" está registrado con el rol "PRODUCT_BUILDER" en el proyecto "Sistema de ventas"

  # Caso normal
  Escenario: Cambiar el nombre y el rol de un integrante (CA-16.1)
    Cuando modifico al integrante "Ana" con el nombre "Ana Pérez" y el rol "AGILE_ENABLER"
    Entonces el integrante queda con el ID 1, el nombre "Ana Pérez" y el rol "AGILE_ENABLER" en el proyecto "Sistema de ventas"

  # Caso alternativo - un solo dato, reenviando el otro
  Esquema del escenario: Cambiar un solo dato reenviando el valor actual del otro (CA-16.1)
    Cuando modifico al integrante "Ana" con el nombre "<nombre>" y el rol "<rol>"
    Entonces el integrante queda con el ID 1, el nombre "<nombre>" y el rol "<rol>" en el proyecto "Sistema de ventas"

    Ejemplos:
      | nombre    | rol               |
      | Ana Pérez | PRODUCT_BUILDER   |
      | Ana       | PRODUCT_ARCHITECT |
      | Ana       | PRODUCT_BUILDER   |

  # Caso límite - normalización
  Escenario: El nombre se guarda sin espacios en los extremos y el rol normalizado (CA-16.1)
    Cuando modifico al integrante "Ana" con el nombre "  Ana Pérez  " y el rol "  agile_enabler "
    Entonces el integrante queda con el ID 1, el nombre "Ana Pérez" y el rol "AGILE_ENABLER" en el proyecto "Sistema de ventas"

  # Casos de error - datos obligatorios
  Esquema del escenario: No se aceptan un nombre vacío ni un rol vacío o inválido (CA-16.2)
    Cuando modifico al integrante "Ana" con el nombre "<nombre>" y el rol "<rol>"
    Entonces la modificación falla con el error "<error>"
    Y el proyecto "Sistema de ventas" tiene al integrante "Ana" con el rol "PRODUCT_BUILDER"

    Ejemplos:
      | nombre | rol             | error                                                              |
      |        | PRODUCT_BUILDER | el nombre del integrante es obligatorio                            |
      | Ana    |                 | el rol del integrante es obligatorio                               |
      | Ana    | SCRUM_MASTER    | el rol debe ser PRODUCT_ARCHITECT, AGILE_ENABLER o PRODUCT_BUILDER |
      | Ana    | Product Builder | el rol debe ser PRODUCT_ARCHITECT, AGILE_ENABLER o PRODUCT_BUILDER |

  # Caso de error - integrante inexistente
  Esquema del escenario: Modificar un integrante inexistente falla (CA-16.3)
    Cuando modifico al integrante con el ID <id> con el nombre "Bruno" y el rol "AGILE_ENABLER"
    Entonces la modificación falla porque el integrante no existe
    Y el proyecto "Sistema de ventas" tiene al integrante "Ana" con el rol "PRODUCT_BUILDER"

    Ejemplos:
      | id |
      | 99 |
      | 0  |
      | -1 |

  # Caso de error - duplicado
  Esquema del escenario: El nombre nuevo no puede ser el de otro integrante del proyecto (CA-16.4)
    Dado que el integrante "Bruno" está registrado con el rol "AGILE_ENABLER" en el proyecto "Sistema de ventas"
    Cuando modifico al integrante "Ana" con el nombre "<nombre>" y el rol "<rol>"
    Entonces la modificación falla con el error "ya hay un integrante con ese nombre en el proyecto"
    Y el proyecto "Sistema de ventas" tiene 2 integrantes

    Ejemplos:
      | nombre | rol             |
      | Bruno  | PRODUCT_BUILDER |
      | BRUNO  | PRODUCT_BUILDER |
      | Bruno  | AGILE_ENABLER   |

  # Caso límite - el propio nombre
  Esquema del escenario: Conservar el propio nombre o cambiar sus mayúsculas es válido (CA-16.4)
    Cuando modifico al integrante "Ana" con el nombre "<nombre>" y el rol "PRODUCT_BUILDER"
    Entonces el integrante queda con el ID 1, el nombre "<nombre>" y el rol "PRODUCT_BUILDER" en el proyecto "Sistema de ventas"

    Ejemplos:
      | nombre |
      | Ana    |
      | ANA    |

  # Caso alternativo - otro proyecto
  Escenario: El nombre de un integrante de otro proyecto se acepta (CA-16.4)
    Dado que existe un proyecto con nombre "Inventario"
    Y que el integrante "Bruno" está registrado con el rol "AGILE_ENABLER" en el proyecto "Inventario"
    Cuando modifico al integrante "Ana" con el nombre "Bruno" y el rol "PRODUCT_BUILDER"
    Entonces el integrante queda con el ID 1, el nombre "Bruno" y el rol "PRODUCT_BUILDER" en el proyecto "Sistema de ventas"

  # Caso límite - copia
  Escenario: Cambiar el integrante devuelto no altera lo guardado (CA-16.5)
    Cuando modifico al integrante "Ana" con el nombre "Ana Pérez" y el rol "AGILE_ENABLER"
    Y cambio el integrante devuelto
    Entonces el proyecto "Sistema de ventas" tiene al integrante "Ana Pérez" con el rol "AGILE_ENABLER"
