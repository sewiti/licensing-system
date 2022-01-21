CREATE TABLE issuers
(
    id       serial                   NOT NULL,
    email    character varying(64)    NOT NULL,
    password bytea                    NOT NULL,
    salt     bytea                    NOT NULL,
    created  timestamp with time zone DEFAULT NOW(),

    CONSTRAINT issuers_pkey PRIMARY KEY (id),
    CONSTRAINT issuers_email_unique UNIQUE (email)
)
