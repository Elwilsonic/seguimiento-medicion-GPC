<!--
Título: "tipo(ID): descripción", igual que los commits (por ejemplo "feat(HU-07): modificar datos de un ítem").
No incluir firmas ni menciones de la IA (AGENTS.md).
-->

## ID · Nombre de la historia o tarea

Closes #N

Historia → SDD → Criterios de Aceptación → BDD → Tests → Código Go

- *SDD:* `docs/especificaciones/HU-XX-nombre.md`
- *BDD:* `features/HU-XX-nombre.feature` (N escenarios; godog ejecuta N)
- *Tests:* archivos y cantidad de tests, contados con un comando
- *Evidencia TDD:* `docs/evidencias/HU-XX-tdd.md`
- *Cobertura:* por paquete

### Qué agrega

Qué funciones, tipos o archivos agrega o cambia, y qué decisiones de la SDD aplica.

### Commits

| Fase | Commit |
|---|---|
| SDD | |
| RED | |
| GREEN | |
| Refactor (si hubo) | |
| Evidencia y cobertura | |

### Para revisar

- Verificar que el commit RED falla solo por la funcionalidad nueva y solo tiene tests (`git checkout <RED>` y `go test ./...`), y que el GREEN pasa y solo tiene código de producción.
- `go test -race ./...` y `go vet ./...` sin problemas (también los corre el CI).
- Desviaciones del ciclo TDD: ninguna, o cuál y dónde está documentada en la evidencia.

### Definition of Done

- [ ] SDD aprobada, escenarios BDD automatizados con godog, tests en verde y evidencia TDD con las salidas de los comandos.
- [ ] Sin código nuevo sin un test que lo pida (`go tool cover -func` revisado antes del GREEN).
- [ ] Sección "## Trazabilidad" agregada al Issue, con links a `blob/main`.
- [ ] Fila de la historia completada en la tabla "Historias por sprint" del README.
- [ ] Sin firmas ni menciones de la IA en commits, PR ni Issue.

Se integra con "Create a merge commit" (nunca squash ni rebase).
