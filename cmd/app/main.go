package main

import (
	"fmt"
	"os"

	_ "alisa-gpt/docs" // регистрация сгенерированных swagger-доков

	app_alisa "alisa-gpt/internal/app/alisa"
	config_pkg "alisa-gpt/internal/config"
)

// @title			alisa-gpt
// @version		1.0
// @description	Навык Яндекс.Алисы для голосового общения с ChatGPT: вебхук протокола Яндекс.Диалогов v1.0, проксирующий реплики пользователя в OpenAI Chat Completions.
// @BasePath		/
func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "run: %s\n", err.Error())
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config_pkg.NewConfigFromEnv()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	if err := app_alisa.NewApp(cfg).Run(); err != nil {
		return fmt.Errorf("app run: %w", err)
	}

	return nil
}
