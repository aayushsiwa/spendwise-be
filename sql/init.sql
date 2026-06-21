CREATE TABLE IF NOT EXISTS categories (
  "ID" TEXT PRIMARY KEY,
  name TEXT NOT NULL UNIQUE,
  icon TEXT,           -- optional
  color TEXT           -- optional
);

CREATE TABLE IF NOT EXISTS records (
  "ID" TEXT PRIMARY KEY,
  date TEXT NOT NULL,  -- format: YYYY-MM-DD
  description TEXT NOT NULL,
  "categoryID" TEXT,
  amount REAL NOT NULL,
  type TEXT NOT NULL CHECK (type IN ('income', 'expense', 'transfer')),
  note TEXT,
  balance REAL NOT NULL,
  FOREIGN KEY ("categoryID") REFERENCES categories("ID") ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS summary (
  month TEXT PRIMARY KEY,  -- format: YYYY-MM
  "totalIncome" REAL NOT NULL DEFAULT 0,
  "totalExpense" REAL NOT NULL DEFAULT 0,
  "openingBalance" REAL NOT NULL DEFAULT 0,
  "netBalance" REAL NOT NULL DEFAULT 0,
  "closingBalance" REAL NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS summary_details (
  "ID" TEXT PRIMARY KEY REFERENCES records("ID"),
  month TEXT NOT NULL,  -- format: YYYY-MM
  type TEXT NOT NULL CHECK (type IN ('income', 'expense','transfer')),
  "categoryID" TEXT NOT NULL,
  "categoryName" TEXT NOT NULL,
  amount REAL NOT NULL
);