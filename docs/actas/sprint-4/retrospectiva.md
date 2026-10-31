# Acta de Retrospectiva — Sprint 4 (La Entrega)

> Issue del repositorio: #136

Las evaluaciones salen del repositorio y de GitHub al 30/10/2026.

## Datos generales

- **Proyecto:** Software Metrics & Estimation
- **Sprint:** Sprint 4 — La Entrega (27/10/2026 al 02/11/2026; desarrollo cerrado el 30/10)
- **Fecha de la Retrospectiva:** 30/10/2026
- **Facilitador:** Peñalbé Hernán (Agile Enabler)

## Participantes

| Rol | Integrante | Asistió |
|---|---|---|
| Product Architect | Profesores de la cátedra | No |
| Agile Enabler | Peñalbé Hernán | Sí |
| Product Builder | Buttini Cristobal | Sí |
| Product Builder | Cardozo Leandro | Sí |
| Product Builder | García Santiago | Sí |
| Product Builder | Sosa Ricardo | Sí |

## Acciones de la Retrospectiva del Sprint 3

| Acción | ¿Se cumplió? | Evidencia |
|---|---|---|
| 1. Quien aprueba un Pull Request revisa también el Issue (criterios tildados y "## Trazabilidad" de 4 líneas) y no aprueba hasta que esté | Parcial | Las 18 historias cerradas tienen la "## Trazabilidad". 13 tienen los criterios tildados; quedaron sin tildar #48 (HU-30), #58 (HU-40, además con "(borrador)"), #28 (HU-14), #31 (HU-17) y #46 (HU-28) |
| 2. Un test que se agrega después del GREEN se registra como desviación en la evidencia, con el motivo | Sí | HU-33 (dos tests de CA-33.8, por la revisión del PR #142) y HU-28 (`841cfc0`, por la revisión del PR #157) lo registran en su evidencia |
| 3. Antes de enviar la aprobación, releer que el comentario sea del PR que se está aprobando | Sí | No se encontró ninguna aprobación con el texto de otro PR; el equipo lo confirmó en la reunión |
| 4. La Daily se publica antes del mediodía y sin campos sin completar | Parcial | Issue #134: antes de las 12:00 el 27/10 (09:52) y el 29/10 (01:50); después, el 28/10 (12:10) y el 30/10 (13:59). En la del 28/10 quedó un `⟨completar⟩` en los bloqueos de Buttini y el campo de bloqueos de García vacío |
| 5. Antes de abrir el Pull Request, traer `main` a la rama; y después de cada merge, avisar en el grupo | Sí | Commits `chore(…): traer main` en HU-14, HU-16, HU-19, HU-24, HU-28, HU-31, HU-36, HU-37, HU-38, HU-41, HU-47 y TEC-07. Los conflictos se resolvieron con merge, sin rebase. El equipo confirmó en la reunión los avisos en el grupo |
| 6. Completar los pendientes menores de la Retrospectiva del Sprint 3 | Parcial | Se tildaron los criterios de #36 (HU-18) y #52 (HU-34). Siguen pendientes: la "## Trazabilidad" de #45 (HU-27) no sigue el formato de `AGENTS.md`; `docs/ia/registro.md` no tiene filas de HU-21, HU-22, HU-26, HU-27, HU-32 ni TEC-03; y el mensaje de `internal/project/features_test.go:209` sigue diciendo "hay %d proyectos guardados" también con uno |

## ¿Qué salió bien?

- Se cumplió el Sprint Goal con 35 de 35 Story Points y, con la capacidad sobrante, entraron 24 Story Points más (HU-41, HU-24, HU-14, HU-15 a HU-17, HU-19 y HU-28): las 42 historias del Product Backlog quedaron cerradas. La velocidad pasó de 34 a 59.
- La cadena larga del Planning (HU-30 → HU-47 → HU-35 → HU-37, HU-38 y HU-39 → HU-40) se integró completa en tres días, sin recortar alcance dentro de las historias.
- La librería de PDF (`go-pdf/fpdf`) se aprobó y quedó registrada antes de usarla, como pide `AGENTS.md` (tema 3 del Planning).
- Las revisiones encontraron problemas reales antes del merge: cambios pedidos en HU-31, HU-33 y HU-35, y observaciones del manual de usuario (TEC-07). El defecto del gráfico de HU-37 se encontró al armar el manual y se corrigió con TDD.
- La cobertura subió a 84.5% sin base y 96.2% con PostgreSQL, y los escenarios BDD ejecutados pasaron de 346 a 656.
- La documentación final avanzó en paralelo con el desarrollo: TEC-06 se actualizó cuatro veces durante el Sprint y TEC-07 entró el 30/10, con 19 capturas de la interfaz final.
- Todos los Pull Requests tuvieron la aprobación de un integrante distinto del autor, CI en verde y merge con "Create a merge commit". Ningún commit lleva firma ni atribución de la IA.
- Cada historia pasó por el ciclo completo (SDD aprobada, RED, GREEN, evidencia) con revisiones intermedias que encontraron errores antes de commitear: tests que pasaban en falso, una fila de tabla que quedó para el GREEN en lugar del RED, y un doble de test que dejaba de compilar al ampliar una interfaz (HU-16).
- Las verificaciones de los revisores con PostgreSQL (HU-16, HU-24, HU-28 y HU-40) cubrieron lo que el CI no ejecuta.

## ¿Qué se puede mejorar?

1. **Cierre de los Issues.** Cinco de las 18 historias quedaron con los criterios sin tildar, aunque desde el Sprint 3 el revisor tiene que revisarlo antes de aprobar (acción 1).
2. **Daily.** Dos de cuatro se publicaron después del mediodía y una quedó con campos sin completar, por tercer Sprint seguido.
3. **Archivos compartidos y aprobaciones anuladas.** Casi todos los PR agregan su fila en el mismo lugar de `docs/ia/registro.md` y tocan `README.md`, `cobertura.md` y `server.go`. Traer `main` para resolver el conflicto anuló la aprobación ya dada en #147, #150, #151, #157 y #162, y el revisor tuvo que aprobar de nuevo.
4. **Revisiones que quedan como comentario.** En #149 y #161 la primera revisión se envió como comentario y no como aprobación (en #161 decía "apruebo"), así que el PR no se podía integrar hasta la aprobación formal.
5. **Interfaz sin tests automáticos.** El JavaScript solo tiene tests de contenido; el defecto del gráfico de HU-37 llegó a `main` y se vio recién al armar el manual.
6. **Pendientes menores que se arrastran.** Tres pendientes de la Retrospectiva del Sprint 3 siguen abiertos (acción 6).
7. **Los tests de PostgreSQL no corren en el CI.** El workflow no levanta una base: esos tests se omiten (34 al cerrar HU-39, 41 al cerrar HU-16) y el SQL nuevo (por ejemplo, `GetByID` y `Update` de HU-16, en 0.0% sin base) depende de que el revisor lo corra a mano con Docker. Conviene sumar un servicio PostgreSQL al CI.
8. **Conversión de métricas duplicada.** `toReportMetricsResponse` (reportes, HU-38) repite la conversión de `projectMetrics` (HU-35) a propósito, mientras HU-37 cambiaba `metrics.go`; con las dos historias integradas, sigue sin unificar.
9. **Entorno de desarrollo en Windows.** En una máquina, `go test` se trababa con el antivirus y la carpeta sincronizada (OneDrive); se resolvió usando `GOTMPDIR=C:/gotmp` con `-p 1`. Además, la base local de Docker anterior a HU-15 no tenía la tabla `members` y hubo que recrearla con `docker compose down -v`.

## Incidencias del Sprint registradas

- Al publicar Pull Requests desde el agente de IA, GitHub agrega un pie "Generated by Claude Code"; se quitó en cada caso y se volvió a verificar.
- La Daily del 29/10 se publicó a la 01:50, de madrugada.
- En Windows, el ejecutable de `internal/httpapi` no arrancaba con la carpeta temporal de Go por defecto (el mismo paquete, compilado aparte, pasaba en 3 s): se usó `GOTMPDIR=C:/gotmp` con `-p 1`. Causa probable: el antivirus; no se verificó.
- Después de HU-15, una base local creada antes no tenía la tabla `members`; se recreó con `docker compose down -v`, que borra los datos locales.

## Acciones hasta la entrega

No hay un Sprint siguiente: el proyecto se entrega el 02/11/2026. Las acciones son para cerrar la entrega. Cada acción tiene un responsable.

| # | Acción | Responsable |
|---|---|---|
| 1 | Tildar los criterios de los Issues cerrados: #48 (HU-30) | Sosa Ricardo |
| 2 | Tildar los criterios de #28 (HU-14) y #31 (HU-17) | Buttini Cristobal |
| 3 | Tildar los criterios de #46 (HU-28) y llevar la "## Trazabilidad" de #45 (HU-27) al formato de `AGENTS.md` | García Santiago |
| 4 | Tildar los criterios de #58 (HU-40), sin "(borrador)" | Peñalbé Hernán |
| 5 | TEC-08: informe final de métricas y cobertura, con el cierre de TEC-06, antes del 02/11 | Peñalbé Hernán |
| 6 | Completar `docs/ia/registro.md`: filas de HU-21 y HU-22 (Buttini Cristobal), HU-26 y HU-27 (García Santiago) y HU-32 y TEC-03 (Peñalbé Hernán), o aclarar si no se usó IA | Cada responsable |
| 7 | Crear el tag `sprint-4` sobre `main` con todo integrado | Peñalbé Hernán |
| 8 | Sumar un servicio PostgreSQL al CI para que los tests de PostgreSQL dejen de omitirse | Peñalbé Hernán |
| 9 | Unificar `toReportMetricsResponse` con la conversión de `projectMetrics`, o dejarlo registrado como límite conocido | Cardozo Leandro |
| 10 | Preparar la presentación y la demo final del 02/11 | Todo el equipo |

## Lecciones para próximos proyectos

- Separar desde el inicio los archivos que todos tocan (por ejemplo, una fila de registro por historia en un archivo propio), para evitar conflictos y aprobaciones anuladas.
- Tener al menos una prueba automática de la interfaz en el navegador, además de los tests de contenido.
- Revisar el Issue (criterios y Trazabilidad) como un paso más del checklist del revisor, no solo de quien cierra la historia.
- Levantar PostgreSQL en el CI desde el inicio, para que los tests de persistencia no dependan de que alguien los corra a mano.

## Pendientes menores

- El mensaje de error de un paso BDD en `internal/project/features_test.go:209` diría "hay 1 proyectos guardados" si quedara un solo proyecto (pendiente desde el Sprint 1). Se deja registrado, sin corregirlo en este Sprint.
