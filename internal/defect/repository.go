package defect

// Repository es el almacenamiento de los defectos. La lógica de negocio
// accede a los datos solo a través de esta interfaz. Por ahora HU-29 solo
// necesita guardar; las consultas y los cambios de estado llegan con HU-30 y
// HU-31.
type Repository interface {
	// Save guarda un defecto nuevo, le asigna el ID y devuelve una copia.
	Save(Defect) (Defect, error)
}
