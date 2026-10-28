# language: es
# HU-35 · Consultar las métricas del proyecto (Issue #53)
#
# Trazabilidad: cada escenario cita su criterio de la SDD
# (docs/especificaciones/HU-35-consulta-metricas.md) y cubre una fila de
# "Casos límite". Los pasos usan los Manager reales en memoria de backlog,
# sprint, effort y defect, con repositorios que cuentan las lecturas y que
# pueden fallar a pedido.
#
# Dos partes de la SDD son solo HTTP y se cubren en
# internal/httpapi/metrics_test.go, no con escenarios godog: CA-35.6 (rutas,
# campos, su orden y códigos) y la fila "HTTP: {id} no numérico" (400).

Característica: HU-35 Consultar las métricas del proyecto
  Como Agile Enabler
  quiero consultar todas las métricas de un proyecto juntas
  para evaluar su estado general

  Antecedentes:
    Dado que existe un proyecto con nombre "Sistema de ventas"

  Escenario: Todas las métricas de un sprint con el ejemplo de TEC-03 (CA-35.1)
    Dado que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Y que el sprint "Sprint 1" tiene los ítems:
      | titulo | sp | horas | estado    | esfuerzo |
      | A      | 5  | 10    | TERMINADO | 12       |
      | B      | 3  | 6     | TERMINADO | 5        |
      | C      | 2  | 0     | EN_CURSO  | 3        |
      | D      | 0  | 4     | PENDIENTE | 0        |
    Y que hay un defecto detectado en el sprint "Sprint 1" y resuelto en el sprint "Sprint 1"
    Cuando consulto las métricas del sprint "Sprint 1"
    Entonces los story points planificados son 10 y los completados son 8
    Y hay 1 ítem sin estimar
    Y el porcentaje de historias completadas es 50 con 2 de 4 ítems
    Y las horas estimadas son 20 y las reales son 20
    Y la desviación es 0 horas y 0 %
    Y hay 1 ítem sin horas estimadas
    Y los defectos detectados son 1 y los resueltos son 1
    Y todas las métricas están disponibles

  Escenario: Todas las métricas del proyecto con ese sprint cerrado (CA-35.1)
    Dado que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Y que el sprint "Sprint 1" tiene los ítems:
      | titulo | sp | horas | estado    | esfuerzo |
      | A      | 5  | 10    | TERMINADO | 12       |
      | B      | 3  | 6     | TERMINADO | 5        |
      | C      | 2  | 0     | EN_CURSO  | 3        |
      | D      | 0  | 4     | PENDIENTE | 0        |
    Y que hay un defecto detectado en el sprint "Sprint 1" y resuelto en el sprint "Sprint 1"
    Y que cierro el sprint "Sprint 1"
    Cuando consulto las métricas del proyecto "Sistema de ventas"
    Entonces los story points planificados son 10 y los completados son 8
    Y hay 1 ítem sin estimar
    Y la velocidad es 8 con 1 sprint cerrado
    Y el porcentaje de historias completadas es 50 con 2 de 4 ítems
    Y las horas estimadas son 20 y las reales son 20
    Y la desviación es 0 horas y 0 %
    Y hay 1 ítem sin horas estimadas
    Y los defectos detectados son 1 y los resueltos son 1
    Y todas las métricas están disponibles

  Escenario: Un proyecto recién creado responde con las cuatro métricas no disponibles (CA-35.2)
    Cuando consulto las métricas del proyecto "Sistema de ventas"
    Entonces los story points planificados son 0 y los completados son 0
    Y hay 0 ítems sin estimar
    Y los ítems totales son 0 y los terminados son 0
    Y las horas estimadas son 0 y las reales son 0
    Y hay 0 ítems sin horas estimadas
    Y los defectos detectados son 0 y los resueltos son 0
    Y las métricas no disponibles son: velocidad, porcentaje, desviación

  Escenario: Un sprint sin ítems, esfuerzos ni defectos (CA-35.2)
    Dado que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Cuando consulto las métricas del sprint "Sprint 1"
    Entonces los story points planificados son 0 y los completados son 0
    Y los ítems totales son 0 y los terminados son 0
    Y las horas estimadas son 0 y las reales son 0
    Y los defectos detectados son 0 y los resueltos son 0
    Y las métricas no disponibles son: porcentaje, desviación

  Escenario: Un proyecto con ítems y sin sprints cerrados solo no tiene velocidad (CA-35.2)
    Dado que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Y que el sprint "Sprint 1" tiene los ítems:
      | titulo | sp | horas | estado    | esfuerzo |
      | A      | 5  | 10    | TERMINADO | 12       |
    Cuando consulto las métricas del proyecto "Sistema de ventas"
    Entonces el porcentaje de historias completadas es 100 con 1 de 1 ítems
    Y la desviación es 2 horas y 20 %
    Y las métricas no disponibles son: velocidad

  Escenario: Ítems estimados sin esfuerzos solo no tienen desviación (CA-35.2)
    Dado que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Y que el sprint "Sprint 1" tiene los ítems:
      | titulo | sp | horas | estado    | esfuerzo |
      | A      | 5  | 10    | PENDIENTE | 0        |
    Cuando consulto las métricas del sprint "Sprint 1"
    Entonces las horas estimadas son 10 y las reales son 0
    Y el porcentaje de historias completadas es 0 con 0 de 1 ítems
    Y las métricas no disponibles son: desviación

  Escenario: Esfuerzos sin horas estimadas solo no tienen desviación (CA-35.2)
    Dado que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Y que el sprint "Sprint 1" tiene los ítems:
      | titulo | sp | horas | estado   | esfuerzo |
      | A      | 5  | 0     | EN_CURSO | 3        |
    Cuando consulto las métricas del sprint "Sprint 1"
    Entonces las horas estimadas son 0 y las reales son 3
    Y hay 1 ítem sin horas estimadas
    Y las métricas no disponibles son: desviación

  Escenario: Un proyecto con tres sprints suma sus esfuerzos y lee los ítems una sola vez (CA-35.4 y CA-35.5)
    Dado que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Y que el sprint "Sprint 1" tiene los ítems:
      | titulo | sp | horas | estado    | esfuerzo |
      | A      | 3  | 4     | TERMINADO | 5        |
    Y que cierro el sprint "Sprint 1"
    Y que existe un sprint abierto "Sprint 2" en el proyecto "Sistema de ventas"
    Y que el sprint "Sprint 2" tiene los ítems:
      | titulo | sp | horas | estado    | esfuerzo |
      | B      | 5  | 6     | TERMINADO | 4        |
    Y que cierro el sprint "Sprint 2"
    Y que existe un sprint abierto "Sprint 3" en el proyecto "Sistema de ventas"
    Y que el sprint "Sprint 3" tiene los ítems:
      | titulo | sp | horas | estado   | esfuerzo |
      | C      | 2  | 10    | EN_CURSO | 2        |
    Cuando consulto las métricas del proyecto "Sistema de ventas"
    Entonces las horas estimadas son 20 y las reales son 11
    Y la desviación es -9 horas y -45 %
    Y la velocidad es 4 con 2 sprints cerrados
    Y los ítems del backlog se leyeron 2 veces en la consulta
    Y cada métrica es igual a la de su operación separada

  Escenario: Los esfuerzos y los defectos de otro proyecto no se incluyen (CA-35.5)
    Dado que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Y que el sprint "Sprint 1" tiene los ítems:
      | titulo | sp | horas | estado    | esfuerzo |
      | A      | 5  | 10    | TERMINADO | 12       |
    Y que hay un defecto detectado en el sprint "Sprint 1" y resuelto en el sprint "Sprint 1"
    Y que existe un proyecto con nombre "Sistema de compras"
    Y que existe un sprint abierto "Sprint X" en el proyecto "Sistema de compras"
    Y que el sprint "Sprint X" tiene los ítems:
      | titulo | sp | horas | estado    | esfuerzo |
      | Z      | 8  | 20    | TERMINADO | 7        |
    Y que hay un defecto abierto detectado en el sprint "Sprint X"
    Cuando consulto las métricas del sprint "Sprint 1"
    Entonces las horas estimadas son 10 y las reales son 12
    Y los defectos detectados son 1 y los resueltos son 1
    Cuando consulto las métricas del proyecto "Sistema de ventas"
    Entonces los story points planificados son 5 y los completados son 5
    Y las horas estimadas son 10 y las reales son 12
    Y los defectos detectados son 1 y los resueltos son 1

  Escenario: Un defecto detectado en un sprint y resuelto en otro (CA-35.5)
    Dado que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Y que el sprint "Sprint 1" tiene los ítems:
      | titulo | sp | horas | estado    | esfuerzo |
      | A      | 5  | 10    | TERMINADO | 12       |
    Y que cierro el sprint "Sprint 1"
    Y que existe un sprint abierto "Sprint 2" en el proyecto "Sistema de ventas"
    Y que el sprint "Sprint 2" tiene los ítems:
      | titulo | sp | horas | estado    | esfuerzo |
      | B      | 3  | 6     | PENDIENTE | 0        |
    Y que hay un defecto detectado en el sprint "Sprint 1" y resuelto en el sprint "Sprint 2"
    Cuando consulto las métricas del sprint "Sprint 1"
    Entonces los defectos detectados son 1 y los resueltos son 0
    Cuando consulto las métricas del sprint "Sprint 2"
    Entonces los defectos detectados son 0 y los resueltos son 1
    Cuando consulto las métricas del proyecto "Sistema de ventas"
    Entonces los defectos detectados son 1 y los resueltos son 1

  Escenario: Horas con decimales dan una desviación de cero sin redondear en el dominio (CA-35.5)
    Dado que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Y que el sprint "Sprint 1" tiene los ítems:
      | titulo | sp | horas | estado    | esfuerzo |
      | A      | 0  | 0.1   | PENDIENTE | 0.3      |
      | B      | 0  | 0.2   | PENDIENTE | 0        |
    Cuando consulto las métricas del sprint "Sprint 1"
    Entonces las horas estimadas son 0.3 y las reales son 0.3
    Y la desviación es 0 horas y 0 %
    Y cada métrica es igual a la de su operación separada

  Escenario: La consulta única da lo mismo que las operaciones separadas (CA-35.5)
    Dado que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Y que el sprint "Sprint 1" tiene los ítems:
      | titulo | sp | horas | estado    | esfuerzo |
      | A      | 5  | 10    | TERMINADO | 12       |
      | B      | 3  | 6     | EN_CURSO  | 2.5      |
    Y que hay un defecto abierto detectado en el sprint "Sprint 1"
    Y que cierro el sprint "Sprint 1"
    Y que existe un sprint abierto "Sprint 2" en el proyecto "Sistema de ventas"
    Y que el sprint "Sprint 2" tiene los ítems:
      | titulo | sp | horas | estado   | esfuerzo |
      | C      | 8  | 0     | EN_CURSO | 1.25     |
    Y que el proyecto "Sistema de ventas" tiene los ítems sin sprint:
      | titulo | sp | horas | estado    |
      | E      | 2  | 3     | PENDIENTE |
    Cuando consulto las métricas del sprint "Sprint 1"
    Entonces cada métrica es igual a la de su operación separada
    Cuando consulto las métricas del sprint "Sprint 2"
    Entonces cada métrica es igual a la de su operación separada
    Cuando consulto las métricas del proyecto "Sistema de ventas"
    Entonces cada métrica es igual a la de su operación separada

  Esquema del escenario: Consultar un sprint inexistente falla (CA-35.3)
    Cuando consulto las métricas del sprint con el ID <id>
    Entonces la consulta falla porque el sprint no existe

    Ejemplos:
      | id |
      | 99 |
      | 0  |
      | -1 |

  Esquema del escenario: Consultar un proyecto inexistente falla (CA-35.3)
    Cuando consulto las métricas del proyecto con el ID <id>
    Entonces la consulta falla porque el proyecto no existe

    Ejemplos:
      | id |
      | 99 |
      | 0  |
      | -1 |

  Esquema del escenario: Si falla la lectura del sprint o de los ítems no se leen esfuerzos ni defectos (CA-35.3)
    Dado que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Y que el sprint "Sprint 1" tiene los ítems:
      | titulo | sp | horas | estado    | esfuerzo |
      | A      | 5  | 10    | TERMINADO | 12       |
    Y que hay un defecto abierto detectado en el sprint "Sprint 1"
    Y que falla la lectura de <fuente>
    Cuando consulto las métricas del <nivel> "<nombre>"
    Entonces la consulta falla con el error de la lectura de <fuente>
    Y no se leyeron los esfuerzos ni los defectos

    Ejemplos:
      | nivel    | nombre            | fuente     |
      | sprint   | Sprint 1          | el sprint  |
      | sprint   | Sprint 1          | los ítems  |
      | proyecto | Sistema de ventas | los ítems  |
      | proyecto | Sistema de ventas | los sprints |

  Esquema del escenario: Si falla la lectura de los esfuerzos no se leen los defectos (CA-35.3)
    Dado que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Y que el sprint "Sprint 1" tiene los ítems:
      | titulo | sp | horas | estado    | esfuerzo |
      | A      | 5  | 10    | TERMINADO | 12       |
    Y que hay un defecto abierto detectado en el sprint "Sprint 1"
    Y que falla la lectura de los esfuerzos
    Cuando consulto las métricas del <nivel> "<nombre>"
    Entonces la consulta falla con el error de la lectura de los esfuerzos
    Y no se leyeron los defectos

    Ejemplos:
      | nivel    | nombre            |
      | sprint   | Sprint 1          |
      | proyecto | Sistema de ventas |

  Esquema del escenario: Si falla la lectura de los defectos la consulta falla (CA-35.3)
    Dado que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Y que el sprint "Sprint 1" tiene los ítems:
      | titulo | sp | horas | estado    | esfuerzo |
      | A      | 5  | 10    | TERMINADO | 12       |
    Y que hay un defecto abierto detectado en el sprint "Sprint 1"
    Y que falla la lectura de los defectos
    Cuando consulto las métricas del <nivel> "<nombre>"
    Entonces la consulta falla con el error de la lectura de los defectos

    Ejemplos:
      | nivel    | nombre            |
      | sprint   | Sprint 1          |
      | proyecto | Sistema de ventas |

  Esquema del escenario: Los ítems del backlog se leen una sola vez por consulta (CA-35.4)
    # 2 lecturas del repositorio de ítems: 1 de metrics y 1 interna de
    # defect.Manager.ListByProject (límite conocido de la SDD).
    Dado que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Y que el sprint "Sprint 1" tiene los ítems:
      | titulo | sp | horas | estado    | esfuerzo |
      | A      | 5  | 10    | TERMINADO | 12       |
      | B      | 3  | 6     | EN_CURSO  | 5        |
    Y que hay un defecto abierto detectado en el sprint "Sprint 1"
    Cuando consulto las métricas del <nivel> "<nombre>"
    Entonces los ítems del backlog se leyeron 2 veces en la consulta

    Ejemplos:
      | nivel    | nombre            |
      | sprint   | Sprint 1          |
      | proyecto | Sistema de ventas |

  Escenario: Consultar no modifica nada y devuelve una copia (CA-35.7)
    Dado que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Y que el sprint "Sprint 1" tiene los ítems:
      | titulo | sp | horas | estado    | esfuerzo |
      | A      | 5  | 10    | TERMINADO | 12       |
      | B      | 3  | 6     | EN_CURSO  | 5        |
    Y que hay un defecto abierto detectado en el sprint "Sprint 1"
    Cuando consulto las métricas del sprint "Sprint 1"
    Entonces modificar el resultado no cambia la próxima consulta
    Cuando consulto las métricas del proyecto "Sistema de ventas"
    Entonces modificar el resultado no cambia la próxima consulta
    Y los datos del proyecto no cambiaron
