-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

CREATE TABLE IF NOT EXISTS pvz (
                                   id UUID PRIMARY KEY,
                                   registration_date TIMESTAMPTZ NOT NULL,
                                   city VARCHAR(50)
    );

CREATE TABLE IF NOT EXISTS receptions (
                                          id UUID PRIMARY KEY,
                                          pvz_id UUID NOT NULL,
                                          datetime TIMESTAMPTZ NOT NULL,
                                          status VARCHAR(20) NOT NULL,
    CONSTRAINT fk_pvz FOREIGN KEY (pvz_id)
    REFERENCES pvz (id)
    ON DELETE CASCADE
    );

CREATE TABLE IF NOT EXISTS products (
                                        id UUID PRIMARY KEY,
                                        reception_id UUID NOT NULL,
                                        datetime TIMESTAMPTZ NOT NULL,
                                        type VARCHAR(20) NOT NULL,
    CONSTRAINT fk_reception FOREIGN KEY (reception_id)
    REFERENCES receptions (id)
    ON DELETE CASCADE
    );

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
