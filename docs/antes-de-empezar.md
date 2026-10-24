# Antes de empezar una historia

Checklist para cualquier integrante, use o no una herramienta de IA. Evita trabajar sobre una copia vieja del repositorio, como pasó con la SDD de HU-12 (tema 7 del acta de Planning del Sprint 3).

## 1. Repositorio al día

```
git checkout main
git pull origin main
git config user.name
git config user.email
```

`user.name` y `user.email` tienen que ser los tuyos, los mismos de tus commits anteriores. Si no, configuralos solo para este repositorio (sin `--global`).

## 2. Leer antes de escribir

- El Issue de la historia: criterios, dependencias y responsable.
- El acta de Planning del Sprint (`docs/actas/sprint-N/planning.md`): Story Points, dependencias y Definition of Done.
- `docs/definition-of-ready-done.md`.
- La SDD de cada historia de la que depende (`docs/especificaciones/`) y el código que vas a reutilizar.
- Para las métricas: `docs/metricas.md` (TEC-03).

## 3. Revisar lo que está en curso

- Los Pull Requests abiertos, para no tocar los mismos archivos que otro sin avisar.
- Que las dependencias de tu historia ya estén integradas en `main`. Si no, avisalo en la Daily.

## 4. Crear la rama

```
git checkout -b feature/HU-XX-nombre-corto
```

Siempre desde `main` actualizado. Después seguí la cadena de `AGENTS.md`: historia y SDD, y esperar la aprobación antes del BDD.

## Si usás una herramienta de IA

- **Con acceso al repositorio** (por ejemplo Claude Code, Copilot, Codex, Cursor o Gemini CLI): leen `AGENTS.md` solos, y su "Paso 0" les pide hacer este checklist y frenar antes de escribir.
- **Sin acceso al repositorio** (un chat web): copiale el mensaje de abajo y pegale el contenido de los archivos que pide.

```
Vamos a trabajar la historia HU-XX, Issue #N, en un proyecto Go con Scrum y SDD → BDD → TDD.
Te paso estos archivos del repositorio, actualizados desde main:
1. AGENTS.md (las reglas del flujo de trabajo: seguilas en orden).
2. El texto del Issue #N.
3. La parte del acta de Planning del Sprint que habla de esta historia.
4. docs/definition-of-ready-done.md.
5. La SDD y el código de las historias de las que depende: <lista>.
No asumas nada que no esté en estos archivos. Hacé solo los pasos 1 y 2 de AGENTS.md (historia y SDD) y frená para que los revise.
```

Al terminar la historia, registrá el uso de la IA en `docs/ia/registro.md`.
