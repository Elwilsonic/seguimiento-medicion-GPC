# language: es
# HU-04 · Crear ítem del backlog (Issue #17)
# Especificación: docs/especificaciones/HU-04-crear-item-backlog.md

Característica: HU-04 Crear ítem del backlog
  Como Agile Enabler,
  quiero agregar un ítem al Product Backlog de un proyecto con sus campos obligatorios,
  para registrar el trabajo pendiente.

  # Caso normal (happy path) - CA-04.1, CA-04.4
  Escenario: Crear un ítem con todos los campos válidos
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Cuando creo un ítem en el proyecto "Sistema de ventas" con título "Login de usuarios", descripción "Acceso con usuario y contraseña", prioridad "ALTA" y criterios de aceptación "Credenciales válidas", "Credenciales inválidas" y "Bloqueo tras 3 intentos"
    Entonces el ítem se crea con el ID 1
    Y el ítem queda en el estado "PENDIENTE"
    Y el ítem tiene los story points en 0
    Y el ítem tiene el título "Login de usuarios"
    Y el ítem tiene la descripción "Acceso con usuario y contraseña"
    Y el ítem tiene la prioridad "ALTA"
    Y el ítem tiene los criterios de aceptación "Credenciales válidas", "Credenciales inválidas" y "Bloqueo tras 3 intentos"
    Y el ítem guardado conserva el título "Login de usuarios"
    Y el ítem guardado conserva los criterios de aceptación "Credenciales válidas", "Credenciales inválidas" y "Bloqueo tras 3 intentos"

  # Caso alternativo - CA-04.1, CA-04.3
  Escenario: Crear un ítem sin descripción y con espacios en los extremos
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Cuando creo un ítem en el proyecto "Sistema de ventas" con título "  Registro de usuarios  ", descripción "", prioridad " media " y criterios de aceptación "  Registro OK  "
    Entonces el ítem se crea con el ID 1
    Y el ítem tiene el título "Registro de usuarios"
    Y el ítem tiene la descripción ""
    Y el ítem tiene la prioridad "MEDIA"
    Y el ítem guardado conserva la prioridad "MEDIA"
    Y el ítem guardado conserva los criterios de aceptación "Registro OK"

  # Caso límite - CA-04.4, RN-04.6
  Escenario: Los IDs son correlativos y globales entre proyectos
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Y que existe un proyecto con nombre "Portal de clientes"
    Cuando creo un ítem en el proyecto "Sistema de ventas" con título "Login de usuarios", descripción "", prioridad "ALTA" y criterios de aceptación "Login OK"
    Y creo un ítem en el proyecto "Portal de clientes" con título "Historial de pedidos", descripción "", prioridad "BAJA" y criterios de aceptación "Historial visible"
    Entonces los ítems se crean con los IDs 1 y 2
    Y hay 2 ítems guardados

  # Caso alternativo - CA-04.5, RN-04.5
  Escenario: Un proyecto finalizado acepta ítems
    Dado que existe un proyecto finalizado con nombre "Proyecto cerrado"
    Cuando creo un ítem en el proyecto "Proyecto cerrado" con título "Ajuste posterior", descripción "", prioridad "BAJA" y criterios de aceptación "Ajuste aplicado"
    Entonces el ítem se crea con el ID 1
    Y el ítem queda en el estado "PENDIENTE"

  # Caso límite - RN-04.1
  Escenario: El título puede repetirse dentro del mismo proyecto
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Cuando creo un ítem en el proyecto "Sistema de ventas" con título "Login de usuarios", descripción "", prioridad "ALTA" y criterios de aceptación "Login OK"
    Y creo un ítem en el proyecto "Sistema de ventas" con título "Login de usuarios", descripción "", prioridad "MEDIA" y criterios de aceptación "Login OK"
    Entonces los ítems se crean con los IDs 1 y 2
    Y hay 2 ítems guardados

  # Caso límite - CA-04.2, RN-04.4
  Escenario: Los criterios vacíos se descartan y se conserva el orden
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Cuando creo un ítem en el proyecto "Sistema de ventas" con título "Login de usuarios", descripción "", prioridad "ALTA" y criterios de aceptación "Primero", "  ", "", "Segundo"
    Entonces el ítem se crea con el ID 1
    Y el ítem guardado conserva los criterios de aceptación "Primero" y "Segundo"

  # Caso límite - CA-04.6, RN-04.9
  Escenario: Modificar el ítem devuelto no altera el ítem guardado
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Cuando creo un ítem en el proyecto "Sistema de ventas" con título "Login de usuarios", descripción "Original", prioridad "ALTA" y criterios de aceptación "Login OK"
    Y modifico el título y los criterios del ítem devuelto
    Entonces el ítem guardado conserva el título "Login de usuarios"
    Y el ítem guardado conserva los criterios de aceptación "Login OK"

  # Caso límite - CA-04.6, RN-04.9
  Escenario: Modificar la lista de criterios recibida no altera el ítem guardado
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Cuando creo un ítem en el proyecto "Sistema de ventas" con título "Login de usuarios", descripción "", prioridad "ALTA" y criterios de aceptación "Login OK"
    Y modifico la lista de criterios que envié
    Entonces el ítem guardado conserva los criterios de aceptación "Login OK"

  # Caso límite - CA-04.7
  Escenario: Creaciones simultáneas con IDs distintos
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Cuando se crean 50 ítems en simultáneo en el proyecto "Sistema de ventas"
    Entonces hay 50 ítems guardados
    Y los IDs de los ítems guardados son todos distintos

  # Caso de error - CA-04.2, RN-04.1
  Escenario: No se puede crear un ítem con el título vacío
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Cuando creo un ítem en el proyecto "Sistema de ventas" con título "   ", descripción "", prioridad "ALTA" y criterios de aceptación "Login OK"
    Entonces la creación falla porque el título es obligatorio
    Y hay 0 ítems guardados

  # Caso de error - CA-04.2, RN-04.4
  Escenario: No se puede crear un ítem sin criterios de aceptación
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Cuando creo un ítem en el proyecto "Sistema de ventas" con título "Login de usuarios", descripción "", prioridad "ALTA" y sin criterios de aceptación
    Entonces la creación falla porque el ítem necesita al menos un criterio de aceptación
    Y hay 0 ítems guardados

  # Caso de error y límite - CA-04.2, RN-04.4
  Esquema del escenario: Los criterios vacíos no cuentan como criterio de aceptación
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Cuando creo un ítem en el proyecto "Sistema de ventas" con título "Login de usuarios", descripción "", prioridad "ALTA" y criterios de aceptación <criterios>
    Entonces la creación falla porque el ítem necesita al menos un criterio de aceptación
    Y hay 0 ítems guardados

    Ejemplos:
      | criterios |
      | ""        |
      | "  "      |

  # Caso de error - CA-04.3, RN-04.3
  Esquema del escenario: No se puede crear un ítem con una prioridad inválida
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Cuando creo un ítem en el proyecto "Sistema de ventas" con título "Login de usuarios", descripción "", prioridad "<prioridad>" y criterios de aceptación "Login OK"
    Entonces la creación falla porque la prioridad debe ser ALTA, MEDIA o BAJA
    Y hay 0 ítems guardados

    Ejemplos:
      | prioridad |
      | ""        |
      | URGENTE   |

  # Caso de error y límite - CA-04.5
  Esquema del escenario: No se puede crear un ítem en un proyecto que no existe
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Cuando creo un ítem en el proyecto con ID <id>, título "Login de usuarios", descripción "", prioridad "ALTA" y criterios de aceptación "Login OK"
    Entonces la creación falla porque el proyecto no existe
    Y hay 0 ítems guardados

    Ejemplos:
      | id  |
      | 999 |
      | 0   |
      | -1  |

  # Caso de error y límite - CA-04.2, orden de validación
  Escenario: Con varios datos inválidos se informa el primer error según el orden
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Cuando creo un ítem en el proyecto con ID 999, título "   ", descripción "", prioridad "URGENTE" y sin criterios de aceptación
    Entonces la creación falla porque el título es obligatorio
    Y hay 0 ítems guardados

  # Caso de error y límite - CA-04.2, orden de validación
  Escenario: Los criterios se validan antes que el proyecto
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Cuando creo un ítem en el proyecto con ID 999, título "Login de usuarios", descripción "", prioridad "ALTA" y sin criterios de aceptación
    Entonces la creación falla porque el ítem necesita al menos un criterio de aceptación
    Y hay 0 ítems guardados
