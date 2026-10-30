# language: es
# HU-19 · Modificar las fechas del proyecto (Issue #37)
# Especificación: docs/especificaciones/HU-19-modificar-fechas-proyecto.md
# El formato AAAA-MM-DD (CA-19.5 en la API) y CA-19.8 (API e interfaz web) se cubren fuera de godog.

Característica: HU-19 Modificar las fechas del proyecto
  Como Agile Enabler,
  quiero modificar las fechas de un proyecto,
  para reflejar cambios en la planificación.

  # Caso normal (happy path)
  Escenario: Postergar las dos fechas de un proyecto (CA-19.1)
    Dado que existe un proyecto con nombre "Sistema de ventas" y descripción "Gestión de ventas"
    Y el proyecto con el ID 1 tiene registradas la fecha de inicio "2026-10-23" y la de fin "2026-12-18"
    Cuando modifico las fechas del proyecto con el ID 1 a inicio "2027-01-11" y fin "2027-03-19"
    Entonces la modificación es exitosa
    Y el proyecto devuelto tiene la fecha de inicio "2027-01-11" y la de fin "2027-03-19"
    Y el proyecto guardado con el ID 1 tiene la fecha de inicio "2027-01-11" y la de fin "2027-03-19"
    Y el proyecto guardado con el ID 1 sigue teniendo el nombre "Sistema de ventas", la descripción "Gestión de ventas" y el estado "PLANIFICADO"

  # Caso alternativo
  Escenario: Adelantar las dos fechas de un proyecto (CA-19.1)
    Dado que existe un proyecto con nombre "Sistema de ventas" y descripción "Gestión de ventas"
    Y el proyecto con el ID 1 tiene registradas la fecha de inicio "2026-10-23" y la de fin "2026-12-18"
    Cuando modifico las fechas del proyecto con el ID 1 a inicio "2026-06-01" y fin "2026-08-31"
    Entonces la modificación es exitosa
    Y el proyecto guardado con el ID 1 tiene la fecha de inicio "2026-06-01" y la de fin "2026-08-31"
    Y el proyecto guardado con el ID 1 sigue teniendo el nombre "Sistema de ventas", la descripción "Gestión de ventas" y el estado "PLANIFICADO"

  # Caso alternativo
  Escenario: Modificar las fechas de un proyecto finalizado (CA-19.1)
    Dado que existe un proyecto con nombre "Proyecto A" en el estado "FINALIZADO"
    Y el proyecto con el ID 1 tiene registradas la fecha de inicio "2026-01-15" y la de fin "2026-06-30"
    Cuando modifico las fechas del proyecto con el ID 1 a inicio "2026-02-01" y fin "2026-07-31"
    Entonces la modificación es exitosa
    Y el proyecto guardado con el ID 1 tiene la fecha de inicio "2026-02-01" y la de fin "2026-07-31"
    Y el proyecto guardado con el ID 1 sigue teniendo el nombre "Proyecto A", la descripción "" y el estado "FINALIZADO"

  # Caso límite y de error
  Esquema del escenario: Rechazar un fin igual o anterior al inicio nuevo (CA-19.2)
    Dado que existe un proyecto con nombre "Proyecto A" y descripción ""
    Y el proyecto con el ID 1 tiene registradas la fecha de inicio "2026-10-23" y la de fin "2026-12-18"
    Cuando modifico las fechas del proyecto con el ID 1 a inicio "2027-01-11" y fin "<fin>"
    Entonces la modificación falla porque la fecha de fin debe ser posterior a la de inicio
    Y el proyecto guardado con el ID 1 tiene la fecha de inicio "2026-10-23" y la de fin "2026-12-18"

    Ejemplos:
      | fin        |
      | 2027-01-11 |
      | 2027-01-10 |
      | 2026-12-31 |

  # Caso de error
  Esquema del escenario: Rechazar un proyecto inexistente (CA-19.3)
    Dado que existe un proyecto con nombre "Proyecto A" y descripción ""
    Y el proyecto con el ID 1 tiene registradas la fecha de inicio "2026-10-23" y la de fin "2026-12-18"
    Cuando modifico las fechas del proyecto con el ID <id> a inicio "2027-01-11" y fin "2027-03-19"
    Entonces la modificación falla porque el proyecto no existe
    Y el proyecto guardado con el ID 1 tiene la fecha de inicio "2026-10-23" y la de fin "2026-12-18"

    Ejemplos:
      | id  |
      | 999 |
      | 0   |
      | -1  |

  # Caso de error
  Esquema del escenario: Rechazar un proyecto sin las dos fechas (CA-19.4)
    Dado que existe un proyecto con nombre "Proyecto A" y descripción ""
    Y el proyecto con el ID 1 tiene registrada <registradas>
    Cuando modifico las fechas del proyecto con el ID 1 a inicio "2027-01-11" y fin "2027-03-19"
    Entonces la modificación falla porque el proyecto no tiene registradas las dos fechas
    Y el proyecto guardado con el ID 1 no cambió

    Ejemplos:
      | registradas                          |
      | ninguna fecha                        |
      | solo la fecha de inicio "2026-10-23" |

  # Caso de error
  Esquema del escenario: Rechazar una fecha ausente (CA-19.5)
    Dado que existe un proyecto con nombre "Proyecto A" y descripción ""
    Y el proyecto con el ID 1 tiene registradas la fecha de inicio "2026-10-23" y la de fin "2026-12-18"
    Cuando modifico las fechas del proyecto con el ID 1 a inicio "<inicio>" y fin "<fin>"
    Entonces la modificación falla porque la fecha de <ausente> es obligatoria
    Y el proyecto guardado con el ID 1 tiene la fecha de inicio "2026-10-23" y la de fin "2026-12-18"

    Ejemplos:
      | inicio     | fin        | ausente |
      |            | 2027-03-19 | inicio  |
      |            |            | inicio  |
      | 2027-01-11 |            | fin     |

  # Caso de error
  Escenario: Error del repositorio (CA-19.6)
    Dado que existe un proyecto con nombre "Proyecto A" y descripción ""
    Y el proyecto con el ID 1 tiene registradas la fecha de inicio "2026-10-23" y la de fin "2026-12-18"
    Y el repositorio falla al actualizar
    Cuando modifico las fechas del proyecto con el ID 1 a inicio "2027-01-11" y fin "2027-03-19"
    Entonces la modificación falla con el error del repositorio
    Y el proyecto guardado con el ID 1 tiene la fecha de inicio "2026-10-23" y la de fin "2026-12-18"

  # Caso límite
  Escenario: Modificar el proyecto devuelto no altera el guardado (CA-19.6)
    Dado que existe un proyecto con nombre "Proyecto A" y descripción ""
    Y el proyecto con el ID 1 tiene registradas la fecha de inicio "2026-10-23" y la de fin "2026-12-18"
    Cuando modifico las fechas del proyecto con el ID 1 a inicio "2027-01-11" y fin "2027-03-19"
    Y cambio las fechas del proyecto devuelto a inicio "2030-01-01" y fin "2030-12-31"
    Entonces el proyecto guardado con el ID 1 tiene la fecha de inicio "2027-01-11" y la de fin "2027-03-19"

  # Caso límite
  Escenario: Modificaciones simultáneas mantienen el orden de las fechas (CA-19.7)
    Dado que existe un proyecto con nombre "Proyecto A" y descripción ""
    Y el proyecto con el ID 1 tiene registradas la fecha de inicio "2026-10-01" y la de fin "2026-12-31"
    Cuando modifico en simultáneo 20 veces las fechas del proyecto con el ID 1 a inicio "2026-10-10" y fin "2026-12-20" y a inicio "2027-02-01" y fin "2027-01-01", y registro las fechas de inicio "2026-10-05" y "2027-01-15" y las de fin "2026-12-25" y "2026-09-01"
    Entonces cada operación simultánea termina con el resultado esperado
    Y la fecha de fin guardada del proyecto con el ID 1 es posterior a la de inicio
