# language: es
# HU-28 · Corregir un registro de esfuerzo (Issue #46)
#
# Trazabilidad: cada escenario cita su criterio de la SDD
# (docs/especificaciones/HU-28-corregir-esfuerzo.md). Los pasos usan los
# Manager reales de proyectos, backlog, sprints y esfuerzos, en memoria.
#
# Tres criterios de la SDD se cubren fuera de godog: CA-28.6 (memoria y
# PostgreSQL) con el test de contrato del repositorio, y CA-28.7 (API e
# interfaz) y la parte de métricas y reportes de CA-28.5 con los tests de
# internal/httpapi.

Característica: HU-28 Corregir un registro de esfuerzo
  Como Product Builder
  quiero corregir un registro de esfuerzo cargado por error
  para que los datos sean confiables

  Antecedentes:
    Dado que existe un sprint abierto "Sprint 1"
    Y que la historia "Login" está en el sprint "Sprint 1"
    Y que "Ana" registró 12 horas en la historia "Login" del sprint "Sprint 1"

  Escenario: Corregir las horas de un esfuerzo de un sprint abierto (CA-28.1)
    Cuando corrijo las horas del esfuerzo 1 a 10
    Entonces la corrección devuelve el esfuerzo 1 con 10 horas
    Y las horas guardadas del esfuerzo 1 son 10
    Y el integrante, la fecha, la actividad, la historia y el sprint del esfuerzo 1 no cambiaron

  Escenario: Las horas con decimales se guardan sin redondear (CA-28.1)
    Cuando corrijo las horas del esfuerzo 1 a 2.75
    Entonces las horas guardadas del esfuerzo 1 son 2.75

  Escenario: Corregir con las mismas horas es válido (CA-28.1)
    Cuando corrijo las horas del esfuerzo 1 a 12
    Entonces la corrección devuelve el esfuerzo 1 con 12 horas
    Y las horas guardadas del esfuerzo 1 son 12

  Esquema del escenario: Las horas corregidas se validan como al registrar (CA-28.2)
    Cuando corrijo las horas del esfuerzo 1 a <horas>
    Entonces la corrección falla porque las horas son inválidas
    Y los esfuerzos guardados no cambiaron

    Ejemplos:
      | horas |
      | 0     |
      | -3    |
      | NaN   |
      | +Inf  |
      | -Inf  |

  Escenario: No se corrigen esfuerzos de un sprint cerrado (CA-28.3)
    Dado que el sprint "Sprint 1" está cerrado
    Cuando corrijo las horas del esfuerzo 1 a 10
    Entonces la corrección falla porque el sprint está cerrado
    Y los esfuerzos guardados no cambiaron

  Esquema del escenario: Corregir un esfuerzo inexistente falla (CA-28.4)
    Cuando corrijo las horas del esfuerzo <id> a 10
    Entonces la corrección falla porque el esfuerzo no existe
    Y los esfuerzos guardados no cambiaron

    Ejemplos:
      | id |
      | 99 |
      | 0  |
      | -1 |

  Escenario: Las horas inválidas se informan antes que un esfuerzo inexistente (CA-28.2)
    Cuando corrijo las horas del esfuerzo 99 a 0
    Entonces la corrección falla porque las horas son inválidas

  Escenario: Las horas reales del sprint reflejan la corrección (CA-28.5)
    Dado que "Luis" registró 5 horas en la historia "Login" del sprint "Sprint 1"
    Cuando corrijo las horas del esfuerzo 1 a 10
    Entonces el total de horas del sprint "Sprint 1" es 15
