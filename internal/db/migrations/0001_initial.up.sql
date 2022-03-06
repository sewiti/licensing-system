CREATE TABLE license_issuer
(
    id            serial                   NOT NULL,
    active        boolean                  NOT NULL DEFAULT true,
    username      character varying(64)    NOT NULL,
    password_hash character varying(128)   NOT NULL,
    max_licenses  integer                  NOT NULL DEFAULT 1,
    created       timestamp with time zone NOT NULL DEFAULT NOW(),
    updated       timestamp with time zone NOT NULL DEFAULT NOW(),

    CONSTRAINT license_issuer_pkey            PRIMARY KEY (id),
    CONSTRAINT license_issuer_username_unique UNIQUE      (username)
);

CREATE TABLE license
(
    id           bytea                    NOT NULL,
    key          bytea                    NOT NULL,
    note         character varying(500)   NOT NULL DEFAULT '',
    data         bytea,
    max_sessions integer                  NOT NULL DEFAULT 1,
    valid_until  timestamp with time zone,
    created      timestamp with time zone NOT NULL DEFAULT NOW(),
    updated      timestamp with time zone NOT NULL DEFAULT NOW(),
    issuer_id    integer                  NOT NULL,

    CONSTRAINT license_pkey           PRIMARY KEY (id),
    CONSTRAINT license_key_unique     UNIQUE      (key),
    CONSTRAINT license_issuer_id_fkey FOREIGN KEY (issuer_id)
        REFERENCES license_issuer (id) MATCH SIMPLE
        ON UPDATE RESTRICT
        ON DELETE RESTRICT
        NOT VALID
);

CREATE TABLE license_session
(
    client_session_id  bytea                    NOT NULL,
    server_session_id  bytea                    NOT NULL,
    server_session_key bytea                    NOT NULL,
    identifier         character varying(300)   NOT NULL,
    machine_id         bytea                    NOT NULL,
    created            timestamp with time zone NOT NULL DEFAULT NOW(),
    expire             timestamp with time zone NOT NULL,
    license_id         bytea                    NOT NULL,

    CONSTRAINT license_session_pkey              PRIMARY KEY (client_session_id),
    CONSTRAINT license_session_machine_id_unique UNIQUE      (machine_id)
        INCLUDE (client_session_id),
    CONSTRAINT license_session_license_id_fkey   FOREIGN KEY (license_id)
        REFERENCES license (id) MATCH SIMPLE
        ON UPDATE RESTRICT
        ON DELETE RESTRICT
        NOT VALID
);

INSERT INTO license_issuer (id, active, username, password_hash, max_licenses) VALUES
    (0, false, 'superadmin', 'nologin', 0);
