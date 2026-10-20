-- Reverts 000001_init.sql. Destroys all data. Local/disposable databases only.
DROP TABLE IF EXISTS payments CASCADE;
DROP TABLE IF EXISTS tickets CASCADE;
DROP TABLE IF EXISTS bookings CASCADE;
DROP TABLE IF EXISTS trip_seat_inventory CASCADE;
DROP TABLE IF EXISTS trips CASCADE;
DROP TABLE IF EXISTS routes CASCADE;
DROP TABLE IF EXISTS vehicles CASCADE;
DROP TABLE IF EXISTS users CASCADE;
DROP TABLE IF EXISTS operators CASCADE;

DROP TYPE IF EXISTS payment_status;
DROP TYPE IF EXISTS payment_gateway;
DROP TYPE IF EXISTS booking_status;
DROP TYPE IF EXISTS seat_inventory_status;
DROP TYPE IF EXISTS trip_status;
DROP TYPE IF EXISTS user_role;
DROP TYPE IF EXISTS transport_type;
