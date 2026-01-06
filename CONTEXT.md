# Teamly - AI Context

Этот файл содержит контекст проекта для AI ассистентов.

## Описание проекта

Teamly - платформа для поиска тиммейтов для онлайн игр и прохождения кооперативных игр. Пользователи создают заявки с указанием игр, платформы и описания, другие игроки могут просматривать эти заявки и связываться с авторами.

## Технологический стек

### Backend
- **Язык:** Go (Golang)
- **Аутентификация:** JWT токены
- **API:** REST API

### База данных
- **СУБД:** PostgreSQL 16
- **Контейнеризация:** Docker
- **Параметры подключения:**
  - Host: localhost
  - Port: 5433
  - Database: teamly_db
  - User: teamly
  - Password: teamly_password
  - Connection String: `postgresql://teamly:teamly_password@localhost:5433/teamly_db`

## Структура базы данных

### Таблица: users
Пользователи системы.

```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    nickname VARCHAR(100) NOT NULL,
    avatar_url VARCHAR(500),
    discord VARCHAR(100),
    telegram VARCHAR(100),
    likes_count INTEGER DEFAULT 0,
    dislikes_count INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

**Поля:**
- `id` - UUID, первичный ключ
- `email` - Email пользователя (уникальный)
- `password_hash` - Хэшированный пароль
- `nickname` - Никнейм пользователя
- `avatar_url` - URL аватарки
- `discord` - Discord username (опционально)
- `telegram` - Telegram username (опционально)
- `likes_count` - Количество положительных оценок (лайков)
- `dislikes_count` - Количество отрицательных оценок (дизлайков)
- `created_at` - Дата регистрации
- `updated_at` - Дата последнего обновления

### Таблица: games
Справочник игр.

```sql
CREATE TABLE games (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) UNIQUE NOT NULL,
    icon_url VARCHAR(500),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

**Поля:**
- `id` - SERIAL, первичный ключ
- `name` - Название игры (уникальное)
- `icon_url` - URL иконки игры (опционально)
- `created_at` - Дата добавления

### Таблица: listings
Заявки на поиск тиммейтов.

```sql
CREATE TYPE listing_status AS ENUM ('active', 'closed', 'in_progress');

CREATE TABLE listings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    platform VARCHAR(50) NOT NULL,
    description TEXT NOT NULL,
    status listing_status DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

**Поля:**
- `id` - UUID, первичный ключ
- `user_id` - Внешний ключ на users.id (автор заявки)
- `platform` - Платформа (PC, PlayStation, Xbox, Nintendo Switch, Mobile, и т.д.)
- `description` - Описание заявки
- `status` - Статус: active (активна), closed (закрыта), in_progress (в процессе)
- `created_at` - Дата создания
- `updated_at` - Дата последнего обновления

### Таблица: listing_games
Связующая таблица many-to-many между заявками и играми.

```sql
CREATE TABLE listing_games (
    id SERIAL PRIMARY KEY,
    listing_id UUID NOT NULL REFERENCES listings(id) ON DELETE CASCADE,
    game_id INTEGER NOT NULL REFERENCES games(id) ON DELETE CASCADE,
    UNIQUE(listing_id, game_id)
);
```

**Поля:**
- `id` - SERIAL, первичный ключ
- `listing_id` - Внешний ключ на listings.id
- `game_id` - Внешний ключ на games.id
- Уникальная пара (listing_id, game_id) - нельзя дважды добавить одну игру в заявку

### Таблица: reviews
Отзывы пользователей (лайки/дизлайки с опциональным комментарием).

```sql
CREATE TYPE review_type AS ENUM ('like', 'dislike');

CREATE TABLE reviews (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    reviewer_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    reviewed_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    review_type review_type NOT NULL,
    comment TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(reviewer_id, reviewed_user_id)
);
```

**Поля:**
- `id` - UUID, первичный ключ
- `reviewer_id` - Внешний ключ на users.id (кто оставил отзыв)
- `reviewed_user_id` - Внешний ключ на users.id (кому оставлен отзыв)
- `review_type` - Тип отзыва: 'like' (плюс) или 'dislike' (минус)
- `comment` - Текстовый комментарий (опционально)
- `created_at` - Дата создания отзыва
- `updated_at` - Дата последнего обновления
- Уникальная пара (reviewer_id, reviewed_user_id) - один пользователь может оставить только один отзыв другому

**Примечания:**
- Пользователь не может оставить отзыв самому себе (проверка на уровне приложения)
- Пользователь может изменить свой отзыв (с like на dislike или наоборот)
- При добавлении/обновлении/удалении отзыва автоматически обновляются счетчики likes_count и dislikes_count у пользователя
- Дизлайк не уменьшает количество лайков - это два независимых показателя

### Схема связей

```
users (1) ----< (N) listings
  |
  └─ Один пользователь может создать много заявок

listings (N) >----< (N) games
  |
  └─ Одна заявка может содержать несколько игр
  └─ Одна игра может быть в нескольких заявках
  └─ Связь через таблицу listing_games

users (1) ----< (N) reviews (as reviewer)
  |
  └─ Один пользователь может оставить много отзывов

users (1) ----< (N) reviews (as reviewed_user)
  |
  └─ Один пользователь может получить много отзывов
```

## Основной функционал (MVP)

### 1. Аутентификация и авторизация
- Регистрация пользователей с email и паролем
- Авторизация с использованием JWT токенов
- Защищенные эндпоинты требуют валидный JWT токен

### 2. Управление пользователями
- Просмотр профиля пользователя
- Редактирование своего профиля (никнейм, аватарка, контакты)
- Отображение количества лайков и дизлайков пользователя
- Просмотр отзывов о пользователе

### 3. Управление заявками
- Создание заявки с указанием:
  - Одной или нескольких игр
  - Платформы
  - Описания (что ищем, требования, пожелания)
- Просмотр списка всех активных заявок
- Фильтрация заявок (по играм, платформе)
- Изменение статуса своей заявки (активна/закрыта/в процессе)
- Редактирование и удаление своих заявок

### 4. Справочник игр
- Просмотр списка доступных игр
- Добавление новых игр (пока без ограничений, потом возможно только админы)

### 5. Система отзывов
- Поставить лайк или дизлайк другому пользователю (с опциональным комментарием)
- Изменить свой отзыв (с лайка на дизлайк или наоборот)
- Удалить свой отзыв
- Просмотр всех отзывов о пользователе
- Автоматическое обновление счетчиков лайков и дизлайков при изменении отзывов

## Планируемый функционал (не в MVP)

- Комментарии к заявкам
- Система откликов на заявки
- Чат между пользователями
- Уведомления (о новых откликах, комментариях)
- Часовой пояс пользователя
- История игр с другими пользователями
- Система оценок после совместной игры
- Черный список пользователей

## API Endpoints (планируемые)

### Auth
- `POST /api/auth/register` - Регистрация
- `POST /api/auth/login` - Авторизация (возвращает JWT)
- `POST /api/auth/refresh` - Обновление токена

### Users
- `GET /api/users/:id` - Получить профиль пользователя
- `PUT /api/users/:id` - Обновить свой профиль (требует JWT)
- `GET /api/users/:id/listings` - Получить заявки пользователя
- `GET /api/users/:id/reviews` - Получить отзывы о пользователе

### Listings
- `GET /api/listings` - Получить список заявок (с пагинацией и фильтрами)
  - Query params: ?game_id=1&platform=PC&status=active&page=1&limit=20
- `GET /api/listings/:id` - Получить заявку по ID
- `POST /api/listings` - Создать заявку (требует JWT)
- `PUT /api/listings/:id` - Обновить заявку (требует JWT, только автор)
- `DELETE /api/listings/:id` - Удалить заявку (требует JWT, только автор)
- `PATCH /api/listings/:id/status` - Изменить статус (требует JWT, только автор)

### Games
- `GET /api/games` - Получить список игр (с поиском)
  - Query params: ?search=counter&page=1&limit=50
- `GET /api/games/:id` - Получить игру по ID
- `POST /api/games` - Добавить игру (требует JWT)

### Reviews
- `GET /api/reviews?user_id=:id` - Получить отзывы о пользователе (с пагинацией)
  - Query params: ?user_id=uuid&page=1&limit=20&type=like (опционально фильтр по типу)
- `POST /api/reviews` - Создать отзыв (требует JWT)
  - Body: { "reviewed_user_id": "uuid", "review_type": "like", "comment": "Отличный тиммейт!" }
- `PUT /api/reviews/:id` - Обновить свой отзыв (требует JWT, только автор)
  - Body: { "review_type": "dislike", "comment": "Изменил мнение" }
- `DELETE /api/reviews/:id` - Удалить свой отзыв (требует JWT, только автор)
- `GET /api/reviews/:id` - Получить отзыв по ID

## Бизнес-правила

1. **Пользователь может:**
   - Создавать неограниченное количество заявок
   - Редактировать и удалять только свои заявки
   - Просматривать все активные заявки других пользователей
   - Добавлять игры в справочник
   - Оставлять отзывы другим пользователям
   - Редактировать и удалять только свои отзывы

2. **Заявка должна содержать:**
   - Минимум одну игру
   - Обязательно указание платформы
   - Обязательно описание (минимум N символов, например 10)

3. **Отзывы:**
   - Один пользователь может оставить только один отзыв другому пользователю
   - Нельзя оставить отзыв самому себе
   - Тип отзыва обязателен (like или dislike), комментарий опционален
   - Пользователь может изменить свой отзыв (с like на dislike или наоборот)
   - При добавлении/обновлении/удалении отзыва автоматически обновляются счетчики likes_count и dislikes_count
   - Лайки и дизлайки - независимые показатели (дизлайк не отнимает лайк)

4. **Валидация:**
   - Email должен быть валидным и уникальным
   - Пароль минимум 8 символов
   - Никнейм от 3 до 50 символов
   - likes_count и dislikes_count >= 0
   - review_type только 'like' или 'dislike'

## Примечания для разработки

- Использовать миграции для работы с БД
- Пароли хэшировать с использованием bcrypt
- JWT токены подписывать секретным ключом из переменных окружения
- Все даты хранить в UTC
- API должен возвращать JSON
- Обрабатывать ошибки и возвращать понятные сообщения
- Добавить логирование запросов

## Переменные окружения

```env
# Database
DB_HOST=localhost
DB_PORT=5433
DB_NAME=teamly_db
DB_USER=teamly
DB_PASSWORD=teamly_password
DB_SSLMODE=disable

# JWT
JWT_SECRET=your_secret_key_here
JWT_EXPIRATION=24h

# Server
PORT=8080
```

## Структура проекта (рекомендуемая)

```
teamly/
├── cmd/
│   └── api/
│       └── main.go
├── internal/
│   ├── config/
│   ├── database/
│   ├── handlers/
│   ├── middleware/
│   ├── models/
│   ├── repository/
│   └── services/
├── migrations/
├── docker-compose.yml
├── go.mod
├── go.sum
└── CONTEXT.md
```
