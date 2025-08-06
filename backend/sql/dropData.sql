-- Remove all rows from the records table
DELETE FROM records;
DELETE FROM categories;
DELETE FROM summary;

-- Reset AUTOINCREMENT counter if used
DELETE FROM sqlite_sequence WHERE name='records';
DELETE FROM sqlite_sequence WHERE name='categories';
DELETE FROM sqlite_sequence WHERE name='summary';
