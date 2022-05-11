ALTER TABLE license
    ADD COLUMN product_id integer DEFAULT NULL;

ALTER TABLE license
    ADD CONSTRAINT license_product_id_fkey FOREIGN KEY (product_id)
        REFERENCES product (id) MATCH SIMPLE
        ON UPDATE RESTRICT
        ON DELETE CASCADE
        NOT VALID;
