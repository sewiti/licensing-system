ALTER TABLE license_issuer
    ADD COLUMN email character varying(128) NOT NULL DEFAULT '';

ALTER TABLE license_issuer
    ADD COLUMN phone_number character varying(24) NOT NULL DEFAULT '';
