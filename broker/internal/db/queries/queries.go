package queries

const SchemaCreationStatement string = `
CREATE TABLE IF NOT EXISTS TopicMsgs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name VARCHAR(255) UNIQUE NOT NULL,
    msg TEXT NULL
);
`

const UpdateLatestMsgMsgStatement string = `
INSERT INTO TopicMsgs (name, msg)
VALUES (?, ?)
ON CONFLICT(name) DO UPDATE SET msg=excluded.msg;
`

const GetLatestMsgStatement string = `
SELECT msg FROM TopicMsgs WHERE name = ?;
`

const GetAllTopicsStatement string = `
SELECT name FROM TopicMsgs;
`
