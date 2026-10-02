\connect cars
CREATE TABLE cars
(
    id                  SERIAL PRIMARY KEY,
    car_uid             uuid UNIQUE NOT NULL,
    brand               VARCHAR(80) NOT NULL,
    model               VARCHAR(80) NOT NULL,
    registration_number VARCHAR(20) NOT NULL,
    power               INT,
    price               INT         NOT NULL,
    type                VARCHAR(20)
        CHECK (type IN ('SEDAN', 'SUV', 'MINIVAN', 'ROADSTER')),
    availability        BOOLEAN     NOT NULL
);

INSERT INTO cars (car_uid, brand, model, registration_number, power, price, type, availability) VALUES
  ('11111111-1111-1111-1111-111111111111', 'Toyota', 'Camry', 'А001АА01', 200, 2500, 'SEDAN', true),
  ('22222222-2222-2222-2222-222222222222', 'BMW', 'X5', 'А002АА01', 300, 5000, 'SUV', true),
  ('33333333-3333-3333-3333-333333333333', 'Kia', 'Rio', 'А003АА01', 120, 1500, 'SEDAN', true),
  ('44444444-4444-4444-4444-444444444444', 'Audi', 'Q7', 'А004АА01', 333, 6000, 'SUV', true),
  ('55555555-5555-5555-5555-555555555555', 'Ford', 'Focus', 'А005АА01', 150, 1800, 'SEDAN', true),
  ('66666666-6666-6666-6666-666666666666', 'Volkswagen', 'Tiguan', 'А006АА01', 180, 3000, 'SUV', true),
  ('77777777-7777-7777-7777-777777777777', 'Skoda', 'Octavia', 'А007АА01', 160, 2200, 'SEDAN', true),
  ('88888888-8888-8888-8888-888888888888', 'Hyundai', 'Tucson', 'А008АА01', 170, 2800, 'SUV', true),
  ('99999999-9999-9999-9999-999999999999', 'Nissan', 'Qashqai', 'А009АА01', 160, 2700, 'SUV', true),
  ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'Mazda', 'CX-5', 'А010АА01', 190, 3200, 'SUV', true);

INSERT INTO cars (car_uid, brand, model, registration_number, power, price, type, availability)
VALUES ('109b42f3-198d-4c89-9276-a7520a7120ab',
        'Mercedes Benz',
        'GLA 250',
        'ЛО777Х799',
        249,
        3500,
        'SEDAN',
        true)
ON CONFLICT (car_uid) DO NOTHING;

\connect rentals
CREATE TABLE IF NOT EXISTS rental
(
    id          SERIAL PRIMARY KEY,
    rental_uid  uuid UNIQUE              NOT NULL,
    username    VARCHAR(80)              NOT NULL,
    payment_uid uuid                     NOT NULL,
    car_uid     uuid                     NOT NULL,
    date_from   TIMESTAMP WITH TIME ZONE NOT NULL,
    date_to     TIMESTAMP WITH TIME ZONE NOT NULL,
    status      VARCHAR(20)              NOT NULL
        CHECK (status IN ('IN_PROGRESS', 'FINISHED', 'CANCELED'))
);

\connect payments
CREATE TABLE IF NOT EXISTS payment
(
    id          SERIAL PRIMARY KEY,
    payment_uid uuid        NOT NULL,
    status      VARCHAR(20) NOT NULL
        CHECK (status IN ('PAID', 'CANCELED')),
    price       INT         NOT NULL
);

\connect cars
GRANT USAGE ON SCHEMA public TO program;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO program;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO program;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO program;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT USAGE, SELECT ON SEQUENCES TO program;

\connect rentals
GRANT USAGE ON SCHEMA public TO program;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO program;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO program;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO program;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT USAGE, SELECT ON SEQUENCES TO program;

\connect payments
GRANT USAGE ON SCHEMA public TO program;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO program;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO program;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO program;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT USAGE, SELECT ON SEQUENCES TO program;