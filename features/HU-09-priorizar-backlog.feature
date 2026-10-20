# language: es
# HU-09 · Priorizar los ítems del backlog (Issue #22)
# Especificación: docs/especificaciones/HU-09-priorizar-backlog.md

Característica: HU-09 Priorizar los ítems del backlog
  Como Agile Enabler,
  quiero ordenar los ítems del backlog por prioridad,
  para trabajar primero lo más importante.

  # Caso normal (happy path) - RN-09.1
  Escenario: Listar por prioridad (CA-09.1)
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Y creo el ítem "Exportar reportes" en el backlog de "Sistema de ventas" con prioridad "BAJA" y criterio "PDF generado"
    Y creo el ítem "Recuperar contraseña" en el backlog de "Sistema de ventas" con prioridad "MEDIA" y criterio "Mail enviado"
    Y creo el ítem "Login de usuarios" en el backlog de "Sistema de ventas" con prioridad "ALTA" y criterio "Login OK"
    Cuando listo por prioridad el backlog del proyecto "Sistema de ventas"
    Entonces obtengo 3 ítems ordenados por prioridad
    Y obtengo los ítems ordenados por prioridad con los IDs 3, 2 y 1
    Y el ítem en la posición 1 del listado por prioridad tiene el título "Login de usuarios", la prioridad "ALTA" y el criterio "Login OK"
    Y el ítem en la posición 2 del listado por prioridad tiene el título "Recuperar contraseña", la prioridad "MEDIA" y el criterio "Mail enviado"
    Y el ítem en la posición 3 del listado por prioridad tiene el título "Exportar reportes", la prioridad "BAJA" y el criterio "PDF generado"

  # Caso límite - RN-09.2 (24 ítems: más de 12, para que un orden no estable se note)
  Escenario: Los ítems con la misma prioridad conservan el orden por ID (CA-09.2)
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Y creo 24 ítems en el backlog de "Sistema de ventas" con las prioridades en ciclo "BAJA", "ALTA" y "MEDIA"
    Cuando listo por prioridad el backlog del proyecto "Sistema de ventas"
    Entonces obtengo 24 ítems ordenados por prioridad
    Y obtengo los ítems ordenados por prioridad con los IDs 2, 5, 8, 11, 14, 17, 20, 23, 3, 6, 9, 12, 15, 18, 21, 24, 1, 4, 7, 10, 13, 16, 19 y 22

  # Caso límite - RN-09.7
  Escenario: Listar por prioridad un backlog vacío (CA-09.3)
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Cuando listo por prioridad el backlog del proyecto "Sistema de ventas"
    Entonces obtengo 0 ítems ordenados por prioridad
    Y el listado por prioridad es una lista vacía

  # Caso alternativo - RN-09.4 (estados, proyecto finalizado y aislamiento por proyecto)
  Escenario: El listado por prioridad tiene los mismos ítems que el listado de HU-06 (CA-09.4)
    Dado que existe un proyecto finalizado con nombre "Proyecto cerrado"
    Y que existe un proyecto con nombre "Portal de clientes"
    Y creo el ítem "Entrega final" en el backlog de "Proyecto cerrado" con prioridad "MEDIA" y criterio "Entregado"
    Y creo el ítem "Login de usuarios" en el backlog de "Portal de clientes" con prioridad "ALTA" y criterio "Login OK"
    Y creo el ítem "Manual de usuario" en el backlog de "Proyecto cerrado" con prioridad "BAJA" y criterio "Manual publicado"
    Y creo el ítem "Corregir defectos" en el backlog de "Proyecto cerrado" con prioridad "ALTA" y criterio "Sin defectos abiertos"
    Y el ítem con el ID 3 pasa al estado "EN_CURSO"
    Y el ítem con el ID 4 pasa al estado "EN_CURSO"
    Y el ítem con el ID 4 pasa al estado "TERMINADO"
    Cuando listo por prioridad el backlog del proyecto "Proyecto cerrado"
    Entonces obtengo los ítems ordenados por prioridad con los IDs 4, 1 y 3
    Y el listado por prioridad tiene los mismos ítems que el listado de HU-06 del proyecto "Proyecto cerrado"

  # Caso alternativo - RN-09.5
  Escenario: El listado de HU-06 sigue ordenado por ID después de listar por prioridad (CA-09.5)
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Y creo el ítem "Exportar reportes" en el backlog de "Sistema de ventas" con prioridad "BAJA" y criterio "PDF generado"
    Y creo el ítem "Login de usuarios" en el backlog de "Sistema de ventas" con prioridad "ALTA" y criterio "Login OK"
    Cuando listo por prioridad el backlog del proyecto "Sistema de ventas"
    Y listo el backlog de HU-06 del proyecto "Sistema de ventas"
    Entonces obtengo los ítems ordenados por prioridad con los IDs 2 y 1
    Y el listado de HU-06 tiene los ítems con los IDs 1 y 2

  # Caso límite - RN-09.3 (ítems guardados sin pasar por el Manager)
  Escenario: Un ítem con una prioridad desconocida va al final (CA-09.6)
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Y guardo directo en el repositorio el ítem "Migrar datos" del proyecto "Sistema de ventas" con prioridad "URGENTE"
    Y creo el ítem "Exportar reportes" en el backlog de "Sistema de ventas" con prioridad "BAJA" y criterio "PDF generado"
    Y guardo directo en el repositorio el ítem "Auditoría" del proyecto "Sistema de ventas" con prioridad "alta"
    Y creo el ítem "Login de usuarios" en el backlog de "Sistema de ventas" con prioridad "ALTA" y criterio "Login OK"
    Y creo el ítem "Recuperar contraseña" en el backlog de "Sistema de ventas" con prioridad "MEDIA" y criterio "Mail enviado"
    Cuando listo por prioridad el backlog del proyecto "Sistema de ventas"
    Entonces obtengo 5 ítems ordenados por prioridad
    Y obtengo los ítems ordenados por prioridad con los IDs 4, 5, 2, 1 y 3

  # Caso de error - RN-09.6
  Esquema del escenario: Listar por prioridad el backlog de un proyecto inexistente (CA-09.7)
    Dado que hay <proyectos> proyectos creados
    Cuando listo por prioridad el backlog del proyecto con ID <id>
    Entonces el listado por prioridad falla porque el proyecto no existe

    Ejemplos:
      | proyectos | id  |
      | 1         | 999 |
      | 1         | 0   |
      | 1         | -1  |
      | 0         | 1   |

  # Caso de error
  Escenario: Si el repositorio falla al listar se devuelve ese error (CA-09.8)
    Dado que existe un proyecto con nombre "Sistema de ventas" y el repositorio de ítems falla al listar
    Cuando listo por prioridad el backlog del proyecto "Sistema de ventas"
    Entonces el listado por prioridad falla por un error del repositorio

  # Caso límite - concurrencia, verificada con go test -race en el CI
  Escenario: Listar por prioridad mientras se crean ítems en simultáneo (CA-09.9)
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Cuando se crean 50 ítems en simultáneo en el proyecto "Sistema de ventas" mientras se lista por prioridad su backlog en simultáneo
    Entonces ningún listado por prioridad simultáneo devuelve un error
    Y al volver a listar por prioridad el backlog de "Sistema de ventas" obtengo 50 ítems

  # Caso límite - RN-09.8
  Escenario: Modificar el listado por prioridad no altera los ítems guardados (CA-09.10)
    Dado que existe un proyecto con nombre "Sistema de ventas"
    Y creo el ítem "Exportar reportes" en el backlog de "Sistema de ventas" con prioridad "BAJA" y criterio "PDF generado"
    Y creo el ítem "Login de usuarios" en el backlog de "Sistema de ventas" con prioridad "ALTA" y criterio "Login OK"
    Cuando listo por prioridad el backlog del proyecto "Sistema de ventas"
    Y modifico el primer ítem del listado por prioridad, sus criterios y el orden de la lista
    Y vuelvo a listar por prioridad el backlog del proyecto "Sistema de ventas"
    Entonces obtengo los ítems ordenados por prioridad con los IDs 2 y 1
    Y el ítem en la posición 1 del listado por prioridad tiene el título "Login de usuarios", la prioridad "ALTA" y el criterio "Login OK"
