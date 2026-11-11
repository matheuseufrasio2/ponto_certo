package i18n_test

import (
	"net/http/httptest"
	"testing"

	"github.com/matheuseufrasio2/ponto_certo/internal/domain"
	"github.com/matheuseufrasio2/ponto_certo/internal/i18n"
)

func TestTranslate(t *testing.T) {
	tests := []struct {
		name     string
		locale   i18n.Locale
		code     domain.ErrorCode
		expected string
	}{
		{
			name:     "pt-BR Barber Not Found",
			locale:   i18n.LocalePTBR,
			code:     domain.ErrCodeBarberNotFound,
			expected: "Profissional barbeiro não encontrado.",
		},
		{
			name:     "en-US Barber Not Found",
			locale:   i18n.LocaleENUS,
			code:     domain.ErrCodeBarberNotFound,
			expected: "Barber professional not found.",
		},
		{
			name:     "pt-BR Tenant Inactive",
			locale:   i18n.LocalePTBR,
			code:     domain.ErrCodeTenantInactive,
			expected: "Acesso bloqueado: esta barbearia está suspensa por pendências financeiras ou cadastrais.",
		},
		{
			name:     "en-US Tenant Inactive",
			locale:   i18n.LocaleENUS,
			code:     domain.ErrCodeTenantInactive,
			expected: "Access blocked: this barbershop is suspended due to pending financial or registration issues.",
		},
		{
			name:     "fallback to pt-BR when locale is unknown",
			locale:   i18n.Locale("es-ES"),
			code:     domain.ErrCodeBarberNotFound,
			expected: "Profissional barbeiro não encontrado.",
		},
		{
			name:     "pt-BR Unit Name Required",
			locale:   i18n.LocalePTBR,
			code:     domain.ErrCodeUnitNameRequired,
			expected: "O nome da unidade é obrigatório.",
		},
		{
			name:     "en-US Unit Name Required",
			locale:   i18n.LocaleENUS,
			code:     domain.ErrCodeUnitNameRequired,
			expected: "Barbershop unit name is required.",
		},
		{
			name:     "pt-BR Invalid Chair Count",
			locale:   i18n.LocalePTBR,
			code:     domain.ErrCodeInvalidChairCount,
			expected: "A quantidade de cadeiras deve ser maior que zero.",
		},
		{
			name:     "en-US Invalid Chair Count",
			locale:   i18n.LocaleENUS,
			code:     domain.ErrCodeInvalidChairCount,
			expected: "Chair count must be greater than zero.",
		},
		{
			name:     "pt-BR Unit Name Already Exists",
			locale:   i18n.LocalePTBR,
			code:     domain.ErrCodeUnitNameAlreadyExists,
			expected: "Já existe uma filial cadastrada com este nome na organização.",
		},
		{
			name:     "en-US Unit Name Already Exists",
			locale:   i18n.LocaleENUS,
			code:     domain.ErrCodeUnitNameAlreadyExists,
			expected: "A branch with this name already exists in your organization.",
		},
		{
			name:     "pt-BR Unit Inactive",
			locale:   i18n.LocalePTBR,
			code:     domain.ErrCodeUnitInactive,
			expected: "Acesso ou operação bloqueada: esta filial está pausada no momento.",
		},
		{
			name:     "en-US Unit Inactive",
			locale:   i18n.LocaleENUS,
			code:     domain.ErrCodeUnitInactive,
			expected: "Access or operation blocked: this branch is currently paused.",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := i18n.Translate(tc.locale, tc.code)
			if got != tc.expected {
				t.Errorf("expected %q, got %q", tc.expected, got)
			}
		})
	}
}

func TestResolveLocale(t *testing.T) {
	t.Run("nil request defaults to pt-BR", func(t *testing.T) {
		if got := i18n.ResolveLocale(nil); got != i18n.LocalePTBR {
			t.Errorf("expected pt-BR, got %s", got)
		}
	})

	t.Run("empty header defaults to pt-BR", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		if got := i18n.ResolveLocale(req); got != i18n.LocalePTBR {
			t.Errorf("expected pt-BR, got %s", got)
		}
	})

	t.Run("Accept-Language with en returns en-US", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Accept-Language", "en-US,en;q=0.9")
		if got := i18n.ResolveLocale(req); got != i18n.LocaleENUS {
			t.Errorf("expected en-US, got %s", got)
		}
	})

	t.Run("Accept-Language with pt returns pt-BR", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Accept-Language", "pt-BR,pt;q=0.9,en;q=0.8")
		if got := i18n.ResolveLocale(req); got != i18n.LocalePTBR {
			t.Errorf("expected pt-BR, got %s", got)
		}
	})

	t.Run("Query param ?lang=en takes precedence", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/?lang=en", nil)
		req.Header.Set("Accept-Language", "pt-BR")
		if got := i18n.ResolveLocale(req); got != i18n.LocaleENUS {
			t.Errorf("expected en-US, got %s", got)
		}
	})
}
