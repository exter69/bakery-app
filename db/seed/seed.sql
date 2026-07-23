-- ============================================================================
-- Dummy data for LOCAL DEV against Postgres.
-- Idempotent: safe to run multiple times (ON CONFLICT DO NOTHING on fixed IDs).
-- Applied automatically by scripts/db-setup.sh after migrations.
--
-- Passwords (bcrypt-hashed below):
--   customer_demo  -> demo-customer   (role 2)
--   baker_jean     -> demo-baker      (role 1, owns Bakery 1 & 3)
--   baker_marie    -> demo-baker      (role 1, owns Bakery 2)
--   admin_demo     -> demo-admin      (role 0)
--   pro_demo       -> demo-b2b        (role 3, B2B / Comptoir)
--
-- NOTE: prices are DECIMAL euros here (e.g. 1.50). The app converts to/from
-- integer cents internally.
-- ============================================================================

BEGIN;

-- ---------- Users ----------------------------------------------------------
INSERT INTO users (id, username, password_hash, role, created_at) VALUES
  ('10000000-0000-0000-0000-000000000001', 'customer_demo', '$2b$12$M32DJq913Zeo9qt8PJ5R0O0vkd7xncIqLgGbLgBq1bARjWGFlS8h.', 2, NOW()),
  ('10000000-0000-0000-0000-000000000002', 'baker_jean',    '$2b$12$7yaOUJIzP./skkMYA0U4Re8ozGQN5i2nbDofCGjZ086iJuEa2EDqO', 1, NOW()),
  ('10000000-0000-0000-0000-000000000003', 'baker_marie',   '$2b$12$7yaOUJIzP./skkMYA0U4Re8ozGQN5i2nbDofCGjZ086iJuEa2EDqO', 1, NOW()),
  ('10000000-0000-0000-0000-000000000004', 'admin_demo',    '$2b$12$C7ZsMdW4YlyX8qbFvakXh..MW2uURRFwRmnE6XgN19tb2xPKBaLFm', 0, NOW()),
  ('10000000-0000-0000-0000-000000000005', 'pro_demo',      '$2b$12$BiwYw05BKxJiQjhzluPfleYa5BvLOtQtCMLaq0BopJk2MLX5ggvj6', 3, NOW())
ON CONFLICT (id) DO NOTHING;

-- ---------- Bakeries -------------------------------------------------------
INSERT INTO bakeries (id, name, photo_url, description, address, latitude, longitude, owner_id, created_at) VALUES
  ('20000000-0000-0000-0000-000000000001', 'La Boulangerie du Coin', 'https://images.unsplash.com/photo-1509440159596-0249088772ff?w=800', 'Traditional French bakery on the corner.', 'Rue de la Paix 12, 1000 Bruxelles', 50.8466, 4.3528, '10000000-0000-0000-0000-000000000002', NOW()),
  ('20000000-0000-0000-0000-000000000002', 'Mie et Beurre',          'https://images.unsplash.com/photo-1517433670267-08bbd4be890f?w=800', 'Artisan sourdough and viennoiseries.',    'Avenue Louise 220, 1050 Ixelles',    50.8270, 4.3720, '10000000-0000-0000-0000-000000000003', NOW()),
  ('20000000-0000-0000-0000-000000000003', 'Le Fournil de Max',      'https://images.unsplash.com/photo-1600788886242-5c96aabe3757?w=800', 'Wood-fired breads from Brittany.',        'Chaussée de Wavre 45, 1040 Etterbeek', 50.8380, 4.3810, '10000000-0000-0000-0000-000000000002', NOW())
ON CONFLICT (id) DO NOTHING;

-- ---------- Products -------------------------------------------------------
-- Bakery 1
INSERT INTO products (id, bakery_id, name, description, price, photo_url, category, is_available, allergens, health_score, created_at) VALUES
  ('30000001-0000-0000-0000-000000000001', '20000000-0000-0000-0000-000000000001', 'Croissant',          'Buttery, flaky classic',                       1.50, 'https://images.unsplash.com/photo-1555507036-ab1f4038808a?w=400', 'Viennoiseries', true, '{gluten,lait}',       2, NOW()),
  ('30000001-0000-0000-0000-000000000002', '20000000-0000-0000-0000-000000000001', 'Pain au Chocolat',   'Two bars of dark chocolate inside pastry',     1.80, 'https://images.unsplash.com/photo-1623334044303-241021148842?w=400', 'Viennoiseries', true, '{gluten,lait}',       2, NOW()),
  ('30000001-0000-0000-0000-000000000003', '20000000-0000-0000-0000-000000000001', 'Baguette Tradition', 'Crusty sourdough baguette, 250g',              1.30, 'https://images.unsplash.com/photo-1608198093002-ad4e005484ec?w=400', 'Breads',        true, '{gluten}',            4, NOW()),
  ('30000001-0000-0000-0000-000000000004', '20000000-0000-0000-0000-000000000001', 'Pain de Campagne',   'Rustic country loaf with rye flour',           4.20, 'https://images.unsplash.com/photo-1589367920969-ab8e050bbb04?w=400', 'Breads',        true, '{gluten}',            5, NOW()),
  ('30000001-0000-0000-0000-000000000005', '20000000-0000-0000-0000-000000000001', 'Tarte aux Pommes',   'Apple tart with almond cream',                 4.50, 'https://images.unsplash.com/photo-1562007908-17c67e878c88?w=400', 'Pastries',      true, '{gluten,lait,oeuf}',  3, NOW()),
  ('30000001-0000-0000-0000-000000000006', '20000000-0000-0000-0000-000000000001', 'Éclair au Café',     'Coffee choux pastry with chocolate glaze',     3.80, 'https://images.unsplash.com/photo-1525059696034-4967a8e1dca2?w=400', 'Pastries',      true, '{gluten,lait,oeuf}',  2, NOW()),
-- Bakery 2
  ('30000002-0000-0000-0000-000000000001', '20000000-0000-0000-0000-000000000002', 'Sourdough Boule',    '48-hour fermented whole wheat sourdough',      5.50, 'https://images.unsplash.com/photo-1586444248902-2f64eddc13df?w=400', 'Breads',        true, '{gluten}',            5, NOW()),
  ('30000002-0000-0000-0000-000000000002', '20000000-0000-0000-0000-000000000002', 'Croissant au Beurre','All-butter croissant, 72-layer lamination',    2.00, 'https://images.unsplash.com/photo-1555507036-ab1f4038808a?w=400', 'Viennoiseries', true, '{gluten,lait}',       2, NOW()),
  ('30000002-0000-0000-0000-000000000003', '20000000-0000-0000-0000-000000000002', 'Brioche Feuilletée', 'Buttery layered brioche, lightly sweet',       3.50, 'https://images.unsplash.com/photo-1620921568790-8adcfb05e7a4?w=400', 'Viennoiseries', true, '{gluten,lait,oeuf}',  2, NOW()),
  ('30000002-0000-0000-0000-000000000004', '20000000-0000-0000-0000-000000000002', 'Fougasse aux Olives','Provençal flatbread with black olives',        3.80, 'https://images.unsplash.com/photo-1509440159596-0249088772ff?w=400', 'Breads',        true, '{gluten}',            4, NOW()),
  ('30000002-0000-0000-0000-000000000005', '20000000-0000-0000-0000-000000000002', 'Tarte au Citron',    'Lemon meringue tart with shortcrust base',     5.00, 'https://images.unsplash.com/photo-1519915028121-7d3463d20b13?w=400', 'Pastries',      true, '{gluten,lait,oeuf}',  3, NOW()),
-- Bakery 3
  ('30000003-0000-0000-0000-000000000001', '20000000-0000-0000-0000-000000000003', 'Pain au Levain',     'Wood-fired sourdough with thick crust',        4.80, 'https://images.unsplash.com/photo-1589367920969-ab8e050bbb04?w=400', 'Breads',        true, '{gluten}',            5, NOW()),
  ('30000003-0000-0000-0000-000000000002', '20000000-0000-0000-0000-000000000003', 'Kouign-Amann',       'Caramelized butter pastry from Brittany',      3.20, 'https://images.unsplash.com/photo-1509365390695-33aee754301f?w=400', 'Pastries',      true, '{gluten,lait}',       1, NOW()),
  ('30000003-0000-0000-0000-000000000003', '20000000-0000-0000-0000-000000000003', 'Pain aux Noix',      'Walnut bread with honey glaze',                5.20, 'https://images.unsplash.com/photo-1608198093002-ad4e005484ec?w=400', 'Breads',        true, '{gluten,fruits_a_coque}', 4, NOW()),
  ('30000003-0000-0000-0000-000000000004', '20000000-0000-0000-0000-000000000003', 'Chausson aux Pommes','Puff pastry turnover with apple compote',      2.80, 'https://images.unsplash.com/photo-1562007908-17c67e878c88?w=400', 'Viennoiseries', true, '{gluten,lait}',       3, NOW())
ON CONFLICT (id) DO NOTHING;

-- ---------- Sample orders (for the customer's history) ---------------------
INSERT INTO orders (id, bakery_id, user_id, scheduled_day, scheduled_start_time, scheduled_end_time, status, total_amount, payment_method, created_at, updated_at) VALUES
  ('40000000-0000-0000-0000-000000000001', '20000000-0000-0000-0000-000000000001', '10000000-0000-0000-0000-000000000001', 2, '09:00', '10:00', 'delivered', 5.80, 'online', NOW() - INTERVAL '5 days', NOW() - INTERVAL '5 days'),
  ('40000000-0000-0000-0000-000000000002', '20000000-0000-0000-0000-000000000002', '10000000-0000-0000-0000-000000000001', 4, '08:00', '09:00', 'confirmed', 9.50, 'online', NOW() - INTERVAL '1 day',  NOW() - INTERVAL '1 day')
ON CONFLICT (id) DO NOTHING;

INSERT INTO order_items (id, order_id, reservation_id, product_id, product_name, quantity, unit_price, subtotal) VALUES
  ('60000000-0000-0000-0000-000000000001', '40000000-0000-0000-0000-000000000001', NULL, '30000001-0000-0000-0000-000000000001', 'Croissant',           3, 1.50, 4.50),
  ('60000000-0000-0000-0000-000000000002', '40000000-0000-0000-0000-000000000001', NULL, '30000001-0000-0000-0000-000000000003', 'Baguette Tradition',  1, 1.30, 1.30),
  ('60000000-0000-0000-0000-000000000003', '40000000-0000-0000-0000-000000000002', NULL, '30000002-0000-0000-0000-000000000002', 'Croissant au Beurre', 2, 2.00, 4.00),
  ('60000000-0000-0000-0000-000000000004', '40000000-0000-0000-0000-000000000002', NULL, '30000002-0000-0000-0000-000000000001', 'Sourdough Boule',     1, 5.50, 5.50)
ON CONFLICT (id) DO NOTHING;

COMMIT;

-- Summary
SELECT 'users' AS table, COUNT(*) FROM users
UNION ALL SELECT 'bakeries', COUNT(*) FROM bakeries
UNION ALL SELECT 'products', COUNT(*) FROM products
UNION ALL SELECT 'orders', COUNT(*) FROM orders
UNION ALL SELECT 'order_items', COUNT(*) FROM order_items;
