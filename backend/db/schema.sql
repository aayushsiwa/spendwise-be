CREATE TABLE IF NOT EXISTS records (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  date TEXT NOT NULL,
  description TEXT,
  category TEXT,
  amount REAL NOT NULL,
  type TEXT CHECK(type IN ('income', 'expense')),
  notes TEXT
);
