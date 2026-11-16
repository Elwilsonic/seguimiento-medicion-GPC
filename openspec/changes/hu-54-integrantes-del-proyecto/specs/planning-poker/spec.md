## ADDED Requirements

### Requirement: Solo votan los integrantes del proyecto
Al votar (HU-22), el sistema SHALL aceptar el voto solo si el integrante está registrado en el proyecto del ítem de la sesión (HU-15). Si no lo es, SHALL rechazarlo con "el integrante no pertenece al proyecto" (400, `ErrNotProjectMember`) sin guardarlo. Reemplaza la decisión 4 de la SDD de HU-21. (HU-54, CA-54.2)

Existencia: es justamente lo que se valida. Se valida al final, después de los errores de HU-22 (D3).

#### Scenario: Vota un integrante del proyecto (caso normal)
- **WHEN** en el proyecto "Sistema de ventas", con el integrante "Sosa Ricardo", se vota 5 como "Sosa Ricardo" en una sesión abierta
- **THEN** el voto se registra y la API responde 200

#### Scenario: Vota alguien que no es integrante (caso de error)
- **WHEN** se vota como "hernan", que no está registrado en el proyecto
- **THEN** la API responde 400 con `{"error":"el integrante no pertenece al proyecto"}` y la sesión no tiene ese voto al revelar

#### Scenario: Integrante de otro proyecto (caso límite)
- **WHEN** "Cardozo Leandro" está registrado solo en el proyecto "Inventario" y vota en una sesión de un ítem de "Sistema de ventas"
- **THEN** la API responde 400 con `{"error":"el integrante no pertenece al proyecto"}`

#### Scenario: Proyecto sin integrantes (caso límite)
- **WHEN** el proyecto del ítem no tiene integrantes y se vota como "Sosa Ricardo"
- **THEN** la API responde 400 con `{"error":"el integrante no pertenece al proyecto"}`

#### Scenario: Se conserva el orden de errores (caso límite)
- **WHEN** alguien que no es integrante vota en una sesión que ya está revelada
- **THEN** la API responde 409 con el error de sesión no abierta, y no el de integrante

### Requirement: El voto queda con el nombre registrado
El integrante del voto SHALL compararse sin espacios en los extremos y sin distinguir mayúsculas (RN-15.4), y el voto SHALL guardarse con el nombre tal como está registrado. Los votos de integrantes del proyecto SHALL seguir registrándose y reemplazándose como en HU-22. (HU-54, CA-54.2 y CA-54.4)

#### Scenario: El nombre se escribe con otras mayúsculas y espacios (caso alternativo)
- **WHEN** se vota 8 como "  sosa ricardo " en una sesión del proyecto
- **THEN** el voto se registra, y al revelar aparece "Sosa Ricardo" con 8

#### Scenario: El mismo integrante cambia su voto (caso alternativo)
- **WHEN** "Sosa Ricardo" vota 5 y después vota 8 como "SOSA RICARDO"
- **THEN** al revelar hay un solo voto de "Sosa Ricardo", con 8
