# GophKeeper - Менеджер паролей

GophKeeper - это безопасный менеджер паролей с клиент-серверной архитектурой, написанный на Go.

## Возможности

- **Безопасное хранение данных**: AES-GCM шифрование для всех пользовательских данных
- **Аутентификация**: JWT токены для безопасной аутентификации
- **Типы данных**: Поддержка паролей, текстовых заметок, банковских карт и бинарных файлов
- **CLI интерфейс**: Удобный интерфейс командной строки
- **Синхронизация**: Возможность синхронизации данных между устройствами
- **Метаданные**: Добавление описаний и тегов к записям

## Архитектура

Проект состоит из двух основных компонентов:
1. **Сервер** - REST API для управления данными
2. **Клиент** - CLI приложение для взаимодействия с сервером

## Установка и запуск

### Подготовка базы данных

1. Установите PostgreSQL
2. Создайте базу данных:
   ```sql
   CREATE DATABASE gophkeeper;
   ```

### Настройка переменных окружения

Создайте файл `.env` в корне проекта:

```env
# Сервер
SERVER_ADDRESS=localhost:8080
JWT_SECRET=your-secret-key-here
POSTGRES_DSN=postgres://username:password@localhost/gophkeeper?sslmode=disable

# Клиент
CLIENT_SERVER_URL=http://localhost:8080
CLIENT_STORAGE_PATH=./client_data
CLIENT_ENCRYPTION_KEY=your-client-encryption-key
```

### Сборка проекта

```bash
go build -o bin/server cmd/server/main.go
go build -o bin/client cmd/client/main.go
```

### Запуск сервера

```bash
./bin/server
```

### Использование клиента

#### Регистрация пользователя
```bash
./bin/client register
```

#### Вход в систему
```bash
./bin/client login
```

#### Добавление данных

Добавить учетные данные:
```bash
./bin/client add credentials --name "My Website"
```

Добавить текстовую заметку:
```bash
./bin/client add text --name "My Note"
```

Добавить банковскую карту:
```bash
./bin/client add card --name "My Credit Card"
```

Добавить бинарный файл:
```bash
./bin/client add binary --name "My Document" --file /path/to/file
```

#### Просмотр данных

Показать все записи:
```bash
./bin/client list
```

Получить конкретную запись:
```bash
./bin/client get <entry-id>
```

#### Удаление данных
```bash
./bin/client delete <entry-id>
```

#### Синхронизация
```bash
./bin/client sync
```

## Структура проекта

```
GophKeeper/
├── cmd/                    # Точки входа приложений
│   ├── server/            # Серверное приложение
│   └── client/            # Клиентское приложение
├── internal/              # Внутренняя логика
│   ├── auth/             # Аутентификация и авторизация
│   ├── client/           # Клиентская логика
│   ├── config/           # Конфигурация
│   ├── crypto/           # Криптографические операции
│   ├── models/           # Модели данных
│   ├── server/           # Серверная логика
│   └── storage/          # Слой работы с данными
├── pkg/                  # Публичные пакеты
│   └── api/             # API структуры
└── bin/                 # Скомпилированные бинарники
```

## API Endpoints

### Аутентификация
- `POST /api/register` - Регистрация пользователя
- `POST /api/login` - Вход в систему

### Управление данными
- `POST /api/entries` - Создание записи
- `GET /api/entries` - Получение всех записей пользователя
- `GET /api/entries/{id}` - Получение конкретной записи
- `PUT /api/entries/{id}` - Обновление записи
- `DELETE /api/entries/{id}` - Удаление записи
- `POST /api/sync` - Синхронизация данных

## Безопасность

- Все пароли хранятся с использованием bcrypt
- Пользовательские данные шифруются AES-GCM
- Используется JWT для аутентификации
- Поддержка HTTPS для защиты передачи данных

## Тестирование

Запуск тестов:
```bash
go test ./...
```

Запуск тестов с покрытием:
```bash
go test -cover ./...
```