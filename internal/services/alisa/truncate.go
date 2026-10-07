package alisa

import (
	"strings"
	"unicode/utf8"
)

// ellipsis — многоточие, добавляемое к обрезанному ответу.
const ellipsis = "…"

// truncate обрезает текст до maxChars символов по границам рун (лимит Алисы
// на response.text — 1024 символа, конфиг ограничивает строже). К обрезанному
// тексту добавляется многоточие, чтобы ответ не выглядел оборванным.
func truncate(text string, maxChars int) string {
	if utf8.RuneCountInString(text) <= maxChars {
		return text
	}

	runes := []rune(text)
	cut := strings.TrimRight(string(runes[:maxChars-1]), " ,.-–—")

	return cut + ellipsis
}
