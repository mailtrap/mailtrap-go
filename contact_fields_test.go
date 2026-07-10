package mailtrap_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/mailtrap/mailtrap-go"
)

func TestContactFields_List(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("GET /api/contacts/fields", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`[{"id":6730,"name":"First name","data_type":"text","merge_tag":"first_name"}]`))
	})

	fields, _, err := client.ContactFields.List(context.Background())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(fields) != 1 || fields[0].DataType != mailtrap.ContactFieldTypeText {
		t.Fatalf("fields = %+v", fields)
	}
}

func TestContactFields_Get(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("GET /api/contacts/fields/6730", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"id":6730,"name":"First name","data_type":"text","merge_tag":"first_name"}`))
	})

	field, _, err := client.ContactFields.Get(context.Background(), 6730)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if field.MergeTag != "first_name" {
		t.Errorf("field = %+v", field)
	}
}

func TestContactFields_Create(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("POST /api/contacts/fields", func(w http.ResponseWriter, r *http.Request) {
		wantJSONBody(t, r, `{"name":"Age","data_type":"integer","merge_tag":"age"}`)
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":6731,"name":"Age","data_type":"integer","merge_tag":"age"}`))
	})

	field, _, err := client.ContactFields.Create(context.Background(), &mailtrap.CreateContactFieldRequest{
		Name:     "Age",
		DataType: mailtrap.ContactFieldTypeInteger,
		MergeTag: "age",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if field.ID != 6731 || field.DataType != mailtrap.ContactFieldTypeInteger {
		t.Errorf("field = %+v", field)
	}
}

func TestContactFields_Update(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("PATCH /api/contacts/fields/6731", func(w http.ResponseWriter, r *http.Request) {
		wantJSONBody(t, r, `{"name":"Years","merge_tag":"years"}`)
		_, _ = w.Write([]byte(`{"id":6731,"name":"Years","data_type":"integer","merge_tag":"years"}`))
	})

	field, _, err := client.ContactFields.Update(context.Background(), 6731, &mailtrap.UpdateContactFieldRequest{
		Name:     "Years",
		MergeTag: "years",
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if field.Name != "Years" || field.MergeTag != "years" {
		t.Errorf("field = %+v", field)
	}
}

func TestContactFields_Delete(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("DELETE /api/contacts/fields/6731", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	resp, err := client.ContactFields.Delete(context.Background(), 6731)
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("status = %d", resp.StatusCode)
	}
}
