package backlog

import (
	"errors"
	"slices"
)

// ErrInvalidStoryPoints indica que el valor no está en la escala Fibonacci
// acordada (RN-20.1 y RN-20.2).
var ErrInvalidStoryPoints = errors.New("los story points deben ser 1, 2, 3, 5, 8, 13 o 21")

// ErrItemFinished indica que el ítem está TERMINADO y sus story points no se
// pueden cambiar (RN-20.5).
var ErrItemFinished = errors.New("no se pueden cambiar los story points de un ítem terminado")

// storyPointScale es la única definición de la escala Fibonacci acordada, en
// orden ascendente (RN-20.1). Es un arreglo no exportado para que nadie la
// modifique; afuera se consulta con StoryPointScale.
var storyPointScale = [...]int{1, 2, 3, 5, 8, 13, 21}

// StoryPointScale devuelve una lista nueva con la escala de story points en
// orden ascendente. Planning Poker (HU-21 y HU-22) tiene que usar esta función
// para las cartas, sin repetir los valores. Modificar la lista devuelta no
// cambia la escala.
func StoryPointScale() []int {
	return slices.Clone(storyPointScale[:])
}

// AssignStoryPoints asigna los story points de un ítem existente (HU-20) y
// devuelve una copia. Se puede reestimar: el valor nuevo reemplaza al anterior
// (RN-20.3). Solo cambian los story points (RN-20.6).
//
// El orden de validación es: escala, sin consultar el repositorio; después,
// con el mutex, existencia (RN-20.4), estado TERMINADO (RN-20.5) y escritura.
// Si falla cualquiera, no se escribe nada.
func (m *Manager) AssignStoryPoints(id, points int) (Item, error) {
	// RN-20.1 y RN-20.2: solo los valores de la escala; el 0 no se asigna.
	if !slices.Contains(storyPointScale[:], points) {
		return Item{}, ErrInvalidStoryPoints
	}

	// La lectura, la verificación del estado y la escritura van bajo el mutex,
	// para no pisar un Update o un ChangeStatus simultáneo (CA-20.7).
	m.mu.Lock()
	defer m.mu.Unlock()

	current, err := m.repo.GetByID(id)
	if err != nil {
		return Item{}, err
	}

	// RN-20.5: los story points de un ítem terminado no se cambian
	// directamente; si se reabre con HU-08, vuelve a ser estimable.
	if current.Status == StatusFinished {
		return Item{}, ErrItemFinished
	}

	current.StoryPoints = points
	return m.repo.Update(current)
}
