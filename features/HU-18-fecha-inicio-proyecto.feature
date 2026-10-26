# language: es
# HU-18 · Registrar la fecha de inicio del proyecto (Issue #36)
# Especificación: docs/especificaciones/HU-18-fecha-inicio-proyecto.md

Característica: HU-18 Registrar la fecha de inicio del proyecto
  Como Agile Enabler,
  quiero registrar la fecha de inicio de un proyecto,
  para saber desde cuándo se cuenta su duración.

  # Caso normal (happy path)
  Escenario: Registrar la fecha de inicio de un proyecto sin fecha (CA-18.1)
    Dado que existe un proyecto con nombre "Sistema de ventas" y descripción "Gestión de ventas"
    Cuando registro la fecha de inicio "2026-10-23" en el proyecto con el ID 1
    Entonces el registro es exitoso
    Y el proyecto devuelto tiene la fecha de inicio "2026-10-23"
    Y el proyecto guardado con el ID 1 tiene la fecha de inicio "2026-10-23"
    Y el proyecto guardado con el ID 1 sigue teniendo el nombre "Sistema de ventas", la descripción "Gestión de ventas" y el estado "PLANIFICADO"

  # Caso alternativo
  Escenario: Registrar la fecha en un proyecto finalizado (CA-18.1)
    Dado que existe un proyecto con nombre "Proyecto A" en el estado "FINALIZADO"
    Cuando registro la fecha de inicio "2026-01-15" en el proyecto con el ID 1
    Entonces el registro es exitoso
    Y el proyecto guardado con el ID 1 tiene la fecha de inicio "2026-01-15"
    Y el proyecto guardado con el ID 1 sigue teniendo el nombre "Proyecto A", la descripción "" y el estado "FINALIZADO"

  # Caso límite
  Escenario: Registrar una fecha futura (CA-18.1)
    Dado que existe un proyecto con nombre "Proyecto A" y descripción ""
    Cuando registro la fecha de inicio "2030-12-31" en el proyecto con el ID 1
    Entonces el registro es exitoso
    Y el proyecto guardado con el ID 1 tiene la fecha de inicio "2030-12-31"

  # Caso alternativo
  Escenario: Volver a registrar la fecha (CA-18.5)
    Dado que existe un proyecto con nombre "Proyecto A" y descripción ""
    Y el proyecto con el ID 1 tiene registrada la fecha de inicio "2026-10-23"
    Cuando registro la fecha de inicio "2026-11-02" en el proyecto con el ID 1
    Entonces el registro es exitoso
    Y el proyecto guardado con el ID 1 tiene la fecha de inicio "2026-11-02"

  # Caso límite
  Escenario: Un proyecto recién creado no tiene fecha de inicio (CA-18.6)
    Dado que existe un proyecto con nombre "Proyecto A" y descripción ""
    Entonces el proyecto guardado con el ID 1 no tiene fecha de inicio

  # Caso alternativo
  Escenario: Modificar el proyecto conserva la fecha de inicio (CA-18.6)
    Dado que existe un proyecto con nombre "Proyecto A" y descripción ""
    Y el proyecto con el ID 1 tiene registrada la fecha de inicio "2026-10-23"
    Cuando modifico el proyecto con el ID 1 con nombre "Proyecto B", descripción "Nueva" y estado "EN_CURSO"
    Entonces el proyecto guardado con el ID 1 tiene la fecha de inicio "2026-10-23"
    Y el proyecto guardado con el ID 1 sigue teniendo el nombre "Proyecto B", la descripción "Nueva" y el estado "EN_CURSO"

  # Caso de error
  Escenario: Rechazar una fecha ausente (CA-18.3)
    Dado que existe un proyecto con nombre "Proyecto A" y descripción ""
    Y el proyecto con el ID 1 tiene registrada la fecha de inicio "2026-10-23"
    Cuando registro una fecha de inicio ausente en el proyecto con el ID 1
    Entonces el registro falla porque la fecha de inicio es obligatoria
    Y el proyecto guardado con el ID 1 tiene la fecha de inicio "2026-10-23"

  # Caso de error
  Esquema del escenario: Rechazar un proyecto inexistente (CA-18.4)
    Dado que existe un proyecto con nombre "Proyecto A" y descripción ""
    Cuando registro la fecha de inicio "2026-10-23" en el proyecto con el ID <id>
    Entonces el registro falla porque el proyecto no existe
    Y el proyecto guardado con el ID 1 no tiene fecha de inicio

    Ejemplos:
      | id  |
      | 999 |
      | 0   |
      | -1  |

  # Caso límite
  Escenario: Modificar el proyecto devuelto no altera el guardado (CA-18.7)
    Dado que existe un proyecto con nombre "Proyecto A" y descripción ""
    Cuando registro la fecha de inicio "2026-10-23" en el proyecto con el ID 1
    Y cambio la fecha de inicio del proyecto devuelto a "2020-01-01"
    Entonces el proyecto guardado con el ID 1 tiene la fecha de inicio "2026-10-23"

  # Caso de error
  Escenario: Error del repositorio al guardar (CA-18.8)
    Dado que existe un proyecto con nombre "Proyecto A" y descripción ""
    Y el repositorio falla al actualizar
    Cuando registro la fecha de inicio "2026-10-23" en el proyecto con el ID 1
    Entonces el registro falla con el error del repositorio
    Y el proyecto guardado con el ID 1 no tiene fecha de inicio
