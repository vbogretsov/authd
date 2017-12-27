CREATE EXTENSION pgcrypto;

CREATE TABLE accounts (
    id        UUID NOT NULL,
    username  VARCHAR(128) NOT NULL,
    password  VARCHAR(256) NOT NULL,
    active    BOOLEAN NOT NULL,
    created   TIMESTAMP NOT NULL,
    lastlogin TIMESTAMP NOT NULL,

    CONSTRAINT pk_accounts PRIMARY KEY (id),
    CONSTRAINT uq_accounts_username UNIQUE (username)
);

CREATE TABLE demands (
    id         UUID NOT NULL,
    created    TIMESTAMP NOT NULL,
    expires    TIMESTAMP NOT NULL,
    account_id UUID NOT NULL,

    CONSTRAINT pk_demands PRIMARY KEY (id),
    CONSTRAINT fk_demands_accounts FOREIGN KEY (account_id) REFERENCES accounts (id) ON DELETE CASCADE
);
