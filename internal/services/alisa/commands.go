package alisa

// farewellText — прощальная реплика при команде завершения.
const farewellText = "Было приятно пообщать. До встречи!"

// resetDoneText — подтверждение сброса истории диалога.
const resetDoneText = "Начинаем новый диалог. Спроси что-нибудь."

// exitCommands — реплики, завершающие навык. Command приходит нормализованным
// из Яндекс.Диалогов: строчные буквы, без знаков препинания.
var exitCommands = map[string]struct{}{
	"стоп":          {},
	"выход":         {},
	"хватит":        {},
	"закончить":     {},
	"на этом всё":   {},
	"на этом все":   {},
	"всего доброго": {},
}

// resetCommands — реплики, очищающие историю диалога.
var resetCommands = map[string]struct{}{
	"новый диалог":  {},
	"сброс":         {},
	"сбросить":      {},
	"начать заново": {},
	"забудь всё":    {},
	"забудь все":    {},
}

func isExitCommand(command string) bool {
	_, ok := exitCommands[command]

	return ok
}

func isResetCommand(command string) bool {
	_, ok := resetCommands[command]

	return ok
}
