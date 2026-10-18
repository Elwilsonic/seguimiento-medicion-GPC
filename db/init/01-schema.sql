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
CREATE TABLE IF NOT EXISTS backlog_items (
    id                  INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    project_id          INTEGER NOT NULL REFERENCES projects (id),
    title               TEXT NOT NULL,
    description         TEXT NOT NULL,
    priority            TEXT NOT NULL,
    acceptance_criteria TEXT[] NOT NULL,
    status              TEXT NOT NULL,
    story_points        INTEGER NOT NULL,
    estimated_hours     DOUBLE PRECISION NOT NULL
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
