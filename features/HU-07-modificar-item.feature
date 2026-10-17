# language: es
# HU-07 · Modificar datos de un ítem (Issue #19)
# Especificación: docs/especificaciones/HU-07-modificar-item.md

Característica: HU-07 Modificar datos de un ítem
  Como Agile Enabler,
  quiero modificar los datos de un ítem del backlog,
  para corregirlo o precisarlo.

  # Caso normal (happy path) - CA-07.1
  Escenario: Modificar todos los datos de un ítem existente
    Dado que existe un ítem con título "Login de usuarios", descripción "Formulario de acceso", prioridad "ALTA" y criterios de aceptación "Login OK" y "Carga correcta"
    Cuando modifico el ítem con el ID 1 con título "Login con email", descripción "Acceso por email", prioridad "MEDIA", criterios de aceptación "Login por email OK" y horas estimadas 6
    Entonces la modificación es exitosa
    Y el ítem modificado tiene el título "Login con email"
    Y el ítem modificado tiene la descripción "Acceso por email"
    Y el ítem modificado tiene la prioridad "MEDIA"
    Y el ítem modificado tiene los criterios de aceptación "Login por email OK"
    Y el ítem modificado tiene las horas estimadas 6
    Y el ítem modificado pertenece al proyecto con ID 1
    Y el ítem modificado está en el estado "PENDIENTE"

  # Caso alternativo - CA-07.1 (cambiar un solo campo reenviando el resto)
  Escenario: Cambiar solo las horas estimadas manteniendo el resto
    Dado que existe un ítem con título "Login de usuarios", descripción "Formulario de acceso", prioridad "ALTA" y criterios de aceptación "Login OK"
    Cuando modifico el ítem con el ID 1 con título "Login de usuarios", descripción "Formulario de acceso", prioridad "ALTA", criterios de aceptación "Login OK" y horas estimadas 4
    Entonces la modificación es exitosa
    Y el ítem modificado tiene las horas estimadas 4
    Y el ítem modificado tiene el título "Login de usuarios"
    Y el ítem modificado tiene la descripción "Formulario de acceso"
    Y el ítem modificado tiene la prioridad "ALTA"
    Y el ítem modificado tiene los criterios de aceptación "Login OK"

  # Caso límite - CA-07.2, RN-07.3 (igual que la creación)
  Esquema del escenario: Aceptar la prioridad sin distinguir mayúsculas ni espacios en los extremos
    Dado que existe un ítem con título "Login de usuarios", descripción "", prioridad "MEDIA" y criterios de aceptación "Login OK"
    Cuando modifico el ítem con el ID 1 con título "Login de usuarios", descripción "", prioridad "<prioridad>", criterios de aceptación "Login OK" y horas estimadas 0
    Entonces la modificación es exitosa
    Y el ítem modificado tiene la prioridad "<guardada>"

    Ejemplos:
      | prioridad | guardada |
      | alta      | ALTA     |
      | baja      | BAJA     |
      | meDia     | MEDIA    |

  # Caso límite - CA-07.1, RN-07.1
  Escenario: Recortar los espacios del título nuevo en los extremos
    Dado que existe un ítem con título "Login de usuarios", descripción "", prioridad "ALTA" y criterios de aceptación "Login OK"
    Cuando modifico el ítem con el ID 1 con título "  Login con email  ", descripción "", prioridad "ALTA", criterios de aceptación "Login OK" y horas estimadas 0
    Entonces la modificación es exitosa
    Y el ítem modificado tiene el título "Login con email"

  # Caso límite - CA-07.4 (fracciones admitidas)
  Escenario: Asignar horas estimadas fraccionadas
    Dado que existe un ítem con título "Login de usuarios", descripción "", prioridad "ALTA" y criterios de aceptación "Login OK"
    Cuando modifico el ítem con el ID 1 con título "Login de usuarios", descripción "", prioridad "ALTA", criterios de aceptación "Login OK" y horas estimadas 2.5
    Entonces la modificación es exitosa
    Y el ítem modificado tiene las horas estimadas 2.5

  # Caso límite - CA-07.4 (0 quita las horas)
  Escenario: Dejar el ítem sin horas estimadas pasando 0
    Dado que existe un ítem con título "Login de usuarios", descripción "", prioridad "ALTA", criterios de aceptación "Login OK" y horas estimadas 8.5
    Cuando modifico el ítem con el ID 1 con título "Login de usuarios", descripción "", prioridad "ALTA", criterios de aceptación "Login OK" y horas estimadas 0
    Entonces la modificación es exitosa
    Y el ítem modificado queda sin horas estimadas

  # Caso de error - CA-07.4
  Esquema del escenario: Rechazar horas estimadas inválidas
    Dado que existe un ítem con título "Login de usuarios", descripción "", prioridad "ALTA" y criterios de aceptación "Login OK"
    Cuando modifico el ítem con el ID 1 con título "Login de usuarios", descripción "", prioridad "ALTA", criterios de aceptación "Login OK" y horas estimadas <horas>
    Entonces la modificación falla porque las horas estimadas son inválidas
    Y el ítem guardado con el ID 1 sigue teniendo las horas estimadas sin horas

    Ejemplos:
      | horas |
      | -1    |
      | NaN   |
      | Inf   |

  # Caso de error - CA-07.2
  Escenario: Rechazar un título vacío
    Dado que existe un ítem con título "Login de usuarios", descripción "Formulario de acceso", prioridad "ALTA" y criterios de aceptación "Login OK"
    Cuando modifico el ítem con el ID 1 con título "", descripción "Formulario de acceso", prioridad "ALTA", criterios de aceptación "Login OK" y horas estimadas 4
    Entonces la modificación falla porque el título es obligatorio
    Y el ítem guardado con el ID 1 sigue teniendo el título "Login de usuarios" y la prioridad "ALTA"
    Y el ítem guardado con el ID 1 sigue teniendo las horas estimadas sin horas

  # Caso límite - CA-07.2
  Escenario: Rechazar un título con solo espacios
    Dado que existe un ítem con título "Login de usuarios", descripción "", prioridad "ALTA" y criterios de aceptación "Login OK"
    Cuando modifico el ítem con el ID 1 con título "   ", descripción "", prioridad "ALTA", criterios de aceptación "Login OK" y horas estimadas 4
    Entonces la modificación falla porque el título es obligatorio
    Y el ítem guardado con el ID 1 sigue teniendo el título "Login de usuarios" y la prioridad "ALTA"

  # Caso de error - CA-07.2
  Escenario: Rechazar una prioridad que no existe
    Dado que existe un ítem con título "Login de usuarios", descripción "", prioridad "ALTA" y criterios de aceptación "Login OK"
    Cuando modifico el ítem con el ID 1 con título "Login de usuarios", descripción "", prioridad "URGENTE", criterios de aceptación "Login OK" y horas estimadas 4
    Entonces la modificación falla porque la prioridad debe ser ALTA, MEDIA o BAJA
    Y el ítem guardado con el ID 1 sigue teniendo el título "Login de usuarios" y la prioridad "ALTA"

  # Caso de error - CA-07.2
  Escenario: Rechazar una modificación sin criterios de aceptación
    Dado que existe un ítem con título "Login de usuarios", descripción "", prioridad "ALTA" y criterios de aceptación "Login OK"
    Cuando modifico el ítem con el ID 1 con título "Login de usuarios", descripción "", prioridad "ALTA", criterios de aceptación vacíos y horas estimadas 4
    Entonces la modificación falla porque el ítem necesita al menos un criterio de aceptación
    Y el ítem guardado con el ID 1 sigue teniendo el título "Login de usuarios" y la prioridad "ALTA"

  # Caso de error - CA-07.3
  Esquema del escenario: Modificar un ítem que no existe
    Dado que existe un ítem con título "Login de usuarios", descripción "", prioridad "ALTA" y criterios de aceptación "Login OK"
    Cuando modifico el ítem con el ID <id> con título "Login con email", descripción "", prioridad "MEDIA", criterios de aceptación "Login OK" y horas estimadas 4
    Entonces la modificación falla porque el ítem no existe

    Ejemplos:
      | id  |
      | 999 |
      | 0   |
      | -1  |

  # Caso de error - CA-07.3
  Escenario: Modificar cuando no hay ítems creados
    Dado que no hay ítems creados
    Cuando modifico el ítem con el ID 1 con título "Login de usuarios", descripción "", prioridad "ALTA", criterios de aceptación "Login OK" y horas estimadas 4
    Entonces la modificación falla porque el ítem no existe

  # Caso límite - CA-07.5
  Escenario: Modificar el ítem devuelto no altera el ítem guardado
    Dado que existe un ítem con título "Login de usuarios", descripción "", prioridad "ALTA" y criterios de aceptación "Login OK"
    Cuando modifico el ítem con el ID 1 con título "Login con email", descripción "", prioridad "ALTA", criterios de aceptación "Login OK" y horas estimadas 4
    Y cambio el título y los criterios del ítem modificado
    Entonces al reconsultar el ítem con el ID 1 obtengo el título "Login con email" y los criterios de aceptación "Login OK"

  # Caso límite - CA-07.6
  Escenario: Varias modificaciones en simultáneo del mismo ítem
    Dado que existen 3 ítems en el proyecto "Sistema de ventas"
    Cuando se modifican en simultáneo 50 veces el ítem con el ID 1 con títulos distintos
    Entonces hay 3 ítems guardados
    Y el ítem guardado con el ID 1 coincide con una de las versiones enviadas
