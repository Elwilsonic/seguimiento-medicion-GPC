# language: es
# HU-31 · Consultar los defectos de un proyecto o de un sprint (Issue #49)
# Especificación: docs/especificaciones/HU-31-consultar-defectos.md
# CA-31.9 (API e interfaz web) se cubre con tests de internal/httpapi, no con escenarios godog.
# El filtro con solo espacios lo decide el handler (D3): se prueba en internal/httpapi.
# Los defectos de partida se guardan con Save del repositorio en memoria: registrarlos es HU-29.

Característica: HU-31 Consultar los defectos de un proyecto o de un sprint
  Como Agile Enabler,
  quiero consultar los defectos de un proyecto o de un sprint,
  para evaluar su calidad.

  Antecedentes:
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Y que existe un ítem "Login de usuarios" en el proyecto "Sistema de ventas"
    Y que existe un ítem "Carrito de compras" en el proyecto "Sistema de ventas"
    Y que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"

  # Caso normal - CA-31.1
  Escenario: Listar los defectos de un proyecto en todos sus estados (CA-31.1)
    Dado que existe un defecto del ítem con el ID 1 detectado en el sprint con el ID 1 en estado "ABIERTO"
    Y que existe un defecto del ítem con el ID 2 detectado en el sprint con el ID 1 en estado "EN_CORRECCION"
    Y que existe un defecto del ítem con el ID 1 detectado en el sprint con el ID 1 en estado "RESUELTO"
    Cuando consulto los defectos del proyecto "Sistema de ventas"
    Entonces la consulta de defectos es exitosa
    Y se listan los defectos con los IDs "1, 2, 3"
    Y los defectos listados están en los estados "ABIERTO, EN_CORRECCION, RESUELTO"
    Y cada defecto listado tiene todos sus datos

  # Caso normal - CA-31.1
  Escenario: Listar los defectos detectados en un sprint (CA-31.1)
    Dado que existe un defecto del ítem con el ID 1 detectado en el sprint con el ID 1 en estado "ABIERTO"
    Y que existe un defecto del ítem con el ID 2 detectado en el sprint con el ID 1 en estado "RESUELTO"
    Cuando consulto los defectos del sprint con el ID 1
    Entonces la consulta de defectos es exitosa
    Y se listan los defectos con los IDs "1, 2"
    Y cada defecto listado tiene todos sus datos

  # Caso alternativo - CA-31.2
  Escenario: No mezclar defectos de otro proyecto (CA-31.2)
    Dado que existe un proyecto con nombre "Sistema de stock"
    Y que existe un ítem "Alta de productos" en el proyecto "Sistema de stock"
    Y que existe un sprint abierto "Sprint A" en el proyecto "Sistema de stock"
    Y que existe un defecto del ítem con el ID 1 detectado en el sprint con el ID 1 en estado "ABIERTO"
    Y que existe un defecto del ítem con el ID 3 detectado en el sprint con el ID 2 en estado "ABIERTO"
    Cuando consulto los defectos del proyecto "Sistema de ventas"
    Entonces se listan los defectos con los IDs "1"
    Cuando consulto los defectos del proyecto "Sistema de stock"
    Entonces se listan los defectos con los IDs "2"

  # Caso alternativo - CA-31.2
  Escenario: Un defecto resuelto en otro sprint aparece solo en el de detección (CA-31.2)
    Dado que el sprint con el ID 1 está cerrado
    Y que existe un sprint abierto "Sprint 2" en el proyecto "Sistema de ventas"
    Y que existe un defecto del ítem con el ID 1 detectado en el sprint con el ID 1 y resuelto en el sprint con el ID 2
    Cuando consulto los defectos del sprint con el ID 1
    Entonces se listan los defectos con los IDs "1"
    Cuando consulto los defectos del sprint con el ID 2
    Entonces la lista de defectos está vacía
    Cuando consulto los defectos del proyecto "Sistema de ventas"
    Entonces se listan los defectos con los IDs "1"

  # Caso alternativo - CA-31.2
  Escenario: Consultar un sprint cerrado (CA-31.2)
    Dado que existe un defecto del ítem con el ID 1 detectado en el sprint con el ID 1 en estado "EN_CORRECCION"
    Y que el sprint con el ID 1 está cerrado
    Cuando consulto los defectos del sprint con el ID 1
    Entonces la consulta de defectos es exitosa
    Y se listan los defectos con los IDs "1"

  # Caso alternativo - CA-31.3
  Esquema del escenario: Filtrar los defectos por estado (CA-31.3)
    Dado que existe un defecto del ítem con el ID 1 detectado en el sprint con el ID 1 en estado "ABIERTO"
    Y que existe un defecto del ítem con el ID 2 detectado en el sprint con el ID 1 en estado "EN_CORRECCION"
    Y que existe un defecto del ítem con el ID 1 detectado en el sprint con el ID 1 en estado "RESUELTO"
    Y que existe un defecto del ítem con el ID 2 detectado en el sprint con el ID 1 en estado "ABIERTO"
    Cuando consulto los defectos del <alcance> con el estado "<estado>"
    Entonces la consulta de defectos es exitosa
    Y se listan los defectos con los IDs "<ids>"

    Ejemplos:
      | alcance                       | estado          | ids  |
      | proyecto "Sistema de ventas"  | ABIERTO         | 1, 4 |
      | proyecto "Sistema de ventas"  | en_correccion   | 2    |
      | sprint con el ID 1            | Resuelto        | 3    |
      | sprint con el ID 1            | abierto         | 1, 4 |

  # Caso límite - CA-31.3
  Escenario: Normalizar el estado del filtro con espacios en los extremos (CA-31.3)
    Dado que existe un defecto del ítem con el ID 1 detectado en el sprint con el ID 1 en estado "ABIERTO"
    Y que existe un defecto del ítem con el ID 2 detectado en el sprint con el ID 1 en estado "EN_CORRECCION"
    Cuando consulto los defectos del proyecto "Sistema de ventas" con el estado "  en_correccion  "
    Entonces la consulta de defectos es exitosa
    Y se listan los defectos con los IDs "2"

  # Caso de error - CA-31.3
  Esquema del escenario: Rechazar un estado de filtro inválido (CA-31.3)
    Dado que existe un defecto del ítem con el ID 1 detectado en el sprint con el ID 1 en estado "ABIERTO"
    Cuando consulto los defectos del <alcance> con el estado "<estado>"
    Entonces la consulta de defectos falla porque el estado no es válido

    Ejemplos:
      | alcance                       | estado        |
      | proyecto "Sistema de ventas"  | CERRADO       |
      | sprint con el ID 1            | PENDIENTE     |
      | proyecto con el ID 999        | CERRADO       |

  # Caso límite - CA-31.4
  Escenario: Proyecto y sprint sin defectos devuelven una lista vacía (CA-31.4)
    Cuando consulto los defectos del proyecto "Sistema de ventas"
    Entonces la consulta de defectos es exitosa
    Y la lista de defectos está vacía
    Cuando consulto los defectos del sprint con el ID 1
    Entonces la consulta de defectos es exitosa
    Y la lista de defectos está vacía

  # Caso límite - CA-31.4
  Escenario: Proyecto sin historias devuelve una lista vacía (CA-31.4)
    Dado que existe un proyecto con nombre "Proyecto vacío"
    Cuando consulto los defectos del proyecto "Proyecto vacío"
    Entonces la consulta de defectos es exitosa
    Y la lista de defectos está vacía

  # Caso límite - CA-31.4
  Escenario: Filtro sin coincidencias devuelve una lista vacía (CA-31.4)
    Dado que existe un defecto del ítem con el ID 1 detectado en el sprint con el ID 1 en estado "ABIERTO"
    Cuando consulto los defectos del proyecto "Sistema de ventas" con el estado "RESUELTO"
    Entonces la consulta de defectos es exitosa
    Y la lista de defectos está vacía

  # Caso de error - CA-31.5
  Esquema del escenario: Rechazar un proyecto y un sprint inexistentes (CA-31.5)
    Dado que existe un defecto del ítem con el ID 1 detectado en el sprint con el ID 1 en estado "ABIERTO"
    Cuando consulto los defectos del proyecto con el ID <id>
    Entonces la consulta de defectos falla porque no existe el proyecto
    Cuando consulto los defectos del sprint con el ID <id>
    Entonces la consulta de defectos falla porque no existe el sprint

    Ejemplos:
      | id  |
      | 999 |
      | 0   |
      | -1  |

  # Caso límite - CA-31.6
  Escenario: Modificar lo devuelto no altera lo guardado (CA-31.6)
    Dado que existe un defecto del ítem con el ID 1 detectado en el sprint con el ID 1 en estado "ABIERTO"
    Y que existe un defecto del ítem con el ID 2 detectado en el sprint con el ID 1 en estado "RESUELTO"
    Cuando consulto los defectos del proyecto "Sistema de ventas"
    Y modifico todos los datos de los defectos listados
    Entonces los defectos guardados no cambiaron
    Y al volver a consultar los defectos del proyecto "Sistema de ventas" se listan los mismos defectos guardados
    Y al volver a consultar los defectos del sprint con el ID 1 se listan los mismos defectos guardados

  # Caso de error - CA-31.8
  Escenario: Error del repositorio al listar (CA-31.8)
    Dado que existe un defecto del ítem con el ID 1 detectado en el sprint con el ID 1 en estado "ABIERTO"
    Y que el repositorio de defectos falla al listar
    Cuando consulto los defectos del proyecto "Sistema de ventas"
    Entonces la consulta de defectos falla con el error del repositorio
    Cuando consulto los defectos del sprint con el ID 1
    Entonces la consulta de defectos falla con el error del repositorio

  # Caso límite - CA-31.7
  Escenario: Consultas simultáneas con registros y cambios de estado (CA-31.7)
    Dado que existe un defecto del ítem con el ID 1 detectado en el sprint con el ID 1 en estado "ABIERTO"
    Cuando consulto en simultáneo 20 veces los defectos del proyecto "Sistema de ventas" y del sprint con el ID 1 mientras registro 20 defectos y cambio el estado del defecto con el ID 1
    Entonces todas las consultas de defectos son exitosas
    Y al final se listan 21 defectos del proyecto "Sistema de ventas"
    Y al final se listan 21 defectos del sprint con el ID 1
