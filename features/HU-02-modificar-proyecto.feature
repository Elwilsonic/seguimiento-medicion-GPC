# language: es
# HU-02 · Modificar datos del proyecto (Issue #15)
# Especificación: docs/especificaciones/HU-02-modificar-proyecto.md

Característica: HU-02 Modificar datos del proyecto
  Como Agile Enabler,
  quiero modificar el nombre, la descripción y el estado de un proyecto,
  para mantener su información actualizada.

  # Caso normal (happy path) - CA-02.1
  Escenario: Modificar el nombre y la descripción de un proyecto existente
    Dado que existe un proyecto con nombre "Sistema de ventas" y descripción "Gestión de ventas"
    Cuando modifico el proyecto con el ID 1 con nombre "Sistema de facturación", descripción "Facturación electrónica" y estado "PLANIFICADO"
    Entonces la modificación es exitosa
    Y el proyecto modificado tiene el nombre "Sistema de facturación"
    Y el proyecto modificado tiene la descripción "Facturación electrónica"
    Y el proyecto modificado está en el estado "PLANIFICADO"

  # Caso normal - CA-02.1, CA-02.2
  Escenario: Cambiar el estado de PLANIFICADO a EN_CURSO manteniendo el resto
    Dado que existe un proyecto con nombre "Sistema de ventas" y descripción "Gestión de ventas"
    Cuando modifico el proyecto con el ID 1 con nombre "Sistema de ventas", descripción "Gestión de ventas" y estado "EN_CURSO"
    Entonces la modificación es exitosa
    Y el proyecto modificado está en el estado "EN_CURSO"
    Y el proyecto modificado tiene el nombre "Sistema de ventas"
    Y el proyecto modificado tiene la descripción "Gestión de ventas"

  # Caso alternativo - CA-02.2 (no hay transiciones prohibidas)
  Esquema del escenario: Pasar a cada uno de los estados válidos
    Dado que existe un proyecto con nombre "Proyecto A" en el estado "EN_CURSO"
    Cuando modifico el proyecto con el ID 1 con nombre "Proyecto A", descripción "" y estado "<estado>"
    Entonces la modificación es exitosa
    Y el proyecto modificado está en el estado "<estado>"

    Ejemplos:
      | estado      |
      | PLANIFICADO |
      | FINALIZADO  |

  # Caso límite - CA-02.2
  Escenario: Aceptar el estado escrito en minúsculas
    Dado que existe un proyecto con nombre "Proyecto A"
    Cuando modifico el proyecto con el ID 1 con nombre "Proyecto A", descripción "" y estado "en_curso"
    Entonces la modificación es exitosa
    Y el proyecto modificado está en el estado "EN_CURSO"

  # Caso límite - CA-02.2
  Escenario: Aceptar el estado con espacios en los extremos
    Dado que existe un proyecto con nombre "Proyecto A"
    Cuando modifico el proyecto con el ID 1 con nombre "Proyecto A", descripción "" y estado " En_Curso "
    Entonces la modificación es exitosa
    Y el proyecto modificado está en el estado "EN_CURSO"

  # Caso límite - CA-02.1, RN-02.2
  Escenario: Recortar los espacios del nombre nuevo en los extremos
    Dado que existe un proyecto con nombre "Proyecto A"
    Cuando modifico el proyecto con el ID 1 con nombre "  Proyecto B  ", descripción "" y estado "PLANIFICADO"
    Entonces la modificación es exitosa
    Y el proyecto modificado tiene el nombre "Proyecto B"

  # Caso límite - RN-02.5
  Escenario: Dejar la descripción vacía
    Dado que existe un proyecto con nombre "Proyecto A" y descripción "Original"
    Cuando modifico el proyecto con el ID 1 con nombre "Proyecto A", descripción "" y estado "PLANIFICADO"
    Entonces la modificación es exitosa
    Y el proyecto modificado tiene la descripción ""

  # Caso de error - CA-02.3
  Escenario: Rechazar un nombre vacío
    Dado que existe un proyecto con nombre "Proyecto A" y descripción "Original"
    Cuando modifico el proyecto con el ID 1 con nombre "", descripción "Original" y estado "PLANIFICADO"
    Entonces la modificación falla porque el nombre es obligatorio
    Y el proyecto guardado con el ID 1 sigue teniendo el nombre "Proyecto A" y la descripción "Original"
    Y el proyecto guardado con el ID 1 está en el estado "PLANIFICADO"

  # Caso límite - CA-02.3
  Escenario: Rechazar un nombre con solo espacios
    Dado que existe un proyecto con nombre "Proyecto A" y descripción "Original"
    Cuando modifico el proyecto con el ID 1 con nombre "   ", descripción "Original" y estado "PLANIFICADO"
    Entonces la modificación falla porque el nombre es obligatorio
    Y el proyecto guardado con el ID 1 sigue teniendo el nombre "Proyecto A" y la descripción "Original"
    Y el proyecto guardado con el ID 1 está en el estado "PLANIFICADO"

  # Caso de error - CA-02.2, CA-02.3
  Escenario: Rechazar un estado que no existe
    Dado que existe un proyecto con nombre "Proyecto A"
    Cuando modifico el proyecto con el ID 1 con nombre "Proyecto A", descripción "" y estado "PAUSADO"
    Entonces la modificación falla porque el estado es inválido
    Y el proyecto guardado con el ID 1 sigue teniendo el nombre "Proyecto A" y la descripción ""
    Y el proyecto guardado con el ID 1 está en el estado "PLANIFICADO"

  # Caso de error - CA-02.5
  Escenario: Rechazar un nombre repetido con otro proyecto
    Dado que existe un proyecto con nombre "Proyecto A"
    Y que existe un proyecto con nombre "Proyecto B"
    Cuando modifico el proyecto con el ID 2 con nombre "proyecto a", descripción "" y estado "PLANIFICADO"
    Entonces la modificación falla porque el nombre está repetido
    Y el proyecto guardado con el ID 2 sigue teniendo el nombre "Proyecto B" y la descripción ""
    Y el proyecto guardado con el ID 2 está en el estado "PLANIFICADO"
    Y el proyecto guardado con el ID 1 sigue teniendo el nombre "Proyecto A" y la descripción ""

  # Caso límite - CA-02.5 (el propio proyecto se excluye del control)
  Escenario: Dejar el nombre del propio proyecto con otras mayúsculas
    Dado que existe un proyecto con nombre "Proyecto A"
    Cuando modifico el proyecto con el ID 1 con nombre " PROYECTO a ", descripción "Actualizado" y estado "PLANIFICADO"
    Entonces la modificación es exitosa
    Y el proyecto modificado tiene el nombre "PROYECTO a"

  # Caso de error - CA-02.4
  Esquema del escenario: Modificar un proyecto que no existe
    Dado que existe un proyecto con nombre "Proyecto A"
    Cuando modifico el proyecto con el ID <id> con nombre "Proyecto B", descripción "" y estado "PLANIFICADO"
    Entonces la modificación falla porque el proyecto no existe

    Ejemplos:
      | id  |
      | 999 |
      | 0   |
      | -1  |

  # Caso de error - CA-02.4
  Escenario: Modificar cuando no hay proyectos creados
    Dado que no hay proyectos creados
    Cuando modifico el proyecto con el ID 1 con nombre "Proyecto A", descripción "" y estado "PLANIFICADO"
    Entonces la modificación falla porque el proyecto no existe

  # Caso límite - CA-02.6
  Escenario: Modificar el proyecto devuelto no altera el proyecto guardado
    Dado que existe un proyecto con nombre "Proyecto A" y descripción "Original"
    Cuando modifico el proyecto con el ID 1 con nombre "Proyecto B", descripción "Nueva" y estado "EN_CURSO"
    Y cambio el nombre del proyecto modificado a "Otro nombre"
    Entonces al reconsultar el proyecto con el ID 1 obtengo el nombre "Proyecto B" y la descripción "Nueva"

  # Caso límite - CA-02.7
  Escenario: Modificaciones y creaciones en simultáneo
    Dado que existen 10 proyectos numerados
    Cuando se modifican en simultáneo los proyectos del 1 al 10 mientras se crean 50 proyectos en simultáneo con nombres distintos
    Entonces hay 60 proyectos guardados
    Y ninguna operación simultánea devuelve un error inesperado
    Y los proyectos del 1 al 10 quedaron con su nombre actualizado y el estado "EN_CURSO"
