# language: es
# HU-36 · Ver el resumen del estado del proyecto (Issue #54)
# Especificación: docs/especificaciones/HU-36-resumen-proyecto.md
# CA-36.6 (API e interfaz) se cubre con tests de Go, no con escenarios.

Característica: HU-36 Ver el resumen del estado del proyecto
  Como Agile Enabler,
  quiero ver un resumen del estado del proyecto,
  para tener una visión rápida del avance.

  Antecedentes:
    Dado que existe un proyecto con nombre "Sistema de ventas"

  # Caso normal
  Escenario: Resumen con ítems en los tres estados y un sprint abierto (CA-36.1)
    Dado que el proyecto "Sistema de ventas" está en estado "EN_CURSO"
    Y que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Y que el proyecto "Sistema de ventas" tiene los ítems:
      | titulo | estado    | sprint   |
      | A      | PENDIENTE | Sprint 1 |
      | B      | PENDIENTE |          |
      | C      | EN_CURSO  | Sprint 1 |
      | D      | TERMINADO | Sprint 1 |
      | E      | TERMINADO | Sprint 1 |
      | F      | TERMINADO |          |
    Cuando consulto el resumen del proyecto "Sistema de ventas"
    Entonces el resumen muestra el proyecto "Sistema de ventas" en estado "EN_CURSO"
    Y el resumen muestra 2 ítems pendientes, 1 en curso y 3 terminados
    Y el sprint actual del resumen es "Sprint 1"

  # Caso alternativo - un sprint cerrado y otro abierto
  Escenario: El sprint actual es el abierto aunque haya sprints cerrados (CA-36.1)
    Dado que existe un sprint cerrado "Sprint 1" en el proyecto "Sistema de ventas"
    Y que existe un sprint abierto "Sprint 2" en el proyecto "Sistema de ventas"
    Cuando consulto el resumen del proyecto "Sistema de ventas"
    Entonces el sprint actual del resumen es "Sprint 2"

  # Caso alternativo - no mezcla proyectos
  Escenario: El resumen no cuenta ítems ni sprints de otro proyecto (CA-36.1)
    Dado que existe un proyecto con nombre "Inventario"
    Y que existe un sprint abierto "Sprint de inventario" en el proyecto "Inventario"
    Y que el proyecto "Inventario" tiene los ítems:
      | titulo | estado    | sprint               |
      | X      | TERMINADO | Sprint de inventario |
    Y que el proyecto "Sistema de ventas" tiene los ítems:
      | titulo | estado    | sprint |
      | A      | PENDIENTE |        |
    Cuando consulto el resumen del proyecto "Sistema de ventas"
    Entonces el resumen muestra 1 ítems pendientes, 0 en curso y 0 terminados
    Y el resumen no tiene sprint actual

  # Caso límite - proyecto sin datos
  Escenario: Un proyecto sin ítems ni sprints muestra el resumen vacío (CA-36.2)
    Cuando consulto el resumen del proyecto "Sistema de ventas"
    Entonces el resumen muestra el proyecto "Sistema de ventas" en estado "PLANIFICADO"
    Y el resumen muestra 0 ítems pendientes, 0 en curso y 0 terminados
    Y el resumen no tiene sprint actual

  # Caso límite - ítems en un solo estado
  Escenario: Los estados sin ítems se muestran en cero (CA-36.1)
    Dado que el proyecto "Sistema de ventas" tiene los ítems:
      | titulo | estado   | sprint |
      | A      | EN_CURSO |        |
      | B      | EN_CURSO |        |
    Cuando consulto el resumen del proyecto "Sistema de ventas"
    Entonces el resumen muestra 0 ítems pendientes, 2 en curso y 0 terminados

  # Caso límite - solo sprints cerrados
  Escenario: Un proyecto con solo sprints cerrados no tiene sprint actual (CA-36.3)
    Dado que existe un sprint cerrado "Sprint 1" en el proyecto "Sistema de ventas"
    Cuando consulto el resumen del proyecto "Sistema de ventas"
    Entonces el resumen muestra el proyecto "Sistema de ventas" en estado "PLANIFICADO"
    Y el resumen no tiene sprint actual

  # Caso alternativo - los datos coinciden con los guardados
  Escenario: El resumen refleja el cambio de estado de un ítem (CA-36.4)
    Dado que el proyecto "Sistema de ventas" tiene los ítems:
      | titulo | estado    | sprint |
      | A      | EN_CURSO  |        |
      | B      | PENDIENTE |        |
    Y que consulto el resumen del proyecto "Sistema de ventas"
    Y que el resumen muestra 1 ítems pendientes, 1 en curso y 0 terminados
    Cuando el ítem "A" pasa al estado "TERMINADO"
    Y consulto el resumen del proyecto "Sistema de ventas"
    Entonces el resumen muestra 1 ítems pendientes, 0 en curso y 1 terminados

  Escenario: El resumen refleja el cierre del sprint actual (CA-36.4)
    Dado que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Y que consulto el resumen del proyecto "Sistema de ventas"
    Y que el sprint actual del resumen es "Sprint 1"
    Cuando se cierra el sprint "Sprint 1"
    Y consulto el resumen del proyecto "Sistema de ventas"
    Entonces el resumen no tiene sprint actual

  # Caso de error - proyecto inexistente
  Esquema del escenario: Consultar el resumen de un proyecto inexistente devuelve error (CA-36.5)
    Cuando consulto el resumen del proyecto con el ID <id>
    Entonces la consulta falla porque el proyecto no existe

    Ejemplos:
      | id |
      | 99 |
      | 0  |
      | -1 |

  # Caso alternativo - solo lectura
  Escenario: Consultar el resumen no modifica nada (CA-36.5)
    Dado que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Y que el proyecto "Sistema de ventas" tiene los ítems:
      | titulo | estado   | sprint   |
      | A      | EN_CURSO | Sprint 1 |
    Cuando consulto el resumen del proyecto "Sistema de ventas"
    Y consulto el resumen del proyecto "Sistema de ventas"
    Entonces el resumen muestra 0 ítems pendientes, 1 en curso y 0 terminados
    Y el proyecto, los ítems y los sprints no cambiaron
