-- Base de pruebas de integración (CA-TEC05.7): se llama como POSTGRES_DB con
-- el sufijo _test y tiene el mismo esquema. Los tests vacían sus tablas, así
-- que nunca tocan los datos de la base de desarrollo.
\getenv dbname POSTGRES_DB
\set testdb :dbname '_test'

CREATE DATABASE :"testdb";

\connect :"testdb"
\ir 01-schema.sql
