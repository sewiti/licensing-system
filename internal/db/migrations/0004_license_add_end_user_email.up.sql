ALTER TABLE license
    ADD COLUMN end_user_email character varying(128) NOT NULL DEFAULT '';
