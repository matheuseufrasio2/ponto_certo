-- ============================================================================
-- Ponto Certo - Development & Testing Realistic Database Seed (90 Tenants)
-- ============================================================================

-- Clean existing data cleanly with cascade
TRUNCATE 
    appointments, 
    billings, 
    products, 
    barber_status_logs,
    barber_services,
    barber_schedule_blocks,
    barber_working_hours,
    barbers, 
    services, 
    unit_operating_hours, 
    barbershop_units, 
    users, 
    clients, 
    organizations 
CASCADE;

-- ============================================================================
-- 1. Organizations (90 Multi-Tenant Franchises & Barbershops)
-- ============================================================================
WITH tenant_data (idx, name, trade_name, plan, fee, sub_status, past_due_days, is_active) AS (
    VALUES
    (1, 'Rede Ponto Certo Franchising LTDA', 'Barbearias Ponto Certo', 'ENTERPRISE', 49900, 'ACTIVE', NULL::INT, TRUE),
    (2, 'Barbearia Don Corleone ME', 'Don Corleone Classic Barbershop', 'PRO', 24900, 'ACTIVE', NULL, TRUE),
    (3, 'Navalha de Ouro Barbearias LTDA', 'Navalha de Ouro Prime', 'ENTERPRISE', 59900, 'ACTIVE', NULL, TRUE),
    (4, 'Viking Beard & Hair Club ME', 'Viking Barbershop', 'PRO', 24900, 'ACTIVE', NULL, TRUE),
    (5, 'Confraria da Barba Franchising', 'Confraria da Barba VIP', 'ENTERPRISE', 49900, 'PAST_DUE', 14, TRUE),
    (6, 'Vintage Retro Barber Shop ME', 'Vintage Barber Retrô', 'BASIC', 19900, 'ACTIVE', NULL, TRUE),
    (7, 'O Mestre Barbeiro Lounge EIRELI', 'Mestre Barbeiro Lounge', 'PRO', 24900, 'ACTIVE', NULL, TRUE),
    (8, 'Cavalera Urban Barber LTDA', 'Cavalera Barbershop', 'PRO', 24900, 'PAST_DUE', 5, TRUE),
    (9, 'Alfa Barbers Club Franchising', 'Alfa Barber Club', 'ENTERPRISE', 49900, 'TRIAL', NULL, TRUE),
    (10, 'El Patron Barber & Beer ME', 'El Patron Barbershop', 'PRO', 24900, 'ACTIVE', NULL, TRUE),
    (11, 'Black Mustache Cortes & Barba', 'Black Mustache', 'BASIC', 14900, 'ACTIVE', NULL, TRUE),
    (12, 'Rei da Navalha Barba e Cabelo', 'Rei da Navalha', 'BASIC', 19900, 'ACTIVE', NULL, TRUE),
    (13, 'Corleone Cuts & Shaves SP', 'Corleone Cuts', 'PRO', 24900, 'ACTIVE', NULL, TRUE),
    (14, 'Barbearia do Zé Tradicional', 'Barbearia do Zé', 'BASIC', 14900, 'ACTIVE', NULL, TRUE),
    (15, 'The Gentlemen Club Barbershop', 'The Gentlemen''s Club', 'ENTERPRISE', 59900, 'ACTIVE', NULL, TRUE),
    (16, 'Barbearia Lenhador Roots ME', 'Lenhador Roots Barbershop', 'PRO', 24900, 'PAST_DUE', 22, TRUE),
    (17, 'La Mafia Barber Studio EIRELI', 'La Mafia Barber', 'PRO', 24900, 'ACTIVE', NULL, TRUE),
    (18, 'Imperium Barbershop VIP LTDA', 'Imperium Barbershop VIP', 'ENTERPRISE', 49900, 'ACTIVE', NULL, TRUE),
    (19, 'Rota 66 Barbearia & Pub ME', 'Rota 66 Barbershop', 'PRO', 24900, 'TRIAL', NULL, TRUE),
    (20, 'Spartacus Barber Academy LTDA', 'Spartacus Barbershop', 'BASIC', 19900, 'CANCELLED', NULL, FALSE),
    (21, 'Barbearia Lord & Crown LTDA', 'Lord & Crown Barbershop', 'ENTERPRISE', 49900, 'ACTIVE', NULL, TRUE),
    (22, 'Old School Barber Co. ME', 'Old School Barbers', 'PRO', 24900, 'ACTIVE', NULL, TRUE),
    (23, 'Barba Negra Cortes Finos ME', 'Barba Negra Barbershop', 'PRO', 24900, 'PAST_DUE', 8, TRUE),
    (24, 'Dom Pedro Barbearia Imperial', 'Dom Pedro Barber Club', 'ENTERPRISE', 59900, 'ACTIVE', NULL, TRUE),
    (25, 'Urban Cuts Barbearia Express', 'Urban Cuts Express', 'BASIC', 14900, 'ACTIVE', NULL, TRUE),
    (26, 'Barbearia Garagem dos Motoqueiros', 'Garagem Moto Barber', 'PRO', 24900, 'ACTIVE', NULL, TRUE),
    (27, 'Royal Shave Barbershop VIP', 'Royal Shave Studio', 'ENTERPRISE', 49900, 'ACTIVE', NULL, TRUE),
    (28, 'Barbearia Raiz Tradicional ME', 'Barbearia Raiz', 'BASIC', 19900, 'ACTIVE', NULL, TRUE),
    (29, 'Distrito Barber & Lounge LTDA', 'Distrito Barber', 'PRO', 24900, 'ACTIVE', NULL, TRUE),
    (30, 'Tattoo & Beard Club Studio ME', 'Tattoo & Beard Club', 'PRO', 24900, 'TRIAL', NULL, TRUE),
    (31, 'Barbearia 1920 Proibicionismo', 'Barbearia 1920', 'ENTERPRISE', 49900, 'ACTIVE', NULL, TRUE),
    (32, 'Mustache & Cigar Barbershop ME', 'Mustache & Cigar Lounge', 'PRO', 24900, 'PAST_DUE', 18, TRUE),
    (33, 'Barbearia do Bairro Matriz ME', 'Barbearia do Bairro', 'BASIC', 14900, 'ACTIVE', NULL, TRUE),
    (34, 'Kingston Reggae Barber Shop ME', 'Kingston Barbers', 'BASIC', 19900, 'ACTIVE', NULL, TRUE),
    (35, 'Diamond Barber Studio VIP LTDA', 'Diamond Barber Studio', 'ENTERPRISE', 59900, 'ACTIVE', NULL, TRUE),
    (36, 'Barbearia Apache Roots ME', 'Apache Barbershop', 'PRO', 24900, 'ACTIVE', NULL, TRUE),
    (37, 'Boutique da Barba & Bigode LTDA', 'Boutique da Barba', 'PRO', 24900, 'ACTIVE', NULL, TRUE),
    (38, 'Gentry Barber Lounge Curitiba', 'Gentry Barbershop', 'ENTERPRISE', 49900, 'ACTIVE', NULL, TRUE),
    (39, 'Barbearia da Estação Express ME', 'Barbearia da Estação', 'BASIC', 14900, 'ACTIVE', NULL, TRUE),
    (40, 'Sanctuary Barber & Male Spa LTDA', 'Sanctuary Male Spa', 'ENTERPRISE', 59900, 'TRIAL', NULL, TRUE),
    (41, 'Barbearia Brooklyn NY Style ME', 'Brooklyn NY Barber', 'PRO', 24900, 'ACTIVE', NULL, TRUE),
    (42, 'Texas Longhorn Barbershop LTDA', 'Texas Longhorn Barber', 'PRO', 24900, 'PAST_DUE', 3, TRUE),
    (43, 'Barbearia do Porto Marítima ME', 'Barbearia do Porto', 'BASIC', 19900, 'ACTIVE', NULL, TRUE),
    (44, 'Golden Blade Grooming Club LTDA', 'Golden Blade Club', 'ENTERPRISE', 49900, 'ACTIVE', NULL, TRUE),
    (45, 'Amnesia Barber Studio Night ME', 'Amnesia Barber Studio', 'BASIC', 19900, 'CANCELLED', NULL, FALSE),
    (46, 'Barbearia Al Capone & Whisky ME', 'Al Capone Barber & Whisky', 'PRO', 24900, 'ACTIVE', NULL, TRUE),
    (47, 'Clube da Navalha & Café LTDA', 'Clube da Navalha Café', 'BASIC', 19900, 'ACTIVE', NULL, TRUE),
    (48, 'Soberano Grooming Barbershop EIRELI', 'Soberano Barbershop', 'ENTERPRISE', 49900, 'ACTIVE', NULL, TRUE),
    (49, 'Barbearia O Padrinho Cuts ME', 'O Padrinho Cuts', 'PRO', 24900, 'PAST_DUE', 11, TRUE),
    (50, 'Barba Urbana Studio & Pub LTDA', 'Barba Urbana Studio', 'PRO', 24900, 'ACTIVE', NULL, TRUE),
    (51, 'Barbearia Machado de Ouro ME', 'Machado de Ouro Barber', 'BASIC', 14900, 'ACTIVE', NULL, TRUE),
    (52, 'Manhattan Classic Barbershop LTDA', 'Manhattan Classic Barber', 'ENTERPRISE', 59900, 'ACTIVE', NULL, TRUE),
    (53, 'Barbearia Maverick Rock & Beard ME', 'Maverick Rock Barbers', 'PRO', 24900, 'TRIAL', NULL, TRUE),
    (54, 'Barbearia Dom Casmurro Imperial', 'Dom Casmurro Barbershop', 'PRO', 24900, 'ACTIVE', NULL, TRUE),
    (55, 'Espaço Homem Moderno Barbearia LTDA', 'Espaço Homem Moderno', 'BASIC', 19900, 'ACTIVE', NULL, TRUE),
    (56, 'Vanguarda Barber Lounge ME', 'Vanguarda Barber Lounge', 'ENTERPRISE', 49900, 'ACTIVE', NULL, TRUE),
    (57, 'Barbearia Rota do Sol Litoral ME', 'Rota do Sol Barbearia', 'BASIC', 14900, 'ACTIVE', NULL, TRUE),
    (58, 'London Bridge Barbershop LTDA', 'London Bridge Barber', 'PRO', 24900, 'PAST_DUE', 9, TRUE),
    (59, 'Barbearia Tradição & Estilo ME', 'Tradição & Estilo Barber', 'PRO', 24900, 'ACTIVE', NULL, TRUE),
    (60, 'Fênix Cortes & Barba Prime EIRELI', 'Fênix Barber Prime', 'BASIC', 19900, 'CANCELLED', NULL, FALSE),
    (61, 'Barbearia Viking Valhalla Club LTDA', 'Valhalla Barber Club', 'ENTERPRISE', 59900, 'ACTIVE', NULL, TRUE),
    (62, 'Estação da Barba & Cerveja ME', 'Estação da Barba', 'PRO', 24900, 'ACTIVE', NULL, TRUE),
    (63, 'Barbearia Black Sheep Urban ME', 'Black Sheep Barbershop', 'PRO', 24900, 'TRIAL', NULL, TRUE),
    (64, 'Bárbaros Barber Studio Brasil LTDA', 'Bárbaros Barber Studio', 'ENTERPRISE', 49900, 'ACTIVE', NULL, TRUE),
    (65, 'Barbearia da Esquina Popular ME', 'Barbearia da Esquina', 'BASIC', 14900, 'ACTIVE', NULL, TRUE),
    (66, 'Gentlemen & Co. Premium Shave LTDA', 'Gentlemen & Co. Shave', 'ENTERPRISE', 59900, 'ACTIVE', NULL, TRUE),
    (67, 'Barbearia Tabacaria & Navalha ME', 'Tabacaria & Navalha Barber', 'PRO', 24900, 'PAST_DUE', 15, TRUE),
    (68, 'Ponto Alto Barbearia Executiva ME', 'Ponto Alto Barbershop', 'BASIC', 19900, 'ACTIVE', NULL, TRUE),
    (69, 'Barbearia Old Harley Davidson ME', 'Old Harley Barbershop', 'PRO', 24900, 'ACTIVE', NULL, TRUE),
    (70, 'Sublime Corte & Estética Masculina', 'Sublime Barbearia VIP', 'ENTERPRISE', 49900, 'ACTIVE', NULL, TRUE),
    (71, 'Barbearia Roots 77 Vintage ME', 'Roots 77 Vintage Barber', 'BASIC', 14900, 'ACTIVE', NULL, TRUE),
    (72, 'Il Maestro Barberia Italiana LTDA', 'Il Maestro Barbershop', 'PRO', 24900, 'ACTIVE', NULL, TRUE),
    (73, 'Barbearia Gran Reserva Lounge ME', 'Gran Reserva Barbershop', 'PRO', 24900, 'PAST_DUE', 6, TRUE),
    (74, 'Centurion Barber Club SP LTDA', 'Centurion Barber Club', 'ENTERPRISE', 59900, 'ACTIVE', NULL, TRUE),
    (75, 'Barbearia Universitária Express ME', 'Universitária Express Barber', 'BASIC', 14900, 'ACTIVE', NULL, TRUE),
    (76, 'Barba & Bigode do Interior ME', 'Barba & Bigode do Interior', 'BASIC', 19900, 'ACTIVE', NULL, TRUE),
    (77, 'The Boss Barber Executive Lounge', 'The Boss Barber Lounge', 'ENTERPRISE', 49900, 'TRIAL', NULL, TRUE),
    (78, 'Barbearia Vintage Blues & Jazz ME', 'Vintage Blues Barbershop', 'PRO', 24900, 'ACTIVE', NULL, TRUE),
    (79, 'Legado Barbeiro Confraria LTDA', 'Legado Barbeiro Confraria', 'PRO', 24900, 'ACTIVE', NULL, TRUE),
    (80, 'Barbearia Nova Era Futurista EIRELI', 'Nova Era Barber Studio', 'BASIC', 19900, 'CANCELLED', NULL, FALSE),
    (81, 'Barbearia Três Irmãos Tradicional', 'Três Irmãos Barbershop', 'BASIC', 14900, 'ACTIVE', NULL, TRUE),
    (82, 'The Barber King Deluxe Studio LTDA', 'The Barber King Deluxe', 'ENTERPRISE', 59900, 'ACTIVE', NULL, TRUE),
    (83, 'Barbearia Faroeste Caboclo ME', 'Faroeste Caboclo Barber', 'PRO', 24900, 'PAST_DUE', 25, TRUE),
    (84, 'Metrópole Hair & Shave Club LTDA', 'Metrópole Barbershop', 'ENTERPRISE', 49900, 'ACTIVE', NULL, TRUE),
    (85, 'Barbearia Alquimia dos Fios ME', 'Alquimia dos Fios Barber', 'BASIC', 19900, 'ACTIVE', NULL, TRUE),
    (86, 'Signature Cut Male Grooming ME', 'Signature Cut Barbershop', 'PRO', 24900, 'ACTIVE', NULL, TRUE),
    (87, 'Barbearia do Silva & Cabelo ME', 'Barbearia do Silva', 'BASIC', 14900, 'ACTIVE', NULL, TRUE),
    (88, 'Cavalheiros de Sangue Azul LTDA', 'Sangue Azul Barber Club', 'ENTERPRISE', 59900, 'ACTIVE', NULL, TRUE),
    (89, 'Barbearia Old Town Shave Shop ME', 'Old Town Shave Shop', 'PRO', 24900, 'PAST_DUE', 4, TRUE),
    (90, 'Zero Hora Night Barber Club ME', 'Zero Hora Night Barber', 'BASIC', 19900, 'CANCELLED', NULL, FALSE)
)
INSERT INTO organizations (
    id, name, trade_name, document, plan, monthly_fee_in_cents, subscription_status, past_due_since, is_active, created_at, updated_at
)
SELECT 
    ('a0000000-0000-0000-0000-' || LPAD(idx::TEXT, 12, '0'))::UUID,
    name,
    trade_name,
    LPAD(idx::TEXT, 2, '0') || '.' || LPAD((idx * 111)::TEXT, 3, '0') || '.' || LPAD((idx * 222)::TEXT, 3, '0') || '/0001-' || LPAD(idx::TEXT, 2, '0'),
    plan,
    fee,
    sub_status,
    CASE WHEN past_due_days IS NOT NULL THEN NOW() - (past_due_days * INTERVAL '1 day') ELSE NULL END,
    is_active,
    NOW() - ((91 - idx) * INTERVAL '3 days'),
    NOW()
FROM tenant_data;

-- ============================================================================
-- 2. Barbershop Units (Branches for all 90 Franchises)
-- Multi-unit for top franchises, single unit for boutique barbershops
-- ============================================================================
-- Main Branches (Unit 1 for all 90 tenants)
INSERT INTO barbershop_units (id, organization_id, name, phone, address, is_active, created_at, updated_at)
SELECT 
    ('b0000000-0000-0000-0000-' || LPAD(idx::TEXT, 12, '0'))::UUID,
    ('a0000000-0000-0000-0000-' || LPAD(idx::TEXT, 12, '0'))::UUID,
    o.trade_name || ' - Matriz Centro',
    '1199' || LPAD(idx::TEXT, 7, '0'),
    'Av. Principal, ' || (idx * 100)::TEXT || ' - Centro, São Paulo - SP',
    o.is_active,
    o.created_at,
    NOW()
FROM organizations o
JOIN (SELECT generate_series(1, 90) AS idx) s 
  ON o.id = ('a0000000-0000-0000-0000-' || LPAD(s.idx::TEXT, 12, '0'))::UUID;

-- Additional Secondary Branches for ENTERPRISE and top PRO franchises (30 additional units)
INSERT INTO barbershop_units (id, organization_id, name, phone, address, is_active, created_at, updated_at)
SELECT 
    ('b0000000-0000-0000-0000-' || LPAD((90 + idx)::TEXT, 12, '0'))::UUID,
    ('a0000000-0000-0000-0000-' || LPAD(idx::TEXT, 12, '0'))::UUID,
    o.trade_name || ' - Filial Shopping',
    '1198' || LPAD(idx::TEXT, 7, '0'),
    'Shopping Plaza Piso L2, Loja ' || (idx * 10)::TEXT || ', São Paulo - SP',
    o.is_active,
    o.created_at + INTERVAL '10 days',
    NOW()
FROM organizations o
JOIN (SELECT generate_series(1, 30) AS idx) s 
  ON o.id = ('a0000000-0000-0000-0000-' || LPAD(s.idx::TEXT, 12, '0'))::UUID;

-- ============================================================================
-- 3. Unit Operating Hours (Segunda a Sábado 09:00 - 20:00)
-- ============================================================================
INSERT INTO unit_operating_hours (unit_id, day_of_week, open_time, close_time, is_closed)
SELECT u.id, d.day, '09:00:00'::TIME, '20:00:00'::TIME, (d.day = 0)
FROM barbershop_units u
CROSS JOIN (SELECT generate_series(0, 6) AS day) d;

-- ============================================================================
-- 4. Users (Super Admin, 90 Franchise Owners, Managers, Receptionists & Barbers)
-- Bcrypt Hash for 'admin123': $2a$10$PsMR5of1qqLgzvViL.JF0ONjpxiVl.T4.dfRoGsPRL.j.8n2gq3t.
-- ============================================================================
-- 1. Global Super Admin
INSERT INTO users (id, organization_id, unit_id, name, email, password_hash, role, is_active, created_at, updated_at)
VALUES 
('c0000000-0000-0000-0000-000000000001', NULL, NULL, 'Carlos Super Admin', 'superadmin@pontocerto.com.br', '$2a$10$PsMR5of1qqLgzvViL.JF0ONjpxiVl.T4.dfRoGsPRL.j.8n2gq3t.', 'SUPER_ADMIN', TRUE, NOW(), NOW());

-- 2. 90 Franchise Owners (ORG_OWNER) - 100% ATIVOS em todas as organizações
INSERT INTO users (id, organization_id, unit_id, name, email, password_hash, role, is_active, created_at, updated_at)
SELECT 
    gen_random_uuid(),
    o.id,
    NULL,
    'Dono ' || o.trade_name,
    CASE 
        WHEN s.idx = 1 THEN 'owner@pontocerto.com.br'
        WHEN s.idx = 2 THEN 'owner.corleone@pontocerto.com.br'
        ELSE 'dono.tenant' || s.idx::TEXT || '@pontocerto.com.br'
    END,
    '$2a$10$PsMR5of1qqLgzvViL.JF0ONjpxiVl.T4.dfRoGsPRL.j.8n2gq3t.',
    'ORG_OWNER',
    TRUE,
    o.created_at,
    NOW()
FROM organizations o
JOIN (SELECT generate_series(1, 90) AS idx) s 
  ON o.id = ('a0000000-0000-0000-0000-' || LPAD(s.idx::TEXT, 12, '0'))::UUID;

-- 3. Managers for all units (UNIT_MANAGER) - 120 gerentes de filial
INSERT INTO users (id, organization_id, unit_id, name, email, password_hash, role, is_active, created_at, updated_at)
SELECT 
    gen_random_uuid(),
    u.organization_id,
    u.id,
    'Gerente ' || u.name,
    CASE 
        WHEN (ROW_NUMBER() OVER()) = 1 THEN 'manager.matriz@pontocerto.com.br'
        ELSE 'gerente.unidade' || (ROW_NUMBER() OVER())::TEXT || '@pontocerto.com.br'
    END,
    '$2a$10$PsMR5of1qqLgzvViL.JF0ONjpxiVl.T4.dfRoGsPRL.j.8n2gq3t.',
    'UNIT_MANAGER',
    TRUE,
    u.created_at,
    NOW()
FROM barbershop_units u;

-- 4. Receptionists for units (RECEPTIONIST) - 160 recepcionistas (10% demitidas/inativas)
WITH unit_ranked AS (
    SELECT id, organization_id, name, created_at, ROW_NUMBER() OVER(ORDER BY id) AS u_idx
    FROM barbershop_units
),
rec_slots AS (
    SELECT u.id AS unit_id, u.organization_id, u.name AS unit_name, u.created_at, u.u_idx, s.n AS rec_num
    FROM unit_ranked u
    CROSS JOIN (SELECT 1 AS n UNION ALL SELECT 2) s
    WHERE s.n = 1 OR u.u_idx <= 40
)
INSERT INTO users (id, organization_id, unit_id, name, email, password_hash, role, is_active, created_at, updated_at)
SELECT 
    gen_random_uuid(),
    rs.organization_id,
    rs.unit_id,
    CASE 
        WHEN (ROW_NUMBER() OVER()) % 10 = 0 THEN 'Ex-Recepcionista Demitida ' || (ROW_NUMBER() OVER())::TEXT
        ELSE 'Recepcionista ' || rs.unit_name || ' ' || rs.rec_num::TEXT
    END,
    CASE 
        WHEN (ROW_NUMBER() OVER()) = 1 THEN 'recepcao.matriz@pontocerto.com.br'
        ELSE 'recepcao.u' || rs.u_idx::TEXT || '.r' || rs.rec_num::TEXT || '@pontocerto.com.br'
    END,
    '$2a$10$PsMR5of1qqLgzvViL.JF0ONjpxiVl.T4.dfRoGsPRL.j.8n2gq3t.',
    'RECEPTIONIST',
    ((ROW_NUMBER() OVER()) % 10 != 0),
    rs.created_at,
    NOW()
FROM rec_slots rs;

-- 5. Operational Barbers (BARBER) - 600 barbeiros (5 por unidade, 10% demitidos/inativos)
WITH unit_ranked AS (
    SELECT id, organization_id, name, created_at, ROW_NUMBER() OVER(ORDER BY id) AS u_idx
    FROM barbershop_units
),
barber_slots AS (
    SELECT u.id AS unit_id, u.organization_id, u.name AS unit_name, u.created_at, u.u_idx, s.n AS barber_num
    FROM unit_ranked u
    CROSS JOIN (SELECT generate_series(1, 5) AS n) s
)
INSERT INTO users (id, organization_id, unit_id, name, email, password_hash, role, is_active, created_at, updated_at)
SELECT 
    gen_random_uuid(),
    bs.organization_id,
    bs.unit_id,
    CASE 
        WHEN (ROW_NUMBER() OVER()) % 10 = 0 THEN 'Ex-Barbeiro Demitido ' || (ROW_NUMBER() OVER())::TEXT
        ELSE 'Barbeiro ' || bs.unit_name || ' ' || bs.barber_num::TEXT
    END,
    CASE 
        WHEN (ROW_NUMBER() OVER()) = 1 THEN 'marcos.barbeiro@pontocerto.com.br'
        ELSE 'barbeiro.u' || bs.u_idx::TEXT || '.b' || bs.barber_num::TEXT || '@pontocerto.com.br'
    END,
    '$2a$10$PsMR5of1qqLgzvViL.JF0ONjpxiVl.T4.dfRoGsPRL.j.8n2gq3t.',
    'BARBER',
    ((ROW_NUMBER() OVER()) % 10 != 0),
    bs.created_at,
    NOW()
FROM barber_slots bs;

-- ============================================================================
-- 6. Barbers Profile
-- ============================================================================
INSERT INTO barbers (id, unit_id, user_id, name, phone, commission_percentage, status, is_active, created_at, updated_at)
SELECT 
    ('d0000000-0000-0000-0000-' || LPAD((ROW_NUMBER() OVER())::TEXT, 12, '0'))::UUID,
    usr.unit_id,
    usr.id,
    usr.name,
    '11988' || LPAD((ROW_NUMBER() OVER())::TEXT, 6, '0'),
    60,
    CASE WHEN usr.is_active THEN 'AVAILABLE' ELSE 'OFF_DUTY' END,
    usr.is_active,
    usr.created_at,
    NOW()
FROM users usr
WHERE usr.role = 'BARBER';

-- ============================================================================
-- 6. Services Catalog (Services for all units)
-- ============================================================================
INSERT INTO services (id, unit_id, name, duration_minutes, price_in_cents, category, is_active, created_at, updated_at)
SELECT 
    gen_random_uuid(),
    u.id,
    'Corte Degradê Tradicional',
    35,
    4500,
    'cabelo',
    TRUE,
    u.created_at,
    NOW()
FROM barbershop_units u;

INSERT INTO services (id, unit_id, name, duration_minutes, price_in_cents, category, is_active, created_at, updated_at)
SELECT 
    gen_random_uuid(),
    u.id,
    'Barba Terapia com Toalha Quente',
    30,
    4000,
    'barba',
    TRUE,
    u.created_at,
    NOW()
FROM barbershop_units u;

INSERT INTO services (id, unit_id, name, duration_minutes, price_in_cents, category, is_active, created_at, updated_at)
SELECT 
    gen_random_uuid(),
    u.id,
    'Combo Cabelo & Barba Executivo',
    60,
    7500,
    'combos',
    TRUE,
    u.created_at,
    NOW()
FROM barbershop_units u;

-- ============================================================================
-- 7. Clients
-- ============================================================================
INSERT INTO clients (id, unit_id, name, phone, email, created_at, updated_at)
SELECT 
    gen_random_uuid(),
    u.id,
    'Cliente VIP ' || u.name,
    '1197777' || LPAD((ROW_NUMBER() OVER())::TEXT, 4, '0'),
    'cliente.unidade' || (ROW_NUMBER() OVER())::TEXT || '@exemplo.com',
    u.created_at,
    NOW()
FROM barbershop_units u;

-- ============================================================================
-- 8. Realistic Appointments & Activity History
-- (Populates operational last_activity_at and counters for all active units)
-- ============================================================================
INSERT INTO appointments (
    id, unit_id, barber_id, client_id, service_id, start_time, end_time, status, payment_status, price_in_cents, notes, created_at, updated_at
)
SELECT 
    gen_random_uuid(),
    u.id,
    b.id,
    c.id,
    s.id,
    NOW() - (seq.n * INTERVAL '2 days'),
    NOW() - (seq.n * INTERVAL '2 days') + INTERVAL '45 minutes',
    'COMPLETED',
    'PAID',
    s.price_in_cents,
    'Atendimento concluído com excelência',
    NOW() - (seq.n * INTERVAL '2 days'),
    NOW() - (seq.n * INTERVAL '2 days')
FROM barbershop_units u
CROSS JOIN (SELECT generate_series(0, 3) AS n) seq
JOIN LATERAL (SELECT id FROM barbers WHERE unit_id = u.id AND is_active = TRUE LIMIT 1) b ON true
JOIN LATERAL (SELECT id FROM clients WHERE unit_id = u.id LIMIT 1) c ON true
JOIN LATERAL (SELECT id, price_in_cents FROM services WHERE unit_id = u.id LIMIT 1) s ON true
WHERE u.is_active = TRUE;

-- Add recent billings for completed appointments
INSERT INTO billings (
    id, unit_id, appointment_id, client_id, barber_id, service_id, amount_in_cents, payment_method, status, date, created_at, updated_at
)
SELECT 
    gen_random_uuid(),
    a.unit_id,
    a.id,
    a.client_id,
    a.barber_id,
    a.service_id,
    a.price_in_cents,
    'PIX',
    'PAID',
    a.end_time,
    a.created_at,
    a.updated_at
FROM appointments a;
