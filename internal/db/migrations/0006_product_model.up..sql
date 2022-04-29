CREATE TABLE product
(
    id            serial                   NOT NULL,
    active        boolean                  NOT NULL DEFAULT true,
    name          character varying(64)    NOT NULL,
    contact_email character varying(128)   NOT NULL DEFAULT '',
    data          bytea,
    created       timestamp with time zone NOT NULL DEFAULT NOW(),
    updated       timestamp with time zone NOT NULL DEFAULT NOW(),
    issuer_id     integer                  NOT NULL,

    CONSTRAINT product_pkey           PRIMARY KEY (id),
    CONSTRAINT product_issuer_id_fkey FOREIGN KEY (issuer_id)
        REFERENCES license_issuer (id) MATCH SIMPLE
        ON UPDATE RESTRICT
        ON DELETE CASCADE
        NOT VALID
);
