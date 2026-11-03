-- One row is one Problem aggregate (solutions, expert match, pool and ledger)
CREATE TABLE IF NOT EXISTS problems (
    id TEXT PRIMARY KEY,
    document TEXT NOT NULL CHECK(json_valid(document))
);

CREATE TABLE IF NOT EXISTS outbox (
    id TEXT PRIMARY KEY,
    document TEXT NOT NULL CHECK(json_valid(document)),
    delivered INTEGER NOT NULL DEFAULT 0 CHECK(delivered IN (0,1))
);

CREATE TABLE IF NOT EXISTS audit (
    event_id TEXT PRIMARY KEY,
    document TEXT NOT NULL CHECK(json_valid(document))
);

PRAGMA user_version = 1;
