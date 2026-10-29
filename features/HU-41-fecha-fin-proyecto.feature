# language: es
# HU-41 · Registrar la fecha de fin del proyecto (Issue #115)
# Especificación: docs/especificaciones/HU-41-fecha-fin-proyecto.md
# CA-41.3 (formato), CA-41.12 (PostgreSQL) y CA-41.13 (API e interfaz web) se cubren fuera de godog.

Característica: HU-41 Registrar la fecha de fin del proyecto
  Como Agile Enabler,
  quiero registrar la fecha de fin de un proyecto,
  para delimitar su duración.

  # Caso normal (happy path)
  Escenario: Registrar la fecha de fin de un proyecto con fecha de inicio (CA-41.1)
    Dado que existe un proyecto con nombre "Sistema de ventas" y descripción "Gestión de ventas"
    Y el proyecto con el ID 1 tiene registrada la fecha de inicio "2026-10-23"
    Cuando registro la fecha de fin "2026-12-18" en el proyecto con el ID 1
    Entonces el registro es exitoso
    Y el proyecto devuelto tiene la fecha de fin "2026-12-18"
    Y el proyecto guardado con el ID 1 tiene la fecha de fin "2026-12-18"
    Y el proyecto guardado con el ID 1 tiene la fecha de inicio "2026-10-23"
    Y el proyecto guardado con el ID 1 sigue teniendo el nombre "Sistema de ventas", la descripción "Gestión de ventas" y el estado "PLANIFICADO"

  # Caso alternativo
  Escenario: Registrar la fecha de fin en un proyecto finalizado (CA-41.1)
    Dado que existe un proyecto con nombre "Proyecto A" en el estado "FINALIZADO"
    Y el proyecto con el ID 1 tiene registrada la fecha de inicio "2026-01-15"
    Cuando registro la fecha de fin "2026-06-30" en el proyecto con el ID 1
    Entonces el registro es exitoso
    Y el proyecto guardado con el ID 1 tiene la fecha de fin "2026-06-30"
    Y el proyecto guardado con el ID 1 sigue teniendo el nombre "Proyecto A", la descripción "" y el estado "FINALIZADO"

  # Caso límite y de error
  Esquema del escenario: Rechazar una fecha de fin igual o anterior a la de inicio (CA-41.2)
    Dado que existe un proyecto con nombre "Proyecto A" y descripción ""
    Y el proyecto con el ID 1 tiene registrada la fecha de inicio "2026-10-23"
    Cuando registro la fecha de fin "<fin>" en el proyecto con el ID 1
    Entonces el registro falla porque la fecha de fin debe ser posterior a la de inicio
    Y el proyecto guardado con el ID 1 no tiene fecha de fin

    Ejemplos:
      | fin        |
      | 2026-10-23 |
      | 2026-10-22 |
      | 2025-12-31 |

  # Caso de error
  Escenario: Rechazar una fecha de fin ausente (CA-41.4)
    Dado que existe un proyecto con nombre "Proyecto A" y descripción ""
    Y el proyecto con el ID 1 tiene registrada la fecha de inicio "2026-10-23"
    Cuando registro una fecha de fin ausente en el proyecto con el ID 1
    Entonces el registro falla porque la fecha de fin es obligatoria
    Y el proyecto guardado con el ID 1 no tiene fecha de fin

  # Caso de error
  Esquema del escenario: Rechazar un proyecto inexistente (CA-41.5)
    Dado que existe un proyecto con nombre "Proyecto A" y descripción ""
    Cuando registro la fecha de fin "2026-12-18" en el proyecto con el ID <id>
    Entonces el registro falla porque el proyecto no existe

    Ejemplos:
      | id  |
      | 999 |
      | 0   |
      | -1  |

  # Caso de error
  Escenario: Rechazar la fecha de fin de un proyecto sin fecha de inicio (CA-41.6)
    Dado que existe un proyecto con nombre "Proyecto A" y descripción ""
    Cuando registro la fecha de fin "2026-12-18" en el proyecto con el ID 1
    Entonces el registro falla porque el proyecto no tiene fecha de inicio
    Y el proyecto guardado con el ID 1 no tiene fecha de fin

  # Caso alternativo
  Escenario: Volver a registrar la fecha de fin (CA-41.7)
    Dado que existe un proyecto con nombre "Proyecto A" y descripción ""
    Y el proyecto con el ID 1 tiene registrada la fecha de inicio "2026-10-23"
    Y el proyecto con el ID 1 tiene registrada la fecha de fin "2026-12-18"
    Cuando registro la fecha de fin "2027-01-29" en el proyecto con el ID 1
    Entonces el registro es exitoso
    Y el proyecto guardado con el ID 1 tiene la fecha de fin "2027-01-29"

  # Caso de error
  Esquema del escenario: Rechazar una fecha de inicio igual o posterior a la de fin registrada (CA-41.8)
    Dado que existe un proyecto con nombre "Proyecto A" y descripción ""
    Y el proyecto con el ID 1 tiene registrada la fecha de inicio "2026-10-23"
    Y el proyecto con el ID 1 tiene registrada la fecha de fin "2026-12-18"
    Cuando registro la fecha de inicio "<inicio>" en el proyecto con el ID 1
    Entonces el registro falla porque la fecha de inicio debe ser anterior a la de fin
    Y el proyecto guardado con el ID 1 tiene la fecha de inicio "2026-10-23"
    Y el proyecto guardado con el ID 1 tiene la fecha de fin "2026-12-18"

    Ejemplos:
      | inicio     |
      | 2026-12-18 |
      | 2026-12-19 |

  # Caso alternativo
  Escenario: Aceptar una fecha de inicio anterior a la de fin registrada (CA-41.8)
    Dado que existe un proyecto con nombre "Proyecto A" y descripción ""
    Y el proyecto con el ID 1 tiene registrada la fecha de inicio "2026-10-23"
    Y el proyecto con el ID 1 tiene registrada la fecha de fin "2026-12-18"
    Cuando registro la fecha de inicio "2026-12-17" en el proyecto con el ID 1
    Entonces el registro es exitoso
    Y el proyecto guardado con el ID 1 tiene la fecha de inicio "2026-12-17"
    Y el proyecto guardado con el ID 1 tiene la fecha de fin "2026-12-18"

  # Caso límite
  Escenario: Un proyecto recién creado no tiene fecha de fin (CA-41.9)
    Dado que existe un proyecto con nombre "Proyecto A" y descripción ""
    Entonces el proyecto guardado con el ID 1 no tiene fecha de fin

  # Caso alternativo
  Escenario: Modificar el proyecto o su fecha de inicio conserva la fecha de fin (CA-41.9)
    Dado que existe un proyecto con nombre "Proyecto A" y descripción ""
    Y el proyecto con el ID 1 tiene registrada la fecha de inicio "2026-10-23"
    Y el proyecto con el ID 1 tiene registrada la fecha de fin "2026-12-18"
    Cuando modifico el proyecto con el ID 1 con nombre "Proyecto B", descripción "Nueva" y estado "EN_CURSO"
    Y registro la fecha de inicio "2026-11-02" en el proyecto con el ID 1
    Entonces el proyecto guardado con el ID 1 tiene la fecha de fin "2026-12-18"
    Y el proyecto guardado con el ID 1 sigue teniendo el nombre "Proyecto B", la descripción "Nueva" y el estado "EN_CURSO"

  # Caso límite
  Escenario: Modificar el proyecto devuelto no altera el guardado (CA-41.10)
    Dado que existe un proyecto con nombre "Proyecto A" y descripción ""
    Y el proyecto con el ID 1 tiene registrada la fecha de inicio "2026-10-23"
    Cuando registro la fecha de fin "2026-12-18" en el proyecto con el ID 1
    Y cambio la fecha de fin del proyecto devuelto a "2030-01-01"
    Entonces el proyecto guardado con el ID 1 tiene la fecha de fin "2026-12-18"

  # Caso de error
  Escenario: Error del repositorio (CA-41.11)
    Dado que existe un proyecto con nombre "Proyecto A" y descripción ""
    Y el proyecto con el ID 1 tiene registrada la fecha de inicio "2026-10-23"
    Y el repositorio falla al actualizar
    Cuando registro la fecha de fin "2026-12-18" en el proyecto con el ID 1
    Entonces el registro falla con el error del repositorio
    Y el proyecto guardado con el ID 1 no tiene fecha de fin

  # Caso límite
  Escenario: Registros de fechas simultáneos mantienen el orden de las fechas (CA-41.11)
    Dado que existe un proyecto con nombre "Proyecto A" y descripción ""
    Y el proyecto con el ID 1 tiene registrada la fecha de inicio "2026-10-01"
    Y el proyecto con el ID 1 tiene registrada la fecha de fin "2026-12-31"
    Cuando registro en simultáneo 20 veces las fechas de inicio "2026-10-05" y "2027-01-15" y las de fin "2026-12-20" y "2026-09-01" en el proyecto con el ID 1
    Entonces cada registro simultáneo termina con el resultado esperado
    Y el proyecto guardado con el ID 1 tiene la fecha de inicio "2026-10-05"
    Y el proyecto guardado con el ID 1 tiene la fecha de fin "2026-12-20"
