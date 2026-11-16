# Tareas: HU-54 · Votar y registrar esfuerzo solo con integrantes del proyecto

Rama `feature/HU-54-integrantes-del-proyecto`, desde `main` (`d0a2b98`). Cada commit y cada push necesitan la autorización explícita del Agile Enabler; cada commit lleva `-m "Refs #219"`.

## 0. Aprobación

- [x] 0.1 Registrar en `proposal.md` la aprobación del Agile Enabler (fecha y ajustes). Verificar con `openspec validate hu-54-integrantes-del-proyecto --strict`. Proponer `git commit -m "docs(HU-54): agregar la propuesta de OpenSpec" -m "Refs #219"`.
- [ ] 0.2 Actualizar los criterios del Issue #219 con los definitivos, solo con el permiso del Agile Enabler.

## 1. Escenarios BDD

- [x] 1.1 Escribir `features/HU-54-integrantes-del-proyecto.feature` (`# language: es`), a nivel de la API, con los escenarios de `specs/planning-poker` y `specs/esfuerzos`: normal, alternativo, límite y error.
- [x] 1.2 Implementar los pasos en `internal/httpapi/hu54_features_test.go` con su runner godog.

## 2. Tests en RED

- [x] 2.1 `internal/member`: `CanonicalName` (nombre exacto, otras mayúsculas y espacios, no integrante, integrante de otro proyecto, proyecto sin integrantes).
- [x] 2.2 `internal/poker` y `internal/effort`: voto y esfuerzo con integrante, con nombre normalizado, rechazados sin guardar, y orden de errores (D3).
- [x] 2.3 `internal/httpapi`: 400 con el mensaje en las dos rutas, y los desplegables de la interfaz en el navegador (chromedp, como HU-53).
- [x] 2.4 Adaptar los tests existentes según D5, sin cambiar lo que verifican. Va con el GREEN: necesita la firma nueva de `NewManager`, que en RED no existe.
- [x] 2.5 Correr `go test ./...` y verificar que fallan solo los tests de HU-54, por función no declarada o por aserción. Proponer `git commit -m "test(HU-54): agregar tests (RED)" -m "Refs #219"`.

## 3. Código en GREEN

- [x] 3.1 `member.CanonicalName` y `ErrNotProjectMember` (D1).
- [x] 3.2 Interfaz `Members` y validación en `poker.Vote` y `effort.Register` (D2, D3); `NewManager` con el parámetro nuevo; `cmd/server/main.go`.
- [x] 3.3 `statusFor`: `ErrNotProjectMember` → 400 (D4).
- [x] 3.4 `index.html` y `app.js`: desplegables de integrante (D6).
- [x] 3.5 `go test ./...` y `go test -race ./...` en verde. Proponer `git commit -m "feat(HU-54): validar el integrante al votar y al registrar esfuerzo (GREEN)" -m "Refs #219"`.

## 4. Refactor

- [x] 4.1 Revisar si hay una mejora clara; si no, registrarlo en la evidencia.

## 5. Evidencia y documentación

- [x] 5.1 `gofmt -w`, `go test ./...` y `go test -cover ./...`.
- [x] 5.2 `docs/evidencias/HU-54-tdd.md` con RED, GREEN, cobertura, commits y la tabla criterio → test → código; `-update` de `internal/trazabilidad`.
- [x] 5.3 `docs/reportes/cobertura.md`, manual de usuario (votar y registrar esfuerzo eligen el integrante) y `docs/ia/registro.md`. Proponer `git commit -m "docs(HU-54): agregar evidencia TDD y cobertura" -m "Refs #219"`.

## 6. Pull Request y trazabilidad

- [ ] 6.1 Abrir el PR hacia `main` con revisión de los cuatro integrantes, y agregar "## Trazabilidad" al Issue #219.
- [ ] 6.2 Después del merge, `/opsx:archive` en su propio commit (`docs(HU-54): archivar el cambio en OpenSpec`) y los links definitivos en el Issue.
