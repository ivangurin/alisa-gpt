package alisa

import "strings"

// markdownRunes — разметка, которую TTS озвучил бы вслух («звёздочка звёздочка»).
const markdownRunes = "*_`#>~|[]()"

// sanitizeSpeech готовит TTS-версию ответа: убирает markdown-разметку,
// которая не озвучивается, и схлопывает лишние пробелы. Text для экрана
// приложения остаётся исходным.
func sanitizeSpeech(text string) string {
	var b strings.Builder
	b.Grow(len(text))

	for _, r := range text {
		if strings.ContainsRune(markdownRunes, r) {
			continue
		}

		// WriteRune на strings.Builder не возвращает ошибок
		_, _ = b.WriteRune(r)
	}

	res := strings.Join(strings.Fields(b.String()), " ")

	return res
}
