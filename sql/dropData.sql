-- Remove all rows from the records table
DELETE FROM records;
DELETE FROM categories;
DELETE FROM summary;
DELETE FROM summary_details;

-- Reset AUTOINCREMENT counter if used
DELETE FROM sqlite_sequence WHERE name='records';
DELETE FROM sqlite_sequence WHERE name='categories';
DELETE FROM sqlite_sequence WHERE name='summary';
DELETE FROM sqlite_sequence WHERE name='summary_details';
