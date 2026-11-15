# Evidencia · TEC-11 · Comentar cada test con su historia, su criterio y qué verifica

> Issue del repositorio: #179

En la revisión del 05/11/2026, el profesor pidió que cada test diga a qué historia pertenece y qué verifica. TEC-11 agrega a cada función de test una línea con ese formato y un test de control que lo exige en la CI de acá en adelante.

- **Mapa de tests por historia y responsable:** [`TEC-11-tests-por-historia.md`](TEC-11-tests-por-historia.md) (PR #184).
- **Test de control:** [`internal/trazabilidad/comentarios_test.go`](../../internal/trazabilidad/comentarios_test.go).

## Formato

Cada función de test lleva arriba una línea con su historia (o tarea técnica), lo que prueba y qué verifica:

| Forma | Cuándo se usa |
|---|---|
| `// HU-XX · CA-XX.N · Verifica que …` | El test prueba un criterio de aceptación |
| `// HU-XX · RN-XX.N · Verifica que …` | Prueba una regla de negocio que no es un criterio |
| `// HU-XX · Error · Verifica que …` | Prueba una condición de error que la especificación describe sin número |
| `// HU-XX · BDD · Verifica que se cumplan los escenarios de features/HU-XX-….feature.` | Es el runner de godog de la historia |
| `// TEC-XX · CA-TECXX.N · Verifica que …` | Es un test de una tarea técnica |

El orden para elegir es: primero un `CA`; si no hay, una `RN`; si no, `Error`. Solo se agregan comentarios: no cambia el código, el nombre ni el orden de ningún test, ni las SDD. Si un test no encaja con ningún criterio, regla ni condición de error, se pregunta al autor de la historia; no se inventa.

## Criterios de aceptación

| Criterio | Cómo se cumple |
|---|---|
| CA-TEC11.1 Cada función `Test` tiene arriba el comentario con el formato | 1148 de 1148 tests, verificado por el test de control (ver "Conteos") |
| CA-TEC11.2 La historia de cada test sale del relevamiento y queda documentada | El mapa (PR #184): la historia de cada test sale del commit que agregó la función, y el autor de cada historia, de las actas de Review |
| CA-TEC11.3 Solo cambian comentarios: el diff no modifica código y `go test ./...` da el mismo resultado | Los cinco Pull Request por responsable solo agregan o corrigen líneas `//`; los 1127 tests del relevamiento siguen siendo 1127 (ver "Conteos") |
| CA-TEC11.4 Un test automático falla si una función `Test` no tiene el comentario | `internal/trazabilidad/comentarios_test.go`, que corre con `go test ./...` y en la CI (ver "Test de control") |
| CA-TEC11.5 Los comentarios se revisan por responsable, cada uno en su Pull Request, aprobados por otro integrante | PR #185, #186, #187, #190 y #191 (ver "Pull Request") |
| CA-TEC11.6 La evidencia tiene los conteos de antes y después; la Definition of Done y `AGENTS.md` incluyen la regla | Este archivo; `docs/definition-of-ready-done.md` y `AGENTS.md` (PR #192) |
| CA-TEC11.7 Las tablas de trazabilidad de las evidencias TDD enlazan cada test y el código con su línea, y un test verifica que los links no queden desactualizados | Agregado después del cierre (#192), a pedido del Agile Enabler el 09/11/2026: ver "Links en las tablas de trazabilidad" |

## Pull Request

| Pull Request | Responsable | Tests comentados | Revisión y aprobación |
|---|---|---|---|
| #184 | Peñalbé Hernán | Mapa de tests por historia y responsable | Sosa Ricardo |
| #185 | Sosa Ricardo | 197 (13 archivos) | Peñalbé Hernán: pidió completar 2 tests de contrato de PostgreSQL |
| #186 | Buttini Cristobal | 238 (18 archivos) | Sosa Ricardo: pidió completar "Verifica que" en 2 tests. El autor corrigió además 3 citas viejas de HU-10 y HU-21 |
| #187 | Peñalbé Hernán | 222 (17 archivos) | Sosa Ricardo, sin cambios |
| #190 | García Santiago | 190 (17 archivos) | Peñalbé Hernán: pidió corregir el formato de 1 test y 3 citas |
| #191 | Cardozo Leandro | 237 (21 archivos) | Peñalbé Hernán: pidió corregir 6 citas |
| Este | Peñalbé Hernán | 43 runners de godog y el test de control | — |

Cada responsable comentó los tests de las historias que hizo, para que el comentario lo escriba quien conoce los criterios. Además de los tests del mapa, cada uno llevó al formato los tests que ya nombraban su historia de otra forma (por ejemplo, `// HU-30 · Cambiar el estado de un defecto (Issue #48).`), porque el test de control también los exige.

En las revisiones se corrigieron citas que existían en la SDD pero no correspondían a lo que prueba el test (por ejemplo, una falla al leer citada con un criterio que habla de la falla al guardar). El test de control no puede detectar esos casos: los detecta la revisión, leyendo el test y la especificación.

## Conteos

| Momento | Tests | Con la línea | Sin la línea |
|---|---|---|---|
| Relevamiento del mapa (08/11/2026) | 1127 | 114 nombraban su historia, pero no con el formato | 1013 |
| Después de los cinco Pull Request por responsable (14/11/2026) | 1142 (los 1127 más 15 de HU-50) | 1099 | 43 runners de godog |
| Con este Pull Request | 1148 (más los 6 del test de control) | 1148 | 0 |

El Issue #179 decía 997 tests para comentar: la suma de sus propias cifras (534 + 199 + 280) da 1013, el número del mapa.

Comentarios por forma, con este Pull Request:

| Forma | Tests |
|---|---|
| `CA-XX.N` | 912 |
| `RN-XX.N` | 80 |
| `Error` | 18 |
| `BDD` | 43 |
| `TEC-XX · CA-TECXX.N` | 95 |
| **Total** | **1148** |

Los 1148 tests están en 135 archivos y citan 48 historias y tareas técnicas.

## Test de control

`internal/trazabilidad/comentarios_test.go` recorre todos los archivos `_test.go` del repositorio con `go/parser` (librería estándar) y lee el comentario de cada función de test. Falla si:

- un test no tiene la línea con alguna de las cinco formas (`TestComentarios_CadaTestTieneLaLineaConElFormato`);
- la historia o tarea citada no tiene especificación (`TestComentarios_LaHistoriaTieneSuEspecificacion`):
  - una historia, su SDD en `docs/especificaciones/HU-XX-*.md` o, para los cambios con OpenSpec (TEC-10), una especificación en `openspec/` que la nombre (por ejemplo, HU-50 en `openspec/specs/proyectos/spec.md`);
  - una tarea técnica, su evidencia en `docs/evidencias/TEC-XX*.md`, donde están sus criterios;
- el `CA` o la `RN` citados no aparecen en esa especificación, o llevan el número de otra historia (`TestComentarios_ElCriterioExisteYEsDeSuHistoria`); `CA-03.1` no cuenta como `CA-03.10`;
- un test con `BDD` es de una historia sin `features/HU-XX-*.feature` (`TestComentarios_BDDTieneSuFeature`).

Con `Error` el control solo puede verificar que la historia tenga su especificación, porque la condición no tiene número: esos tests se revisan en el Pull Request. Dos tests más verifican el formato con ejemplos válidos e inválidos y la búsqueda exacta del criterio.

Sus propios tests llevan el comentario `// TEC-11 · CA-TEC11.4 · Verifica que …`, y este archivo es la especificación de TEC-11 que el control encuentra.

### RED

Con el test de control escrito y antes de comentar los runners ni escribir esta evidencia:

```
$ go test -count=1 ./internal/trazabilidad/
--- FAIL: TestComentarios_CadaTestTieneLaLineaConElFormato (0.05s)
    comentarios_test.go:251: internal/backlog/create_features_test.go:18: TestFeatures_HU04CrearItem no tiene arriba la línea "// HU-XX · CA-XX.N · Verifica que …" (ver docs/evidencias/TEC-11-tests-por-historia.md)
    comentarios_test.go:251: internal/backlog/get_features_test.go:17: TestFeatures_HU05ConsultarItem no tiene arriba la línea "// HU-XX · CA-XX.N · Verifica que …" (ver docs/evidencias/TEC-11-tests-por-historia.md)
    … (43 runners de godog en total)
--- FAIL: TestComentarios_ElCriterioExisteYEsDeSuHistoria (0.01s)
    comentarios_test.go:288: internal/trazabilidad/comentarios_test.go:247: TestComentarios_CadaTestTieneLaLineaConElFormato cita CA-TEC11.4, que no aparece en la especificación de TEC-11
    … (los 6 tests del control: todavía no existía esta evidencia)
FAIL
FAIL	seguimiento-medicion-gpc/internal/trazabilidad	0.070s
```

Fallaron solo los 43 runners de godog (el de HU-50 ya tenía la línea) y los 6 tests del control. Los otros 1099 tests ya cumplían el formato, la especificación y el criterio.

### GREEN

Con los 43 runners comentados con `BDD` y esta evidencia:

```
$ go test -count=1 -v ./internal/trazabilidad/
=== RUN   TestComentarios_CadaTestTieneLaLineaConElFormato
--- PASS: TestComentarios_CadaTestTieneLaLineaConElFormato
=== RUN   TestComentarios_LaHistoriaTieneSuEspecificacion
--- PASS: TestComentarios_LaHistoriaTieneSuEspecificacion
=== RUN   TestComentarios_ElCriterioExisteYEsDeSuHistoria
--- PASS: TestComentarios_ElCriterioExisteYEsDeSuHistoria
=== RUN   TestComentarios_BDDTieneSuFeature
--- PASS: TestComentarios_BDDTieneSuFeature
=== RUN   TestComentarios_FormatoAceptaSoloLasFormasAcordadas
--- PASS: TestComentarios_FormatoAceptaSoloLasFormasAcordadas
=== RUN   TestComentarios_ReferenciaExactaDelCriterio
--- PASS: TestComentarios_ReferenciaExactaDelCriterio
PASS
ok  	seguimiento-medicion-gpc/internal/trazabilidad
```

## Validación final

```
$ gofmt -l .
(sin salida)

$ go vet ./...
(sin salida)

$ go test ./...
ok  	seguimiento-medicion-gpc/cmd/server
ok  	seguimiento-medicion-gpc/internal/backlog
ok  	seguimiento-medicion-gpc/internal/dashboard
ok  	seguimiento-medicion-gpc/internal/defect
ok  	seguimiento-medicion-gpc/internal/effort
ok  	seguimiento-medicion-gpc/internal/httpapi
ok  	seguimiento-medicion-gpc/internal/member
ok  	seguimiento-medicion-gpc/internal/metrics
ok  	seguimiento-medicion-gpc/internal/poker
ok  	seguimiento-medicion-gpc/internal/postgres
ok  	seguimiento-medicion-gpc/internal/postgrestest
ok  	seguimiento-medicion-gpc/internal/project
ok  	seguimiento-medicion-gpc/internal/report
ok  	seguimiento-medicion-gpc/internal/sprint
ok  	seguimiento-medicion-gpc/internal/trazabilidad
?   	seguimiento-medicion-gpc/internal/web	[no test files]
```

## Links en las tablas de trazabilidad (CA-TEC11.7)

Las tablas "Trazabilidad" de las 44 evidencias TDD (`docs/evidencias/HU-XX-tdd.md`) dicen, para cada criterio, qué tests y qué código lo cubren. Ahora cada nombre es un link que abre el archivo en GitHub, y los tests y las funciones abren en su línea: [`TestRegister_Exitoso`](../../internal/defect/register_test.go#L207) abre `register_test.go` en la línea del test, con su comentario `// HU-29 · CA-29.1 · …` arriba.

| Link | Cantidad |
|---|---|
| Tests, a la línea de su `func` | 1060 |
| Archivos de código | 345 |
| Funciones, tipos y errores, a la línea de su declaración | 276 |
| **Total** | **1681** |

Cómo se armaron:

- **Tests:** cada nombre de test de las tablas. Si el mismo nombre existe en varios paquetes (por ejemplo, `TestMemoryRepository_Contrato`), se eligió el de los archivos que tocaron los commits de la historia o, si no alcanzaba, el del paquete de los demás tests de la evidencia.
- **Código:** los archivos y las funciones que ya nombraba la columna "Código", con el mismo criterio para los nombres repetidos (por ejemplo, `manager.go`).
- **La columna "Código" que faltaba** en 27 evidencias se armó con datos, no a mano. Primero se completaron HU-21, HU-22, HU-33 y HU-49 (PR #194) y después las otras 23: HU-14 a HU-19, HU-23 a HU-25, HU-27, HU-28, HU-30 a HU-32, HU-34 a HU-41 y HU-47 (191 filas, de las que 2 quedaron con "—" porque el criterio no es código: CA-40.4, la librería aprobada, y CA-24.11, que usa el contrato de repositorio que ya existía). El método: para cada fila, las funciones que la historia agregó o cambió en sus commits y que ejecutan los tests de esa fila (cobertura de cada test con `go test -run '^Test…$' -coverprofile`). En las filas de interfaz se suman `index.html` y las funciones de `app.js` de la historia, y en las de "memoria y PostgreSQL", las del repositorio de PostgreSQL, porque sus tests se omiten sin base.
- **HU-01:** su tabla nombraba `TestCreate_NombreVacioSeInformaAntesQueRepetido`, que se reemplazó en la revisión del PR #68 (ver "Ajustes por la revisión del PR #68" de esa evidencia). Se dejó como texto, sin link, aclarando que se quitó.

`internal/trazabilidad/enlaces_test.go` verifica que:

- cada test nombrado en una tabla de trazabilidad tenga su link (`TestEnlaces_CadaTestDeLasTablasTieneSuLink`);
- cada link abra un archivo que existe y, si lleva línea, que esa línea declare lo que el link nombra (`TestEnlaces_CadaLinkApuntaALaLineaDeLoQueNombra`). Un campo de un struct no cuenta como declaración.

**Cuando un cambio mueve líneas** (por ejemplo, HU-51 en los tests de proyectos), el error dice en qué línea está ahora el test y qué comando corregirlo:

```
go test ./internal/trazabilidad/ -run TestEnlaces -args -update
```

Ese comando corrige el número de línea de cada link corrido y agrega el link a los tests nuevos de nombre único; los links que siguen bien no se tocan. Si un nombre se repite en varios paquetes, el link se agrega a mano. `TestEnlaces_ActualizarCorrigeLineasYAgregaLinks` lo verifica. Se probó moviendo dos líneas en `internal/project/get_test.go` y `get.go`: el test marcó 12 links corridos, `-update` corrigió solo los de HU-03 y el test volvió a pasar (la observación de García Santiago en la revisión del #194).

### RED

```
$ go test -count=1 ./internal/trazabilidad/
--- FAIL: TestComentarios_ElCriterioExisteYEsDeSuHistoria (0.01s)
    … (los 3 tests nuevos citan CA-TEC11.7, que todavía no estaba en esta evidencia)
--- FAIL: TestEnlaces_CadaTestDeLasTablasTieneSuLink (0.01s)
    enlaces_test.go:78: docs/evidencias/HU-01-tdd.md:20: TestCreate_ProyectoValidoQuedaConIDNombreDescripcionYEstadoInicial no tiene link a su test
    enlaces_test.go:78: docs/evidencias/HU-01-tdd.md:20: TestCreate_IDsCorrelativosEmpiezanEnUno no tiene link a su test
    … (1061 nombres de test sin link)
--- FAIL: TestEnlaces_CadaLinkApuntaALaLineaDeLoQueNombra (0.00s)
    enlaces_test.go:121: las tablas de trazabilidad no tienen ningún link
FAIL
```

### GREEN

```
$ go test -count=1 -v -run 'TestEnlaces|TestComentarios' ./internal/trazabilidad/
--- PASS: TestComentarios_CadaTestTieneLaLineaConElFormato
--- PASS: TestComentarios_LaHistoriaTieneSuEspecificacion
--- PASS: TestComentarios_ElCriterioExisteYEsDeSuHistoria
--- PASS: TestComentarios_BDDTieneSuFeature
--- PASS: TestComentarios_FormatoAceptaSoloLasFormasAcordadas
--- PASS: TestComentarios_ReferenciaExactaDelCriterio
--- PASS: TestEnlaces_CadaTestDeLasTablasTieneSuLink
--- PASS: TestEnlaces_CadaLinkApuntaALaLineaDeLoQueNombra
--- PASS: TestEnlaces_ReconoceLasDeclaraciones
PASS
ok  	seguimiento-medicion-gpc/internal/trazabilidad
```

| Fase | Commit |
|---|---|
| Test de los links (RED) | `3373ce4` |
| Links en las 44 evidencias y columna "Código" en 4 (GREEN) | `04be497` |

## Commits del Pull Request de cierre (#192)

| Fase | Commit |
|---|---|
| Test de control (RED) | `a265916` |
| Runners de godog con `BDD` y esta evidencia (GREEN) | `abd5c75` |

El commit de cierre, con la regla en `AGENTS.md`, la Definition of Done, el mapa, el README y el registro de IA, lleva `Closes #179` y no se lista acá porque un commit no puede contener su propio hash.
