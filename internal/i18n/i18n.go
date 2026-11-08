package i18n

import (
	"embed"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/matheuseufrasio2/ponto_certo/internal/domain"
)

//go:embed locales/*.json
var localesFS embed.FS

// Locale represents a supported IETF language tag.
type Locale string

const (
	LocalePTBR Locale = "pt-BR"
	LocaleENUS Locale = "en-US"
)

var translations = make(map[Locale]map[domain.ErrorCode]string)

func init() {
	loadLocale(LocalePTBR, "locales/pt-br.json")
	loadLocale(LocaleENUS, "locales/en-us.json")
}

func loadLocale(loc Locale, path string) {
	data, err := localesFS.ReadFile(path)
	if err != nil {
		return
	}
	var dict map[string]string
	if err := json.Unmarshal(data, &dict); err == nil {
		translations[loc] = make(map[domain.ErrorCode]string)
		for k, v := range dict {
			translations[loc][domain.ErrorCode(k)] = v
		}
	}
}

// ResolveLocale extracts target locale from ?lang= query param or Accept-Language header.
func ResolveLocale(r *http.Request) Locale {
	if r == nil {
		return LocalePTBR
	}

	if lang := r.URL.Query().Get("lang"); lang != "" {
		if strings.HasPrefix(strings.ToLower(lang), "en") {
			return LocaleENUS
		}
		return LocalePTBR
	}

	accept := strings.ToLower(r.Header.Get("Accept-Language"))
	if strings.Contains(accept, "en") && !strings.HasPrefix(accept, "pt") {
		return LocaleENUS
	}

	return LocalePTBR
}

// Translate returns the localized message for the given error code.
func Translate(locale Locale, code domain.ErrorCode) string {
	if dict, ok := translations[locale]; ok {
		if msg, found := dict[code]; found {
			return msg
		}
	}
	// Fallback to PT-BR
	if dict, ok := translations[LocalePTBR]; ok {
		if msg, found := dict[code]; found {
			return msg
		}
	}
	return "Ocorreu um erro ao processar a solicitação."
}

