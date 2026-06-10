package graphql_test

import (
	"context"
	"encoding/json"
	"testing"

	gql "github.com/graphql-go/graphql"
	"github.com/tekindet/matryoshka/internal/domain"
	"github.com/tekindet/matryoshka/internal/graphql"
)

type mockManager struct {
	projects []*domain.Project
	services []*domain.Service
}

func (m *mockManager) CreateProject(ctx context.Context, name, description string) (*domain.Project, error) {
	proj := &domain.Project{ID: "proj-1", Name: name, Description: description}
	m.projects = append(m.projects, proj)
	return proj, nil
}

func (m *mockManager) GetProject(ctx context.Context, id string) (*domain.Project, error) {
	for _, p := range m.projects {
		if p.ID == id {
			return p, nil
		}
	}
	return nil, nil
}

func (m *mockManager) ListProjects(ctx context.Context) ([]*domain.Project, error) {
	return m.projects, nil
}

func (m *mockManager) CreateService(ctx context.Context, projectID, name, svcType string) (*domain.Service, error) {
	svc := &domain.Service{
		ID:        "svc-1",
		ProjectID: projectID,
		Name:      name,
		Type:      svcType,
		Status:    domain.ServiceStatusRunning,
	}
	m.services = append(m.services, svc)
	return svc, nil
}

func (m *mockManager) ListServices(ctx context.Context, projectID string) ([]*domain.Service, error) {
	return m.services, nil
}

func execute(t *testing.T, s gql.Schema, query string) map[string]interface{} {
	t.Helper()
	r := gql.Do(gql.Params{Schema: s, RequestString: query})
	if r.HasErrors() {
		t.Fatalf("unexpected errors: %v", r.Errors)
	}
	var result map[string]interface{}
	b, _ := json.Marshal(r.Data)
	json.Unmarshal(b, &result)
	return result
}

func TestCreateProjectMutation(t *testing.T) {
	mgr := &mockManager{}
	r := graphql.New(mgr)
	s, err := gql.NewSchema(gql.SchemaConfig{
		Query:    r.QueryType(),
		Mutation: r.MutationType(),
	})
	if err != nil {
		t.Fatal(err)
	}

	query := `mutation { createProject(input: { name: "test-proj", description: "a test project" }) { id name description } }`
	result := execute(t, s, query)

	proj := result["createProject"].(map[string]interface{})
	if proj["name"] != "test-proj" {
		t.Errorf("expected name test-proj, got %v", proj["name"])
	}
	if proj["description"] != "a test project" {
		t.Errorf("expected description 'a test project', got %v", proj["description"])
	}
}

func TestCreateServiceMutation(t *testing.T) {
	mgr := &mockManager{}
	r := graphql.New(mgr)
	s, err := gql.NewSchema(gql.SchemaConfig{
		Query:    r.QueryType(),
		Mutation: r.MutationType(),
	})
	if err != nil {
		t.Fatal(err)
	}

	query := `mutation { createService(input: { projectId: "proj-1", name: "my-redis", type: "redis" }) { id projectId name type status } }`
	result := execute(t, s, query)

	svc := result["createService"].(map[string]interface{})
	if svc["name"] != "my-redis" {
		t.Errorf("expected name my-redis, got %v", svc["name"])
	}
	if svc["type"] != "redis" {
		t.Errorf("expected type redis, got %v", svc["type"])
	}
	if svc["status"] != "running" {
		t.Errorf("expected status running, got %v", svc["status"])
	}
}

func TestProjectQuery(t *testing.T) {
	mgr := &mockManager{
		projects: []*domain.Project{
			{ID: "p1", Name: "alpha", Description: "first project"},
		},
	}
	r := graphql.New(mgr)
	s, err := gql.NewSchema(gql.SchemaConfig{
		Query:    r.QueryType(),
		Mutation: r.MutationType(),
	})
	if err != nil {
		t.Fatal(err)
	}

	query := `query { project(id: "p1") { id name description } }`
	result := execute(t, s, query)

	proj := result["project"].(map[string]interface{})
	if proj["id"] != "p1" {
		t.Errorf("expected id p1, got %v", proj["id"])
	}
}

func TestProjectsQuery(t *testing.T) {
	mgr := &mockManager{
		projects: []*domain.Project{
			{ID: "p1", Name: "alpha", Description: "first"},
			{ID: "p2", Name: "beta", Description: "second"},
		},
	}
	r := graphql.New(mgr)
	s, err := gql.NewSchema(gql.SchemaConfig{
		Query:    r.QueryType(),
		Mutation: r.MutationType(),
	})
	if err != nil {
		t.Fatal(err)
	}

	query := `query { projects { id name } }`
	result := execute(t, s, query)

	projects := result["projects"].([]interface{})
	if len(projects) != 2 {
		t.Fatalf("expected 2 projects, got %d", len(projects))
	}
}
