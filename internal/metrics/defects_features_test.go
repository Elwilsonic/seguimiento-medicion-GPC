package metrics

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/cucumber/godog"

	"seguimiento-medicion-gpc/internal/backlog"
	"seguimiento-medicion-gpc/internal/defect"
	"seguimiento-medicion-gpc/internal/project"
	"seguimiento-medicion-gpc/internal/sprint"
)

// Ejecuta los escenarios BDD de
// features/HU-47-defectos-detectados-resueltos.feature con godog.
func TestFeatures_HU47DefectosDetectadosResueltos(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: func(ctx *godog.ScenarioContext) { initializeDefectsScenario(ctx, t) },
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"../../features/HU-47-defectos-detectados-resueltos.feature"},
			Strict:   true,
			TestingT: t,
		},
	}
	if suite.Run() != 0 {
		t.Fatal("fallaron escenarios BDD de HU-47")
	}
}

// defectsScenario guarda el estado de un escenario de HU-47.
type defectsScenario struct {
	t        *testing.T
	env      *metricsEnv
	projects map[string]int
	sprints  map[string]sprint.Sprint

	defects Defects
	err     error

	defectsBefore []defect.Defect
	projectID     int
}

func initializeDefectsScenario(ctx *godog.ScenarioContext, t *testing.T) {
	s := &defectsScenario{t: t}
	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		s.env = newMetricsEnv()
		s.projects = map[string]int{}
		s.sprints = map[string]sprint.Sprint{}
		s.defects = Defects{}
		s.err = nil
		s.defectsBefore = nil
		s.projectID = 0
		return ctx, nil
	})

	ctx.Step(`^que existe un proyecto con nombre "([^"]*)"$`, s.existingProject)
	ctx.Step(`^que existe un sprint abierto "([^"]*)" en el proyecto "([^"]*)"$`, s.existingOpenSprint)
	ctx.Step(`^que existe un sprint cerrado "([^"]*)" en el proyecto "([^"]*)"$`, s.existingClosedSprint)
	ctx.Step(`^que el proyecto "([^"]*)" tiene los defectos:$`, s.projectDefects)
	ctx.Step(`^consulto los defectos del sprint "([^"]*)"$`, s.querySprintByName)
	ctx.Step(`^consulto los defectos del sprint con el ID (-?\d+)$`, s.querySprint)
	ctx.Step(`^consulto los defectos del proyecto "([^"]*)"$`, s.queryProjectByName)
	ctx.Step(`^consulto los defectos del proyecto con el ID (-?\d+)$`, s.queryProject)
	ctx.Step(`^hay (\d+) defectos detectados y (\d+) resueltos$`, s.defectsAre)
	ctx.Step(`^la consulta falla porque el sprint no existe$`, s.failsSprintNotFound)
	ctx.Step(`^la consulta falla porque el proyecto no existe$`, s.failsProjectNotFound)
	ctx.Step(`^los defectos del proyecto no cambiaron$`, s.nothingChanged)
}

func (s *defectsScenario) existingProject(name string) error {
	s.projects[name] = s.env.project(s.t, name)
	if s.projectID == 0 {
		s.projectID = s.projects[name]
	}
	return nil
}

func (s *defectsScenario) existingOpenSprint(name, projectName string) error {
	s.sprints[name] = s.env.openSprint(s.t, s.projects[projectName], name)
	return nil
}

func (s *defectsScenario) existingClosedSprint(name, projectName string) error {
	s.sprints[name] = s.env.closedSprint(s.t, s.projects[projectName], name)
	return nil
}

// sprintID devuelve el ID del sprint de la tabla; "-" es ningún sprint (0).
func (s *defectsScenario) sprintID(name string) (int, error) {
	if name == "-" {
		return 0, nil
	}
	sp, ok := s.sprints[name]
	if !ok {
		return 0, fmt.Errorf("no existe el sprint %q en el escenario", name)
	}
	return sp.ID, nil
}

// projectDefects guarda los defectos de la tabla (descripcion, detectado_en,
// estado, resuelto_en), cada uno sobre una historia nueva del proyecto.
func (s *defectsScenario) projectDefects(projectName string, table *godog.Table) error {
	projectID := s.projects[projectName]
	for _, row := range table.Rows[1:] {
		detection, err := s.sprintID(row.Cells[1].Value)
		if err != nil {
			return err
		}
		resolution, err := s.sprintID(row.Cells[3].Value)
		if err != nil {
			return err
		}
		it := s.env.item(s.t, projectID, "Historia de "+row.Cells[0].Value, 0, backlog.StatusPending, 0)
		if _, err := s.env.defectRepo.Save(defect.Defect{
			Description:        row.Cells[0].Value,
			Severity:           defect.SeverityMedium,
			Status:             defect.Status(row.Cells[2].Value),
			ItemID:             it.ID,
			DetectionSprintID:  detection,
			ResolutionSprintID: resolution,
		}); err != nil {
			return fmt.Errorf("guardar el defecto %q: %w", row.Cells[0].Value, err)
		}
	}
	return nil
}

// snapshot guarda, antes de la primera consulta, los defectos guardados para
// comparar después (CA-47.5). Lee el repositorio en memoria directamente.
func (s *defectsScenario) snapshot() {
	if s.defectsBefore != nil {
		return
	}
	s.defectsBefore, _ = s.env.defectRepo.List()
}

func (s *defectsScenario) querySprintByName(name string) error {
	return s.querySprint(s.sprints[name].ID)
}

func (s *defectsScenario) querySprint(id int) error {
	s.snapshot()
	s.defects, s.err = s.env.metrics.SprintDefects(id)
	return nil
}

func (s *defectsScenario) queryProjectByName(name string) error {
	return s.queryProject(s.projects[name])
}

func (s *defectsScenario) queryProject(id int) error {
	s.snapshot()
	s.defects, s.err = s.env.metrics.ProjectDefects(id)
	return nil
}

func (s *defectsScenario) defectsAre(detected, resolved int) error {
	if s.err != nil {
		return fmt.Errorf("la consulta falló: %v", s.err)
	}
	if want := (Defects{Detected: detected, Resolved: resolved}); s.defects != want {
		return fmt.Errorf("defectos = %+v, se esperaba %+v", s.defects, want)
	}
	return nil
}

func (s *defectsScenario) failsSprintNotFound() error {
	if !errors.Is(s.err, sprint.ErrSprintNotFound) {
		return fmt.Errorf("err = %v, se esperaba %v", s.err, sprint.ErrSprintNotFound)
	}
	return nil
}

func (s *defectsScenario) failsProjectNotFound() error {
	if !errors.Is(s.err, project.ErrProjectNotFound) {
		return fmt.Errorf("err = %v, se esperaba %v", s.err, project.ErrProjectNotFound)
	}
	return nil
}

func (s *defectsScenario) nothingChanged() error {
	if s.err != nil {
		return fmt.Errorf("la consulta falló: %v", s.err)
	}
	after, err := s.env.defectRepo.List()
	if err != nil {
		return err
	}
	if len(s.defectsBefore) == 0 || !reflect.DeepEqual(after, s.defectsBefore) {
		return fmt.Errorf("la consulta modificó los defectos: antes %+v, después %+v", s.defectsBefore, after)
	}
	return nil
}
