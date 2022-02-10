CREATE TABLE license_issuers
(
    id            serial                   NOT NULL,
    active        boolean                  NOT NULL DEFAULT true,
    username      character varying(64)    NOT NULL,
    password_hash character varying(128)   NOT NULL,
    max_licenses  integer                  NOT NULL DEFAULT 1,
    created       timestamp with time zone NOT NULL DEFAULT NOW(),
    last_active   timestamp with time zone NOT NULL DEFAULT NOW(),

    CONSTRAINT license_issuers_pkey            PRIMARY KEY (id),
    CONSTRAINT license_issuers_username_unique UNIQUE      (username)
);

CREATE TABLE licenses
(
    id           bytea                    NOT NULL,
    key          bytea                    NOT NULL,
    note         character varying(256)   NOT NULL DEFAULT '',
    data         bytea,
    max_sessions integer                  NOT NULL DEFAULT 1,
    valid_until  timestamp with time zone,
    created      timestamp with time zone NOT NULL DEFAULT NOW(),
    updated      timestamp with time zone NOT NULL DEFAULT NOW(),
    issuer_id    integer                  NOT NULL,

    CONSTRAINT licenses_pkey           PRIMARY KEY (id),
    CONSTRAINT licenses_key_unique     UNIQUE      (key),
    CONSTRAINT licenses_issuer_id_fkey FOREIGN KEY (issuer_id)
        REFERENCES license_issuers (id) MATCH SIMPLE
        ON UPDATE RESTRICT
        ON DELETE RESTRICT
        NOT VALID
);

CREATE TABLE license_sessions
(
    client_session_id  bytea                    NOT NULL,
    server_session_id  bytea                    NOT NULL,
    server_session_key bytea                    NOT NULL,
    machine_uuid       bytea                    NOT NULL,
    created            timestamp with time zone NOT NULL DEFAULT NOW(),
    expire             timestamp with time zone NOT NULL,
    license_id         bytea                    NOT NULL,

    CONSTRAINT license_sessions_pkey                PRIMARY KEY (client_session_id),
    CONSTRAINT license_sessions_machine_uuid_unique UNIQUE      (machine_uuid)
        INCLUDE (client_session_id),
    CONSTRAINT license_sessions_license_id_fkey     FOREIGN KEY (license_id)
        REFERENCES licenses (id) MATCH SIMPLE
        ON UPDATE RESTRICT
        ON DELETE RESTRICT
        NOT VALID
);

INSERT INTO license_issuers (id, active, username, password_hash, max_licenses) VALUES
    (0, false, 'superadmin', 'nologin', -1);
