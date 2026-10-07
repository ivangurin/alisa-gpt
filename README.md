# alisa-gpt

Навык Яндекс.Алисы для голосового общения с ChatGPT: вебхук на Go (chi) принимает
запросы Яндекс.Диалогов, ходит в OpenAI Chat Completions и отвечает голосом.

## Стек

- Go 1.27, chi-роутер, slog (обёртка `internal/pkg/logger`)
- OpenAI-совместимый API (base URL и модель настраиваются)
- Swagger UI (`/swagger/index.html`), генерация `make gen-swagger`
- Docker-образ в GitHub Container Registry, на VPS — `docker run` + nginx (HTTPS)
- Тесты: testify + mockery + gofakeit, линт: golangci-lint (revive enable-all)

## Команды

```
make run          # запуск локально (нужен .env с OPENAI_API_KEY)
make test         # тесты с -race и покрытием
make lint         # golangci-lint --fix
make generate     # моки + swagger + go fix + gofumpt
make build        # бинарник в bin/
make docker-build
make docker-run   # контейнер на 127.0.0.1:8080 (ключ из .env)
```

## Конфигурация

Только переменные окружения: дефолты заданы `envDefault`-тегами, после парсинга —
fail-fast валидация. Полный список с дефолтами — `.env.example`;
локально удобно держать их в `.env`. Обязательна одна переменная:

```
OPENAI_API_KEY=sk-...
```

Часто переопределяют: `APP_PORT`, `OPENAI_MODEL` (быстрее — `gpt-4o-mini`),
`OPENAI_BASE_URL` (агрегатор/прокси), `OPENAI_TIMEOUT_MS` (≤ 4000).

Жёсткое ограничение платформы: **Алиса ждёт ответ вебхука не дольше 4.5 с**.
`OPENAI_TIMEOUT_MS` по умолчанию 3500 с запасом на транспорт; при ошибке/таймауте
LLM навык отвечает fallback-фразой и сохраняет сессию.

## API

| Метод | Путь | Назначение |
|---|---|---|
| POST | `/` | вебхук навыка (протокол Яндекс.Диалогов v1.0) |
| GET | `/healthcheck` | живость |
| GET | `/swagger/index.html` | Swagger UI |

Спецификация генерируется swaggo из аннотаций хендлеров (`make gen-swagger`),
артефакты лежат в `docs/`.

## Деплой

### 1. CI → GitHub Container Registry

Пуш в `main` запускает `.github/workflows/docker.yml`: тесты → сборка →
`ghcr.io/<owner>/alisa-gpt:latest` (+ тег по SHA, версии `v*` → semver). Образ
приватный по умолчанию — на VPS нужен `docker login ghcr.io` с PAT
(scope `read:packages`) либо сделайте образ публичным в настройках пакета.

### 2. VPS: запуск контейнера

```bash
docker login ghcr.io -u <github-user>
docker pull ghcr.io/<owner>/alisa-gpt:latest
docker run -d --name alisa-gpt \
  --restart unless-stopped \
  -e OPENAI_API_KEY=sk-... \
  -p 127.0.0.1:8080:8080 \
  ghcr.io/<owner>/alisa-gpt:latest
```

Прямой `api.openai.com` не отвечает с российских IP — VPS должен быть
за рубежом (или смените `OPENAI_BASE_URL` на доступный из РФ прокси/агрегатор).

### 3. VPS: nginx + HTTPS

Яндекс.Диалоги валидируют SSL-сертификат вебхука, поэтому спереди ставим
nginx с Let's Encrypt. Готовый конфиг — `deploy/nginx.conf.example`:

```bash
sudo cp deploy/nginx.conf.example /etc/nginx/sites-available/alisa-gpt
# заменить alice.example.com на свой поддомен, A-запись → IP VPS
sudo ln -s /etc/nginx/sites-available/alisa-gpt /etc/nginx/sites-enabled/
sudo nginx -t && sudo systemctl reload nginx
sudo certbot --nginx -d alice.example.com
```

Проверка: `curl https://alice.example.com/healthcheck` → `{"status":"ok"}`.

## Регистрация навыка в Яндекс.Диалогах

1. Зайдите на [dialogs.yandex.ru/developer](https://dialogs.yandex.ru/developer)
   **под аккаунтом, к которому привязана колонка**.
2. «Создать навык» → тип «Навык в Алисе».
3. Параметры:
   - **Название** — минимум два слова, без «GPT» (Алиса отвечает, что уже
     умеет GPT); например «Умный собеседник».
   - **Backend → Webhook URL** — `https://alice.example.com/`.
   - **Голос** — отличимый от голоса Алисы (например Джейн).
   - **Тип доступа** — «Приватный».
4. Вкладка **«Тестирование»** — проверьте диалог в веб-консоли.
5. Опубликуйте приватный навык / включите тумблер «на отладке» — модерация
   каталога для приватного личного использования не нужна.

## Запуск на колонке

Скажите: **«Алиса, запусти навык Умный собеседник»** — дальше обычный
разговор. Служебные команды навыка:

- «стоп», «выход», «хватит» — завершить навык;
- «новый диалог», «сброс», «начать заново» — очистить историю.

## Проверка вебхука вручную

```bash
curl -X POST https://alice.example.com/ -H 'Content-Type: application/json' -d '{
  "session": {"message_id": 0, "session_id": "test", "user_id": "u1", "new": true},
  "request": {"command": "", "original_utterance": "", "type": "SimpleUtterance"},
  "version": "1.0"
}'
```

Ответ — JSON с `response.text` (welcome-фраза). Полный цикл: `new: false` +
`command: "привет"` — придёт ответ GPT.

## Ограничения

- История диалога — in-memory: рестарт контейнера начинает диалог заново
  (для личного навыка приемлемо).
- Лимит Алисы на `response.text` — 1024 символа: ответы обрезаются
  (`DIALOG_MAX_CHARS`, по умолчанию 1000).
