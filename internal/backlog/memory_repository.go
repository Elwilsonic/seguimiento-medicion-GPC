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

// ListByProject devuelve una lista nueva con copias de los ítems del
// proyecto indicado, en orden de guardado, que es el orden de ID ascendente
// (RN-06.6). Solo aparecen los ítems de ese proyecto (RN-06.3) y cada ítem
// tiene su propia copia de la lista de criterios (RN-06.5). Un proyecto sin
// ítems devuelve una lista vacía, no nil (RN-06.2).
func (r *MemoryRepository) ListByProject(projectID int) ([]Item, error) {
	result := make([]Item, 0, len(r.items))
	for _, item := range r.items {
		if item.ProjectID == projectID {
			item.AcceptanceCriteria = slices.Clone(item.AcceptanceCriteria)
			result = append(result, item)
		}
	}
	return result, nil
}
