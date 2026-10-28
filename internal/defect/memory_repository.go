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

// GetByID (HU-30). Firma vacía del RED: se implementa en el GREEN.
func (r *MemoryRepository) GetByID(id int) (Defect, error) {
	return Defect{}, nil
}

// Update (HU-30). Firma vacía del RED: se implementa en el GREEN.
func (r *MemoryRepository) Update(defect Defect) (Defect, error) {
	return Defect{}, nil
}
