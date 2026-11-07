-- ============================================================================
-- Ponto Certo - Development & Testing Realistic Database Seed
-- ============================================================================

-- 1. Organizations (Root Tenant)
INSERT INTO organizations (id, name, trade_name, document, is_active, created_at, updated_at)
VALUES 
(
    'a0000000-0000-0000-0000-000000000001',
    'Rede Ponto Certo Franchising LTDA',
    'Barbearias Ponto Certo',
    '12.345.678/0001-90',
    TRUE,
    NOW(),
    NOW()
),
(
    'a0000000-0000-0000-0000-000000000002',
    'Barbearia Don Corleone ME',
    'Don Corleone Classic Barbershop',
    '98.765.432/0001-10',
    TRUE,
    NOW(),
    NOW()
) ON CONFLICT (id) DO NOTHING;

-- 2. Barbershop Units (Branches)
INSERT INTO barbershop_units (id, organization_id, name, phone, address, is_active, created_at, updated_at)
VALUES 
(
    'b0000000-0000-0000-0000-000000000001',
    'a0000000-0000-0000-0000-000000000001',
    'Ponto Certo - Matriz Centro',
    '11999990001',
    'Av. Paulista, 1000 - Bela Vista, São Paulo - SP',
    TRUE,
    NOW(),
    NOW()
),
(
    'b0000000-0000-0000-0000-000000000002',
    'a0000000-0000-0000-0000-000000000001',
    'Ponto Certo - Shopping Sul',
    '11999990002',
    'Av. das Nações Unidas, 4777 - Shopping Sul Piso L2, São Paulo - SP',
    TRUE,
    NOW(),
    NOW()
),
(
    'b0000000-0000-0000-0000-000000000003',
    'a0000000-0000-0000-0000-000000000002',
    'Don Corleone - Jardins',
    '11999990003',
    'Rua Oscar Freire, 1200 - Jardins, São Paulo - SP',
    TRUE,
    NOW(),
    NOW()
) ON CONFLICT (id) DO NOTHING;

-- 3. Unit Operating Hours (Sunday = 0 to Saturday = 6: 10:00 - 22:00)
INSERT INTO unit_operating_hours (unit_id, day_of_week, open_time, close_time, is_closed)
SELECT u.id, d.day, '10:00:00'::TIME, '22:00:00'::TIME, FALSE
FROM barbershop_units u
CROSS JOIN (SELECT generate_series(0, 6) AS day) d
ON CONFLICT (unit_id, day_of_week) DO UPDATE 
SET open_time = EXCLUDED.open_time, close_time = EXCLUDED.close_time, is_closed = EXCLUDED.is_closed;

-- 4. Users (All RBAC Roles with bcrypt hash for 'admin123')
-- Bcrypt Hash: $2a$10$PsMR5of1qqLgzvViL.JF0ONjpxiVl.T4.dfRoGsPRL.j.8n2gq3t.
INSERT INTO users (id, organization_id, unit_id, name, email, password_hash, role, is_active, created_at, updated_at)
VALUES
(
    'c0000000-0000-0000-0000-000000000001',
    NULL,
    NULL,
    'Carlos Super Admin',
    'superadmin@pontocerto.com.br',
    '$2a$10$PsMR5of1qqLgzvViL.JF0ONjpxiVl.T4.dfRoGsPRL.j.8n2gq3t.',
    'SUPER_ADMIN',
    TRUE,
    NOW(),
    NOW()
),
(
    'c0000000-0000-0000-0000-000000000002',
    'a0000000-0000-0000-0000-000000000001',
    NULL,
    'Roberto Franchising Owner',
    'owner@pontocerto.com.br',
    '$2a$10$PsMR5of1qqLgzvViL.JF0ONjpxiVl.T4.dfRoGsPRL.j.8n2gq3t.',
    'ORG_OWNER',
    TRUE,
    NOW(),
    NOW()
),
(
    'c0000000-0000-0000-0000-000000000003',
    'a0000000-0000-0000-0000-000000000001',
    'b0000000-0000-0000-0000-000000000001',
    'Juliana Gerente Matriz',
    'manager.matriz@pontocerto.com.br',
    '$2a$10$PsMR5of1qqLgzvViL.JF0ONjpxiVl.T4.dfRoGsPRL.j.8n2gq3t.',
    'UNIT_MANAGER',
    TRUE,
    NOW(),
    NOW()
),
(
    'c0000000-0000-0000-0000-000000000004',
    'a0000000-0000-0000-0000-000000000001',
    'b0000000-0000-0000-0000-000000000001',
    'Lucas Recepcionista Matriz',
    'recepcao.matriz@pontocerto.com.br',
    '$2a$10$PsMR5of1qqLgzvViL.JF0ONjpxiVl.T4.dfRoGsPRL.j.8n2gq3t.',
    'RECEPTIONIST',
    TRUE,
    NOW(),
    NOW()
),
(
    'c0000000-0000-0000-0000-000000000005',
    'a0000000-0000-0000-0000-000000000001',
    'b0000000-0000-0000-0000-000000000001',
    'Marcos Mestre Barbeiro',
    'marcos.barbeiro@pontocerto.com.br',
    '$2a$10$PsMR5of1qqLgzvViL.JF0ONjpxiVl.T4.dfRoGsPRL.j.8n2gq3t.',
    'BARBER',
    TRUE,
    NOW(),
    NOW()
),
(
    'c0000000-0000-0000-0000-000000000006',
    'a0000000-0000-0000-0000-000000000002',
    NULL,
    'Vito Corleone Owner',
    'owner.corleone@pontocerto.com.br',
    '$2a$10$PsMR5of1qqLgzvViL.JF0ONjpxiVl.T4.dfRoGsPRL.j.8n2gq3t.',
    'ORG_OWNER',
    TRUE,
    NOW(),
    NOW()
),
(
    'c0000000-0000-0000-0000-000000000007',
    'a0000000-0000-0000-0000-000000000002',
    'b0000000-0000-0000-0000-000000000003',
    'Santino Gerente Jardins',
    'manager.jardins@pontocerto.com.br',
    '$2a$10$PsMR5of1qqLgzvViL.JF0ONjpxiVl.T4.dfRoGsPRL.j.8n2gq3t.',
    'UNIT_MANAGER',
    TRUE,
    NOW(),
    NOW()
),
(
    'c0000000-0000-0000-0000-000000000008',
    'a0000000-0000-0000-0000-000000000002',
    'b0000000-0000-0000-0000-000000000003',
    'Connie Recepcionista Jardins',
    'recepcao.jardins@pontocerto.com.br',
    '$2a$10$PsMR5of1qqLgzvViL.JF0ONjpxiVl.T4.dfRoGsPRL.j.8n2gq3t.',
    'RECEPTIONIST',
    TRUE,
    NOW(),
    NOW()
),
(
    'c0000000-0000-0000-0000-000000000009',
    'a0000000-0000-0000-0000-000000000002',
    'b0000000-0000-0000-0000-000000000003',
    'Enzo Barbeiro Jardins',
    'barbeiro.jardins@pontocerto.com.br',
    '$2a$10$PsMR5of1qqLgzvViL.JF0ONjpxiVl.T4.dfRoGsPRL.j.8n2gq3t.',
    'BARBER',
    TRUE,
    NOW(),
    NOW()
) ON CONFLICT (email) DO NOTHING;

-- 5. Barbers
INSERT INTO barbers (id, unit_id, user_id, name, phone, commission_percentage, is_active, created_at, updated_at)
VALUES
(
    'd0000000-0000-0000-0000-000000000001',
    'b0000000-0000-0000-0000-000000000001',
    'c0000000-0000-0000-0000-000000000005',
    'Marcos Mestre Barbeiro',
    '11988880001',
    60,
    TRUE,
    NOW(),
    NOW()
),
(
    'd0000000-0000-0000-0000-000000000002',
    'b0000000-0000-0000-0000-000000000001',
    NULL,
    'Thiago Especialista Barba',
    '11988880002',
    50,
    TRUE,
    NOW(),
    NOW()
),
(
    'd0000000-0000-0000-0000-000000000003',
    'b0000000-0000-0000-0000-000000000001',
    NULL,
    'Gabriel Barbeiro Junior',
    '11988880003',
    45,
    TRUE,
    NOW(),
    NOW()
),
(
    'd0000000-0000-0000-0000-000000000004',
    'b0000000-0000-0000-0000-000000000002',
    NULL,
    'Rafael Barbeiro Shopping Sul',
    '11988880004',
    50,
    TRUE,
    NOW(),
    NOW()
),
(
    'd0000000-0000-0000-0000-000000000005',
    'b0000000-0000-0000-0000-000000000003',
    'c0000000-0000-0000-0000-000000000009',
    'Enzo Barbeiro Jardins',
    '11988880005',
    55,
    TRUE,
    NOW(),
    NOW()
),
(
    'd0000000-0000-0000-0000-000000000006',
    'b0000000-0000-0000-0000-000000000003',
    NULL,
    'Tom Barbeiro Jardins',
    '11988880006',
    50,
    TRUE,
    NOW(),
    NOW()
) ON CONFLICT (id) DO NOTHING;

-- 6. Barber Working Hours (Weekly shift: Mon-Sat 10:00 - 19:00 with lunch 13:00 - 14:00, Sun off)
INSERT INTO barber_working_hours (barber_id, day_of_week, start_time, end_time, break_start_time, break_end_time, is_off)
SELECT b.id, d.day, 
    CASE WHEN d.day = 0 THEN '00:00:00'::TIME ELSE '10:00:00'::TIME END,
    CASE WHEN d.day = 0 THEN '00:00:00'::TIME ELSE '19:00:00'::TIME END,
    CASE WHEN d.day = 0 THEN NULL ELSE '13:00:00'::TIME END,
    CASE WHEN d.day = 0 THEN NULL ELSE '14:00:00'::TIME END,
    CASE WHEN d.day = 0 THEN TRUE ELSE FALSE END
FROM barbers b
CROSS JOIN (SELECT generate_series(0, 6) AS day) d
ON CONFLICT (barber_id, day_of_week) DO NOTHING;

-- 7. Barber Schedule Blocks (Temporary blackout: Doctor appointment for Thiago tomorrow)
INSERT INTO barber_schedule_blocks (barber_id, start_time, end_time, reason)
VALUES (
    'd0000000-0000-0000-0000-000000000002',
    (CURRENT_DATE + INTERVAL '1 day' + TIME '14:00:00')::TIMESTAMPTZ,
    (CURRENT_DATE + INTERVAL '1 day' + TIME '16:00:00')::TIMESTAMPTZ,
    'Consulta Médica de Rotina'
) ON CONFLICT DO NOTHING;

-- 8. Services Catalog (All 34 services from Downloads/servi_os_de_barbearia.md loaded for Unit 1)
INSERT INTO services (id, unit_id, name, category, description, duration_minutes, price_in_cents, is_active)
VALUES
-- CABELO
('e0000000-0000-0000-0000-000000000001', 'b0000000-0000-0000-0000-000000000001', 'Corte de Cabelo', 'CABELO', 'Corte com máquina e tesoura respeitando o formato do rosto.', 30, 6000, TRUE),
('e0000000-0000-0000-0000-000000000002', 'b0000000-0000-0000-0000-000000000001', 'Corte de Cabelo e Botox Capilar', 'CABELO', 'Corte com máquina/tesoura mais tratamento botox redutor de volume.', 60, 18500, TRUE),
('e0000000-0000-0000-0000-000000000003', 'b0000000-0000-0000-0000-000000000001', 'Corte de Cabelo e Cabelo Platinado', 'CABELO', 'Corte completo com descoloração global platinada profissional.', 120, 30500, TRUE),
('e0000000-0000-0000-0000-000000000004', 'b0000000-0000-0000-0000-000000000001', 'Corte de Cabelo e Luzes', 'CABELO', 'Corte e aplicação de técnica de luzes para iluminação dos fios.', 90, 24000, TRUE),
('e0000000-0000-0000-0000-000000000005', 'b0000000-0000-0000-0000-000000000001', 'Corte de Cabelo e Penteado Blindado', 'CABELO', 'Corte e penteado resistente com mechas e fixação extrema.', 45, 8000, TRUE),
('e0000000-0000-0000-0000-000000000006', 'b0000000-0000-0000-0000-000000000001', 'Corte de Cabelo e Relaxamento Capilar', 'CABELO', 'Corte e relaxamento para cabelos cacheados e crespos.', 60, 17000, TRUE),
('e0000000-0000-0000-0000-000000000007', 'b0000000-0000-0000-0000-000000000001', 'Corte de Cabelo e Selagem', 'CABELO', 'Corte com selagem térmica anti-frizz e fortalecimento.', 60, 19500, TRUE),
('e0000000-0000-0000-0000-000000000008', 'b0000000-0000-0000-0000-000000000001', 'Corte de Cabelo e Sobrancelha', 'CABELO', 'Corte e design de sobrancelha com navalha e tesoura.', 40, 8500, TRUE),
('e0000000-0000-0000-0000-000000000009', 'b0000000-0000-0000-0000-000000000001', 'Corte de Cabelo Nudred', 'CABELO', 'Corte e estilização de cachos com esponja nudred específica.', 35, 6000, TRUE),
('e0000000-0000-0000-0000-000000000010', 'b0000000-0000-0000-0000-000000000001', 'Corte Pai e Filho', 'CABELO', 'Dois cortes completos simultâneos ou consecutivos.', 60, 11000, TRUE),
('e0000000-0000-0000-0000-000000000011', 'b0000000-0000-0000-0000-000000000001', 'Lavagem e Penteado', 'CABELO', 'Lavagem com shampoo/condicionador e modelagem dos fios.', 25, 4000, TRUE),
('e0000000-0000-0000-0000-000000000012', 'b0000000-0000-0000-0000-000000000001', 'Penteado', 'CABELO', 'Penteado clássico ou contemporâneo com pomada/gel.', 20, 3500, TRUE),
('e0000000-0000-0000-0000-000000000013', 'b0000000-0000-0000-0000-000000000001', 'Pezinho', 'CABELO', 'Acabamento e alinhamento do contorno da nuca.', 15, 2500, TRUE),

-- BARBA
('e0000000-0000-0000-0000-000000000014', 'b0000000-0000-0000-0000-000000000001', 'Aplicação de Tintura na Barba', 'BARBA', 'Cobertura de fios brancos com tintura específica e hidratação.', 30, 6000, TRUE),
('e0000000-0000-0000-0000-000000000015', 'b0000000-0000-0000-0000-000000000001', 'Barba', 'BARBA', 'Barba com toalha quente, navalha, balm hidratante e pós-barba.', 30, 6000, TRUE),
('e0000000-0000-0000-0000-000000000016', 'b0000000-0000-0000-0000-000000000001', 'Barba e Higienização de Barba', 'BARBA', 'Limpeza profunda anti-frizz e corte completo de barba.', 40, 8000, TRUE),
('e0000000-0000-0000-0000-000000000017', 'b0000000-0000-0000-0000-000000000001', 'Barba e Pezinho', 'BARBA', 'Alinhamento completo da barba e contorno do pescoço/nuca.', 35, 8000, TRUE),
('e0000000-0000-0000-0000-000000000018', 'b0000000-0000-0000-0000-000000000001', 'Barba e Sobrancelha', 'BARBA', 'Barba com toalha quente e design completo de sobrancelhas.', 40, 8500, TRUE),
('e0000000-0000-0000-0000-000000000019', 'b0000000-0000-0000-0000-000000000001', 'Barba e Tintura', 'BARBA', 'Corte preciso de barba e aplicação de tintura tonalizante.', 45, 11500, TRUE),
('e0000000-0000-0000-0000-000000000020', 'b0000000-0000-0000-0000-000000000001', 'Barboterapia com Vapor e Óleos Essenciais', 'BARBA', 'Spa facial para barba com vapor de ozônio, esfoliação e óleos.', 45, 9000, TRUE),
('e0000000-0000-0000-0000-000000000021', 'b0000000-0000-0000-0000-000000000001', 'Higienização de Barba', 'BARBA', 'Limpeza profunda dos fios faciais com xampu e esfoliante.', 20, 3000, TRUE),
('e0000000-0000-0000-0000-000000000022', 'b0000000-0000-0000-0000-000000000001', 'Pigmentação de Barba', 'BARBA', 'Correção de falhas com pigmentação temporária.', 20, 2500, TRUE),

-- CABELO E BARBA
('e0000000-0000-0000-0000-000000000023', 'b0000000-0000-0000-0000-000000000001', 'Corte de Cabelo e Barba', 'CABELO_E_BARBA', 'Combo tradicional de corte completo e barba com toalha quente.', 60, 11500, TRUE),
('e0000000-0000-0000-0000-000000000024', 'b0000000-0000-0000-0000-000000000001', 'Corte de Cabelo e Barboterapia com Vapor e Óleos Essenciais', 'CABELO_E_BARBA', 'Corte estilizado e barboterapia completa com vapor de ozônio.', 60, 15000, TRUE),
('e0000000-0000-0000-0000-000000000025', 'b0000000-0000-0000-0000-000000000001', 'Corte de Cabelo, Barba e Botox Capilar', 'CABELO_E_BARBA', 'Corte, barba e alinhamento com botox capilar.', 90, 24000, TRUE),
('e0000000-0000-0000-0000-000000000026', 'b0000000-0000-0000-0000-000000000001', 'Corte de Cabelo, Barba e Cabelo Platinado', 'CABELO_E_BARBA', 'Transformação total com corte, barba e platinado global.', 150, 36500, TRUE),
('e0000000-0000-0000-0000-000000000027', 'b0000000-0000-0000-0000-000000000001', 'Corte de Cabelo, Barba e Depilação Nariz e Orelha na Cera', 'CABELO_E_BARBA', 'Corte, barba e cera quente higiênica em nariz e orelha.', 60, 17000, TRUE),
('e0000000-0000-0000-0000-000000000028', 'b0000000-0000-0000-0000-000000000001', 'Corte de Cabelo, Barba e Luzes', 'CABELO_E_BARBA', 'Corte, barba completa e luzes iluminadas nos fios.', 120, 30000, TRUE),
('e0000000-0000-0000-0000-000000000029', 'b0000000-0000-0000-0000-000000000001', 'Corte de Cabelo, Barba e Relaxamento Capilar', 'CABELO_E_BARBA', 'Corte, barba e relaxamento químico nos cabelos.', 90, 22500, TRUE),
('e0000000-0000-0000-0000-000000000030', 'b0000000-0000-0000-0000-000000000001', 'Corte de Cabelo, Barba e Selagem', 'CABELO_E_BARBA', 'Corte, barba e selagem térmica reconstrutora.', 90, 25000, TRUE),
('e0000000-0000-0000-0000-000000000031', 'b0000000-0000-0000-0000-000000000001', 'Corte de Cabelo, Barba e Sobrancelha', 'CABELO_E_BARBA', 'Trio clássico: cabelo, barba e sobrancelha com navalha.', 60, 13500, TRUE),
('e0000000-0000-0000-0000-000000000032', 'b0000000-0000-0000-0000-000000000001', 'Corte de Cabelo, Barba, Depilação com Cera Nariz e Orelha e Sobrancelha', 'CABELO_E_BARBA', 'Cuidados integrais masculinos em um único atendimento.', 75, 19000, TRUE),
('e0000000-0000-0000-0000-000000000033', 'b0000000-0000-0000-0000-000000000001', 'Corte de Cabelo, Barboterapia com Vapor e Óleos Essenciais e Sobrancelha', 'CABELO_E_BARBA', 'Corte, barboterapia relaxante e design de sobrancelha.', 75, 17000, TRUE),

-- COMBOS
('e0000000-0000-0000-0000-000000000034', 'b0000000-0000-0000-0000-000000000001', 'Combo Experiência Trezentos', 'COMBOS', 'Corte, barboterapia, higienização, sobrancelha, cera e hidratação.', 120, 27900, TRUE),
('e0000000-0000-0000-0000-000000000035', 'b0000000-0000-0000-0000-000000000001', 'Combo Leonidas', 'COMBOS', 'Corte de cabelo, sobrancelha e botox capilar.', 75, 19300, TRUE),
('e0000000-0000-0000-0000-000000000036', 'b0000000-0000-0000-0000-000000000001', 'Combo Persa', 'COMBOS', 'Corte na máquina pente único, barboterapia e sobrancelha.', 60, 14500, TRUE),
('e0000000-0000-0000-0000-000000000037', 'b0000000-0000-0000-0000-000000000001', 'Combo Sparta', 'COMBOS', 'Corte, barba, sobrancelha, hidratação e depilação com cera.', 90, 19500, TRUE),

-- TRATAMENTOS CAPILARES
('e0000000-0000-0000-0000-000000000038', 'b0000000-0000-0000-0000-000000000001', 'Corte de Cabelo e Hidratação', 'TRATAMENTOS_CAPILARES', 'Corte com reposição de nutrientes e água para os fios.', 50, 10000, TRUE),
('e0000000-0000-0000-0000-000000000039', 'b0000000-0000-0000-0000-000000000001', 'Hidratação', 'TRATAMENTOS_CAPILARES', 'Hidratação profunda reconstrutora com máscara capilar.', 30, 5000, TRUE),

-- CUIDADOS ESPECIAIS
('e0000000-0000-0000-0000-000000000040', 'b0000000-0000-0000-0000-000000000001', 'Depilação com Máquina Orelha e Nariz', 'CUIDADOS_ESPECIAIS', 'Remoção prática e rápida de pelos superficiais.', 15, 2500, TRUE),
('e0000000-0000-0000-0000-000000000041', 'b0000000-0000-0000-0000-000000000001', 'Depilação Nariz na Cera', 'CUIDADOS_ESPECIAIS', 'Depilação nasal interna higiênica e indolor com cera morna.', 15, 3500, TRUE),
('e0000000-0000-0000-0000-000000000042', 'b0000000-0000-0000-0000-000000000001', 'Depilação Nariz e Orelha na Cera', 'CUIDADOS_ESPECIAIS', 'Cera quente para nariz e orelha com limpeza completa.', 20, 6000, TRUE),
('e0000000-0000-0000-0000-000000000043', 'b0000000-0000-0000-0000-000000000001', 'Depilação Orelha na Cera', 'CUIDADOS_ESPECIAIS', 'Depilação auricular de orelhas com cera suave.', 15, 3500, TRUE),
('e0000000-0000-0000-0000-000000000044', 'b0000000-0000-0000-0000-000000000001', 'Lavagem com Shampoo Don Alcides', 'CUIDADOS_ESPECIAIS', 'Lavagem especial com produtos premium para homens.', 15, 2500, TRUE),

-- QUÍMICAS CAPILARES
('e0000000-0000-0000-0000-000000000045', 'b0000000-0000-0000-0000-000000000001', 'Botox Capilar', 'QUIMICAS_CAPILARES', 'Redução de volume e controle do frizz na fibra capilar.', 60, 13000, TRUE),
('e0000000-0000-0000-0000-000000000046', 'b0000000-0000-0000-0000-000000000001', 'Cabelo Platinado', 'QUIMICAS_CAPILARES', 'Descoloração global técnica com proteção antiqueda.', 120, 24500, TRUE),
('e0000000-0000-0000-0000-000000000047', 'b0000000-0000-0000-0000-000000000001', 'Coloração Tonalização', 'QUIMICAS_CAPILARES', 'Coloração profissional para realce de cor ou camuflagem.', 60, 8500, TRUE),
('e0000000-0000-0000-0000-000000000048', 'b0000000-0000-0000-0000-000000000001', 'Luzes', 'QUIMICAS_CAPILARES', 'Mechas e reflexos loiros ou platinados com touca/papel.', 90, 18000, TRUE),
('e0000000-0000-0000-0000-000000000049', 'b0000000-0000-0000-0000-000000000001', 'Relaxamento Capilar', 'QUIMICAS_CAPILARES', 'Abertura de cachos e controle de volume com hidróxidos.', 60, 11500, TRUE),
('e0000000-0000-0000-0000-000000000050', 'b0000000-0000-0000-0000-000000000001', 'Selagem', 'QUIMICAS_CAPILARES', 'Tratamento térmico de alinhamento com selagem de cutículas.', 60, 14000, TRUE),

-- VISAGISMO
('e0000000-0000-0000-0000-000000000051', 'b0000000-0000-0000-0000-000000000001', 'Corte Análise Visagista', 'VISAGISMO', 'Consultoria de imagem pessoal combinada com corte harmônico.', 60, 13000, TRUE),

-- ESTÉTICA FACIAL
('e0000000-0000-0000-0000-000000000052', 'b0000000-0000-0000-0000-000000000001', 'Limpeza Facial Black', 'ESTETICA_FACIAL', 'Limpeza profunda com esfoliação e máscara negra de carvão.', 30, 4000, TRUE),
('e0000000-0000-0000-0000-000000000053', 'b0000000-0000-0000-0000-000000000001', 'Sobrancelha', 'ESTETICA_FACIAL', 'Design e limpeza com navalha e tesoura.', 20, 2500, TRUE),

-- SERVIÇOS UNIDADE 3 (DON CORLEONE - JARDINS - ORG 2)
('e0000000-0000-0000-0000-000000000101', 'b0000000-0000-0000-0000-000000000003', 'Corte Don Corleone Clássico', 'CABELO', 'Corte clássico estilo italiano com tesoura e acabamento impecável.', 35, 7500, TRUE),
('e0000000-0000-0000-0000-000000000102', 'b0000000-0000-0000-0000-000000000003', 'Barba Tradicional Italiana', 'BARBA', 'Barba desenhada com toalha quente, navalha e loção siciliana.', 30, 6500, TRUE),
('e0000000-0000-0000-0000-000000000103', 'b0000000-0000-0000-0000-000000000003', 'Combo Il Padrino (Cabelo e Barba)', 'CABELO_E_BARBA', 'Experiência completa italiana: corte clássico e barba com toalha quente.', 60, 13000, TRUE),
('e0000000-0000-0000-0000-000000000104', 'b0000000-0000-0000-0000-000000000003', 'Tratamento Capilar Don', 'TRATAMENTOS_CAPILARES', 'Hidratação profunda reconstrutora com máscara capilar italiana.', 40, 9000, TRUE)
ON CONFLICT (id) DO NOTHING;

-- 9. Barber Services Capabilities (N:N Skill Matrix)
-- Marcos (Master) is qualified for ALL services
INSERT INTO barber_services (barber_id, service_id)
SELECT 'd0000000-0000-0000-0000-000000000001', s.id
FROM services s
WHERE s.unit_id = 'b0000000-0000-0000-0000-000000000001'
ON CONFLICT (barber_id, service_id) DO NOTHING;

-- Thiago (Barba & Combos Specialist)
INSERT INTO barber_services (barber_id, service_id)
SELECT 'd0000000-0000-0000-0000-000000000002', s.id
FROM services s
WHERE s.unit_id = 'b0000000-0000-0000-0000-000000000001'
  AND s.category IN ('BARBA', 'CABELO_E_BARBA', 'COMBOS', 'CUIDADOS_ESPECIAIS')
ON CONFLICT (barber_id, service_id) DO NOTHING;

-- Gabriel (Junior: Hair, Treatments & Facial)
INSERT INTO barber_services (barber_id, service_id)
SELECT 'd0000000-0000-0000-0000-000000000003', s.id
FROM services s
WHERE s.unit_id = 'b0000000-0000-0000-0000-000000000001'
  AND s.category IN ('CABELO', 'TRATAMENTOS_CAPILARES', 'ESTETICA_FACIAL')
ON CONFLICT (barber_id, service_id) DO NOTHING;

-- Enzo Barbeiro Jardins (Master Unit 3) is qualified for ALL Unit 3 services
INSERT INTO barber_services (barber_id, service_id)
SELECT 'd0000000-0000-0000-0000-000000000005', s.id
FROM services s
WHERE s.unit_id = 'b0000000-0000-0000-0000-000000000003'
ON CONFLICT (barber_id, service_id) DO NOTHING;

-- Tom Barbeiro Jardins (Unit 3) is qualified for Cabelo, Barba e Combo
INSERT INTO barber_services (barber_id, service_id)
SELECT 'd0000000-0000-0000-0000-000000000006', s.id
FROM services s
WHERE s.unit_id = 'b0000000-0000-0000-0000-000000000003'
  AND s.category IN ('CABELO', 'BARBA', 'CABELO_E_BARBA')
ON CONFLICT (barber_id, service_id) DO NOTHING;

-- 10. Products Catalog (All 20 retail products loaded for Unit 1)
INSERT INTO products (id, unit_id, name, category, description, price_in_cents, cost_in_cents, stock_quantity, is_active)
VALUES
-- KITS
('f0000000-0000-0000-0000-000000000001', 'b0000000-0000-0000-0000-000000000001', 'Kit Bio Extratus Homem', 'KITS', '5 itens essenciais com ingredientes naturais para limpar e modelar.', 16000, 8000, 8, TRUE),
('f0000000-0000-0000-0000-000000000002', 'b0000000-0000-0000-0000-000000000001', 'Kit Condado Oceano', 'KITS', '6 produtos de hidratação masculina com necessaire.', 14000, 7000, 5, TRUE),
('f0000000-0000-0000-0000-000000000003', 'b0000000-0000-0000-0000-000000000001', 'Kit Tratamento Barbershop', 'KITS', 'Focado em peles sensíveis, cabelos loiros e grisalhos.', 11000, 5500, 6, TRUE),
('f0000000-0000-0000-0000-000000000004', 'b0000000-0000-0000-0000-000000000001', 'Kit All Black Treeliss', 'KITS', 'Shampoo anticaspa, gel de barbear, balm e loção amadeirada.', 13000, 6500, 10, TRUE),
('f0000000-0000-0000-0000-000000000005', 'b0000000-0000-0000-0000-000000000001', 'Kit Men Barber Plancton', 'KITS', 'Shampoo 3 em 1, óleo e shaving gel refrescante.', 11500, 6000, 7, TRUE),

-- SHAMPOO E HIGIENE DIÁRIA
('f0000000-0000-0000-0000-000000000006', 'b0000000-0000-0000-0000-000000000001', 'Granado Shampoo Cabelo Barba e Bigode', 'SHAMPOO_HIGIENE', 'Opção vegana cruelty-free com limpeza profunda sem ressecamento.', 4500, 2200, 15, TRUE),
('f0000000-0000-0000-0000-000000000007', 'b0000000-0000-0000-0000-000000000001', 'QOD Barber Shop Shampoo Whiskey', 'SHAMPOO_HIGIENE', 'Aroma amadeirado exclusivo para fios secos e danificados.', 5500, 2800, 12, TRUE),
('f0000000-0000-0000-0000-000000000008', 'b0000000-0000-0000-0000-000000000001', 'Shampoo Shark Barber Profissional 2L', 'SHAMPOO_HIGIENE', 'Embalagem de alto rendimento para uso na bancada e revenda.', 7500, 4000, 4, TRUE),

-- FINALIZADORES & CUIDADOS
('f0000000-0000-0000-0000-000000000009', 'b0000000-0000-0000-0000-000000000001', 'Pomada Modeladora Efeito Matte 100g', 'FINALIZADORES', 'Fixação forte com acabamento seco sem resíduos.', 5000, 2000, 25, TRUE),
('f0000000-0000-0000-0000-000000000010', 'b0000000-0000-0000-0000-000000000001', 'Óleo Hidratante para Barba 30ml', 'FINALIZADORES', 'Hidratação profunda para barbas longas e controle de frizz.', 5500, 2500, 18, TRUE),
('f0000000-0000-0000-0000-000000000011', 'b0000000-0000-0000-0000-000000000001', 'Balm Refrescante para Barba 120g', 'FINALIZADORES', 'Redução de volume e sensação de frescor na pele.', 6000, 2800, 14, TRUE),
('f0000000-0000-0000-0000-000000000012', 'b0000000-0000-0000-0000-000000000001', 'Shaving Gel de Barbear Transparente 500g', 'BARBEARIA_CUIDADOS', 'Gel transparente que facilita o alinhamento da navalha.', 4000, 1800, 20, TRUE),
('f0000000-0000-0000-0000-000000000013', 'b0000000-0000-0000-0000-000000000001', 'Loção Pós-Barba Efervescente 150ml', 'BARBEARIA_CUIDADOS', 'Fecha os poros e previne foliculite e irritação.', 5000, 2200, 16, TRUE),
('f0000000-0000-0000-0000-000000000014', 'b0000000-0000-0000-0000-000000000001', 'Cera Modeladora de Bigode 30g', 'FINALIZADORES', 'Fixação clássica para modelar bigodes e pontas.', 3500, 1500, 10, TRUE),
('f0000000-0000-0000-0000-000000000015', 'b0000000-0000-0000-0000-000000000001', 'Tônico Fator de Crescimento Barba 60ml', 'BARBEARIA_CUIDADOS', 'Blend de vitaminas que estimula preenchimento de falhas.', 9500, 4500, 9, TRUE),
('f0000000-0000-0000-0000-000000000016', 'b0000000-0000-0000-0000-000000000001', 'Pó Texturizador Hair Dust Volume 10g', 'FINALIZADORES', 'Volume instantâneo e controle de oleosidade na raiz.', 6500, 3000, 11, TRUE),

-- ACESSÓRIOS & MERCH
('f0000000-0000-0000-0000-000000000017', 'b0000000-0000-0000-0000-000000000001', 'Pente de Madeira Anti-Estática', 'ACESSORIOS', 'Pente artesanal de madeira para barba e cabelo.', 3000, 1200, 22, TRUE),
('f0000000-0000-0000-0000-000000000018', 'b0000000-0000-0000-0000-000000000001', 'Necessaire Masculina Couro Sintético', 'ACESSORIOS', 'Resistente à água e com compartimentos para viagens.', 7500, 3500, 8, TRUE),
('f0000000-0000-0000-0000-000000000019', 'b0000000-0000-0000-0000-000000000001', 'Sabonete Facial Esfoliante Skincare 100g', 'SKINCARE', 'Esfoliação suave anti-oleosidade e anti-cravos.', 6500, 2800, 12, TRUE),
('f0000000-0000-0000-0000-000000000020', 'b0000000-0000-0000-0000-000000000001', 'Cerveja Artesanal Ponto Certo IPA 500ml', 'BEBIDAS_MERCHANDISING', 'Cerveja artesanal com lúpulo aromático para espera e consumo.', 1800, 800, 30, TRUE),

-- PRODUTOS UNIDADE 3 (DON CORLEONE - JARDINS - ORG 2)
('f0000000-0000-0000-0000-000000000101', 'b0000000-0000-0000-0000-000000000003', 'Pomada Clássica Don Corleone 100g', 'FINALIZADORES', 'Pomada italiana base d água fixação forte brilho natural.', 6500, 3000, 15, TRUE),
('f0000000-0000-0000-0000-000000000102', 'b0000000-0000-0000-0000-000000000003', 'Óleo Hidratante de Barba Siciliano 30ml', 'FINALIZADORES', 'Óleo aromático com extrato de bergamota e sândalo.', 6000, 2500, 12, TRUE)
ON CONFLICT (id) DO NOTHING;

-- 11. Clients
INSERT INTO clients (id, unit_id, name, phone, email, created_at, updated_at)
VALUES
(
    '00000000-0000-0000-0000-000000000001',
    'b0000000-0000-0000-0000-000000000001',
    'Felipe Costa',
    '11977770001',
    'felipe.costa@cliente.com',
    NOW(),
    NOW()
),
(
    '00000000-0000-0000-0000-000000000002',
    'b0000000-0000-0000-0000-000000000001',
    'Bruno Alves',
    '11977770002',
    'bruno.alves@cliente.com',
    NOW(),
    NOW()
),
(
    '00000000-0000-0000-0000-000000000003',
    'b0000000-0000-0000-0000-000000000001',
    'Rodrigo Martins',
    '11977770003',
    'rodrigo.martins@cliente.com',
    NOW(),
    NOW()
),
(
    '00000000-0000-0000-0000-000000000004',
    'b0000000-0000-0000-0000-000000000001',
    'André Lima',
    '11977770004',
    'andre.lima@cliente.com',
    NOW(),
    NOW()
),
-- CLIENTES UNIDADE 3 (DON CORLEONE - JARDINS - ORG 2)
(
    '00000000-0000-0000-0000-000000000005',
    'b0000000-0000-0000-0000-000000000003',
    'Giovanni Rossi',
    '11977770005',
    'giovanni.rossi@cliente.com',
    NOW(),
    NOW()
),
(
    '00000000-0000-0000-0000-000000000006',
    'b0000000-0000-0000-0000-000000000003',
    'Matteo Bianchi',
    '11977770006',
    'matteo.bianchi@cliente.com',
    NOW(),
    NOW()
) ON CONFLICT (id) DO NOTHING;

-- 12. Appointments (Realistic Appointments across past, today, and future)
INSERT INTO appointments (id, unit_id, barber_id, client_id, service_id, start_time, end_time, status, payment_status, price_in_cents, notes, created_at, updated_at)
VALUES
-- Appointment 1: Completed and Paid today (Corte + Barba com Marcos)
(
    '11111111-0000-0000-0000-000000000001',
    'b0000000-0000-0000-0000-000000000001',
    'd0000000-0000-0000-0000-000000000001',
    '00000000-0000-0000-0000-000000000001',
    'e0000000-0000-0000-0000-000000000023', -- Corte de Cabelo e Barba
    (CURRENT_DATE + TIME '10:00:00')::TIMESTAMPTZ,
    (CURRENT_DATE + TIME '11:00:00')::TIMESTAMPTZ,
    'COMPLETED',
    'PAID',
    11500,
    'Cliente preferiu degradê navalhado e barba alinhada.',
    NOW(),
    NOW()
),
-- Appointment 2: Scheduled today (Corte com Marcos)
(
    '11111111-0000-0000-0000-000000000002',
    'b0000000-0000-0000-0000-000000000001',
    'd0000000-0000-0000-0000-000000000001',
    '00000000-0000-0000-0000-000000000002',
    'e0000000-0000-0000-0000-000000000001', -- Corte de Cabelo
    (CURRENT_DATE + TIME '11:30:00')::TIMESTAMPTZ,
    (CURRENT_DATE + TIME '12:00:00')::TIMESTAMPTZ,
    'SCHEDULED',
    'PENDING',
    6000,
    'Primeira visita à unidade.',
    NOW(),
    NOW()
),
-- Appointment 3: Scheduled tomorrow (Barboterapia com Thiago)
(
    '11111111-0000-0000-0000-000000000003',
    'b0000000-0000-0000-0000-000000000001',
    'd0000000-0000-0000-0000-000000000002',
    '00000000-0000-0000-0000-000000000003',
    'e0000000-0000-0000-0000-000000000020', -- Barboterapia
    (CURRENT_DATE + INTERVAL '1 day' + TIME '10:30:00')::TIMESTAMPTZ,
    (CURRENT_DATE + INTERVAL '1 day' + TIME '11:15:00')::TIMESTAMPTZ,
    'SCHEDULED',
    'PENDING',
    9000,
    'Solicitou essência de eucalipto.',
    NOW(),
    NOW()
),
-- Appointment 4: No-Show yesterday
(
    '11111111-0000-0000-0000-000000000004',
    'b0000000-0000-0000-0000-000000000001',
    'd0000000-0000-0000-0000-000000000003',
    '00000000-0000-0000-0000-000000000004',
    'e0000000-0000-0000-0000-000000000001', -- Corte de Cabelo
    (CURRENT_DATE - INTERVAL '1 day' + TIME '15:00:00')::TIMESTAMPTZ,
    (CURRENT_DATE - INTERVAL '1 day' + TIME '15:30:00')::TIMESTAMPTZ,
    'NO_SHOW',
    'PENDING',
    6000,
    'Cliente não compareceu e não atendeu contato telefônico.',
    NOW(),
    NOW()
),
-- Appointment 5: Cancelled yesterday
(
    '11111111-0000-0000-0000-000000000005',
    'b0000000-0000-0000-0000-000000000001',
    'd0000000-0000-0000-0000-000000000002',
    '00000000-0000-0000-0000-000000000002',
    'e0000000-0000-0000-0000-000000000015', -- Barba
    (CURRENT_DATE - INTERVAL '1 day' + TIME '16:00:00')::TIMESTAMPTZ,
    (CURRENT_DATE - INTERVAL '1 day' + TIME '16:30:00')::TIMESTAMPTZ,
    'CANCELLED',
    'PENDING',
    6000,
    'Cancelado pelo cliente com 4 horas de antecedência.',
    NOW(),
    NOW()
),
-- AGENDAMENTOS UNIDADE 3 (DON CORLEONE - JARDINS - ORG 2)
-- Unit 3 Appointment 1: Completed and Paid today (Il Padrino com Enzo)
(
    '11111111-0000-0000-0000-000000000006',
    'b0000000-0000-0000-0000-000000000003',
    'd0000000-0000-0000-0000-000000000005',
    '00000000-0000-0000-0000-000000000005',
    'e0000000-0000-0000-0000-000000000103', -- Combo Il Padrino
    (CURRENT_DATE + TIME '10:00:00')::TIMESTAMPTZ,
    (CURRENT_DATE + TIME '11:00:00')::TIMESTAMPTZ,
    'COMPLETED',
    'PAID',
    13000,
    'Cliente VIP Jardins.',
    NOW(),
    NOW()
),
-- Unit 3 Appointment 2: Scheduled today (Corte Clássico com Enzo)
(
    '11111111-0000-0000-0000-000000000007',
    'b0000000-0000-0000-0000-000000000003',
    'd0000000-0000-0000-0000-000000000005',
    '00000000-0000-0000-0000-000000000006',
    'e0000000-0000-0000-0000-000000000101', -- Corte Clássico
    (CURRENT_DATE + TIME '11:30:00')::TIMESTAMPTZ,
    (CURRENT_DATE + TIME '12:05:00')::TIMESTAMPTZ,
    'SCHEDULED',
    'PENDING',
    7500,
    'Agendado pela recepção dos Jardins.',
    NOW(),
    NOW()
) ON CONFLICT (id) DO NOTHING;

-- 13. Billings (Financial transactions: Appointment checkout + Product sale)
INSERT INTO billings (id, unit_id, date, barber_id, client_id, service_id, product_id, appointment_id, quantity, amount_in_cents, payment_method, status, notes, created_at, updated_at)
VALUES
-- Billing 1: Checkout do Agendamento 1 (Corte + Barba)
(
    '22222222-0000-0000-0000-000000000001',
    'b0000000-0000-0000-0000-000000000001',
    CURRENT_DATE,
    'd0000000-0000-0000-0000-000000000001',
    '00000000-0000-0000-0000-000000000001',
    'e0000000-0000-0000-0000-000000000023',
    NULL,
    '11111111-0000-0000-0000-000000000001',
    1,
    11500,
    'PIX',
    'PAID',
    'Checkout de agendamento finalizado via PIX no balcão.',
    NOW(),
    NOW()
),
-- Billing 2: Venda Avulsa de Produto (Pomada Matte)
(
    '22222222-0000-0000-0000-000000000002',
    'b0000000-0000-0000-0000-000000000001',
    CURRENT_DATE,
    'd0000000-0000-0000-0000-000000000001',
    '00000000-0000-0000-0000-000000000001',
    NULL,
    'f0000000-0000-0000-0000-000000000009',
    NULL,
    1,
    5000,
    'CREDIT_CARD',
    'PAID',
    'Venda de produto após atendimento.',
    NOW(),
    NOW()
),
-- Billing 3: Checkout Unidade 3 (Il Padrino com Enzo)
(
    '22222222-0000-0000-0000-000000000003',
    'b0000000-0000-0000-0000-000000000003',
    CURRENT_DATE,
    'd0000000-0000-0000-0000-000000000005',
    '00000000-0000-0000-0000-000000000005',
    'e0000000-0000-0000-0000-000000000103',
    NULL,
    '11111111-0000-0000-0000-000000000006',
    1,
    13000,
    'PIX',
    'PAID',
    'Checkout de agendamento finalizado na unidade Jardins.',
    NOW(),
    NOW()
) ON CONFLICT (id) DO NOTHING;
