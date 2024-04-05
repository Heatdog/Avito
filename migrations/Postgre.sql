CREATE TABLE IF NOT EXISTS tags(
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) UNIQUE
);

CREATE TABLE IF NOT EXISTS features(
    id SERIAL PRIMARY KEY,
    name VARCHAR(255)
);

CREATE TABLE IF NOT EXISTS banners(
    id SERIAL PRIMARY KEY,
    content json NOT NULL,
    is_active BOOLEAN DEFAULT true,
    feature_id INTEGER NOT NULL REFERENCES features(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS banners_to_tags(
    banner_id INTEGER NOT NULL REFERENCES banners(id) ON DELETE CASCADE,
    tag_id INTEGER NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    CONSTRAINT banners_to_tags_pk PRIMARY KEY(banner_id,tag_id)
);

CREATE OR REPLACE PROCEDURE new_banner_row(banner_id integer, tag_id integer)
LANGUAGE SQL
AS $$
    IF EXISTS(SELECT
    FROM banners b
    JOIN banners_to_tags bt ON b.id = bt.banner_id AND b.id = banner_id AND bt.tag_id = tag_id
    GROUP BY b.feature_id, bt.tag_id
    HAVING COUNT(*) = 1) THEN
        ROLLBACK;
    END IF;
$$;

CREATE OR REPLACE TRIGGER new_banner 
BEFORE INSERT ON banners_to_tags
FOR EACH ROW 
EXECUTE new_banner_row(NEW.banner_id, NEW.tag_id);