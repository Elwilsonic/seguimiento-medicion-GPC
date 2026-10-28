package defect

import "sync"

// MemoryRepository guarda los defectos en memoria.
type MemoryRepository struct {
	mu      sync.Mutex
	defects []Defect
	lastID  int
}

// NewMemoryRepository crea un repositorio en memoria vacío.
func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{}
}

// Save asigna el siguiente ID correlativo global, guarda una copia del defecto
// y devuelve otra copia. Defect no tiene slices ni mapas, así que copiar el
// struct alcanza (RN-29.8 y RN-29.9).
func (r *MemoryRepository) Save(defect Defect) (Defect, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.lastID++
	stored := defect
	stored.ID = r.lastID
	r.defects = append(r.defects, stored)
	return stored, nil
}

// GetByID devuelve una copia del defecto con ese ID o ErrDefectNotFound si
// no existe, incluidos los IDs 0 y negativos (RN-30.3 y RN-30.10).
func (r *MemoryRepository) GetByID(id int) (Defect, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, defect := range r.defects {
		if defect.ID == id {
			return defect, nil
		}
	}
	return Defect{}, ErrDefectNotFound
}

// Update reemplaza el defecto con ese ID y devuelve una copia (RN-30.9 y
// RN-30.10). Si no existe, devuelve ErrDefectNotFound y no modifica nada.
func (r *MemoryRepository) Update(defect Defect) (Defect, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i := range r.defects {
		if r.defects[i].ID == defect.ID {
			r.defects[i] = defect
			return defect, nil
		}
	}
	return Defect{}, ErrDefectNotFound
}
