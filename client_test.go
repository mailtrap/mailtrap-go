package mailtrap_test

import (
	"testing"

	"github.com/mailtrap/mailtrap-go"
)

func TestNewClient_validation(t *testing.T) {
	tests := []struct {
		name    string
		token   string
		opts    []mailtrap.Option
		wantErr bool
	}{
		{name: "token required", token: "", wantErr: true},
		{name: "token only", token: "tok"},
		{name: "nil HTTP client", token: "tok", opts: []mailtrap.Option{mailtrap.WithHTTPClient(nil)}, wantErr: true},
		{name: "empty user agent", token: "tok", opts: []mailtrap.Option{mailtrap.WithUserAgent("")}, wantErr: true},
		{name: "negative account id", token: "tok", opts: []mailtrap.Option{mailtrap.WithAccountID(-1)}, wantErr: true},
		{name: "empty base url", token: "tok", opts: []mailtrap.Option{mailtrap.WithBaseURL(mailtrap.HostGeneral, "")}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := mailtrap.NewClient(tt.token, tt.opts...)
			if (err != nil) != tt.wantErr {
				t.Fatalf("NewClient err = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}
