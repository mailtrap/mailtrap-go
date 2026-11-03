package mailtrap_test

import (
	"context"
	"go/ast"
	"go/doc"
	"go/parser"
	"go/token"
	"net/http"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mailtrap/mailtrap-go"
)

func TestDomains_List(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("GET /api/domains", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"data":[{"id":435,"domain_name":"mailtrap.io","compliance_status":"compliant","dns_records":[{"key":"spf","type":"TXT","status":"pass"}]}]}`))
	})

	domains, _, err := client.Domains.List(context.Background())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(domains) != 1 || domains[0].ID != 435 || domains[0].DomainName != "mailtrap.io" {
		t.Fatalf("domains = %+v", domains)
	}
	if len(domains[0].DNSRecords) != 1 || domains[0].DNSRecords[0].Key != "spf" {
		t.Errorf("dns_records = %+v", domains[0].DNSRecords)
	}
}

func TestDomains_Get(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("GET /api/domains/435", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"id":435,"domain_name":"mailtrap.io","permissions":{"can_read":true,"can_update":true}}`))
	})

	domain, _, err := client.Domains.Get(context.Background(), 435)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if domain.ID != 435 || domain.Permissions == nil || !domain.Permissions.CanUpdate {
		t.Errorf("domain = %+v", domain)
	}
}

func TestDomains_Create(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("POST /api/domains", func(w http.ResponseWriter, r *http.Request) {
		wantJSONBody(t, r, `{"domain":{"domain_name":"example.com"}}`)
		_, _ = w.Write([]byte(`{"id":9,"domain_name":"example.com","compliance_status":"unverified_dns"}`))
	})

	domain, _, err := client.Domains.Create(context.Background(), "example.com")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if domain.ID != 9 || domain.ComplianceStatus != "unverified_dns" {
		t.Errorf("domain = %+v", domain)
	}
}

func TestDomains_Update(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("PATCH /api/domains/9", func(w http.ResponseWriter, r *http.Request) {
		wantJSONBody(t, r, `{"domain":{"open_tracking_enabled":false,"click_tracking_enabled":true}}`)
		_, _ = w.Write([]byte(`{"id":9,"open_tracking_enabled":false,"click_tracking_enabled":true}`))
	})

	domain, _, err := client.Domains.Update(context.Background(), 9, &mailtrap.UpdateDomainRequest{
		OpenTrackingEnabled:  mailtrap.Ptr(false),
		ClickTrackingEnabled: mailtrap.Ptr(true),
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if domain.OpenTrackingEnabled || !domain.ClickTrackingEnabled {
		t.Errorf("domain = %+v", domain)
	}
}

func TestDomains_Delete(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("DELETE /api/domains/9", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	if _, err := client.Domains.Delete(context.Background(), 9); err != nil {
		t.Fatalf("Delete: %v", err)
	}
}

func TestDomains_SendSetupInstructions(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("POST /api/domains/9/send_setup_instructions", func(w http.ResponseWriter, r *http.Request) {
		wantJSONBody(t, r, `{"email":"devops@example.com"}`)
		w.WriteHeader(http.StatusNoContent)
	})

	if _, err := client.Domains.SendSetupInstructions(context.Background(), 9, "devops@example.com"); err != nil {
		t.Fatalf("SendSetupInstructions: %v", err)
	}
}

func TestDomains_CompanyInfo(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("GET /api/domains/9/company_info", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"data":{"name":"Mailtrap","country":"US","info_level":"business"}}`))
	})

	info, _, err := client.Domains.CompanyInfo(context.Background(), 9)
	if err != nil {
		t.Fatalf("CompanyInfo: %v", err)
	}
	if info.Name != "Mailtrap" || info.InfoLevel != "business" {
		t.Errorf("company info = %+v", info)
	}
}

func TestDomains_CreateCompanyInfo(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("POST /api/domains/9/company_info", func(w http.ResponseWriter, r *http.Request) {
		wantJSONBody(t, r, `{"company_info":{"name":"Mailtrap","address":"123 Main St","city":"SF","country":"US","zip_code":"94105","website_url":"https://mailtrap.io"}}`)
		_, _ = w.Write([]byte(`{"data":{"name":"Mailtrap","country":"US"}}`))
	})

	info, _, err := client.Domains.CreateCompanyInfo(context.Background(), 9, &mailtrap.CompanyInfoRequest{
		Name:       "Mailtrap",
		Address:    "123 Main St",
		City:       "SF",
		Country:    "US",
		ZipCode:    "94105",
		WebsiteURL: "https://mailtrap.io",
	})
	if err != nil {
		t.Fatalf("CreateCompanyInfo: %v", err)
	}
	if info.Name != "Mailtrap" {
		t.Errorf("company info = %+v", info)
	}
}

func TestDomains_UpdateCompanyInfo(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("PATCH /api/domains/9/company_info", func(w http.ResponseWriter, r *http.Request) {
		wantJSONBody(t, r, `{"company_info":{"city":"New York"}}`)
		_, _ = w.Write([]byte(`{"data":{"name":"Mailtrap","city":"New York"}}`))
	})

	info, _, err := client.Domains.UpdateCompanyInfo(context.Background(), 9, &mailtrap.CompanyInfoRequest{City: "New York"})
	if err != nil {
		t.Fatalf("UpdateCompanyInfo: %v", err)
	}
	if info.City != "New York" {
		t.Errorf("company info = %+v", info)
	}
}

// The deprecated SendingDomain names are aliases, so code written against them
// keeps compiling with values of the new types.
var (
	_ *mailtrap.SendingDomainsService    = (*mailtrap.DomainsService)(nil)
	_ *mailtrap.SendingDomain            = &mailtrap.Domain{}
	_ *mailtrap.SendingDomainPermissions = &mailtrap.DomainPermissions{}
)

func TestSendingDomains_DeprecatedAlias(t *testing.T) {
	_, client := setup(t)

	if client.SendingDomains != client.Domains {
		t.Errorf("SendingDomains = %p, want Domains %p", client.SendingDomains, client.Domains)
	}
}

// Each deprecated name has a paragraph starting "Deprecated:" that names its
// replacement, which is what makes staticcheck (SA1019) and gopls flag callers.
func TestSendingDomains_DeprecationNotices(t *testing.T) {
	want := map[string]string{
		"Client.SendingDomains":    "Deprecated: Use Domains instead.",
		"SendingDomainsService":    "Deprecated: Use DomainsService instead.",
		"SendingDomain":            "Deprecated: Use Domain instead.",
		"SendingDomainPermissions": "Deprecated: Use DomainPermissions instead.",
	}

	paths, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatalf("glob: %v", err)
	}
	fset := token.NewFileSet()
	var files []*ast.File
	for _, path := range paths {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			t.Fatalf("parse %s: %v", path, err)
		}
		files = append(files, f)
	}
	// go/doc attaches doc comments to declarations the way go doc and gopls do.
	pkg, err := doc.NewFromFiles(fset, files, "github.com/mailtrap/mailtrap-go")
	if err != nil {
		t.Fatalf("doc: %v", err)
	}
	docs := map[string]string{}
	for _, typ := range pkg.Types {
		docs[typ.Name] = typ.Doc
		if st, ok := typ.Decl.Specs[0].(*ast.TypeSpec).Type.(*ast.StructType); ok {
			for _, field := range st.Fields.List {
				for _, name := range field.Names {
					docs[typ.Name+"."+name.Name] = field.Doc.Text()
				}
			}
		}
	}

	for name, notice := range want {
		doc := docs[name]
		if !strings.HasPrefix(doc, notice) && !strings.Contains(doc, "\n\n"+notice) {
			t.Errorf("doc comment of %s = %q, want a paragraph starting %q", name, doc, notice)
		}
	}
}
