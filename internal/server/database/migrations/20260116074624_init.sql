-- +goose Up
-- +goose StatementBegin
-- Таблица для метрик с историей изменений
CREATE TABLE metrics (
    -- Идентификаторы
    id VARCHAR(255) NOT NULL,
    m_type VARCHAR(50) NOT NULL,
    
    -- Значения (одно из них заполняется в зависимости от типа)
    delta BIGINT,
    value DOUBLE PRECISION,
    
    -- Хэш для проверки целостности
    hash VARCHAR(128),
    
    -- Метаданные
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    -- Индексы
    PRIMARY KEY (id, m_type),
    
    -- Ограничения
    CONSTRAINT check_metric_values CHECK (
        (m_type = 'counter' AND delta IS NOT NULL AND value IS NULL) OR
        (m_type = 'gauge' AND value IS NOT NULL AND delta IS NULL)
    ),
    CONSTRAINT check_positive_delta CHECK (
        delta IS NULL OR delta >= 0
    )
);

-- Индексы для быстрого поиска
CREATE INDEX idx_metrics_id ON metrics(id);
CREATE INDEX idx_metrics_type ON metrics(m_type);
CREATE INDEX idx_metrics_updated ON metrics(updated_at DESC);

-- Триггер для обновления updated_at
create or replace function update_updated_at_column()
returns trigger
as $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$
language 'plpgsql'
;

CREATE TRIGGER update_metrics_updated_at
    BEFORE UPDATE ON metrics
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS metrics;
-- +goose StatementEnd
