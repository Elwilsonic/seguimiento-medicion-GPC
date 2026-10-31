# Evidencia — TEC-09 · Despliegue en Azure Container Apps

> Issue del repositorio: #172

- **Responsable:** Peñalbé Hernán
- **Rama:** `feature/TEC-09-despliegue-azure`, creada desde `main` en `9e85d72`
- **Fecha:** 31/10/2026
- **Entorno de la verificación local:** Linux, Docker 29.8.2 y go1.27.1

TEC-09 es una tarea técnica: no tiene historia de usuario, SDD ni escenarios BDD propios (decisión 5 del [acta de Planning del Sprint 1](../actas/sprint-1/planning.md)). Su alcance son los criterios del Issue #172.

**Por qué no hay tests de Go nuevos:** TEC-09 no agrega código Go. Agrega un `Dockerfile`, un `.dockerignore`, un workflow y documentación. Como en TEC-02 (CI) y TEC-05 (Docker Compose), la configuración se verificó ejecutándola: construyendo y corriendo la imagen, y probando los pasos del workflow que se pueden correr sin Azure.

## Trazabilidad

Issue #172 → Criterios de Aceptación → Verificación → Configuración

| Criterio | Cómo se verificó | Resultado |
|---|---|---|
| CA-TEC09.1 — imagen multi-etapa, mínima, sin root, puerto 8080 y memoria | `docker build` e inspección de la imagen | Ver "Imagen" |
| CA-TEC09.2 — la imagen funciona localmente | `docker run -p 8088:8080` y pedidos con `curl` | Ver "Prueba local" |
| CA-TEC09.3 — workflow de construcción y despliegue | Lectura del YAML con Python, `bash -n` sobre cada paso y ejecución de los pasos que no usan Azure | Ver "Workflow" |
| CA-TEC09.4 — secretos solo en Secrets | Búsqueda de fragmentos de los valores en el repositorio y prueba del ocultamiento en los logs | Ver "Secretos" |
| CA-TEC09.5 — `.dockerignore` y CI sin cambios | Lista de archivos de la imagen; `ci.yml` sin cambios | Ver "Imagen" y "CI" |
| CA-TEC09.6 — documentación | [README](../../README.md#despliegue-en-azure) y [documentación técnica, sección 8](../documentacion-tecnica.md#8-despliegue-en-azure-container-apps) | — |
| CA-TEC09.7 — despliegue real y URL | Pendiente: necesita el merge a `main` y los Secrets cargados | Se registra en el Issue #172 |

## Imagen

Por cómo funciona esta sesión, la construcción local necesitó agregar a la etapa de compilación el certificado del proxy que usa para salir a internet. Sin eso, `go mod download` fallaba con `x509: certificate signed by unknown authority`. Se usó una copia del `Dockerfile` fuera del repositorio con una línea más (`COPY --from=ca …`) en esa etapa. La imagen final y el `Dockerfile` del repositorio no cambian, y en GitHub Actions no hace falta.

```text
$ docker build --build-arg GO_VERSION=1.27.1 -t gpc-local:test .     (con la copia descrita arriba)
sha256:31589d4bbd935160d971115582260d1e80c9ee761b747e60a6dfa4304f1094ee
$ docker images gpc-local:test --format '{{.Repository}}:{{.Tag}} {{.Size}}'
gpc-local:test 22.9MB
$ docker image inspect gpc-local:test --format 'user={{.Config.User}} ports={{.Config.ExposedPorts}} entry={{.Config.Entrypoint}}'
user=nonroot:nonroot ports=map[8080/tcp:{}] entry=[/server]
```

Las variables de la imagen son `PORT=8080` y `STORAGE=memory`. La copia local además tiene `SSL_CERT_FILE`, que viene de la imagen de compilación.

La imagen tiene el binario `/server` y los archivos base de distroless, y ningún archivo `.go`. `docker export | tar -t | grep '\.go$'` no devuelve nada.

## Prueba local

```text
$ docker run -d --name gpc-test -p 8088:8080 gpc-local:test
$ curl -s -o /dev/null -w '%{http_code}' http://localhost:8088/
200
$ curl -s -XPOST localhost:8088/api/proyectos -d '{"nombre":"Demo Azure","descripcion":"prueba"}'
{"id":1,"nombre":"Demo Azure","descripcion":"prueba","estado":"PLANIFICADO","fecha_inicio":"","fecha_fin":""}   [201]
$ curl -s localhost:8088/api/proyectos/1
{"id":1,"nombre":"Demo Azure","descripcion":"prueba","estado":"PLANIFICADO","fecha_inicio":"","fecha_fin":""}   [200]
$ docker logs gpc-test
2026/10/31 21:37:16 servidor escuchando en :8080 (almacenamiento: memory)
```

## Workflow

El archivo `.github/workflows/deploy.yml` se leyó con `yaml.safe_load`. Tiene los disparadores `push` a `main` y `workflow_dispatch`, y estos 7 pasos:

1. Descargar el código.
2. Verificar que estén los Secrets.
3. Construir la imagen (Go de go.mod).
4. Subir la imagen al registro.
5. Iniciar sesión en Azure.
6. Crear o actualizar la Container App.
7. Verificar que la aplicación responda.

`bash -n` sobre el script de cada paso no dio errores de sintaxis. Los pasos que no necesitan Azure se ejecutaron con valores de ejemplo, sin secretos reales:

```text
Verificar que estén los Secrets, sin AZURE_CREDENTIALS
  -> salida 1: Faltan los Secrets: AZURE_CREDENTIALS (Settings → Secrets and variables → Actions).
Verificar que estén los Secrets, con los dos
  -> salida 0
Iniciar sesión en Azure, hasta antes de az login, con {"clientId": "c-1", "clientSecret": "s-2", "subscriptionId": "sub-3", "tenantId": "t-4"}
  -> ::add-mask::c-1 / ::add-mask::s-2 / ::add-mask::t-4 / ::add-mask::sub-3, y los cuatro campos leídos bien
```

Lo que solo se puede probar contra Azure es el login, la subida de la imagen, la creación o actualización de la Container App y la URL. Queda para la primera ejecución después del merge (CA-TEC09.7).

No se usó `actionlint`, porque bajar y ejecutar esa herramienta externa no estaba autorizado en esta sesión.

## Secretos

- **No están en el repositorio:** se buscaron fragmentos de la contraseña del registro y de los cuatro campos de las credenciales de la planilla, y no aparecen en ningún archivo versionado ni en los archivos nuevos. Los valores no se copian en esta evidencia.
- **En el workflow:** los Secrets entran por `env` solo en los pasos que los usan. La contraseña del registro entra a `docker login` por `--password-stdin`. Cada campo de `AZURE_CREDENTIALS` se oculta en los logs con `::add-mask::` antes de usarlo, porque GitHub oculta el Secret completo pero no sus partes.

## CI

`.github/workflows/ci.yml` no cambia. El workflow nuevo no corre en los Pull Requests, solo en los push a `main` y a mano.

## Límites

1. **Almacenamiento en memoria.** La cátedra no provee una base de datos, así que en Azure los datos se pierden al reiniciar el contenedor o al volver a desplegar, en cada integración a `main`. Por eso la Container App tiene una sola réplica, siempre encendida.
2. **Credenciales compartidas.** La contraseña del registro y las credenciales de Azure son las mismas para todos los grupos de la cátedra. El repositorio no las guarda, pero quien las tenga puede usar la cuenta.
3. **Primera ejecución.** El workflow supone que el service principal de la cátedra puede crear y actualizar Container Apps en `frsrutn`. Si la cátedra ya creó la Container App, el workflow le vuelve a configurar el registro y el ingreso antes de actualizarla.
