package utils

import (
	"strings"
)

var accentReplacer = strings.NewReplacer(
	"á", "a", "â", "a", "ã", "a", "à", "a", "ä", "a",
	"é", "e", "ê", "e", "è", "e", "ë", "e",
	"í", "i", "î", "i", "ì", "i", "ï", "i",
	"ó", "o", "ô", "o", "õ", "o", "ò", "o", "ö", "o",
	"ú", "u", "û", "u", "ù", "u", "ü", "u",
	"ç", "c", "ñ", "n",
	"Á", "a", "Â", "a", "Ã", "a", "À", "a", "Ä", "a",
	"É", "e", "Ê", "e", "È", "e", "Ë", "e",
	"Í", "i", "Î", "i", "Ì", "i", "Ï", "i",
	"Ó", "o", "Ô", "o", "Õ", "o", "Ò", "o", "Ö", "o",
	"Ú", "u", "Û", "u", "Ù", "u", "Ü", "u",
	"Ç", "c", "Ñ", "n",
)

// NormalizeSlug converte uma string para lowercase e remove acentos,
// facilitando o match de slugs digitados pelo usuário.
// Ex: "Saúde" → "saude", "Alimentação" → "alimentacao"
func NormalizeSlug(s string) string {
	return accentReplacer.Replace(strings.ToLower(s))
}
