# language: es
# HU-54 · Votar y registrar esfuerzo solo con integrantes del proyecto (Issue #219)
# Especificación: openspec/changes/hu-54-integrantes-del-proyecto/specs/

Característica: Votar y registrar esfuerzo solo con integrantes del proyecto
  Como integrante del equipo
  quiero que solo los integrantes del proyecto puedan votar en Planning Poker y registrar horas
  para que los votos y las horas sean siempre de personas del equipo

  Antecedentes:
    Dado que existe el proyecto "Sistema de ventas" con los integrantes "Sosa Ricardo" y "García Santiago"
    Y que el proyecto "Sistema de ventas" tiene el ítem "Registrar venta" en su sprint abierto

  # Planning Poker (CA-54.2 y CA-54.4)

  Escenario: Vota un integrante del proyecto (CA-54.2)
    Dado que hay una sesión de Planning Poker abierta para "Registrar venta"
    Cuando "Sosa Ricardo" vota 5
    Entonces la respuesta tiene el código 200
    Y al revelar, los votos son:
      | integrante   | puntos |
      | Sosa Ricardo | 5      |

  Escenario: El nombre se escribe con otras mayúsculas y espacios (CA-54.4)
    Dado que hay una sesión de Planning Poker abierta para "Registrar venta"
    Cuando "  sosa ricardo " vota 8
    Entonces la respuesta tiene el código 200
    Y al revelar, los votos son:
      | integrante   | puntos |
      | Sosa Ricardo | 8      |

  Escenario: El mismo integrante cambia su voto (CA-54.4)
    Dado que hay una sesión de Planning Poker abierta para "Registrar venta"
    Cuando "Sosa Ricardo" vota 5
    Y "SOSA RICARDO" vota 8
    Entonces la respuesta tiene el código 200
    Y al revelar, los votos son:
      | integrante   | puntos |
      | Sosa Ricardo | 8      |

  Escenario: Vota alguien que no es integrante (CA-54.2)
    Dado que hay una sesión de Planning Poker abierta para "Registrar venta"
    Cuando "hernan" vota 5
    Entonces la respuesta tiene el código 400
    Y el error es "el integrante no pertenece al proyecto"
    Y al revelar después de que "García Santiago" vota 3, los votos son:
      | integrante      | puntos |
      | García Santiago | 3      |

  Escenario: Integrante de otro proyecto (CA-54.2)
    Dado que existe el proyecto "Inventario" con el integrante "Cardozo Leandro"
    Y que hay una sesión de Planning Poker abierta para "Registrar venta"
    Cuando "Cardozo Leandro" vota 5
    Entonces la respuesta tiene el código 400
    Y el error es "el integrante no pertenece al proyecto"

  Escenario: Proyecto sin integrantes (CA-54.2)
    Dado que existe el proyecto "Inventario" sin integrantes
    Y que el proyecto "Inventario" tiene el ítem "Contar stock" en su sprint abierto
    Y que hay una sesión de Planning Poker abierta para "Contar stock"
    Cuando "Sosa Ricardo" vota 5
    Entonces la respuesta tiene el código 400
    Y el error es "el integrante no pertenece al proyecto"

  Escenario: Se conserva el orden de errores al votar (CA-54.2)
    Dado que hay una sesión de Planning Poker revelada para "Registrar venta"
    Cuando "hernan" vota 5
    Entonces la respuesta tiene el código 409
    Y el error es "la sesión de Planning Poker no está abierta"

  # Esfuerzos (CA-54.3 y CA-54.4)

  Escenario: Registra horas un integrante del proyecto (CA-54.3)
    Cuando "García Santiago" registra 4 horas de "Desarrollo" en "Registrar venta"
    Entonces la respuesta tiene el código 201
    Y el esfuerzo registrado es de "García Santiago"

  Escenario: El nombre del esfuerzo se escribe con otras mayúsculas (CA-54.4)
    Cuando "GARCÍA SANTIAGO" registra 4 horas de "Desarrollo" en "Registrar venta"
    Entonces la respuesta tiene el código 201
    Y el esfuerzo registrado es de "García Santiago"

  Escenario: Registra horas alguien que no es integrante (CA-54.3)
    Cuando "hernan" registra 4 horas de "Desarrollo" en "Registrar venta"
    Entonces la respuesta tiene el código 400
    Y el error es "el integrante no pertenece al proyecto"
    Y el sprint abierto de "Sistema de ventas" no tiene esfuerzos

  Escenario: Se conserva el orden de errores al registrar horas (CA-54.3)
    Dado que el proyecto "Sistema de ventas" tiene el ítem "Emitir factura" sin sprint
    Cuando "hernan" registra 2 horas de "Desarrollo" en "Emitir factura"
    Entonces la respuesta tiene el código 409
    Y el error es "el ítem no está asignado al sprint"
