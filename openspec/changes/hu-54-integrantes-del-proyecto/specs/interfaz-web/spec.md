## ADDED Requirements

### Requirement: Elegir el integrante de una lista al votar y al registrar esfuerzo
En "Votar" (Planning Poker) y en "Registrar esfuerzo", «Integrante» SHALL ser un desplegable con los integrantes del proyecto actual, por nombre, sin escribir el nombre a mano. Sin integrantes, SHALL mostrar "Registrá integrantes primero" y no enviar el formulario. (HU-54, CA-54.1)

La lista se actualiza al cambiar el proyecto actual y después de registrar o modificar un integrante, como las demás listas (CA-53.3).

Existencia: la valida la API (CA-54.2 y CA-54.3) y la interfaz muestra su error.

#### Scenario: Votar eligiendo el integrante (caso normal)
- **WHEN** con el proyecto "Sistema de ventas" y sus integrantes "Sosa Ricardo" y "García Santiago", se abre una sesión, se elige "Sosa Ricardo" en «Integrante» de "Votar" y se vota 5
- **THEN** la interfaz muestra el mensaje de voto registrado

#### Scenario: Registrar esfuerzo eligiendo el integrante (caso normal)
- **WHEN** en "Registrar esfuerzo" se elige "García Santiago" en «Integrante» y se completan los demás campos
- **THEN** el esfuerzo se registra con "García Santiago"

#### Scenario: La lista sigue al proyecto actual (caso alternativo)
- **WHEN** se cambia el proyecto actual a "Inventario", que tiene solo a "Cardozo Leandro"
- **THEN** «Integrante» de "Votar" y de "Registrar esfuerzo" ofrece solo "Cardozo Leandro"

#### Scenario: Proyecto sin integrantes (caso límite)
- **WHEN** el proyecto actual no tiene integrantes
- **THEN** «Integrante» muestra "Registrá integrantes primero" y no se puede votar ni registrar el esfuerzo
