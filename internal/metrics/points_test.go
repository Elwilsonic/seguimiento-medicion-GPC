package metrics

import "testing"

// HU-32 · Calcular Story Points planificados, completados y velocidad (Issue #50)
//
// Cálculo de la velocidad (M3 de docs/metricas.md) sobre los sprints
// cerrados, sin repositorios (RN-32.5 a RN-32.7).

// CA-32.5 y CA-32.6 · Promedio de los Story Points completados de los últimos
// 3 sprints cerrados, por ID descendente, o de todos si hay menos.
func TestVelocity_UltimosTresSprintsCerrados(t *testing.T) {
	cases := []struct {
		name     string
		closed   []closedSprint
		want     float64
		wantUsed int
	}{
		{"sin sprints cerrados", nil, 0, 0},
		{"un sprint cerrado", []closedSprint{{ID: 1, Completed: 8}}, 8, 1},
		{"ejemplo de TEC-03", []closedSprint{{ID: 1, Completed: 8}, {ID: 2, Completed: 10}, {ID: 3, Completed: 12}}, 10, 3},
		{"cuatro cerrados usa los IDs 4, 3 y 2", []closedSprint{{ID: 1, Completed: 2}, {ID: 2, Completed: 8}, {ID: 3, Completed: 10}, {ID: 4, Completed: 12}}, 10, 3},
		{"el orden de entrada no importa", []closedSprint{{ID: 4, Completed: 12}, {ID: 1, Completed: 2}, {ID: 3, Completed: 10}, {ID: 2, Completed: 8}}, 10, 3},
		{"un sprint cerrado sin ítems cuenta con 0", []closedSprint{{ID: 1, Completed: 0}, {ID: 2, Completed: 10}}, 5, 2},
		{"sin redondear", []closedSprint{{ID: 1, Completed: 10}, {ID: 2, Completed: 15}, {ID: 3, Completed: 12}}, 37.0 / 3, 3},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, used := velocity(tc.closed)
			if got != tc.want || used != tc.wantUsed {
				t.Errorf("velocity = %v con %d sprints, se esperaba %v con %d", got, used, tc.want, tc.wantUsed)
			}
		})
	}
}

// RN-32.8 y CA-32.7 · Calcular no modifica nada: velocity no reordena la
// lista que recibe.
func TestVelocity_NoModificaLaLista(t *testing.T) {
	closed := []closedSprint{{ID: 3, Completed: 10}, {ID: 1, Completed: 2}, {ID: 2, Completed: 8}}
	want := append([]closedSprint(nil), closed...)

	velocity(closed)

	for i := range want {
		if closed[i] != want[i] {
			t.Fatalf("velocity modificó la lista: %+v, se esperaba %+v", closed, want)
		}
	}
}
