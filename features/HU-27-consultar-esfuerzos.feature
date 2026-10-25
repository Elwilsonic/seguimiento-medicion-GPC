# language: es
# HU-27 — Consultar los registros de esfuerzo (Issue #45)
# Especificación: docs/especificaciones/HU-27-consultar-esfuerzos.md

Característica: HU-27 Consultar los registros de esfuerzo
  Como Agile Enabler,
  quiero consultar los registros de esfuerzo por historia o por sprint,
  para ver cuánto se trabajó.

  # Caso normal - CA-27.1 y CA-27.2
  Escenario: Listar los esfuerzos de un sprint con su total (CA-27.1 y CA-27.2)
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Y que existe un ítem "Login de usuarios" en el proyecto "Sistema de ventas"
    Y que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Y que el ítem con el ID 1 está asignado al sprint con el ID 1
    Y que existe un esfuerzo de 2.5 horas del integrante "García Santiago" con actividad "Implementar el registro" para el ítem con el ID 1 del sprint con el ID 1
    Y que existe un esfuerzo de 1 horas del integrante "Buttini" con actividad "Revisar el registro" para el ítem con el ID 1 del sprint con el ID 1
    Cuando consulto los esfuerzos del sprint con el ID 1
    Entonces la consulta es exitosa
    Y se listan 2 esfuerzos
    Y los esfuerzos listados tienen IDs 1, 2
    Y el total de horas es 3.5
    Y el primer esfuerzo listado tiene integrante "García Santiago" y actividad "Implementar el registro"

  # Caso normal - CA-27.1 y CA-27.2
  Escenario: Listar los esfuerzos de una historia con su total (CA-27.1 y CA-27.2)
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Y que existe un ítem "Login de usuarios" en el proyecto "Sistema de ventas"
    Y que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Y que el ítem con el ID 1 está asignado al sprint con el ID 1
    Y que existe un esfuerzo de 2.5 horas del integrante "García Santiago" con actividad "Implementar el registro" para el ítem con el ID 1 del sprint con el ID 1
    Y que existe un esfuerzo de 1 horas del integrante "Buttini" con actividad "Revisar el registro" para el ítem con el ID 1 del sprint con el ID 1
    Cuando consulto los esfuerzos del ítem con el ID 1
    Entonces la consulta es exitosa
    Y se listan 2 esfuerzos
    Y los esfuerzos listados tienen IDs 1, 2
    Y el total de horas es 3.5

  # Caso límite - CA-27.3
  Escenario: Sprint existente sin esfuerzos (CA-27.3)
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Y que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Cuando consulto los esfuerzos del sprint con el ID 1
    Entonces la consulta es exitosa
    Y la lista de esfuerzos no es nil
    Y no se devuelve ningún esfuerzo
    Y el total de horas es 0

  # Caso límite - CA-27.3
  Escenario: Historia existente sin esfuerzos (CA-27.3)
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Y que existe un ítem "Login de usuarios" en el proyecto "Sistema de ventas"
    Cuando consulto los esfuerzos del ítem con el ID 1
    Entonces la consulta es exitosa
    Y la lista de esfuerzos no es nil
    Y no se devuelve ningún esfuerzo
    Y el total de horas es 0

  # Caso alternativo - CA-27.5
  Escenario: Aislar los esfuerzos por sprint (CA-27.5)
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Y que existe un ítem "Login de usuarios" en el proyecto "Sistema de ventas"
    Y que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Y que existe un sprint abierto "Sprint 2" en el proyecto "Sistema de ventas"
    Y que existe un esfuerzo de 2.5 horas del integrante "García Santiago" con actividad "Implementar el registro" para el ítem con el ID 1 del sprint con el ID 1
    Y que existe un esfuerzo de 4 horas del integrante "Buttini" con actividad "Preparar el despliegue" para el ítem con el ID 1 del sprint con el ID 2
    Cuando consulto los esfuerzos del sprint con el ID 1
    Entonces la consulta es exitosa
    Y se listan 1 esfuerzos
    Y los esfuerzos listados tienen IDs 1
    Y el total de horas es 2.5

  # Caso alternativo - CA-27.5
  Escenario: Aislar los esfuerzos por historia (CA-27.5)
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Y que existe un ítem "Login de usuarios" en el proyecto "Sistema de ventas"
    Y que existe un ítem "Cierre de sesión" en el proyecto "Sistema de ventas"
    Y que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Y que existe un esfuerzo de 2.5 horas del integrante "García Santiago" con actividad "Implementar el registro" para el ítem con el ID 1 del sprint con el ID 1
    Y que existe un esfuerzo de 4 horas del integrante "Buttini" con actividad "Implementar el cierre" para el ítem con el ID 2 del sprint con el ID 1
    Cuando consulto los esfuerzos del ítem con el ID 1
    Entonces la consulta es exitosa
    Y se listan 1 esfuerzos
    Y los esfuerzos listados tienen IDs 1
    Y el total de horas es 2.5

  # Caso alternativo - CA-27.6
  Escenario: Consultar un sprint cerrado (CA-27.6)
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Y que existe un ítem "Login de usuarios" en el proyecto "Sistema de ventas"
    Y que existe un sprint cerrado "Sprint 1" en el proyecto "Sistema de ventas"
    Y que existe un esfuerzo de 2 horas del integrante "García Santiago" con actividad "Implementar el registro" para el ítem con el ID 1 del sprint con el ID 1
    Cuando consulto los esfuerzos del sprint con el ID 1
    Entonces la consulta es exitosa
    Y se listan 1 esfuerzos
    Y el total de horas es 2

  # Caso de error - CA-27.4
  Escenario: Rechazar un sprint y una historia inexistentes (CA-27.4)
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Cuando consulto los esfuerzos del sprint con el ID 999
    Entonces la consulta falla porque no existe el sprint
    Y no se devuelve ningún resumen
    Cuando consulto los esfuerzos del ítem con el ID 999
    Entonces la consulta falla porque no existe el ítem
    Y no se devuelve ningún resumen

  # Caso límite - CA-27.7
  Escenario: Modificar lo devuelto no altera lo guardado (CA-27.7)
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Y que existe un ítem "Login de usuarios" en el proyecto "Sistema de ventas"
    Y que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Y que existe un esfuerzo de 2.5 horas del integrante "García Santiago" con actividad "Implementar el registro" para el ítem con el ID 1 del sprint con el ID 1
    Cuando consulto los esfuerzos del sprint con el ID 1
    Entonces la consulta es exitosa
    Y la lista de esfuerzos no es nil
    Cuando modifico el primer esfuerzo devuelto
    Y vuelvo a consultar los esfuerzos del sprint con el ID 1
    Entonces la consulta es exitosa
    Y los esfuerzos guardados no cambian

  # Caso de error - CA-27.9
  Escenario: Error del repositorio al listar (CA-27.9)
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Y que existe un ítem "Login de usuarios" en el proyecto "Sistema de ventas"
    Y que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Y que el repositorio de esfuerzos falla al listar
    Cuando consulto los esfuerzos del ítem con el ID 1
    Entonces la consulta falla con el error del repositorio
    Y no se devuelve ningún resumen

  # Caso límite - CA-27.8
  Escenario: Listados simultáneos con registros (CA-27.8)
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Y que existe un ítem "Login de usuarios" en el proyecto "Sistema de ventas"
    Y que existe un sprint abierto "Sprint 1" en el proyecto "Sistema de ventas"
    Y que el ítem con el ID 1 está asignado al sprint con el ID 1
    Cuando registro 10 esfuerzos y consulto en simultáneo para el ítem con el ID 1 y el sprint con el ID 1
    Entonces todos los registros y listados son exitosos
    Y quedan 10 esfuerzos guardados
    Y el total final es 10
