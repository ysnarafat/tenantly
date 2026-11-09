CREATE TABLE support_case_messages (
    id SERIAL PRIMARY KEY,
    case_id INTEGER NOT NULL REFERENCES support_cases(id) ON DELETE CASCADE,
    author_id INTEGER NOT NULL REFERENCES users(id),
    body TEXT NOT NULL CHECK (length(btrim(body)) > 0),
    created_at TIMESTAMP NOT NULL DEFAULT (NOW() AT TIME ZONE 'UTC')
);

CREATE INDEX idx_support_case_messages_case ON support_case_messages(case_id, created_at, id);
