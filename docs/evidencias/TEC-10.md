# Evidencia · TEC-10 · Adoptar SDD con OpenSpec en el proyecto heredado

> Issue del repositorio: #177

En la revisión del 05/11/2026, el profesor pidió aplicar SDD con una herramienta reconocida y tratar el proyecto como heredado: un sistema que ya existe y al que ahora se le hacen cambios. TEC-10 adopta OpenSpec 1.16.0 para especificar esos cambios, y HU-50 y HU-51 recorren el ciclo completo con la herramienta.

- **Decisión:** [decisión 22 de `docs/decisiones-tecnicas.md`](../decisiones-tecnicas.md), con la alternativa descartada (GitHub Spec Kit) y la revisión de seguridad.
- **Instalación y uso:** sección "Especificación de cambios con OpenSpec" del [README](../../README.md).
- **Cadena de trabajo:** sección "Cambios con OpenSpec (desde TEC-10)" de [`AGENTS.md`](../../AGENTS.md).

## Criterios de aceptación

| Criterio | Cómo se cumple |
|---|---|
| CA-TEC10.1 La decisión queda registrada, con la alternativa descartada y la revisión de seguridad (versión, licencia, archivos que crea, conexiones y cómo se desactivan) | Decisión 22 de `docs/decisiones-tecnicas.md` (PR #178) |
| CA-TEC10.2 Versión fija (1.16.0), telemetría desactivada y los pasos en el README | README, "Especificación de cambios con OpenSpec": `npm install -g @fission-ai/openspec@1.16.0` y `openspec config set telemetry.enabled false` (PR #178) |
| CA-TEC10.3 `openspec init` aplicado y revisado, sin cambiar el código ni `AGENTS.md` | `openspec/` (configuración, specs y cambios) y `.claude/` (6 comandos `/opsx:*` y 6 skills), commit `84a5684`. Lo que crea está revisado en la decisión 22 (PR #178) |
| CA-TEC10.4 `openspec/config.yaml` como "constitución": stack, Scrum, BDD y TDD, la Definition of Done, y que cada requisito cite su historia, cada escenario se automatice con godog y cada test lleve su comentario | [`openspec/config.yaml`](../../openspec/config.yaml) (PR #178) |
| CA-TEC10.5 `AGENTS.md` y `docs/documentacion-tecnica.md` describen la cadena nueva y su relación con las 43 SDD anteriores | `AGENTS.md`, "Cambios con OpenSpec (desde TEC-10)", y `docs/documentacion-tecnica.md`, "Cambios con OpenSpec (TEC-10)" (PR #178) |
| CA-TEC10.6 El primer cambio con OpenSpec, el listado de proyectos, recorre el ciclo completo y se archiva | HU-50 (PR #188 y #189). Además, HU-51 recorrió el mismo ciclo (PR #195 y #196). Ver "Ciclos completos" |

## Ciclos completos con OpenSpec

| Fase | HU-50 · Listar los proyectos (#183) | HU-51 · Identificadores UUID (#193) |
|---|---|---|
| Propuesta (`proposal.md`, `specs/`, `design.md` y `tasks.md`) | `openspec/changes/hu-50-listar-proyectos/` | `openspec/changes/hu-51-identificadores-uuid/` |
| Aprobación del Agile Enabler | 08/11/2026, con tres ajustes | 14/11/2026, con dos arreglos y las decisiones D1 y D9 |
| BDD, RED, GREEN y evidencia | PR #188 | PR #195 |
| Revisión de otro integrante | Sosa Ricardo | García Santiago y Cardozo Leandro |
| Archivo (`/opsx:archive`) | PR #189, 12/11/2026 | PR #196, 15/11/2026 |
| Requisitos sumados a `openspec/specs/proyectos/spec.md` | 7 (`ADDED`) | 9 (`ADDED`) |

Los dos cambios están archivados en [`openspec/changes/archive/`](../../openspec/changes/archive/), y la especificación de la capability `proyectos` quedó con 16 requisitos:

```
$ openspec --version
1.16.0

$ openspec list
No active changes found.

$ openspec list --specs
Specs:
  proyectos     requirements 16

$ openspec validate --specs --strict
✓ spec/proyectos
Totals: 1 passed, 0 failed (1 items)
```

La trazabilidad de los tests llega también a OpenSpec: el test de control de TEC-11 (`internal/trazabilidad`) busca el criterio de HU-50 y de HU-51 en la especificación de `openspec/`.

## Alcance: qué pasa por OpenSpec

OpenSpec describe requisitos del sistema ("el sistema SHALL…"). Por eso se especifican con OpenSpec las historias nuevas y las tareas que cambian el comportamiento del sistema. Las tareas que no lo cambian (documentación, comentarios de los tests, CI) se especifican con su Issue, como antes. Por ejemplo, TEC-11 (comentar los tests) no agrega requisitos y no pasó por OpenSpec. `AGENTS.md` y la decisión 22 lo dicen desde este Pull Request.

## Lo que quedó fuera

- **Las 43 historias heredadas no se migraron a `openspec/specs/`.** Sus SDD de `docs/especificaciones/` siguen como documentación del comportamiento actual, como decidió TEC-10. Migrar los requisitos que un cambio modifica es lo que permite usar `MODIFIED`: OpenSpec rechaza un `MODIFIED` sobre una especificación que no existe (se probó en una copia aparte con HU-03 el 08/11/2026).
- **HU-51 usó `ADDED` en lugar de `MODIFIED`** por el plazo, y lo dice su decisión D9: los requisitos nuevos nombran qué reemplazan (por ejemplo, RN-01.5), y la base sigue en las SDD.
- **HU-52** (#197, los identificadores UUID del resto de las entidades) está en curso con el mismo ciclo.

## Commits

| Pull Request | Commit | Qué hace |
|---|---|---|
| #178 | `84a5684` | `openspec init` y `openspec/config.yaml` |
| #178 | `9716854` | Decisión 22, README, `AGENTS.md` y documentación técnica |
| #178 | `a836b3e` | La trazabilidad de los cambios con OpenSpec en el Issue (dos momentos) |
| #178 | `0e490cb` | Correcciones de la revisión: seguridad de las skills y telemetría |

El commit de cierre de este Pull Request, con `Closes #177`, no se lista porque un commit no puede contener su propio hash.
