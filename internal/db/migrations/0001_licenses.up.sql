CREATE TABLE licenses
(
    id        serial  NOT NULL,
    key       bytea   NOT NULL,
    issuer_id integer NOT NULL,

    CONSTRAINT licenses_pkey PRIMARY KEY (id),
    CONSTRAINT licenses_key_unique UNIQUE (key)
)
