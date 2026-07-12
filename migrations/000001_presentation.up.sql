CREATE SCHEMA presentation;


CREATE TABLE presentation.brands (
    name            VARCHAR(256)    UNIQUE NOT NULL,
    password        TEXT            NOT NULL
);

CREATE TABLE presentation.works (
    brand               VARCHAR(256)    NOT NULL,
    workName            VARCHAR(256)    NOT NULL,
    url                 TEXT            DEFAULT '',
    description         TEXT            NOT NULL DEFAULT '',

    CONSTRAINT presentation_work FOREIGN KEY (brand) REFERENCES presentation.brands(name)
        ON DELETE CASCADE
        ON UPDATE CASCADE,
    CONSTRAINT unique_brand_work UNIQUE (brand, workName)
);

CREATE INDEX work_name_idx ON presentation.works(workName);