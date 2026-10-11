# HU-01 · Crear proyecto

> Issue del repositorio: #14

## Historia de usuario

Como Agile Enabler, quiero crear un proyecto con nombre, descripción y estado inicial, para poder empezar a planificar su trabajo.

## Objetivo

Registrar un proyecto nuevo en el sistema, con un identificador único, un nombre que no se repita y el estado inicial `PLANIFICADO`, para que las demás historias (modificarlo, consultarlo y cargarle el Product Backlog) tengan sobre qué trabajar.

Esta historia también define la organización del código y la interfaz de repositorio que usan las demás historias del Sprint 1 (ver "Restricciones").

## Entradas

| Campo | Tipo | Obligatorio | Descripción |
|---|---|---|---|
| Nombre | texto | Sí | Nombre del proyecto |
| Descripción | texto | No | Descripción libre del proyecto; puede estar vacía |

El identificador y el estado no son entradas: los asigna el sistema.

## Salidas esperadas

- **Si la creación es válida:** el proyecto creado, con:
  - **ID:** número entero correlativo asignado por el sistema (el primero es 1, el siguiente 2, y así).
  - **Nombre:** el recibido, sin los espacios del principio y del final.
  - **Descripción:** la recibida, tal como llegó.
  - **Estado:** `PLANIFICADO`.
- **Si la creación no es válida:** un error que indica el motivo (ver "Condiciones de error"), y no se crea ningún proyecto.

## Reglas de negocio

1. **RN-01.1 Nombre obligatorio.** El nombre no puede estar vacío ni contener solo espacios. Se consideran espacios los espacios, tabs y saltos de línea.
2. **RN-01.2 Nombre sin espacios en los extremos.** El nombre se guarda sin los espacios del principio y del final. Los espacios internos se conservan tal como llegan. Se consideran espacios los espacios, tabs y saltos de línea.
3. **RN-01.3 Nombre único.** No puede haber dos proyectos con el mismo nombre. Dos nombres son iguales si coinciden después de quitar los espacios de los extremos y sin distinguir mayúsculas de minúsculas (por ejemplo, "Proyecto A" y " proyecto a " son el mismo nombre).
4. **RN-01.4 Descripción opcional.** La descripción puede estar vacía y se guarda tal como llega. No tiene longitud máxima.
5. **RN-01.5 Identificador único.** Cada proyecto recibe un ID numérico correlativo que empieza en 1. El ID lo asigna el repositorio al guardar el proyecto.
6. **RN-01.6 Estados del proyecto.** Un proyecto puede estar en uno de estos tres estados: `PLANIFICADO`, `EN_CURSO` o `FINALIZADO`. Todo proyecto nuevo se crea en `PLANIFICADO`. El cambio de estado lo define HU-02.

## Restricciones

- **Organización del código.** El código de proyectos va en el paquete `internal/project`. Cada historia agrega su propio archivo para evitar conflictos de merge:
  - `project.go`: el tipo `Project`, los estados y los errores.
  - `repository.go`: la interfaz `Repository`.
  - `memory_repository.go`: la implementación en memoria del repositorio.
  - `manager.go`: el `Manager`, que aplica las reglas de negocio sobre el repositorio, con su mutex (ver "Concurrencia").
  - `create.go`: la creación (HU-01). HU-02 y HU-03 agregan sus propios archivos con sus operaciones del `Manager`.
- **Interfaz de repositorio.** La lógica de negocio (el `Manager` del paquete) accede al almacenamiento solo a través de la interfaz `Repository`, nunca directamente. En el Sprint 1 se implementa en memoria; en el Sprint 2, TEC-05 agrega la implementación sobre PostgreSQL sin cambiar las reglas de negocio ni los tests. HU-01 define solo las operaciones que necesita: guardar un proyecto nuevo (el repositorio asigna el ID) y listar los proyectos guardados (para controlar que el nombre no se repita). Las historias siguientes agregan las operaciones que necesiten.
- **Copias por valor.** `Create` devuelve una copia del proyecto (un valor, no un puntero), y el repositorio en memoria guarda y devuelve copias. Modificar el proyecto devuelto no altera el proyecto guardado.
- **Errores identificables.** Cada condición de error tiene su propio error exportado, comparable con `errors.Is`, para que TEC-01 y los tests puedan distinguirlos.
- **Concurrencia.** Crear un proyecto hace dos pasos seguidos: listar los proyectos para controlar el nombre y guardar el nuevo. Como TEC-01 atiende pedidos en paralelo, el `Manager` protege esos dos pasos con un `sync.Mutex`, para que dos creaciones simultáneas no repitan nombre ni ID. Las historias que agreguen operaciones al `Manager` (HU-02, HU-03) usan el mismo mutex.
- **Librerías.** Solo la librería estándar de Go en el código, y godog en los tests (se agrega al `go.mod` en esta historia, según la decisión 14 del acta de Planning).
- **Entidades referenciadas.** HU-01 no referencia ninguna otra entidad por ID, así que no hay que validar la existencia de otra entidad.

## Casos límite

| Caso | Resultado esperado |
|---|---|
| Nombre con espacios al principio y al final ("  Proyecto A  ") | Se crea con el nombre "Proyecto A" |
| Nombre con solo espacios ("   ") | Error: nombre vacío |
| Nombre igual a uno existente con otras mayúsculas ("PROYECTO A" si existe "Proyecto A") | Error: nombre repetido |
| Nombre igual a uno existente con espacios en los extremos (" Proyecto A ") | Error: nombre repetido |
| Nombres que solo difieren en los espacios internos ("Proyecto  A" y "Proyecto A") | Se consideran distintos: se crean los dos |
| Descripción vacía | Se crea el proyecto con la descripción vacía |
| Primer proyecto del sistema | Recibe el ID 1 |
| Varias creaciones simultáneas con el mismo nombre | Se crea un solo proyecto; las demás devuelven error de nombre repetido |
| Varias creaciones simultáneas con nombres distintos | Se crean todas, cada una con un ID distinto |

## Condiciones de error

| Condición | Error | Efecto |
|---|---|---|
| Nombre vacío o con solo espacios | `ErrEmptyName` ("el nombre del proyecto es obligatorio") | No se crea el proyecto |
| Nombre repetido (según RN-01.3) | `ErrDuplicateName` ("ya existe un proyecto con ese nombre") | No se crea el proyecto |

Si el nombre está vacío, se informa `ErrEmptyName` sin controlar repetidos.

## Criterios de aceptación

- **CA-01.1** Se crea un proyecto con nombre y descripción, y queda con un ID numérico único asignado por el sistema: el primero es 1 y cada proyecto nuevo recibe el siguiente.
- **CA-01.2** El nombre es obligatorio: si está vacío o tiene solo espacios, se devuelve `ErrEmptyName` y no se crea el proyecto.
- **CA-01.3** No se pueden crear dos proyectos con el mismo nombre, sin distinguir mayúsculas y sin contar los espacios de los extremos: se devuelve `ErrDuplicateName` y no se crea el proyecto.
- **CA-01.4** El proyecto nuevo queda en el estado inicial `PLANIFICADO`.
- **CA-01.5** El nombre se guarda sin los espacios de los extremos y la descripción, que es opcional, se guarda tal como llega.
- **CA-01.6** Modificar el proyecto devuelto por la creación no altera el proyecto guardado.
- **CA-01.7** Varias creaciones simultáneas no repiten nombre ni ID.
