/* Copyright 2014 The Cactus Authors. All rights reserved. */

PRAGMA journal_mode = WAL;

CREATE TABLE IF NOT EXISTS accounts (
    "id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    "handle_lower" TEXT NOT NULL UNIQUE,
    "name_lower" TEXT NOT NULL,
    "level" INTEGER NOT NULL,
    "struct" BLOB NOT NULL,
    "created" TIMESTAMP NOT NULL,
    "modified" TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS activities (
    "id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    "actor_id" INTEGER NOT NULL,
    "struct" BLOB NOT NULL,
    "created" TIMESTAMP NOT NULL,
    "modified" TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS clarifications (
    "id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    "asker_id" INTEGER NOT NULL,
    "response" INTEGER NOT NULL,
    "struct" BLOB NOT NULL,
    "created" TIMESTAMP NOT NULL,
    "modified" TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS contests (
    "id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    "struct" BLOB NOT NULL,
    "salt" BLOB NOT NULL,
    "ready" BOOLEAN NOT NULL,
    "created" TIMESTAMP NOT NULL,
    "modified" TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS executions (
    "id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    "struct" BLOB NOT NULL,
    "created" TIMESTAMP NOT NULL,
    "modified" TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS problems (
    "id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    "slug" TEXT NOT NULL UNIQUE,
    "char" TEXT NOT NULL,
    "struct" BLOB NOT NULL,
    "created" TIMESTAMP NOT NULL,
    "modified" TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS notifications (
    "id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    "account_id" INTEGER NOT NULL,
    "account_level" INTEGER NOT NULL,
    "struct" BLOB NOT NULL,
    "created" TIMESTAMP NOT NULL,
    "modified" TIMESTAMP NOT NULL
);


CREATE TABLE IF NOT EXISTS standings (
    "id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    "author_id" INTEGER NOT NULL,
    "score" INTEGER NOT NULL,
    "penalty" INTEGER NOT NULL,
    "struct" BLOB NOT NULL,
    "created" TIMESTAMP NOT NULL,
    "modified" TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS submissions (
    "id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    "author_id" INTEGER NOT NULL,
    "struct" BLOB NOT NULL,
    "created" TIMESTAMP NOT NULL,
    "modified" TIMESTAMP NOT NULL
);

INSERT OR IGNORE INTO contests(id, struct, salt, ready, created, modified) VALUES (1, "", RANDOMBLOB(32), 0, STRFTIME('%s', 'now'), STRFTIME('%s', 'now'));
