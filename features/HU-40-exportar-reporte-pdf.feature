# language: es
# HU-40 · Exportar un reporte a un archivo PDF (Issue #58)
# Especificación: docs/especificaciones/HU-40-exportar-reporte-pdf.md
# CA-40.1 en la API y la interfaz, y CA-40.4 (librería) se cubren con tests de
# Go y con la SDD, no con escenarios.

Característica: HU-40 Exportar un reporte a un archivo PDF
  Como Agile Enabler,
  quiero exportar un reporte a un archivo PDF,
  para compartirlo o entregarlo.

  # Caso normal - ejemplo de TEC-03
  Escenario: Exportar el reporte del ejemplo de TEC-03 (CA-40.1, CA-40.3)
    Dado que existe el proyecto "Software Metrics" del ejemplo de TEC-03
    Cuando exporto a PDF el reporte del proyecto "Software Metrics"
    Entonces se genera un archivo PDF
    Y el PDF contiene las líneas:
      | Reporte del proyecto «Software Metrics»                                                 |
      | Inicio: 2026-10-05                                                                      |
      | Velocidad: 8.0 SP por sprint (1 sprint(s) cerrado(s)).                                  |
      | Historias completadas: 50.0 % (2 de 4).                                                 |
      | Desviación: 0.0 h (0.0 %).                                                              |
      | #1 · 2026-10-06 · Ana · Desarrollo · ítem 1 · sprint 1 · 12 h                           |
      | Total: 20.0 horas                                                                       |
      | #1 · ALTA · RESUELTO · ítem 1 · detectado en sprint 1 · resuelto en sprint 1 · Defecto |
      | Integrantes que registraron esfuerzo: Ana, Luis                                         |
    Y el PDF contiene cada línea del reporte del proyecto

  # Caso alternativo - proyecto sin datos
  Escenario: Exportar el reporte de un proyecto sin datos (CA-40.5)
    Dado que existe un proyecto con nombre "Proyecto vacío"
    Cuando exporto a PDF el reporte del proyecto "Proyecto vacío"
    Entonces se genera un archivo PDF
    Y el PDF contiene las líneas:
      | Inicio: sin registrar                            |
      | Fin: sin registrar                               |
      | Velocidad: sin datos (0 sprint(s) cerrado(s)).   |
      | Historias completadas: sin datos.                |
      | Desviación: sin datos.                           |
      | Sin esfuerzos registrados.                       |
      | Total: 0.0 horas                                 |
      | Sin defectos registrados.                        |
      | Cantidad: 0 defecto(s)                           |
      | Integrantes que registraron esfuerzo: ninguno    |

  # Caso límite - redondeo a un decimal
  Escenario: Las horas con decimales se redondean como en la API (CA-40.3)
    Dado que existe un proyecto con nombre "Decimales"
    Y que el proyecto "Decimales" tiene un sprint abierto con los esfuerzos:
      | integrante | horas |
      | Ana        | 0.1   |
      | Luis       | 0.2   |
    Cuando exporto a PDF el reporte del proyecto "Decimales"
    Entonces se genera un archivo PDF
    Y el PDF contiene las líneas:
      | Horas: 0.0 estimadas, 0.3 reales, 1 ítem(s) sin horas estimadas. |
      | Total: 0.3 horas                                                  |
      | Integrantes que registraron esfuerzo: Ana, Luis                   |

  # Caso límite - caracteres
  Escenario: Las tildes, la ñ y los signos se ven en el PDF (CA-40.7)
    Dado que existe un proyecto con nombre "Gestión de ñandúes «Sur»"
    Cuando exporto a PDF el reporte del proyecto "Gestión de ñandúes «Sur»"
    Entonces se genera un archivo PDF
    Y el PDF contiene las líneas:
      | Nombre: Gestión de ñandúes «Sur» |

  Escenario: Un carácter que no se puede mostrar no impide generar el PDF (CA-40.7)
    Dado que existe un proyecto con nombre "Proyecto 中"
    Cuando exporto a PDF el reporte del proyecto "Proyecto 中"
    Entonces se genera un archivo PDF
    Y el PDF contiene las líneas:
      | Nombre: Proyecto . |

  # Casos de error
  Escenario: Si no se puede escribir el archivo, se devuelve error (CA-40.2)
    Dado que existe el proyecto "Software Metrics" del ejemplo de TEC-03
    Cuando exporto a PDF el reporte del proyecto "Software Metrics" en un destino que falla
    Entonces la exportación falla con el error del destino

  Esquema del escenario: Exportar el reporte de un proyecto inexistente falla (CA-40.6)
    Cuando exporto a PDF el reporte del proyecto con el ID <id>
    Entonces la exportación falla porque el proyecto no existe
    Y no se escribió ningún PDF

    Ejemplos:
      | id |
      | 99 |
      | 0  |
      | -1 |

  Escenario: Exportar el reporte no modifica nada (CA-40.8)
    Dado que existe el proyecto "Software Metrics" del ejemplo de TEC-03
    Cuando exporto a PDF el reporte del proyecto "Software Metrics"
    Y exporto a PDF el reporte del proyecto "Software Metrics"
    Entonces los dos PDF tienen el mismo contenido
    Y los datos guardados del proyecto no cambiaron
