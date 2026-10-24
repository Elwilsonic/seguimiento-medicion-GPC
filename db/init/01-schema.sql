-- Esquema de la base (TEC-05). La imagen de PostgreSQL lo ejecuta sobre la
-- base POSTGRES_DB la primera vez que arranca con el volumen vacío
-- (CA-TEC05.2); 02-test-database.sql lo vuelve a aplicar en la base de
-- pruebas. Las reglas de negocio las aplica cada Manager: el esquema solo
-- guarda los datos y las relaciones entre tablas.

-- Proyectos (HU-01). El ID es correlativo (RN-01.5).
CREATE TABLE IF NOT EXISTS projects (
    id          INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name        TEXT NOT NULL,
    description TEXT NOT NULL,
    status      TEXT NOT NULL
);

-- Ítems del Product Backlog (HU-04). El ID es correlativo y global
-- (RN-04.6). Los criterios de aceptación se guardan en orden en un arreglo.
-- sprint_id es el sprint asignado por HU-11; 0 significa sin asignar
-- (RN-11.4). No hay clave foránea porque el 0 no es un sprint.
CREATE TABLE IF NOT EXISTS backlog_items (
    id                  INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    project_id          INTEGER NOT NULL REFERENCES projects (id),
    title               TEXT NOT NULL,
    description         TEXT NOT NULL,
    priority            TEXT NOT NULL,
    acceptance_criteria TEXT[] NOT NULL,
    status              TEXT NOT NULL,
    story_points        INTEGER NOT NULL,
    estimated_hours     DOUBLE PRECISION NOT NULL,
    sprint_id            INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS backlog_items_project_id_idx ON backlog_items (project_id);

-- Sprints (HU-10). El ID es correlativo y global (RN-10.6).
CREATE TABLE IF NOT EXISTS sprints (
    id         INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    project_id INTEGER NOT NULL REFERENCES projects (id),
    name       TEXT NOT NULL,
    start_date TIMESTAMPTZ NOT NULL,
    end_date   TIMESTAMPTZ NOT NULL,
    goal       TEXT NOT NULL,
    status     TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS sprints_project_id_idx ON sprints (project_id);

-- Esfuerzos reales (HU-26). Las claves foraneas apuntan a items y sprints; la
-- pertenencia del item al sprint la valida el Manager (RN-26.6).
CREATE TABLE IF NOT EXISTS efforts (
    id          INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    member      TEXT NOT NULL,
    effort_date TIMESTAMPTZ NOT NULL,
    activity    TEXT NOT NULL,
    item_id     INTEGER NOT NULL REFERENCES backlog_items (id),
    sprint_id   INTEGER NOT NULL REFERENCES sprints (id),
    hours       DOUBLE PRECISION NOT NULL
);

CREATE INDEX IF NOT EXISTS efforts_sprint_id_idx ON efforts (sprint_id);
CREATE INDEX IF NOT EXISTS efforts_item_id_idx ON efforts (item_id);

-- Sesiones de Planning Poker (HU-21). Cada sesión pertenece a un ítem del
-- backlog y guarda sus votos como JSON; el ID es correlativo global
-- (RN-21.7).
CREATE TABLE IF NOT EXISTS poker_sessions (
    id           INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    item_id      INTEGER NOT NULL REFERENCES backlog_items (id),
    status       TEXT NOT NULL,
    round_number INTEGER NOT NULL,
    votes        JSONB NOT NULL DEFAULT '[]'::jsonb
);

CREATE INDEX IF NOT EXISTS poker_sessions_item_id_idx ON poker_sessions (item_id);
