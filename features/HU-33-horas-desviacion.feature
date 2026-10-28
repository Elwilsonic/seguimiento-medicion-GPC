# language: es
# HU-33 · Calcular horas estimadas, horas reales y desviación (Issue #51)
#
# Trazabilidad: cada escenario cita su criterio de la SDD
# (docs/especificaciones/HU-33-horas-desviacion.md). Los pasos usan los
# Manager reales en memoria de backlog, sprint y effort.

Característica: HU-33 Calcular horas estimadas, horas reales y desviación
  Como Agile Enabler
  quiero calcular las horas estimadas, las horas reales y la desviación entre ambas
  para mejorar las estimaciones futuras

  Antecedentes:
    Dado que existe un proyecto con nombre "Sistema de ventas"

  Escenario: Horas y desviación de un sprint con el ejemplo de TEC-03 (CA-33.1, CA-33.2 y CA-33.3)
    Dado que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Y que el sprint "Sprint 1" tiene los ítems con horas y esfuerzos:
      | titulo | horas | esfuerzo |
      | A      | 10    | 12       |
      | B      | 6     | 5        |
      | C      | 0     | 3        |
      | D      | 4     | 0        |
    Cuando consulto las horas del sprint "Sprint 1"
    Entonces las horas estimadas son 20
    Y las horas reales son 20
    Y la desviación en horas es 0
    Y la desviación porcentual es 0
    Y hay 1 ítem sin horas estimadas

  Escenario: Un sprint sin ítems no tiene desviación (CA-33.4 y CA-33.7)
    Dado que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Cuando consulto las horas del sprint "Sprint 1"
    Entonces las horas estimadas son 0
    Y las horas reales son 0
    Y la desviación no está disponible
    Y hay 0 ítems sin horas estimadas

  Escenario: Un proyecto sin ítems no tiene desviación (CA-33.7)
    Cuando consulto las horas del proyecto "Sistema de ventas"
    Entonces las horas estimadas son 0
    Y las horas reales son 0
    Y la desviación no está disponible
    Y hay 0 ítems sin horas estimadas

  Escenario: El proyecto suma todo el backlog y los esfuerzos de sus sprints (CA-33.6)
    Dado que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Y que el sprint "Sprint 1" tiene los ítems con horas y esfuerzos:
      | titulo | horas | esfuerzo |
      | A      | 10    | 12       |
      | B      | 6     | 5        |
    Y que el proyecto "Sistema de ventas" tiene los ítems sin sprint con horas:
      | titulo | horas |
      | E      | 4     |
      | F      | 0     |
    Cuando consulto las horas del proyecto "Sistema de ventas"
    Entonces las horas estimadas son 20
    Y las horas reales son 17
    Y la desviación en horas es -3
    Y la desviación porcentual es -15
    Y hay 1 ítem sin horas estimadas

  Escenario: Sin horas estimadas la desviación no está disponible (CA-33.4)
    Dado que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Y que el sprint "Sprint 1" tiene los ítems con horas y esfuerzos:
      | titulo | horas | esfuerzo |
      | A      | 0     | 3        |
    Cuando consulto las horas del sprint "Sprint 1"
    Entonces las horas estimadas son 0
    Y las horas reales son 3
    Y la desviación no está disponible
    Y hay 1 ítem sin horas estimadas

  Escenario: Sin esfuerzos la desviación no está disponible (CA-33.4)
    Dado que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Y que el sprint "Sprint 1" tiene los ítems con horas y esfuerzos:
      | titulo | horas | esfuerzo |
      | A      | 10    | 0        |
    Cuando consulto las horas del sprint "Sprint 1"
    Entonces las horas estimadas son 10
    Y las horas reales son 0
    Y la desviación no está disponible
    Y hay 0 ítems sin horas estimadas

  Escenario: Un sprint solo suma sus propios esfuerzos (CA-33.2)
    Dado que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Y que el sprint "Sprint 1" tiene los ítems con horas y esfuerzos:
      | titulo | horas | esfuerzo |
      | A      | 10    | 2        |
    Y que cierro el sprint "Sprint 1"
    Y que existe un sprint abierto "Sprint 2" en el proyecto "Sistema de ventas"
    Y que el sprint "Sprint 2" tiene los ítems con horas y esfuerzos:
      | titulo | horas | esfuerzo |
      | B      | 6     | 4        |
    Cuando consulto las horas del sprint "Sprint 1"
    Entonces las horas estimadas son 10
    Y las horas reales son 2
    Y la desviación en horas es -8
    Y la desviación porcentual es -80
    Y hay 0 ítems sin horas estimadas

  Esquema del escenario: Consultar un sprint inexistente falla (CA-33.8)
    Cuando consulto las horas del sprint con el ID <id>
    Entonces la consulta falla porque el sprint no existe

    Ejemplos:
      | id |
      | 99 |
      | 0  |
      | -1 |

  Esquema del escenario: Consultar un proyecto inexistente falla (CA-33.8)
    Cuando consulto las horas del proyecto con el ID <id>
    Entonces la consulta falla porque el proyecto no existe

    Ejemplos:
      | id |
      | 99 |
      | 0  |
      | -1 |

  Escenario: Consultar las métricas no modifica los datos (CA-33.8)
    Dado que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Y que el sprint "Sprint 1" tiene los ítems con horas y esfuerzos:
      | titulo | horas | esfuerzo |
      | A      | 10    | 12       |
      | B      | 6     | 5        |
    Cuando consulto las horas del sprint "Sprint 1"
    Y consulto las horas del proyecto "Sistema de ventas"
    Entonces los ítems, los sprints y los esfuerzos del proyecto no cambiaron
