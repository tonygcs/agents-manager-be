package handler

import (
	"testing"
)

func TestApplyTemplate(t *testing.T) {
	tests := []struct {
		name     string
		template string
		tplName  string
		domain   string
		want     string
		wantErr  bool
	}{
		{
			name:     "plain string, no template",
			template: "hello world",
			tplName:  "myname",
			domain:   "example.com",
			want:     "hello world",
		},
		{
			name:     "name variable",
			template: "worker-{{.name}}",
			tplName:  "abc123",
			domain:   "example.com",
			want:     "worker-abc123",
		},
		{
			name:     "domain variable",
			template: "host.{{.domain}}",
			tplName:  "myname",
			domain:   "example.com",
			want:     "host.example.com",
		},
		{
			name:     "split func",
			template: `{{index (split .domain ".") 0}}`,
			tplName:  "myname",
			domain:   "sub.example.com",
			want:     "sub",
		},
		{
			name:     "rootdomain func - subdomain",
			template: "{{rootdomain .domain}}",
			tplName:  "myname",
			domain:   "sub.example.com",
			want:     "example.com",
		},
		{
			name:     "rootdomain func - hyphenated subdomain under known TLD",
			template: "{{rootdomain .domain}}",
			tplName:  "myname",
			domain:   "a-b.example.com",
			want:     "example.com",
		},
		{
			name:     "rootdomain func - two parts only",
			template: "{{rootdomain .domain}}",
			tplName:  "myname",
			domain:   "example.com",
			want:     "example.com",
		},
		{
			name:     "rootdomain func - two parts without known TLD",
			template: "{{rootdomain .domain}}",
			tplName:  "myname",
			domain:   "sub.example",
			want:     "example",
		},
		{
			name:     "rootdomain func - hyphenated subdomain without known TLD",
			template: "{{rootdomain .domain}}",
			tplName:  "myname",
			domain:   "a-b.example",
			want:     "example",
		},
		{
			name:     "rootdomain func - deep subdomain",
			template: "{{rootdomain .domain}}",
			tplName:  "myname",
			domain:   "a.b.c.example.com",
			want:     "example.com",
		},
		{
			name:     "combined name and domain",
			template: "{{.name}}.{{rootdomain .domain}}",
			tplName:  "worker42",
			domain:   "sub.example.com",
			want:     "worker42.example.com",
		},
		{
			name:     "invalid template syntax",
			template: "{{.name",
			tplName:  "myname",
			domain:   "example.com",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := applyTemplate(tt.template, tt.tplName, tt.domain)
			if (err != nil) != tt.wantErr {
				t.Fatalf("applyTemplate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("applyTemplate() = %q, want %q", got, tt.want)
			}
		})
	}
}
