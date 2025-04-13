# avito-pvz
Сервис для работы с ПВЗ

## Технологии

- Go 1.24+
- PostgreSQL 16+
- Chi Router
- JWT аутентификация
- Bcrypt для хеширования паролей

## Установка и запуск

1. Клонирование репозитория:
```bash
git clone https://github.com/mi4r/avito-pvz.git
cd avito-pvz
```

2. Установка и обновление зависимостей
```bash
go mod tidy
```

3. Для использования API внесите ваши данные в конфиг ".env"
```bash
cp .env.bak .env # Замените значения
```

4. Миграции применяются автоматически.
Запуск миграций вручную при необходимости:
```bash
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest 
```
```bash
migrate -path internal/storage/migrations -database postgres://$DB_USER:$DB_PASS@$DB_HOST:$DB_PORT/$DB_NAME?sslmode=disable up
```
 Откат миграции:
 ```bash
migrate -path internal/storage/migrations -database postgres://$DB_USER:$DB_PASS@$DB_HOST:$DB_PORT/$DB_NAME?sslmode=disable down 1
```

5. Запуск API через Docker:
```bash
docker compose up
```

## API Endpoints

### Простая авторизация
```POST /dummyLogin```
Пример вводных данных:
```json
{
  "role": "moderator"
}
```
Ответ:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

### Регистрация с почтой и паролем
```POST /register```
Пример вводных данных:
```json
{
  "email": "user1@gmail.com",
  "password": "pass123",
  "role": "employee"
}
```
Ответ:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

### Авторизация с почтой и паролем
```POST /login```
Пример вводных данных:
```json
{
  "email": "user1@gmail.com",
  "password": "pass123",
}
```
Ответ:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

### Получение данных о ПВЗ
```GET /pvz```
Пример вводных данных:
Загаловок
```
Authorization: Bearer <token>
```
Ответ:
```json
[
    {
        "PVZ": {
            "id": "997ff497-1799-40a9-9d48-be3a5c15b093",
            "registrationDate": "2025-04-14T01:12:15.556496Z",
            "city": "Казань"
        },
        "Receptions": []
    },
    {
        "PVZ": {
            "id": "4a8cc5b1-5584-4d2a-a2d5-bc4c4e71120a",
            "registrationDate": "2025-04-13T23:45:03.099288Z",
            "city": "Москва"
        },
        "Receptions": [
            {
                "Reception": {
                    "id": "e76cbb36-b00c-437c-a8e4-7b71bdd6ba29",
                    "createdAt": "2025-04-13T23:51:21.463983Z",
                    "pvzId": "4a8cc5b1-5584-4d2a-a2d5-bc4c4e71120a",
                    "status": "closed"
                },
                "Products": [
                    {
                        "id": "0fd0dc35-e7cd-4f6c-8c2f-ab193bffb8ef",
                        "createdAt": "2025-04-13T23:57:51.636232Z",
                        "type": "одежда",
                        "receptionId": "e76cbb36-b00c-437c-a8e4-7b71bdd6ba29"
                    },
                    {
                        "id": "e8e5db0d-90af-442a-8fb3-183af59925eb",
                        "createdAt": "2025-04-13T23:57:22.51872Z",
                        "type": "обувь",
                        "receptionId": "e76cbb36-b00c-437c-a8e4-7b71bdd6ba29"
                    }
                ]
            }
        ]
    }
]
```

### Заведение ПВЗ
```POST /pvz```
Пример вводных данных:
```json
{
    "city":"Казань"
}
```
Ответ:
```json
{
    "id": "997ff497-1799-40a9-9d48-be3a5c15b093",
    "registrationDate": "2025-04-14T01:12:15.556496Z",
    "city": "Казань"
}
```  
Статус успешного выполнения или код ошибки с комментарием.

### Добавление информации о приёмке товаров
```POST /receptions```
Пример вводных данных:
```json
{
    "pvzId":"4a8cc5b1-5584-4d2a-a2d5-bc4c4e71120a"
}
```
Ответ:
```json
{
    "id": "2549f7ca-6640-4194-8fee-bf37d1f0584c",
    "createdAt": "2025-04-14T02:22:13.259367Z",
    "pvzId": "4a8cc5b1-5584-4d2a-a2d5-bc4c4e71120a",
    "status": "in_progress"
}
```
Статус успешного выполнения или код ошибки с комментарием.

### Добавление товаров в рамках одной приёмки
```POST /products```
Пример вводных данных:
```json
{
    "type":"одежда",
    "pvzId":"4a8cc5b1-5584-4d2a-a2d5-bc4c4e71120a"
}
```
Ответ:
```json
{
    "ID": "b7429d85-aa62-427c-9416-22e3b2a558ac",
    "CreatedAt": "2025-04-14T00:00:18.120574Z",
    "Type": "одежда",
    "ReceptionID": "e76cbb36-b00c-437c-a8e4-7b71bdd6ba29"
}
```
Статус успешного выполнения или код ошибки с комментарием.


### Удаление товаров в рамках не закрытой приёмки
```POST /pvz/{pvzId}/delete_last_product```
Пример вводных данных:
```
localhost:8080/pvz/4a8cc5b1-5584-4d2a-a2d5-bc4c4e71120a/delete_last_product
```
Ответ:
Статус успешного выполнения или код ошибки с комментарием.

### Закрытие приёмки
```POST /pvz/{pvzId}/close_last_reception```
Пример вводных данных:
```
localhost:8080/pvz/4a8cc5b1-5584-4d2a-a2d5-bc4c4e71120a/close_last_reception
```
Ответ:
Статус успешного выполнения или код ошибки с комментарием.


## Тестирование
```bash
go test ./... -coverprofile profiles/cover.out && go tool cover -func=profiles/cover.out
```

## Лицензия
```
MIT License

Copyright (c) 2025 Tiko

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.

```
