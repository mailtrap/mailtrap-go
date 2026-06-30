package mailtrap_test

import (
	"context"
	"net/http"
	"testing"
)

func TestProjects_List(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("GET /api/projects", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`[{"id":1,"name":"Demo"}]`))
	})

	projects, _, err := client.Projects.List(context.Background())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(projects) != 1 || projects[0].ID != 1 || projects[0].Name != "Demo" {
		t.Fatalf("projects = %+v", projects)
	}
}

func TestProjects_Get(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("GET /api/projects/5", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"id":5,"name":"Five"}`))
	})

	project, _, err := client.Projects.Get(context.Background(), 5)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if project.ID != 5 || project.Name != "Five" {
		t.Errorf("project = %+v", project)
	}
}

func TestProjects_Create(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("POST /api/projects", func(w http.ResponseWriter, r *http.Request) {
		wantJSONBody(t, r, `{"project":{"name":"New"}}`)
		_, _ = w.Write([]byte(`{"id":9,"name":"New"}`))
	})

	project, _, err := client.Projects.Create(context.Background(), "New")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if project.ID != 9 {
		t.Errorf("project = %+v", project)
	}
}

func TestProjects_Update(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("PATCH /api/projects/9", func(w http.ResponseWriter, r *http.Request) {
		wantJSONBody(t, r, `{"project":{"name":"Renamed"}}`)
		_, _ = w.Write([]byte(`{"id":9,"name":"Renamed"}`))
	})

	project, _, err := client.Projects.Update(context.Background(), 9, "Renamed")
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if project.Name != "Renamed" {
		t.Errorf("project = %+v", project)
	}
}

func TestProjects_Delete(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("DELETE /api/projects/9", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"id":9}`))
	})

	if _, err := client.Projects.Delete(context.Background(), 9); err != nil {
		t.Fatalf("Delete: %v", err)
	}
}
