-- Для больших таблиц: партиционирование по дате
-- Это существенно ускорит удаление старых данных
CREATE TABLE users (
                            id SERIAL,
                            created_at TIMESTAMP NOT NULL,
                            data JSONB
    -- другие колонки
) PARTITION BY RANGE (created_at);

-- Создаем оптимальный индекс для таблицы
CREATE INDEX IF NOT EXISTS idx_table_created_at ON users (created_at);

-- Создание партиций по месяцам
CREATE TABLE users_y2025m01 PARTITION OF users
    FOR VALUES FROM ('2025-01-01') TO ('2025-02-01');

CREATE TABLE users_y2025m02 PARTITION OF users
    FOR VALUES FROM ('2025-02-01') TO ('2025-03-01');

CREATE TABLE users_y2025m03 PARTITION OF users
    FOR VALUES FROM ('2025-03-01') TO ('2025-04-01');

CREATE TABLE users_y2025m04 PARTITION OF users
    FOR VALUES FROM ('2025-04-01') TO ('2025-05-01');

CREATE TABLE users_y2025m05 PARTITION OF users
    FOR VALUES FROM ('2025-05-01') TO ('2025-06-01');

CREATE TABLE users_y2025m06 PARTITION OF users
    FOR VALUES FROM ('2025-06-01') TO ('2025-07-01');


-- Настройка autovacuum для каждой партиции
ALTER TABLE users_y2025m01 SET (
    autovacuum_vacuum_scale_factor = 0.05,
    autovacuum_analyze_scale_factor = 0.02
    );

ALTER TABLE users_y2025m02 SET (
    autovacuum_vacuum_scale_factor = 0.05,
    autovacuum_analyze_scale_factor = 0.02
    );

ALTER TABLE users_y2025m03 SET (
    autovacuum_vacuum_scale_factor = 0.05,
    autovacuum_analyze_scale_factor = 0.02
    );

ALTER TABLE users_y2025m04 SET (
    autovacuum_vacuum_scale_factor = 0.05,
    autovacuum_analyze_scale_factor = 0.02
    );

ALTER TABLE users_y2025m05 SET (
    autovacuum_vacuum_scale_factor = 0.05,
    autovacuum_analyze_scale_factor = 0.02
    );

ALTER TABLE users_y2025m06 SET (
    autovacuum_vacuum_scale_factor = 0.05,
    autovacuum_analyze_scale_factor = 0.02
    );

-- Генерируем 100000 строк, равномерно распределённых по датам
INSERT INTO users (created_at, data)
SELECT
    timestamp '2025-02-01' + (random() * (timestamp '2025-07-01' - timestamp '2025-02-01')),
    jsonb_build_object('field', gen_random_uuid()::text)
FROM generate_series(1, 1000000);