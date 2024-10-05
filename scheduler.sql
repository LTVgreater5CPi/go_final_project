CREATE TABLE IF NOT EXISTS scheduler (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    date TEXT NOT NULL,
    title TEXT NOT NULL,
    comment TEXT,
    repeat TEXT CHECK(LENGTH(repeat) <= 128)
);

-- Создаем индекс по дате для сортировки задач по дате
CREATE INDEX idx_scheduler_date ON scheduler(date);
