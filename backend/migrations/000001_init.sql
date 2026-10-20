-- Ticket-Katon: initial schema (academic demo).
-- Matches the schema already running on Supabase, so it is safe to run there
-- too: every statement is IF NOT EXISTS / duplicate-safe.

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- ---------------------------------------------------------------- enums
DO $$ BEGIN CREATE TYPE transport_type        AS ENUM ('BUS', 'LAUNCH', 'TRAIN', 'AIR');                          EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN CREATE TYPE user_role             AS ENUM ('CUSTOMER', 'OPERATOR_ADMIN', 'SUPER_ADMIN');              EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN CREATE TYPE trip_status           AS ENUM ('SCHEDULED', 'RUNNING', 'COMPLETED', 'CANCELLED');         EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN CREATE TYPE seat_inventory_status AS ENUM ('AVAILABLE', 'HELD', 'BOOKED');                            EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN CREATE TYPE booking_status       AS ENUM ('PENDING', 'CONFIRMED', 'CANCELLED', 'REFUNDED');           EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN CREATE TYPE payment_gateway      AS ENUM ('BKASH', 'NAGAD', 'SSLCOMMERZ', 'MOCK');                    EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN CREATE TYPE payment_status       AS ENUM ('INITIATED', 'SUCCESS', 'FAILED', 'REFUNDED');              EXCEPTION WHEN duplicate_object THEN NULL; END $$;

-- ---------------------------------------------------------------- tables
CREATE TABLE IF NOT EXISTS operators (
    id            UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name          VARCHAR(255) NOT NULL,
    type          transport_type NOT NULL,
    trade_license VARCHAR(255),
    contact_phone VARCHAR(50) NOT NULL,
    created_at    TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE IF NOT EXISTS users (
    id            UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    full_name     VARCHAR(255) NOT NULL,
    phone_number  VARCHAR(50) NOT NULL UNIQUE,
    email         VARCHAR(255),
    password_hash VARCHAR(255),
    role          user_role DEFAULT 'CUSTOMER',
    nid_number    VARCHAR(50),
    created_at    TIMESTAMPTZ DEFAULT now(),
    updated_at    TIMESTAMPTZ DEFAULT now(),
    operator_id   UUID REFERENCES operators(id) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS vehicles (
    id                 UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    operator_id        UUID NOT NULL REFERENCES operators(id) ON DELETE CASCADE,
    registration_no    VARCHAR(100) NOT NULL,
    type               transport_type NOT NULL,
    seat_layout_config JSONB NOT NULL,
    created_at         TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE IF NOT EXISTS routes (
    id                 UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    origin_city        VARCHAR(100) NOT NULL,
    destination_city   VARCHAR(100) NOT NULL,
    intermediate_stops JSONB DEFAULT '[]'::jsonb,
    created_at         TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE IF NOT EXISTS trips (
    id             UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    vehicle_id     UUID NOT NULL REFERENCES vehicles(id),
    route_id       UUID NOT NULL REFERENCES routes(id),
    departure_time TIMESTAMPTZ NOT NULL,
    arrival_time   TIMESTAMPTZ NOT NULL,
    base_fare      NUMERIC(12,2) NOT NULL CHECK (base_fare >= 0),
    status         trip_status DEFAULT 'SCHEDULED',
    created_at     TIMESTAMPTZ DEFAULT now(),
    updated_at     TIMESTAMPTZ DEFAULT now(),
    CHECK (arrival_time > departure_time)
);

CREATE TABLE IF NOT EXISTS trip_seat_inventory (
    id            UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    trip_id       UUID NOT NULL REFERENCES trips(id) ON DELETE CASCADE,
    seat_number   VARCHAR(50) NOT NULL,
    deck_or_class VARCHAR(100) NOT NULL,
    fare          NUMERIC(12,2) NOT NULL CHECK (fare >= 0),
    status        seat_inventory_status DEFAULT 'AVAILABLE',
    held_until    TIMESTAMPTZ,
    updated_at    TIMESTAMPTZ DEFAULT now(),
    CONSTRAINT uq_trip_seat UNIQUE (trip_id, seat_number)
);

CREATE TABLE IF NOT EXISTS bookings (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id         UUID NOT NULL REFERENCES users(id),
    trip_id         UUID NOT NULL REFERENCES trips(id),
    pnr_code        VARCHAR(30) NOT NULL UNIQUE,
    total_amount    NUMERIC(12,2) NOT NULL CHECK (total_amount >= 0),
    discount_amount NUMERIC(12,2) DEFAULT 0.00 CHECK (discount_amount >= 0),
    status          booking_status DEFAULT 'PENDING',
    created_at      TIMESTAMPTZ DEFAULT now(),
    updated_at      TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE IF NOT EXISTS tickets (
    id               UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    booking_id       UUID NOT NULL REFERENCES bookings(id) ON DELETE CASCADE,
    seat_number      VARCHAR(50) NOT NULL,
    passenger_name   VARCHAR(255) NOT NULL,
    passenger_phone  VARCHAR(50) NOT NULL,
    passenger_gender VARCHAR(20),
    ticket_number    VARCHAR(60) NOT NULL UNIQUE,
    created_at       TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE IF NOT EXISTS payments (
    id                 UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    booking_id         UUID NOT NULL REFERENCES bookings(id),
    gateway            payment_gateway NOT NULL,
    transaction_id     VARCHAR(255) UNIQUE,
    gateway_payment_id VARCHAR(255),
    amount             NUMERIC(12,2) NOT NULL CHECK (amount >= 0),
    currency           VARCHAR(10) DEFAULT 'BDT',
    status             payment_status DEFAULT 'INITIATED',
    raw_payload        JSONB DEFAULT '{}'::jsonb,
    created_at         TIMESTAMPTZ DEFAULT now(),
    updated_at         TIMESTAMPTZ DEFAULT now()
);

-- --------------------------------------------------------------- indexes
CREATE UNIQUE INDEX IF NOT EXISTS users_email_lower_key      ON users (lower(email)) WHERE email IS NOT NULL;
CREATE INDEX        IF NOT EXISTS idx_users_operator_id      ON users (operator_id);
CREATE INDEX        IF NOT EXISTS idx_vehicles_operator_id   ON vehicles (operator_id);
CREATE INDEX        IF NOT EXISTS idx_trips_search           ON trips (route_id, departure_time, status);
CREATE INDEX        IF NOT EXISTS idx_trip_seat_status       ON trip_seat_inventory (trip_id, status);
CREATE INDEX        IF NOT EXISTS idx_bookings_user_created  ON bookings (user_id, created_at DESC);
CREATE INDEX        IF NOT EXISTS idx_bookings_trip_id       ON bookings (trip_id);
CREATE INDEX        IF NOT EXISTS idx_tickets_booking_id     ON tickets (booking_id);
CREATE INDEX        IF NOT EXISTS idx_payments_booking_id    ON payments (booking_id);
