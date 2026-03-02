package queries

const SchemaCreationStatement string = `
CREATE TABLE IF NOT EXISTS TopicMsgs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name VARCHAR(255) UNIQUE NOT NULL,
    msg VARCHAR(255) NULL,
);
`
