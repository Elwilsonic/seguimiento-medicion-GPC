# Definition of Ready y Definition of Done

Criterios vigentes para todas las historias y tareas, acordados en la Retrospectiva del Sprint 1 (acciones 1 a 5 y 9) y en el [acta de Planning del Sprint 2](actas/sprint-2/planning.md). Cada Planning puede ajustarlos; los cambios se registran en su acta y en este archivo.

## Definition of Ready

Una historia puede empezar a implementarse cuando:

- Tiene su Issue con los criterios de aceptación definitivos (sin "(borrador)").
- Su SDD está aprobada por el Agile Enabler.
- Tiene sus Story Points estimados.
- Las historias de las que depende están integradas en `main`.
- Tiene un único responsable asignado en el Issue.

Las SDD se pueden escribir en paralelo, aunque la dependencia no esté en `main`. La implementación espera a que la dependencia esté integrada: si no, el RED fallaría por código que todavía no existe y no solo por la funcionalidad nueva (paso 4 de `AGENTS.md`), y habría que crear ramas desde otras ramas en lugar de desde `main`.

## Definition of Done

- Cada historia tiene su SDD aprobada, escenarios BDD automatizados con godog, tests en verde, código, evidencia TDD (`docs/evidencias/HU-XX-tdd.md`) y trazabilidad en el Issue.
- La fila de la historia está completada en la tabla "Historias por sprint" del README.
- Los tests pasan (`go test ./...`, también con `-race` en el CI), el código está formateado (`gofmt`) y `go vet ./...` no informa problemas.
- No queda código nuevo sin un test que lo pida: se revisa `go tool cover -func` antes del commit GREEN.
- Las cifras de la evidencia salen de la salida de comandos, pegada en la evidencia.
- La funcionalidad se puede usar desde la interfaz (TEC-04) cuando corresponde.
- La rama está integrada a `main` por Pull Request, con "Create a merge commit" y la aprobación de un integrante distinto del responsable, que verifica el RED y el GREEN.
- El Issue se cierra con `Closes #N` al hacer merge, nunca a mano.
- No hay firmas ni menciones de la IA en commits, Pull Requests, Issues ni comentarios; el uso de la IA se registra en `docs/ia/registro.md`.
