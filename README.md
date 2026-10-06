# M96 API

Самостоятельный Go-сервис. Требует Go 1.25 или Docker. HTTP API, JSON-хранилище и загрузки файлов находятся только в этой папке.

## Docker на отдельном сервере

```bash
cp .env.example .env
# В .env замените ADMIN_TOKEN на результат: openssl rand -hex 32
docker compose -f compose.yml up --build -d
```

Go слушает `HTTP_ADDR=:8080`; Docker публикует `BACKEND_PUBLIC_PORT` (по умолчанию 8080). Проверьте `http://АДРЕС-СЕРВЕРА:8080/healthz` и `/api/content`. Для публичного использования направьте домен с HTTPS на порт сервиса. Данные живут в Docker volume `content`; начальный файл берётся из `data/content.json`. `ADMIN_TOKEN` остаётся только на этом сервере.

При прямом запуске без Docker используйте локальные `DATA_FILE=data/content.json`, `UPLOAD_DIR=data/uploads`, `HTTP_ADDR=:8080` и задайте `ADMIN_TOKEN` в окружении.
