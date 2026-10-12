package backlog

import "slices"

// MemoryRepository guarda los ítems en memoria (Sprint 1). No es seguro para
// uso concurrente: se accede a través del Manager, que lo protege.
type MemoryRepository struct {
	items  []Item
	lastID int
}

// NewMemoryRepository crea un repositorio en memoria vacío.
func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{}
}

// Save asigna el siguiente ID correlativo global (RN-04.6), guarda una copia
// del ítem y devuelve otra copia, sin compartir la lista de criterios ni con
// el llamador ni con lo guardado (RN-04.9).
func (r *MemoryRepository) Save(item Item) (Item, error) {
	r.lastID++
	stored := item
	stored.ID = r.lastID
	stored.AcceptanceCriteria = slices.Clone(item.AcceptanceCriteria)
	r.items = append(r.items, stored)

	fresh := r.items[len(r.items)-1]
	fresh.AcceptanceCriteria = slices.Clone(stored.AcceptanceCriteria)
	return fresh, nil
}
