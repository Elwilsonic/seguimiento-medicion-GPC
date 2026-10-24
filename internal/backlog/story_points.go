package backlog

import "errors"

// ErrInvalidStoryPoints indica que el valor no está en la escala Fibonacci
// acordada (RN-20.1 y RN-20.2).
var ErrInvalidStoryPoints = errors.New("los story points deben ser 1, 2, 3, 5, 8, 13 o 21")

// ErrItemFinished indica que el ítem está TERMINADO y sus story points no se
// pueden cambiar (RN-20.5).
var ErrItemFinished = errors.New("no se pueden cambiar los story points de un ítem terminado")

// StoryPointScale devuelve la escala de story points (HU-20).
// Stub de la fase RED: todavía no devuelve la escala.
func StoryPointScale() []int {
	return nil
}

// AssignStoryPoints asigna los story points de un ítem (HU-20).
// Stub de la fase RED: todavía no valida ni guarda nada.
func (m *Manager) AssignStoryPoints(id, points int) (Item, error) {
	return Item{}, nil
}
