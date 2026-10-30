package effort

import "sync"

// MemoryRepository guarda los esfuerzos en memoria.
type MemoryRepository struct {
	mu      sync.Mutex
	efforts []Effort
	lastID  int
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{}
}

// Save asigna el siguiente ID correlativo global, guarda el esfuerzo y
// devuelve una copia. Effort no tiene slices ni mapas, así que copiar el
// struct alcanza (RN-26.9).
func (r *MemoryRepository) Save(effort Effort) (Effort, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.lastID++
	stored := effort
	stored.ID = r.lastID
	r.efforts = append(r.efforts, stored)
	return stored, nil
}

// ListByItem devuelve los esfuerzos de una historia, ordenados por ID.
func (r *MemoryRepository) ListByItem(itemID int) ([]Effort, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	efforts := make([]Effort, 0)
	for _, effort := range r.efforts {
		if effort.ItemID == itemID {
			efforts = append(efforts, effort)
		}
	}
	return efforts, nil
}

// ListBySprint devuelve los esfuerzos de un sprint, ordenados por ID.
func (r *MemoryRepository) ListBySprint(sprintID int) ([]Effort, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	efforts := make([]Effort, 0)
	for _, effort := range r.efforts {
		if effort.SprintID == sprintID {
			efforts = append(efforts, effort)
		}
	}
	return efforts, nil
}

// GetByID devuelve una copia del esfuerzo con ese ID o ErrEffortNotFound si
// no existe, incluidos los IDs 0 y negativos (HU-28).
func (r *MemoryRepository) GetByID(id int) (Effort, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, effort := range r.efforts {
		if effort.ID == id {
			return effort, nil
		}
	}
	return Effort{}, ErrEffortNotFound
}

// Update reemplaza el esfuerzo con ese ID y devuelve una copia (HU-28). Si no
// existe, devuelve ErrEffortNotFound y no modifica nada.
func (r *MemoryRepository) Update(effort Effort) (Effort, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i := range r.efforts {
		if r.efforts[i].ID == effort.ID {
			r.efforts[i] = effort
			return effort, nil
		}
	}
	return Effort{}, ErrEffortNotFound
}
