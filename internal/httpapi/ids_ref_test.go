package httpapi_test

import (
	"strconv"
	"strings"
	"testing"
)

// HU-52 · Identificadores UUID del resto de las entidades (Issue #197).
//
// idRef es el id de un ítem, sprint, defecto, esfuerzo o sesión en una
// respuesta (id, sprint_id, item_id, sprint_deteccion_id o
// sprint_resolucion_id), como texto. Acepta un texto JSON o un número, como
// projectRef en HU-51, así que los tests adaptados verifican lo mismo que
// antes sin depender del tipo: hoy la API devuelve enteros, y cuando una
// entidad pase a UUID estos campos siguen decodificando. Que sea un UUID y no
// un número lo verifican los tests de HU-52 de cada entidad.
type idRef string

func (r *idRef) UnmarshalJSON(b []byte) error {
	*r = idRef(strings.Trim(string(b), `"`))
	return nil
}

// idInt devuelve el id como entero, para los tests que necesitan el número
// (por ejemplo, para comparar listas de IDs correlativos). Hoy la API
// devuelve enteros; cada responsable lo adapta cuando su entidad pase a UUID.
func idInt(t *testing.T, r idRef) int {
	t.Helper()
	n, err := strconv.Atoi(string(r))
	if err != nil {
		t.Fatalf("id = %q, se esperaba un número entero", string(r))
	}
	return n
}
