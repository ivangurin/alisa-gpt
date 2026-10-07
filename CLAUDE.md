# CLAUDE.md

## Commands

- `make run` — запуск вебхука локально (нужен `.env` с `OPENAI_API_KEY`)
- `make test` — тесты с `-race` и покрытием; один тест:
  `go test -race -count=1 -run TestName ./internal/...`
- `make lint` — golangci-lint `--fix` по всему проекту (конфиг `.golangci.yaml`)
- `make generate` — моки mockery + swagger + `go fix` + gofumpt `-extra`
- `make gen-swagger` — перегенерация `docs/` из аннотаций хендлеров
  (`go tool swag fmt; go tool swag init --quiet --parseDependency --parseInternal -g cmd/app/main.go`)
- `make docker-build` / `make docker-run` — образ и контейнер (8080 на 127.0.0.1)

## Architecture

Слои и направление зависимостей:

```
cmd/app/main.go                entry: config.NewConfigFromEnv → app.Run();
                               топ-аннотации swagger + import docs
internal/app/alisa             composition root: config → logger → closer →
                               LLM-клиент → service_provider → роутер →
                               HTTP-сервер (Start/Stop)
internal/config                env-only (caarlos0/env, envDefault-теги) +
                               .env + fail-fast валидация
internal/models                DTO протокола Алисы (v1.0), Message чата,
                               ErrorResponse API
internal/api/
  handlers/handlers.go         chi-роутер, константы путей, адаптер
                               HandlerFunc → http.Handler (safety-net 400)
  handlers/base                базовый хендлер: WriteJSON / WriteErrorResponse*
  handlers/meta                /healthcheck
  handlers/alisa               POST / — вебхук (swagger-аннотации)
  middleware/logging.go        request/response-логи (request_id из ctx)
internal/clients/llm           ИНТЕРФЕЙС LLM-клиента (у потребителя) + сентинелы
internal/clients/openai        реализация: POST {base}/chat/completions,
                               маппинг ошибок в llm.ErrTimeout/ErrUnavailable
internal/services/alisa        сервис: welcome/стоп/сброс/ответ, история, fallback
internal/service_provider      продовый DI: ленивые Get* под sync.Mutex
internal/pkg/
  servers/http                 обёртка net/http.Server: Start/Stop
  history                      in-memory диалоги по user_id (лимит, TTL)
  logger / closer              slog-обёртка, LIFO-closer
  metadata                     мостик к chi RequestID (request_id в ctx)
  suite/provider               тестовый DI: NewProvider() + cleanup, мок LLM
  suite/factory                gofakeit-билдеры запросов Алисы
docs/                          сгенерированные swagger-артефакты (DO NOT EDIT)
```

Модель запуска (`internal/app/alisa/alisa.go`): `Run` собирает всё руками,
HTTP-сервер стартует в горутине (`Start` блокируется), фатальная ошибка —
`errCh` + `cl.Stop()`, сигнал — `cl.Wait()`. Порядок shutdown (LIFO closer):
зарегистрировано `cancel` → `httpServer.Stop`; исполняется наоборот — сначала
сервер перестаёт принимать запросы и дренирует in-flight (ответы Алисе уходят
целиком), затем гасится корневой ctx.

Middleware-цепочка роутера: `chi Recoverer` → `chi RequestID` → `Logging`.
Хендлеры: сигнатура `func(w, r) error`, ответ пишут сами и возвращают `nil`;
адаптер `NewHandler` — safety-net 400 на не-nil ошибке.

## Conventions

- Один файл на метод/тип; конструктор `New*`; структуры по указателю.
- Интерфейсы у потребителя: `llm.Client`, `alisa_handler.Service`.
- Сентинелы `ErrXxx` + `fmt.Errorf("...: %w", err)` + `errors.Is`; маппинг
  внешних ошибок через `errors.Join(err, llm.ErrTimeout)`.
- Логи: map-поля с `field*`-константами, `log.With("component", ...)`,
  ошибки — полем `"error": err.Error()`, request_id из ctx (chi RequestID →
  metadata).
- Конфиг: env-only, дефолты в `envDefault`-тегах; длинные тексты (system
  prompt, welcome, fallback) — константами-дефолтами на пустое значение;
  `,required` не используется — секреты валидируются сентинелом ErrNoAPIKey.
- Тесты — только через `suite_provider.NewProvider()` (+ `defer cleanup()`);
  мок LLM подключается к тесту через `sp.GetLLMClientMock().Test(t)`;
  фикстуры — `suite_factory.NewRequestFactory()`; в хендлерах httptest —
  `assert`, не `require` (паника из горутины/хендлера).
- Swagger: аннотации табами над методом хендлера; после изменения —
  `make gen-swagger`; docs/ коммитится, CI не перегенерирует.
- Форматирование gofumpt `-extra`; `nlreturn` (пустая строка перед `return`);
  godox запрещён (без TODO в коде); без Co-Authored-By в коммитах.

## База знаний

- **Лимит Алисы — 4.5 с** на ответ вебхука. Весь путь (nginx → сервис → OpenAI)
  обязан укладываться в бюджет: `OPENAI_TIMEOUT_MS ≤ 4000` (дефолт 3500),
  fallback-фраза при ошибке, ретраев внутри запроса нет. Stop HTTP-сервера
  держит 5-секундный бюджет дрена (тоже с запасом против 4.5 с).
- **Порядок shutdown closer'а**: http-stop регистрируется ПОСЛЕДНИМ (выполняется
  первым по LIFO), cancel корневого ctx — первым зарегистрирован (исполняется
  последним). Всё, от чего зависит дрен, регистрируется в closer ПОЗЖЕ дрена.
- `response.text` лимит 1024 символа — конфиг валидирует `max-chars ≤ 1024`,
  сервис обрезает по рунам с многоточием.
- `command` из Яндекс.Диалогов приходит нормализованным (строчные, без
  пунктуации) — служебные команды сравниваются точным совпадением.
- История диалога ключуется `user_id` (стабилен между запусками навыка),
  не `session_id` (меняется на каждый запуск).
- nilerr на `return nil` после записанного ответа в хендлере подавляется
  точечным `//nolint:nilerr` — это контракт хендлеров (ответ уже записан), не баг.
