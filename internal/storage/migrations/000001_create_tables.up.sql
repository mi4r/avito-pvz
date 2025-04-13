CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Пользователи системы
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(20) NOT NULL CHECK (role IN ('employee', 'moderator')),
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Пункты выдачи заказов (ПВЗ)
CREATE TABLE pvz (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    city VARCHAR(50) NOT NULL CHECK (city IN ('Москва', 'Санкт-Петербург', 'Казань')),
    registration_date TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Приёмки товаров
CREATE TABLE receptions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    pvz_id UUID NOT NULL REFERENCES pvz(id) ON DELETE CASCADE,
    status VARCHAR(20) NOT NULL DEFAULT 'in_progress' CHECK (status IN ('in_progress', 'closed')),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Частичный уникальный индекс для активных приёмок
CREATE UNIQUE INDEX unique_active_reception 
ON receptions (pvz_id) 
WHERE status = 'in_progress';

-- Товары
CREATE TABLE products (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    reception_id UUID NOT NULL REFERENCES receptions(id) ON DELETE CASCADE,
    type VARCHAR(20) NOT NULL CHECK (type IN ('электроника', 'одежда', 'обувь')),
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Индексы для ускорения поиска
CREATE INDEX idx_receptions_pvz_status ON receptions(pvz_id, status);
CREATE INDEX idx_products_reception ON products(reception_id);
CREATE INDEX idx_users_email ON users(email);