# Diseño: HU-54 · Votar y registrar esfuerzo solo con integrantes del proyecto

## Context

- `poker.Manager.Vote(sessionID, voter, points)` guarda el votante como texto, sin validarlo (decisión 4 de la SDD de HU-21). Lo llama `internal/httpapi/poker.go`.
- `effort.Manager.Register(sprintID, itemID, member, …)` guarda el integrante como texto (RN-26.1). Valida el sprint y el ítem dentro de `sprint.Manager.WithOpenSprint`. Lo llama `internal/httpapi/efforts.go`.
- `member.Manager` (HU-15 a HU-17) ya guarda los integrantes por proyecto y compara nombres con `nameTaken`: sin espacios en los extremos y sin distinguir mayúsculas (RN-15.4).
- La interfaz (HU-53) ya carga `datos.integrantes` del proyecto actual para "Modificar integrante".

## Goals / Non-Goals

**Goals:** que la API no acepte votos ni esfuerzos de quien no es integrante del proyecto, que la interfaz ofrezca la lista, y no cambiar el orden de errores ni lo que verifican los tests existentes.

**Non-Goals:** guardar el UUID del integrante en votos y esfuerzos, dar de baja integrantes, migrar datos guardados.

## Decisions

### D1. `member.Manager.CanonicalName` resuelve el nombre registrado

`CanonicalName(projectID int, name string) (string, error)`: con el mutex del `Manager`, lista los integrantes del proyecto y devuelve el `Name` del que coincide con `strings.EqualFold(strings.TrimSpace(name), m.Name)`, o `ErrNotProjectMember` ("el integrante no pertenece al proyecto"). Reutiliza la regla de RN-15.4, así "sosa ricardo" es "Sosa Ricardo" en todos lados.

*Alternativa descartada: `IsMember(...) bool`.* No permite guardar el nombre con sus mayúsculas registradas, y el voto de "sosa ricardo" quedaría distinto del de "Sosa Ricardo" en los reportes.

### D2. Una interfaz chica en `poker` y en `effort`

Cada paquete declara `type Members interface { CanonicalName(projectID int, name string) (string, error) }`, que `*member.Manager` cumple. Así `poker` y `effort` no importan `member` (no hay ciclo y los tests usan un falso).

### D3. La pertenencia se valida al final

- **Vote:** después de leer la sesión y verificar que está ABIERTA, se busca el ítem de la sesión (`items.GetItem`) para obtener su proyecto, y se llama a `CanonicalName`. Todo dentro del mutex de `poker.Manager`, como hoy. El voto se guarda con el nombre devuelto, y el reemplazo de un voto previo (HU-22) compara contra ese nombre.
- **Register:** dentro de `WithOpenSprint`, después de validar el ítem y que esté asignado al sprint, `CanonicalName(item.ProjectID, member)`. El esfuerzo se guarda con el nombre devuelto.

Así los errores que ya existían salen igual y en el mismo orden. Orden de mutex: sprint → member (en `Register`) y poker → member (en `Vote`). `member.Manager` no toma ningún otro mutex mientras tiene el suyo, así que no hay bloqueo cruzado.

*Alternativa descartada: validar en `httpapi` antes de llamar al `Manager`.* Cambia el orden de errores (el integrante saldría antes que la sesión cerrada) y deja la regla fuera del dominio.

### D4. 400 para `ErrNotProjectMember`

Es un valor inválido del cuerpo, como el integrante vacío (`ErrEmptyVoter`, `ErrEmptyMember`), que ya responden 400. No es 404 porque el integrante no viaja como identificador de la ruta.

### D5. Tests existentes: un falso que acepta a todos

`NewManager` de `poker` y de `effort` recibe el nuevo parámetro `members`. Los tests de otras historias (HU-21 a HU-28) que arman el `Manager` a mano reciben un falso que devuelve el nombre tal cual (`anyMember`), para que sigan verificando lo mismo sin registrar integrantes. Los handlers de `internal/httpapi` de los tests usan el `member.Manager` real: los tests de API que votan o registran horas registran antes a los integrantes que usan, en sus helpers. Ningún test cambia lo que verifica, y la evidencia lista cada uno.

### D6. Interfaz

`<input name="integrante">` de `#form-votar-poker` y de "Registrar esfuerzo" pasan a `<select name="integrante">`, llenados con `llenarDesplegable` desde `datos.integrantes` (HU-53), con el nombre como valor y como texto. Sin integrantes, una sola opción deshabilitada "Registrá integrantes primero" y el botón no envía (validación `required` del navegador).

## Risks / Trade-offs

- **[R1] Cambio de firma de `NewManager`.** Toca todos los tests que arman los managers de `poker` y `effort`. → Es mecánico (D5) y la CI lo verifica.
- **[R2] Votos y esfuerzos viejos con nombres libres.** Quedan como están. Azure usa memoria y se recarga; PostgreSQL local se puede recrear.
- **[R3] Integrante renombrado (HU-16).** Sus votos y esfuerzos previos conservan el nombre anterior. → Fuera de alcance; se documenta.
- **[Trade-off] Se compara por nombre y no por UUID.** Es lo que guardan hoy votos y esfuerzos; pasar a UUID cambia el esquema y las respuestas.

## Migration Plan

Merge con "Create a merge commit" y la CI en verde. Sin cambios de esquema. Después del despliegue en Azure, cargar los datos de la demo (los integrantes se registran antes de votar y de cargar horas).
