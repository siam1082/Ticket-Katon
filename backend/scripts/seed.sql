-- Ticket-Katon demo seed. ALL DATA IS FICTIONAL (academic demo).
-- Safe to run repeatedly: fixed ids / deterministic ids + ON CONFLICT.
-- Re-running later adds trips for newly covered dates (today + next 7 days).
-- Requires 000001_init.sql and 000002_email_verification.sql to have been applied.

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- ---------------------------------------------------------- operators
INSERT INTO operators (id, name, type, contact_phone) VALUES
  ('a0000000-0000-4000-8000-000000000001', 'Demo Green Line Coaches', 'BUS'::transport_type,    '01800000001'),
  ('a0000000-0000-4000-8000-000000000002', 'Demo Padma Launch Co.',   'LAUNCH'::transport_type, '01800000002'),
  ('a0000000-0000-4000-8000-000000000003', 'Demo Inter-City Rail',    'TRAIN'::transport_type,  '01800000003')
ON CONFLICT (id) DO NOTHING;

-- ----------------------------------------------------------- vehicles
INSERT INTO vehicles (id, operator_id, registration_no, type, seat_layout_config) VALUES
  ('b0000000-0000-4000-8000-000000000001', 'a0000000-0000-4000-8000-000000000001', 'DEMO-BUS-101',    'BUS'::transport_type,    '{"type":"2x2"}'::jsonb),
  ('b0000000-0000-4000-8000-000000000002', 'a0000000-0000-4000-8000-000000000001', 'DEMO-BUS-102',    'BUS'::transport_type,    '{"type":"2x2"}'::jsonb),
  ('b0000000-0000-4000-8000-000000000003', 'a0000000-0000-4000-8000-000000000002', 'DEMO-LAUNCH-201', 'LAUNCH'::transport_type, '{"type":"cabin_and_deck"}'::jsonb),
  ('b0000000-0000-4000-8000-000000000004', 'a0000000-0000-4000-8000-000000000003', 'DEMO-TRAIN-301',  'TRAIN'::transport_type,
   '{"coaches":[{"class":"SNIGDHA","layout":"2x2","coach_name":"KA","total_seats":60},{"class":"SHOVON_CHAIR","layout":"2x3","coach_name":"KHA","total_seats":60}]}'::jsonb)
ON CONFLICT (id) DO NOTHING;

-- ------------------------------------------------------------- routes
INSERT INTO routes (id, origin_city, destination_city) VALUES
  ('c0000000-0000-4000-8000-000000000001', 'Dhaka',      'Sylhet'),
  ('c0000000-0000-4000-8000-000000000002', 'Sylhet',     'Dhaka'),
  ('c0000000-0000-4000-8000-000000000003', 'Dhaka',      'Barishal'),
  ('c0000000-0000-4000-8000-000000000004', 'Barishal',   'Dhaka'),
  ('c0000000-0000-4000-8000-000000000005', 'Dhaka',      'Chattogram'),
  ('c0000000-0000-4000-8000-000000000006', 'Chattogram', 'Dhaka')
ON CONFLICT (id) DO NOTHING;

-- -------------------------------------------------------------- trips
-- Today + next 7 days (Asia/Dhaka). Trip ids are deterministic, so re-runs do not duplicate.
WITH days AS (
  SELECT ((now() AT TIME ZONE 'Asia/Dhaka')::date + g.n) AS d
  FROM generate_series(0, 7) AS g(n)
),
plan(vehicle_id, route_id, dep_time, duration_h, base_fare) AS (
  VALUES
    ('b0000000-0000-4000-8000-000000000001'::uuid, 'c0000000-0000-4000-8000-000000000001'::uuid, time '08:00',  7,  750),
    ('b0000000-0000-4000-8000-000000000002'::uuid, 'c0000000-0000-4000-8000-000000000001'::uuid, time '22:00',  7,  850),
    ('b0000000-0000-4000-8000-000000000002'::uuid, 'c0000000-0000-4000-8000-000000000002'::uuid, time '09:00',  7,  750),
    ('b0000000-0000-4000-8000-000000000003'::uuid, 'c0000000-0000-4000-8000-000000000003'::uuid, time '19:00', 10,  600),
    ('b0000000-0000-4000-8000-000000000003'::uuid, 'c0000000-0000-4000-8000-000000000004'::uuid, time '19:30', 10,  600),
    ('b0000000-0000-4000-8000-000000000004'::uuid, 'c0000000-0000-4000-8000-000000000005'::uuid, time '07:00',  6,  400),
    ('b0000000-0000-4000-8000-000000000004'::uuid, 'c0000000-0000-4000-8000-000000000006'::uuid, time '15:00',  6,  400)
)
INSERT INTO trips (id, vehicle_id, route_id, departure_time, arrival_time, base_fare, status)
SELECT
  uuid_generate_v5(uuid_ns_dns(), p.vehicle_id::text || '|' || p.route_id::text || '|' || (d.d + p.dep_time)::text),
  p.vehicle_id,
  p.route_id,
  (d.d + p.dep_time) AT TIME ZONE 'Asia/Dhaka',
  ((d.d + p.dep_time) AT TIME ZONE 'Asia/Dhaka') + make_interval(hours => p.duration_h),
  p.base_fare,
  'SCHEDULED'::trip_status
FROM days d CROSS JOIN plan p
ON CONFLICT (id) DO NOTHING;

-- -------------------------------------------------------------- seats
-- Bus: 2x2 coach, rows A-J x seats 1-4 (A1..J4), 40 seats.
INSERT INTO trip_seat_inventory (trip_id, seat_number, deck_or_class, fare)
SELECT t.id, chr(64 + r.n) || c.n::text, 'AC_CHAIR', t.base_fare
FROM trips t
JOIN vehicles v ON v.id = t.vehicle_id AND v.type = 'BUS'::transport_type AND v.registration_no LIKE 'DEMO-%'
CROSS JOIN generate_series(1, 10) AS r(n)
CROSS JOIN generate_series(1, 4)  AS c(n)
ON CONFLICT (trip_id, seat_number) DO NOTHING;

-- Train: coach KA (Snigdha, 2x2, 60 seats) and coach KHA (Shovon chair, 2x3, 60 seats).
INSERT INTO trip_seat_inventory (trip_id, seat_number, deck_or_class, fare)
SELECT t.id, k.prefix || '-' || s.n::text, k.class, round(t.base_fare * k.mult, 2)
FROM trips t
JOIN vehicles v ON v.id = t.vehicle_id AND v.type = 'TRAIN'::transport_type AND v.registration_no LIKE 'DEMO-%'
CROSS JOIN (VALUES ('KA', 'SNIGDHA', 60, 1.4), ('KHA', 'SHOVON_CHAIR', 60, 1.0)) AS k(prefix, class, total, mult)
CROSS JOIN LATERAL generate_series(1, k.total) AS s(n)
ON CONFLICT (trip_id, seat_number) DO NOTHING;

-- Launch: 40 deck seats, 8 single cabins, 4 double cabins, 2 VIP cabins.
INSERT INTO trip_seat_inventory (trip_id, seat_number, deck_or_class, fare)
SELECT t.id, g.prefix || s.n::text, g.class, round(t.base_fare * g.mult, 2)
FROM trips t
JOIN vehicles v ON v.id = t.vehicle_id AND v.type = 'LAUNCH'::transport_type AND v.registration_no LIKE 'DEMO-%'
CROSS JOIN (VALUES
    ('DECK-',  1,  40, 'DECK',         1.0),
    ('CABIN-', 201, 208, 'CABIN_SINGLE', 2.0),
    ('CABIN-', 301, 304, 'CABIN_DOUBLE', 3.5),
    ('CABIN-', 401, 402, 'VIP_CABIN',    5.0)
) AS g(prefix, first_no, last_no, class, mult)
CROSS JOIN LATERAL generate_series(g.first_no, g.last_no) AS s(n)
ON CONFLICT (trip_id, seat_number) DO NOTHING;

-- ------------------------------------------------------- demo accounts
-- Fictional presentation accounts. Do not reuse these passwords anywhere else.
--   admin@ticketkaton.demo     / Admin@12345  (SUPER_ADMIN)
--   operator@ticketkaton.demo  / Oper@12345   (OPERATOR_ADMIN, Demo Green Line Coaches)
--   passenger@ticketkaton.demo / Demo@12345   (CUSTOMER)
INSERT INTO users (full_name, email, phone_number, password_hash, role, operator_id, email_verified_at) VALUES
  ('Demo Super Admin',    'admin@ticketkaton.demo',     '01700000001', crypt('Admin@12345', gen_salt('bf', 10)), 'SUPER_ADMIN'::user_role,    NULL, NOW()),
  ('Demo Operator Admin', 'operator@ticketkaton.demo',  '01700000003', crypt('Oper@12345',  gen_salt('bf', 10)), 'OPERATOR_ADMIN'::user_role, 'a0000000-0000-4000-8000-000000000001', NOW()),
  ('Demo Passenger',      'passenger@ticketkaton.demo', '01700000002', crypt('Demo@12345',  gen_salt('bf', 10)), 'CUSTOMER'::user_role,       NULL, NOW())
ON CONFLICT (phone_number) DO UPDATE
SET full_name     = EXCLUDED.full_name,
    email         = EXCLUDED.email,
    password_hash = EXCLUDED.password_hash,
    role          = EXCLUDED.role,
    operator_id   = EXCLUDED.operator_id,
    email_verified_at = COALESCE(users.email_verified_at, NOW()),
    updated_at    = now();
