package backlog

import (
	"errors"
	"fmt"
	"slices"
	"sync"
	"testing"

	"seguimiento-medicion-gpc/internal/project"
)

// HU-08 · Cambiar el estado de un ítem (Issue #20)

// seedStatusItem guarda un ítem completo con ese estado, para verificar que el
// cambio de estado no toca el resto de los campos (RN-08.5).
func seedStatusItem(t *testing.T, repo *MemoryRepository, status Status) Item {
	t.Helper()
	item, err := repo.Save(Item{
		ProjectID:          1,
		Title:              validTitle,
		Description:        "Formulario de acceso",
		Priority:           PriorityHigh,
		AcceptanceCriteria: []string{"Login OK", "Carga correcta"},
		Status:             status,
		StoryPoints:        5,
		EstimatedHours:     6,
		SprintID:           3,
	})
	if err != nil {
		t.Fatalf("guardar el ítem de prueba: %v", err)
	}
	return item
}

// assertOtherFieldsUnchanged compara todos los campos menos el estado.
func assertOtherFieldsUnchanged(t *testing.T, got, want Item) {
	t.Helper()
	if got.ID != want.ID || got.ProjectID != want.ProjectID || got.Title != want.Title ||
		got.Description != want.Description || got.Priority != want.Priority ||
		got.StoryPoints != want.StoryPoints || got.EstimatedHours != want.EstimatedHours ||
		got.SprintID != want.SprintID {
		t.Errorf("el ítem cambió en un campo que no es el estado: %+v, se esperaba %+v", got, want)
	}
	if !slices.Equal(got.AcceptanceCriteria, want.AcceptanceCriteria) {
		t.Errorf("criterios = %q, se esperaban %q", got.AcceptanceCriteria, want.AcceptanceCriteria)
	}
}

// CA-08.1
func TestManagerChangeStatus_PendienteAEnCursoConservaElRestoDeLosCampos(t *testing.T) {
	manager, repo, _, _ := newTestManager()
	created := seedStatusItem(t, repo, StatusPending)

	changed, err := manager.ChangeStatus(created.ID, "EN_CURSO")

	if err != nil {
		t.Fatalf("ChangeStatus() devolvió un error inesperado: %v", err)
	}
	if changed.Status != StatusInProgress {
		t.Errorf("Status = %q, se esperaba %q", changed.Status, StatusInProgress)
	}
	assertOtherFieldsUnchanged(t, changed, created)
	if saved := savedItem(t, repo, created.ID); saved.Status != StatusInProgress {
		t.Errorf("estado guardado = %q, se esperaba %q", saved.Status, StatusInProgress)
	}
}

// CA-08.2
func TestManagerChangeStatus_EnCursoATerminado(t *testing.T) {
	manager, repo, _, _ := newTestManager()
	created := seedStatusItem(t, repo, StatusInProgress)

	changed, err := manager.ChangeStatus(created.ID, "TERMINADO")

	if err != nil {
		t.Fatalf("ChangeStatus() devolvió un error inesperado: %v", err)
	}
	if changed.Status != StatusFinished {
		t.Errorf("Status = %q, se esperaba %q", changed.Status, StatusFinished)
	}
	if saved := savedItem(t, repo, created.ID); saved.Status != StatusFinished {
		t.Errorf("estado guardado = %q, se esperaba %q", saved.Status, StatusFinished)
	}
}

// CA-08.3
func TestManagerChangeStatus_VolverAtrasYReabrirSonValidos(t *testing.T) {
	cases := []struct {
		name   string
		origen Status
	}{
		{"volver atrás de EN_CURSO a PENDIENTE", StatusInProgress},
		{"reabrir de TERMINADO a EN_CURSO", StatusFinished},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			manager, repo, _, _ := newTestManager()
			created := seedStatusItem(t, repo, tc.origen)
			destino := StatusPending
			if tc.origen == StatusFinished {
				destino = StatusInProgress
			}

			changed, err := manager.ChangeStatus(created.ID, string(destino))

			if err != nil {
				t.Fatalf("ChangeStatus() devolvió un error inesperado: %v", err)
			}
			if changed.Status != destino {
				t.Errorf("Status = %q, se esperaba %q", changed.Status, destino)
			}
			if saved := savedItem(t, repo, created.ID); saved.Status != destino {
				t.Errorf("estado guardado = %q, se esperaba %q", saved.Status, destino)
			}
		})
	}
}

// CA-08.4
func TestManagerChangeStatus_RechazaSaltosDeDosEstados(t *testing.T) {
	cases := []struct {
		name    string
		origen  Status
		destino string
	}{
		{"PENDIENTE a TERMINADO", StatusPending, "TERMINADO"},
		{"TERMINADO a PENDIENTE", StatusFinished, "PENDIENTE"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			manager, repo, _, _ := newTestManager()
			created := seedStatusItem(t, repo, tc.origen)

			_, err := manager.ChangeStatus(created.ID, tc.destino)

			if !errors.Is(err, ErrInvalidTransition) {
				t.Errorf("error = %v, se esperaba ErrInvalidTransition", err)
			}
			if saved := savedItem(t, repo, created.ID); saved.Status != tc.origen {
				t.Errorf("estado guardado = %q, se esperaba %q: el ítem no debía cambiar", saved.Status, tc.origen)
			}
		})
	}
}

// CA-08.5
func TestManagerChangeStatus_AceptaElEstadoSinMayusculasNiEspacios(t *testing.T) {
	cases := []struct {
		name  string
		seed  Status
		input string
		want  Status
	}{
		{"minúsculas", StatusPending, "en_curso", StatusInProgress},
		{"mayúsculas mezcladas y espacios", StatusInProgress, "  termiNADO  ", StatusFinished},
		{"volver atrás normalizado", StatusInProgress, "pEnDiEnTe", StatusPending},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			manager, repo, _, _ := newTestManager()
			created := seedStatusItem(t, repo, tc.seed)

			changed, err := manager.ChangeStatus(created.ID, tc.input)

			if err != nil {
				t.Fatalf("ChangeStatus() devolvió un error inesperado: %v", err)
			}
			if changed.Status != tc.want {
				t.Errorf("Status = %q, se esperaba %q", changed.Status, tc.want)
			}
			if saved := savedItem(t, repo, created.ID); saved.Status != tc.want {
				t.Errorf("estado guardado = %q, se esperaba %q", saved.Status, tc.want)
			}
		})
	}
}

// CA-08.5
func TestManagerChangeStatus_RechazaEstadosInvalidos(t *testing.T) {
	for _, status := range []string{"", "   ", "BLOQUEADO", "en curso"} {
		t.Run("estado "+status, func(t *testing.T) {
			manager, repo, _, _ := newTestManager()
			created := seedStatusItem(t, repo, StatusPending)

			_, err := manager.ChangeStatus(created.ID, status)

			if !errors.Is(err, ErrInvalidStatus) {
				t.Errorf("error = %v, se esperaba ErrInvalidStatus", err)
			}
			if saved := savedItem(t, repo, created.ID); saved.Status != StatusPending {
				t.Errorf("estado guardado = %q, se esperaba %q: el ítem no debía cambiar", saved.Status, StatusPending)
			}
		})
	}
}

// Restricción "Orden de validación": el estado se valida antes de consultar el
// repositorio, así que sobre un ítem inexistente igual gana ErrInvalidStatus.
func TestManagerChangeStatus_EstadoInvalidoSobreItemInexistenteDaErrInvalidStatus(t *testing.T) {
	manager, _, _, _ := newTestManager()

	_, err := manager.ChangeStatus(999, "BLOQUEADO")

	if !errors.Is(err, ErrInvalidStatus) {
		t.Errorf("error = %v, se esperaba ErrInvalidStatus", err)
	}
}

// CA-08.6
func TestManagerChangeStatus_ItemInexistenteDaErrItemNotFound(t *testing.T) {
	for _, id := range []int{999, 0, -1} {
		t.Run(fmt.Sprintf("ID %d", id), func(t *testing.T) {
			manager, repo, _, _ := newTestManager()
			seedStatusItem(t, repo, StatusPending)

			changed, err := manager.ChangeStatus(id, "EN_CURSO")

			if !errors.Is(err, ErrItemNotFound) {
				t.Errorf("ID %d: error = %v, se esperaba ErrItemNotFound", id, err)
			}
			if changed.ID != 0 {
				t.Errorf("ítem devuelto = %+v, se esperaba un ítem vacío", changed)
			}
		})
	}
}

// CA-08.6
func TestManagerChangeStatus_SinItemsCreadosDaErrItemNotFound(t *testing.T) {
	manager, _, _, _ := newTestManager()

	_, err := manager.ChangeStatus(1, "EN_CURSO")

	if !errors.Is(err, ErrItemNotFound) {
		t.Errorf("error = %v, se esperaba ErrItemNotFound", err)
	}
}

// CA-08.7 (el repositorio falla al actualizar: si el no-op llamara a Update,
// el cambio devolvería errRepository)
func TestManagerChangeStatus_MismoEstadoEsNoOpYNoLlamaAUpdate(t *testing.T) {
	projects := project.NewManager(project.NewMemoryRepository())
	repo := NewMemoryRepository()
	created := seedStatusItem(t, repo, StatusPending)
	manager := NewManager(failingUpdateStatusRepository{repo}, projects)

	changed, err := manager.ChangeStatus(created.ID, " pendiente ")

	if err != nil {
		t.Fatalf("ChangeStatus() devolvió un error inesperado: %v", err)
	}
	if changed.Status != StatusPending {
		t.Errorf("Status = %q, se esperaba %q", changed.Status, StatusPending)
	}
	assertOtherFieldsUnchanged(t, changed, created)
	if saved := savedItem(t, repo, created.ID); saved.Status != StatusPending {
		t.Errorf("estado guardado = %q, se esperaba %q", saved.Status, StatusPending)
	}
}

// CA-08.8
func TestManagerChangeStatus_CambioQuedaGuardadoYElResultadoEsCopia(t *testing.T) {
	manager, repo, _, _ := newTestManager()
	created := seedStatusItem(t, repo, StatusPending)

	changed, err := manager.ChangeStatus(created.ID, "EN_CURSO")
	if err != nil {
		t.Fatalf("ChangeStatus() devolvió un error inesperado: %v", err)
	}

	changed.Status = StatusFinished
	changed.Title = "Otro título"
	changed.AcceptanceCriteria[0] = "Otro criterio"

	if saved := savedItem(t, repo, created.ID); saved.Status != StatusInProgress {
		t.Errorf("estado guardado = %q, se esperaba %q", saved.Status, StatusInProgress)
	}
	if saved := savedItem(t, repo, created.ID); saved.Title != created.Title {
		t.Errorf("título guardado = %q, se esperaba %q", saved.Title, created.Title)
	}
	if saved := savedItem(t, repo, created.ID); !slices.Equal(saved.AcceptanceCriteria, created.AcceptanceCriteria) {
		t.Errorf("criterios guardados = %q, se esperaban %q", saved.AcceptanceCriteria, created.AcceptanceCriteria)
	}

	requeried, err := manager.GetItem(created.ID)
	if err != nil {
		t.Fatalf("GetItem() devolvió un error inesperado: %v", err)
	}
	if requeried.Status != StatusInProgress {
		t.Errorf("estado reconsultado = %q, se esperaba %q", requeried.Status, StatusInProgress)
	}
}

// CA-08.9
func TestManagerChangeStatus_CambiosSimultaneosAlMismoEstado(t *testing.T) {
	manager, repo, _, _ := newTestManager()
	created := seedStatusItem(t, repo, StatusPending)

	const attempts = 50
	errs := make([]error, attempts)
	var wg sync.WaitGroup
	for i := range attempts {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, errs[i] = manager.ChangeStatus(created.ID, "EN_CURSO")
		}(i)
	}
	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Errorf("el cambio simultáneo %d falló: %v", i+1, err)
		}
	}
	if saved := savedItem(t, repo, created.ID); saved.Status != StatusInProgress {
		t.Errorf("estado guardado = %q, se esperaba %q", saved.Status, StatusInProgress)
	}
}

// CA-08.10 (error del repositorio al buscar)
func TestManagerChangeStatus_ErrorDelRepositorioAlBuscarSeDevuelve(t *testing.T) {
	projects := project.NewManager(project.NewMemoryRepository())
	manager := NewManager(failingRepository{getErr: errRepository}, projects)

	changed, err := manager.ChangeStatus(1, "EN_CURSO")

	if !errors.Is(err, errRepository) {
		t.Errorf("error = %v, se esperaba errRepository", err)
	}
	if changed.ID != 0 {
		t.Errorf("ítem devuelto = %+v, se esperaba un ítem vacío", changed)
	}
}

// CA-08.10 (error del repositorio al actualizar)
func TestManagerChangeStatus_ErrorDelRepositorioAlActualizarSeDevuelveYNoCambiaElEstado(t *testing.T) {
	projects := project.NewManager(project.NewMemoryRepository())
	repo := NewMemoryRepository()
	created := seedStatusItem(t, repo, StatusPending)
	manager := NewManager(failingUpdateStatusRepository{repo}, projects)

	_, err := manager.ChangeStatus(created.ID, "EN_CURSO")

	if !errors.Is(err, errRepository) {
		t.Errorf("error = %v, se esperaba errRepository", err)
	}
	if saved := savedItem(t, repo, created.ID); saved.Status != StatusPending {
		t.Errorf("estado guardado = %q, se esperaba %q", saved.Status, StatusPending)
	}
}
