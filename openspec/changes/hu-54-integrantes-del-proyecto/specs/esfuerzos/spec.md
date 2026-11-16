## ADDED Requirements

### Requirement: Solo registran esfuerzo los integrantes del proyecto
Al registrar un esfuerzo (HU-26), el sistema SHALL aceptarlo solo si el integrante está registrado en el proyecto del ítem (HU-15). Si no lo es, SHALL rechazarlo con "el integrante no pertenece al proyecto" (400, `ErrNotProjectMember`) sin guardarlo. Reemplaza RN-26.1; corregir horas (HU-28) no cambia. (HU-54, CA-54.3)

Existencia: es justamente lo que se valida. Se valida al final, después de los errores de HU-26 (D3).

#### Scenario: Registra horas un integrante del proyecto (caso normal)
- **WHEN** "García Santiago", integrante de "Sistema de ventas", registra 4 horas de "Desarrollo" sobre un ítem del sprint abierto
- **THEN** la API responde 201 y el esfuerzo tiene `"integrante": "García Santiago"`

#### Scenario: Registra horas alguien que no es integrante (caso de error)
- **WHEN** se registra un esfuerzo con el integrante "hernan"
- **THEN** la API responde 400 con `{"error":"el integrante no pertenece al proyecto"}` y la lista de esfuerzos del sprint no cambia

#### Scenario: Se conserva el orden de errores (caso límite)
- **WHEN** alguien que no es integrante registra horas sobre un ítem que no está en el sprint
- **THEN** la API responde con el error de HU-26 del ítem no asignado al sprint, y no el de integrante

### Requirement: El esfuerzo queda con el nombre registrado
El integrante del esfuerzo SHALL compararse sin espacios en los extremos y sin distinguir mayúsculas (RN-15.4), y el esfuerzo SHALL guardarse con el nombre tal como está registrado. Los esfuerzos de integrantes del proyecto SHALL seguir registrándose como en HU-26. (HU-54, CA-54.3 y CA-54.4)

#### Scenario: El nombre se escribe con otras mayúsculas (caso alternativo)
- **WHEN** se registra un esfuerzo con el integrante "GARCÍA SANTIAGO"
- **THEN** la API responde 201 y el esfuerzo tiene `"integrante": "García Santiago"`
