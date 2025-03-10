CREATE TABLE slot (
                      id bigint PRIMARY KEY,
                      name text,
                      description text,
                      "npHandle" text,
                      "publishedIn" text,
                      game smallint,
                      "firstPublished" bigint,
                      "lastUpdated" bigint,
                      "heartCount" bigint,
                      background text,
                      icon bytea,
                      "rootLevel" bytea,
                      missing_root_level boolean
);

CREATE UNIQUE INDEX slot_pkey ON slot(id int8_ops);
CREATE INDEX slot_description_gin_trgm_idx ON slot USING GIN (description gin_trgm_ops);
CREATE INDEX slot_name_gin_trgm_idx ON slot USING GIN (name gin_trgm_ops);
CREATE INDEX slot_nphandle_gin_trgm_idx ON slot USING GIN ("npHandle" gin_trgm_ops);