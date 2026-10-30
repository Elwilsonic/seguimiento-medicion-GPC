package poker

import (
	"slices"
	"testing"
)

// HU-24 · Detectar diferencias y repetir ronda (Issue #42)

// expectGroup verifica los puntos y los integrantes de un extremo.
func expectGroup(t *testing.T, what string, got VoteGroup, points int, voters ...string) {
	t.Helper()
	if got.Points != points || !slices.Equal(got.Voters, voters) {
		t.Errorf("%s = %+v, se esperaba %d de %v", what, got, points, voters)
	}
}

// CA-24.1 · Con votos distintos se informa la diferencia, el menor y el mayor.
func TestCompareVotes_VotosDistintosInformaDiferenciaMenorYMayor(t *testing.T) {
	got := CompareVotes([]Vote{{Voter: "Ana", Points: 3}, {Voter: "Bruno", Points: 8}, {Voter: "Carla", Points: 8}})

	if !got.HasDifferences {
		t.Error("HasDifferences = false, se esperaba true")
	}
	expectGroup(t, "menor", got.Lowest, 3, "Ana")
	expectGroup(t, "mayor", got.Highest, 8, "Bruno", "Carla")
}

// CA-24.1 · Con todos los votos iguales no hay diferencia y el menor y el
// mayor son iguales, con todos los integrantes.
func TestCompareVotes_TodosIgualesNoInformaDiferencia(t *testing.T) {
	got := CompareVotes([]Vote{{Voter: "Ana", Points: 5}, {Voter: "Bruno", Points: 5}})

	if got.HasDifferences {
		t.Error("HasDifferences = true, se esperaba false")
	}
	expectGroup(t, "menor", got.Lowest, 5, "Ana", "Bruno")
	expectGroup(t, "mayor", got.Highest, 5, "Ana", "Bruno")
}

// CA-24.1 · Un único voto no tiene diferencias.
func TestCompareVotes_UnSoloVotoNoInformaDiferencia(t *testing.T) {
	got := CompareVotes([]Vote{{Voter: "Ana", Points: 13}})

	if got.HasDifferences {
		t.Error("HasDifferences = true, se esperaba false")
	}
	expectGroup(t, "menor", got.Lowest, 13, "Ana")
	expectGroup(t, "mayor", got.Highest, 13, "Ana")
}

// CA-24.1 · Se informan todos los integrantes de cada extremo, en el orden de
// los votos.
func TestCompareVotes_VariosIntegrantesConElMismoExtremo(t *testing.T) {
	got := CompareVotes([]Vote{
		{Voter: "Carla", Points: 2},
		{Voter: "Ana", Points: 13},
		{Voter: "Eva", Points: 5},
		{Voter: "Bruno", Points: 2},
		{Voter: "Diego", Points: 13},
	})

	if !got.HasDifferences {
		t.Error("HasDifferences = false, se esperaba true")
	}
	expectGroup(t, "menor", got.Lowest, 2, "Carla", "Bruno")
	expectGroup(t, "mayor", got.Highest, 13, "Ana", "Diego")
}

// CA-24.1 · Sin votos no hay diferencia; los puntos son 0 y las listas de
// integrantes están vacías, no nil.
func TestCompareVotes_SinVotosDevuelveComparacionVacia(t *testing.T) {
	for _, votes := range [][]Vote{nil, {}} {
		got := CompareVotes(votes)

		if got.HasDifferences || got.Lowest.Points != 0 || got.Highest.Points != 0 {
			t.Errorf("CompareVotes(%v) = %+v, se esperaba sin diferencias y puntos 0", votes, got)
		}
		if got.Lowest.Voters == nil || got.Highest.Voters == nil || len(got.Lowest.Voters) != 0 || len(got.Highest.Voters) != 0 {
			t.Errorf("CompareVotes(%v) = %+v, se esperaban listas de integrantes vacías no nil", votes, got)
		}
	}
}

// CA-24.6 · Las listas devueltas son nuevas: modificarlas no altera los votos
// ni el otro extremo.
func TestCompareVotes_DevuelveListasNuevas(t *testing.T) {
	votes := []Vote{{Voter: "Ana", Points: 5}, {Voter: "Bruno", Points: 5}}

	got := CompareVotes(votes)
	if len(got.Lowest.Voters) != 2 || len(got.Highest.Voters) != 2 {
		t.Fatalf("comparación = %+v, se esperaban dos integrantes en cada extremo", got)
	}
	got.Lowest.Voters[0] = "Intruso"
	got.Lowest.Voters = append(got.Lowest.Voters, "Otro")

	if want := []string{"Ana", "Bruno"}; !slices.Equal(got.Highest.Voters, want) {
		t.Errorf("mayor = %v, modificar el menor no debía cambiarlo", got.Highest.Voters)
	}
	if want := []Vote{{Voter: "Ana", Points: 5}, {Voter: "Bruno", Points: 5}}; !slices.Equal(votes, want) {
		t.Errorf("votos = %+v, la comparación no debía modificarlos", votes)
	}
}
