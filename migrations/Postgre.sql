CREATE TABLE IF NOT EXISTS tags(
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) UNIQUE
);

CREATE TABLE IF NOT EXISTS features(
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) UNIQUE
);

CREATE TABLE IF NOT EXISTS banners(
    id SERIAL PRIMARY KEY,
    content json NOT NULL,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now()
);

CREATE TABLE IF NOT EXISTS features_tags_to_banners(
    feature_id INTEGER NOT NULL REFERENCES features(id) ON DELETE CASCADE,
    tag_id INTEGER NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    banner_id INTEGER NOT NULL REFERENCES banners(id) ON DELETE CASCADE,
    CONSTRAINT features_tags_to_banners_pk PRIMARY KEY(feature_id,tag_id)
);

CREATE INDEX banners_idx ON features_tags_to_banners(banner_id);

