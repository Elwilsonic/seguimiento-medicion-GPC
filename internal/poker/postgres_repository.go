package poker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresRepository guarda las sesiones en PostgreSQL (TEC-05). Cumple el
// mismo contrato que MemoryRepository.
type PostgresRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresRepository crea un repositorio sobre pool. El pool lo cierra
// quien lo abrió.
func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

// pokerSessionColumns son las columnas de una sesión, en el orden de
// scanSession.
const pokerSessionColumns = "id, item_id, status, round_number, votes"

// Save guarda la sesión; la base asigna el siguiente ID correlativo global
// (RN-21.7). Los votos se codifican con encoding/json en la columna JSONB.
func (r *PostgresRepository) Save(session Session) (Session, error) {
	votes := session.Votes
	if votes == nil {
		votes = []Vote{}
	}
	encoded, err := json.Marshal(votes)
	if err != nil {
		return Session{}, fmt.Errorf("codificar los votos de la sesión: %w", err)
	}
	row := r.pool.QueryRow(context.Background(),
		`INSERT INTO poker_sessions (item_id, status, round_number, votes)
		VALUES ($1, $2, $3, $4::jsonb) RETURNING `+pokerSessionColumns,
		session.ItemID, session.Status, session.Round, string(encoded))
	saved, err := scanSession(row)
	if err != nil {
		return Session{}, fmt.Errorf("guardar la sesión de Planning Poker: %w", err)
	}
	return saved, nil
}

// OpenByItem devuelve una copia de la sesión ABIERTA del ítem y si existe
// (RN-21.2). Una sesión de otro ítem o con otro estado no aparece.
func (r *PostgresRepository) OpenByItem(itemID int) (Session, bool, error) {
	row := r.pool.QueryRow(context.Background(),
		"SELECT "+pokerSessionColumns+" FROM poker_sessions WHERE item_id = $1 AND status = $2 ORDER BY id LIMIT 1",
		itemID, StatusOpen)
	session, err := scanSession(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return Session{}, false, nil
	}
	if err != nil {
		return Session{}, false, fmt.Errorf("buscar la sesión abierta: %w", err)
	}
	return session, true, nil
}

// GetByID devuelve una sesión por ID.
func (r *PostgresRepository) GetByID(id int) (Session, error) {
	row := r.pool.QueryRow(context.Background(),
		"SELECT "+pokerSessionColumns+" FROM poker_sessions WHERE id = $1", id)
	session, err := scanSession(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return Session{}, ErrSessionNotFound
	}
	if err != nil {
		return Session{}, fmt.Errorf("buscar la sesión de Planning Poker: %w", err)
	}
	return session, nil
}

// Update reemplaza una sesión existente.
func (r *PostgresRepository) Update(session Session) (Session, error) {
	votes := session.Votes
	if votes == nil {
		votes = []Vote{}
	}
	encoded, err := json.Marshal(votes)
	if err != nil {
		return Session{}, fmt.Errorf("codificar los votos de la sesión: %w", err)
	}
	row := r.pool.QueryRow(context.Background(),
		`UPDATE poker_sessions
		SET status = $2, round_number = $3, votes = $4::jsonb
		WHERE id = $1
		RETURNING `+pokerSessionColumns,
		session.ID, session.Status, session.Round, string(encoded))
	updated, err := scanSession(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return Session{}, ErrSessionNotFound
	}
	if err != nil {
		return Session{}, fmt.Errorf("actualizar la sesión de Planning Poker: %w", err)
	}
	return updated, nil
}

// scanSession lee una sesión con las columnas de pokerSessionColumns y
// decodifica los votos con encoding/json.
func scanSession(row pgx.Row) (Session, error) {
	var session Session
	var votes []byte
	if err := row.Scan(&session.ID, &session.ItemID, &session.Status, &session.Round, &votes); err != nil {
		return Session{}, err
	}
	if err := json.Unmarshal(votes, &session.Votes); err != nil {
		return Session{}, fmt.Errorf("decodificar los votos de la sesión: %w", err)
	}
	return cloneSession(session), nil
}
