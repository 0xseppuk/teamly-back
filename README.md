# Teamly Backend

## Setup

### Local Development

1. Скопируйте `.env.example` в `.env`:
```bash
cp .env.example .env
```

2. Измените переменные в `.env` на свои значения:
   - `POSTGRES_PASSWORD` - сильный пароль для БД
   - `JWT_SECRET` - случайная строка для JWT токенов
   - `JWT_REFRESH_SECRET` - другая случайная строка для refresh токенов

3. Запустите сервисы:
```bash
docker-compose up -d
```

### Production

**ВАЖНО:**
- Никогда не коммитьте `.env` файл в git
- Используйте сильные пароли для продакшена
- Для генерации секретов можно использовать:
```bash
openssl rand -base64 32
```

### Переменные окружения

Все необходимые переменные описаны в `.env.example`

## Commands

```bash
# Запустить все сервисы
docker-compose up -d

# Посмотреть логи
docker-compose logs -f

# Остановить
docker-compose down

# Пересобрать после изменений
docker-compose up -d --build

# Локальная разработка без Docker
make run
```
