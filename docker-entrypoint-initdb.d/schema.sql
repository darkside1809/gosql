CREATE TABLE customers (
   id         BIGSERIAL PRIMARY KEY,
   name       TEXT      NOT NULL,
   phone      TEXT      NOT NULL UNIQUE,
   password   TEXT      NOT NULL,
   active     BOOLEAN   NOT NULL DEFAULT TRUE,
   created    TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);


CREATE TABLE managers (
   id         BIGSERIAL PRIMARY KEY,
   name       TEXT      NOT NULL,
   salary     INTEGER   NOT NULL DEFAULT 0,
   plan       INTEGER   NOT NULL DEFAULT 0,
   boss_id    BIGINT    REFERENCES managers,
   department TEXT,
   phone      TEXT      NOT NULL UNIQUE,
   password   TEXT      NOT NULL,
   is_admin   BOOLEAN   NOT NULL DEFAULT TRUE,
   active     BOOLEAN   NOT NULL DEFAULT TRUE,
   created    TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
); 

CREATE TABLE users (
   id         BIGSERIAL PRIMARY KEY,
   name       TEXT      NOT NULL,
   phone      TEXT      NOT NULL UNIQUE,
   password   TEXT      NOT NULL,
   roles      TEXT[]    NOT NULL DEFAULT '{}',
   active     BOOLEAN   NOT NULL DEFAULT TRUE,
   created    TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE customers_tokens (
   token       TEXT       NOT NULL UNIQUE,
   customer_id BIGINT     NOT NULL REFERENCES customers,
   expire      TIMESTAMP  NOT NULL DEFAULT CURRENT_TIMESTAMP + INTERVAL '1 hour',
   created     TIMESTAMP  NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE managers_tokens (
   token       TEXT NOT NULL      UNIQUE,
   manager_id  BIGINT NOT NULL    REFERENCES managers,
   expire      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP + INTERVAL '1 hour',
   created     TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE products (
   id      BIGSERIAL PRIMARY KEY,
   name    TEXT      NOT NULL,
   price   BIGINT    NOT NULL CHECK(price >= 0),
   qty     BIGINT    NOT NULL DEFAULT 0 CHECK(qty >= 0),
   active  BOOLEAN   NOT NULL DEFAULT TRUE,
   created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP);
   
CREATE TABLE sales (
   id          BIGSERIAL          PRIMARY KEY,
   manager_id  BIGINT NOT NULL    REFERENCES managers,
   customer_id BIGINT NOT NULL    REFERENCES customers,
   created     TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE sale_positions (
    id          BIGSERIAL           PRIMARY KEY,
    product_id  BIGINT    NOT NULL  REFERENCES products,
    sale_id     BIGINT    NOT NULL  REFERENCES sales,
    price       BIGINT    NOT NULL  CHECK(price >= 0),
    qty         BIGINT    NOT NULL  DEFAULT 0 CHECK(qty >= 0),
    created     TIMESTAMP NOT NULL  DEFAULT CURRENT_TIMESTAMP
);
