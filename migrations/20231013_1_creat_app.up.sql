CREATE TABLE ygm_user (
    id BIGSERIAL PRIMARY KEY,
    login_hash VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL UNIQUE
);
CREATE TABLE ygm_order_status (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL UNIQUE,
    ident VARCHAR(50) NOT NULL UNIQUE
);
CREATE TABLE ygm_order (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES ygm_user (id) NOT NULL,
    status_id BIGINT REFERENCES ygm_order_status (id) NOT NULL,
    num BIGINT NOT NULL UNIQUE,
    order_date TIMESTAMP NOT NULL
);
CREATE TABLE ygm_balls_operation (
    id BIGSERIAL PRIMARY KEY,
    order_id BIGINT REFERENCES ygm_order (id) NOT NULL,
    amount NUMERIC(2),
);