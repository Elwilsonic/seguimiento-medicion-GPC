# language: es
# HU-15 · Registrar integrante del proyecto (Issue #29)
# Especificación: docs/especificaciones/HU-15-registrar-integrante.md
# CA-15.6 (PostgreSQL) y CA-15.7 (API e interfaz) se cubren con tests de Go,
# no con escenarios.

Característica: HU-15 Registrar integrante del proyecto
  Como Agile Enabler,
  quiero registrar un integrante con nombre y rol en un proyecto,
  para saber quién participa.

  Antecedentes:
    Dado que existe un proyecto con nombre "Sistema de ventas"

  # Caso normal
  Escenario: Registrar un integrante con nombre y rol (CA-15.1)
    Cuando registro al integrante "Ana Pérez" con el rol "PRODUCT_BUILDER" en el proyecto "Sistema de ventas"
    Entonces el integrante queda registrado con el ID 1, el nombre "Ana Pérez" y el rol "PRODUCT_BUILDER"

  # Caso alternativo - los tres roles
  Esquema del escenario: Se acepta cada uno de los tres roles (CA-15.1)
    Cuando registro al integrante "Ana" con el rol "<rol>" en el proyecto "Sistema de ventas"
    Entonces el integrante queda registrado con el ID 1, el nombre "Ana" y el rol "<guardado>"

    Ejemplos:
      | rol               | guardado          |
      | PRODUCT_ARCHITECT | PRODUCT_ARCHITECT |
      | AGILE_ENABLER     | AGILE_ENABLER     |
      | product_builder   | PRODUCT_BUILDER   |

  # Caso límite - espacios en los extremos
  Escenario: El nombre se guarda sin espacios en los extremos (CA-15.1)
    Cuando registro al integrante "  Ana Pérez  " con el rol "AGILE_ENABLER" en el proyecto "Sistema de ventas"
    Entonces el integrante queda registrado con el ID 1, el nombre "Ana Pérez" y el rol "AGILE_ENABLER"

  # Casos de error - datos obligatorios
  Esquema del escenario: Nombre y rol son obligatorios y el rol tiene que ser válido (CA-15.2)
    Cuando registro al integrante "<nombre>" con el rol "<rol>" en el proyecto "Sistema de ventas"
    Entonces el registro falla con el error "<error>"
    Y el proyecto "Sistema de ventas" no tiene integrantes

    Ejemplos:
      | nombre | rol             | error                                                                    |
      |        | PRODUCT_BUILDER | el nombre del integrante es obligatorio                                  |
      | Ana    |                 | el rol del integrante es obligatorio                                     |
      | Ana    | SCRUM_MASTER    | el rol debe ser PRODUCT_ARCHITECT, AGILE_ENABLER o PRODUCT_BUILDER       |
      | Ana    | Product Builder | el rol debe ser PRODUCT_ARCHITECT, AGILE_ENABLER o PRODUCT_BUILDER       |

  # Caso de error - duplicado
  Esquema del escenario: No se registra dos veces al mismo integrante en el proyecto (CA-15.3)
    Dado que el integrante "Ana" está registrado con el rol "PRODUCT_BUILDER" en el proyecto "Sistema de ventas"
    Cuando registro al integrante "<nombre>" con el rol "<rol>" en el proyecto "Sistema de ventas"
    Entonces el registro falla con el error "ya hay un integrante con ese nombre en el proyecto"
    Y el proyecto "Sistema de ventas" tiene 1 integrante

    Ejemplos:
      | nombre | rol             |
      | Ana    | PRODUCT_BUILDER |
      | ana    | PRODUCT_BUILDER |
      | Ana    | AGILE_ENABLER   |

  # Caso alternativo - otro proyecto
  Escenario: El mismo nombre en otro proyecto se acepta (CA-15.3)
    Dado que existe un proyecto con nombre "Inventario"
    Y que el integrante "Ana" está registrado con el rol "PRODUCT_BUILDER" en el proyecto "Sistema de ventas"
    Cuando registro al integrante "Ana" con el rol "PRODUCT_BUILDER" en el proyecto "Inventario"
    Entonces el integrante queda registrado con el ID 2, el nombre "Ana" y el rol "PRODUCT_BUILDER"

  # Caso de error - proyecto inexistente
  Esquema del escenario: Registrar en un proyecto inexistente falla (CA-15.4)
    Cuando registro al integrante "Ana" con el rol "PRODUCT_BUILDER" en el proyecto con el ID <id>
    Entonces el registro falla porque el proyecto no existe

    Ejemplos:
      | id |
      | 99 |
      | 0  |
      | -1 |

  # Caso límite - copia
  Escenario: Modificar el integrante devuelto no cambia lo guardado (CA-15.5)
    Cuando registro al integrante "Ana" con el rol "PRODUCT_BUILDER" en el proyecto "Sistema de ventas"
    Y modifico el integrante devuelto
    Entonces el proyecto "Sistema de ventas" tiene al integrante "Ana" con el rol "PRODUCT_BUILDER"
