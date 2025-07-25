-- Remove all rows from the records table
DELETE FROM records;

-- Reset AUTOINCREMENT counter if used
DELETE FROM sqlite_sequence WHERE name='records';
