CREATE TABLE IF NOT EXISTS categories (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL UNIQUE,
  icon TEXT,           -- optional
  color TEXT           -- optional
);

CREATE TABLE IF NOT EXISTS records (
  id INTEGER PRIMARY KEY,
  date TEXT NOT NULL,  -- format: YYYY-MM-DD
  description TEXT NOT NULL,
  category_id INTEGER,
  amount REAL NOT NULL,
  type TEXT NOT NULL CHECK (type IN ('income', 'expense', 'transfer')),
  notes TEXT,
  FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS summary (
  month TEXT PRIMARY KEY,  -- format: YYYY-MM
  total_income REAL DEFAULT 0,
  total_expense REAL DEFAULT 0,
  opening_balance REAL DEFAULT 0,
  net_balance REAL DEFAULT 0,
  closing_balance REAL DEFAULT 0
);
