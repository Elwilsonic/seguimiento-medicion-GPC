# Propuesta: HU-54 · Votar y registrar esfuerzo solo con integrantes del proyecto

> Issue del repositorio: #219 · Épica E4 · Registro de integrantes, con alcance en E5 (Planning Poker) y E7 (Esfuerzo) · Vinculada a HU-15 (#29), HU-17 (#31), HU-21 (#39), HU-22 (#40), HU-26 (#44) y HU-53 (#211)
>
> Aprobada por: Peñalbé Hernán (Agile Enabler), 16/11/2026, sin ajustes: error 400, el voto y el esfuerzo se guardan con el nombre registrado, y el desarrollo sigue sin integrarse en `main` antes de la presentación.
>
> Acta de Planning: el acta del Sprint 5 (`docs/actas/sprint-5/planning.md`) no menciona esta historia; no hay alcance con qué compararla. Surge de la prueba de la demo del 16/11/2026.

## Historia de usuario

Como integrante del equipo, quiero elegir al integrante de una lista al votar en Planning Poker y al registrar un esfuerzo, y que el sistema rechace a quien no es integrante del proyecto, para que los votos y las horas sean siempre de personas del equipo.

## Why

El votante de Planning Poker y el integrante de un esfuerzo son texto libre: la decisión 4 de la SDD de HU-21 (que siguen HU-22, HU-23 y HU-25) y RN-26.1 de HU-26 lo decidieron porque la épica E4 (integrantes) no estaba implementada, y la SDD de HU-26 deja la migración para "una historia posterior". E4 ya está en `main` (HU-15 a HU-17), pero nadie conectó las dos partes. Resultado, visto en la prueba de la demo: se puede votar como "hernan", que no es integrante del proyecto, y la API lo acepta. Lo mismo pasa con las horas, que alimentan las métricas de esfuerzo (HU-33 y HU-37).

## What Changes

- **Votar (HU-22):** `POST /api/poker-sessions/{id}/votes` SHALL rechazar un integrante que no está registrado en el proyecto del ítem de la sesión, con 400 `{"error":"el integrante no pertenece al proyecto"}`, sin guardar el voto. El nombre se compara sin espacios en los extremos y sin distinguir mayúsculas, la misma regla que usa HU-15 para los duplicados (RN-15.4). El voto se guarda con el nombre tal como está registrado.
- **Registrar esfuerzo (HU-26):** `POST /api/sprints/{id}/items/{itemId}/esfuerzos` SHALL rechazar un integrante que no está registrado en el proyecto del ítem, con el mismo error y la misma comparación, sin guardar el esfuerzo. El esfuerzo se guarda con el nombre registrado.
- **Interfaz:** «Integrante» de "Votar" y de "Registrar esfuerzo" pasa de campo de texto a desplegable con los integrantes del proyecto actual, por nombre. Sin integrantes, el desplegable avisa "Registrá integrantes primero" y no se puede enviar.
- **Orden de errores:** se conserva el de cada historia; la pertenencia se valida al final, justo antes de guardar (D3 de `design.md`).
- **BREAKING (API):** un cliente que votaba o registraba horas con cualquier nombre deja de poder hacerlo. El único cliente es la interfaz propia, que se adapta en este cambio.

## Capabilities

### New Capabilities

Ninguna.

### Modified Capabilities

- `planning-poker`: se agrega el requisito de que el votante sea integrante del proyecto.
- `esfuerzos`: se agrega el requisito de que el integrante del esfuerzo sea integrante del proyecto.
- `interfaz-web`: se agrega el requisito de elegir el integrante de una lista en "Votar" y "Registrar esfuerzo".

## Fuera de alcance

- Dar de baja integrantes (excluido en las SDD de HU-15 y HU-16; el manual dice que no se borra nada).
- Migrar o validar votos y esfuerzos ya guardados: quedan como están.
- Qué pasa con los votos y esfuerzos de un integrante que después cambia de nombre (HU-16): conservan el nombre con el que se guardaron, como hoy.
- Guardar el UUID del integrante en lugar del nombre en votos y esfuerzos (cambia el esquema y las respuestas; queda como mejora posible).
- Ocultar los UUID de la interfaz (se habló como otra historia).

## Impact

- `internal/member`: una operación nueva, `CanonicalName(projectID int, name string) (string, error)`, que devuelve el nombre registrado o `ErrNotProjectMember` (D1).
- `internal/poker` y `internal/effort`: el `Manager` recibe quién resuelve integrantes (una interfaz chica, D2) y valida en `Vote` y `Register`. Cambia la firma de `NewManager` en los dos paquetes (D5).
- `internal/httpapi`: `ErrNotProjectMember` responde 400 en `statusFor`.
- `cmd/server/main.go`: arma los managers con el de integrantes.
- `internal/web`: `index.html` y `app.js` (desplegables).
- Tests: los existentes de otras historias se adaptan sin cambiar lo que verifican (D5). Nuevos: dominio, API, interfaz y BDD `features/HU-54-integrantes-del-proyecto.feature`.
- Sin cambios en el esquema de la base ni en dependencias.
