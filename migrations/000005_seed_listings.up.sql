-- Demo Survival
INSERT INTO listings (game_id, seller_id, item_code, quantity, price, status, buyer_id, created_at, closed_at)
VALUES ('11111111-1111-1111-1111-111111111111', 'vasya', 'iron_sword',    1,  100, 'active',    NULL,    now() - interval '2 hours',  NULL),
       ('11111111-1111-1111-1111-111111111111', 'vasya', 'wolf_pelt',    50,  200, 'active',    NULL,    now() - interval '90 minutes', NULL),
       ('11111111-1111-1111-1111-111111111111', 'petya', 'health_potion', 10,  45, 'active',    NULL,    now() - interval '1 hour',   NULL),
       ('11111111-1111-1111-1111-111111111111', 'petya', 'leather_armor', 1,  150, 'active',    NULL,    now() - interval '30 minutes', NULL),
       ('11111111-1111-1111-1111-111111111111', 'masha', 'iron_sword',    1,   90, 'active',    NULL,    now() - interval '10 minutes', NULL),
       ('11111111-1111-1111-1111-111111111111', 'masha', 'wood',        200,   60, 'sold',      'petya', now() - interval '1 day',    now() - interval '20 hours'),
       ('11111111-1111-1111-1111-111111111111', 'vasya', 'stone_axe',     1,   40, 'cancelled', NULL,    now() - interval '1 day',    now() - interval '22 hours');

-- Demo RPG
INSERT INTO listings (game_id, seller_id, item_code, quantity, price, status, buyer_id, created_at, closed_at)
VALUES ('22222222-2222-2222-2222-222222222222', 'arthas',  'fire_staff',     1, 1200, 'active',    NULL,      now() - interval '3 hours',  NULL),
       ('22222222-2222-2222-2222-222222222222', 'arthas',  'mana_potion',   20,  300, 'active',    NULL,      now() - interval '2 hours',  NULL),
       ('22222222-2222-2222-2222-222222222222', 'jaina',   'dragon_scale',   3, 2500, 'active',    NULL,      now() - interval '45 minutes', NULL),
       ('22222222-2222-2222-2222-222222222222', 'thrall',  'steel_helmet',   1,  400, 'active',    NULL,      now() - interval '15 minutes', NULL),
       ('22222222-2222-2222-2222-222222222222', 'jaina',   'mana_potion',   10,  140, 'sold',      'thrall',  now() - interval '2 days',   now() - interval '1 day'),
       ('22222222-2222-2222-2222-222222222222', 'thrall',  'fire_staff',     1, 1500, 'cancelled', NULL,      now() - interval '2 days',   now() - interval '36 hours');
