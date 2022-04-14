ALTER TABLE license
    ADD COLUMN last_used timestamp with time zone DEFAULT NULL;

ALTER TABLE license
    ADD COLUMN tags character varying(64)[] NOT NULL DEFAULT '{}';

ALTER TABLE license
    ADD COLUMN name character varying(64) NOT NULL DEFAULT '';

ALTER TABLE license_session
    ADD COLUMN app_version character varying(32) NOT NULL DEFAULT '';
