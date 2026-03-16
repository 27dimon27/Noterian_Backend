CREATE EXTENSION IF NOT EXISTS pgcrypto;

INSERT INTO accounts (id, username, password, token_version, created_at, updated_at) 
VALUES (
    '11111111-1111-1111-1111-111111111111',
    'testuser',
    crypt('Password123', gen_salt('bf'))::bytea,
    1,
    NOW() - INTERVAL '1 month',
    NOW()
) ON CONFLICT (id) DO NOTHING;

INSERT INTO notes (id, user_id, title, parent_id, created_at, updated_at) VALUES
(
    gen_random_uuid(),
    '11111111-1111-1111-1111-111111111111',
    'The Forgotten City',
    NULL,
    NOW() - INTERVAL '1 month',
    NOW()
),
(
    gen_random_uuid(),
    '11111111-1111-1111-1111-111111111111',
    'Our goals',
    NULL,
    NOW() - INTERVAL '2 months',
    NOW()
),
(
    gen_random_uuid(),
    '11111111-1111-1111-1111-111111111111',
    'Eagle''s history',
    NULL,
    NOW() - INTERVAL '3 months',
    NOW()
),
(
    gen_random_uuid(),
    '11111111-1111-1111-1111-111111111111',
    'Metro and all about it',
    NULL,
    NOW() - INTERVAL '4 months',
    NOW()
),
(
    gen_random_uuid(),
    '11111111-1111-1111-1111-111111111111',
    'Shopping',
    NULL,
    NOW() - INTERVAL '5 months',
    NOW()
);

CREATE TEMP TABLE temp_note_ids AS
SELECT id, title, ROW_NUMBER() OVER (ORDER BY created_at) as rn
FROM notes 
WHERE user_id = '11111111-1111-1111-1111-111111111111'
ORDER BY created_at;

DO $$
DECLARE
    note_record RECORD;
    block_id UUID;
BEGIN
    SELECT id INTO note_record FROM temp_note_ids WHERE title = 'The Forgotten City';
    
    block_id := gen_random_uuid();
    INSERT INTO blocks (id, note_id, block_type_id, position, content, created_at, updated_at) VALUES
    (block_id, note_record.id, 1, 0, 
     'There is no game like Outer Wilds. That doesn''t stop fans from search for the elusive Wilds-like. One game that keeps popping up is The Forgotten City. Being very fond of flying into the sun and eating burned marshmallows, I was intrigued to try another knowledge based game.',
     NOW() - INTERVAL '1 month', NOW());
    
    INSERT INTO block_states (id, block_id, formatting, created_at, updated_at) VALUES
    (gen_random_uuid(), block_id, '{"format":"text","tags":["пример","тест","мок"]}'::jsonb, 
     NOW() - INTERVAL '1 month', NOW());
    
    block_id := gen_random_uuid();
    INSERT INTO blocks (id, note_id, block_type_id, position, content, created_at, updated_at) VALUES
    (block_id, note_record.id, 1, 1,
     'I''d like to start by stating my expectations when I opened TFC. Because I feel this review, and my experience with the game at large, are in big part a result of them. When the internets sold me on The Forgotten City, I was painted an image of a knowledge-based time loop mystery with lots of philosophy set in the roman empire. A period drama whodunnit Outer Wilds meets The Talos Principle?! Count me in! My experience was shot through this lens, with expectations you might''ve not had.',
     NOW() - INTERVAL '1 month', NOW());
    
    INSERT INTO block_states (id, block_id, formatting, created_at, updated_at) VALUES
    (gen_random_uuid(), block_id, '{"format":"text","tags":["пример","тест","мок"]}'::jsonb, 
     NOW() - INTERVAL '1 month', NOW());
    
    block_id := gen_random_uuid();
    INSERT INTO blocks (id, note_id, block_type_id, position, content, created_at, updated_at) VALUES
    (block_id, note_record.id, 1, 2,
     'I usually begin reviews by talking about presentation. Alas, the graphics are often wonky and the mechanics are stiff. The NPC models are aggressively Bethesda-esque, and that''s not a compliment. Part of game''s fame comes from it starting life as a Skyrim mod made by 3 people, and unfortunately it shows. I can''t deny it''s an impressive feat. Few could create such an experience, but as I paid more for this than for some of my all-time favourite stories, I can''t see such context as an excuse for rocks with poor clipping and countless opportunities to get soft-locked in invisible walls. I think part of my negative perception in this regard stems from TFC going for a "realistic" look, which beside being subjectively boring is hard to do well on a budget. I genuinely think I would''ve liked the exact same game measurably more if it was styled as well as indies tend to be.',
     NOW() - INTERVAL '1 month', NOW());
    
    INSERT INTO block_states (id, block_id, formatting, created_at, updated_at) VALUES
    (gen_random_uuid(), block_id, '{"format":"text","tags":["пример","тест","мок"]}'::jsonb, 
     NOW() - INTERVAL '1 month', NOW());

    SELECT id INTO note_record FROM temp_note_ids WHERE title = 'Our goals';
    
    block_id := gen_random_uuid();
    INSERT INTO blocks (id, note_id, block_type_id, position, content, created_at, updated_at) VALUES
    (block_id, note_record.id, 1, 0,
     'People talk a lot these days about the need to have goals. Many books have been written about achieving goals. The speakers give a lot of advice on this topic. Do you have a goal?',
     NOW() - INTERVAL '2 months', NOW());
    
    INSERT INTO block_states (id, block_id, formatting, created_at, updated_at) VALUES
    (gen_random_uuid(), block_id, '{"format":"text","tags":["пример","тест","мок"]}'::jsonb, 
     NOW() - INTERVAL '2 months', NOW());
    
    block_id := gen_random_uuid();
    INSERT INTO blocks (id, note_id, block_type_id, position, content, created_at, updated_at) VALUES
    (block_id, note_record.id, 1, 1,
     'There are special technologies for achieving a goal. Successful people, when they try to achieve their goals, met many obstacles. In order not to stop halfway, they developed their own techniques. Their experience can serve as a good example for others. Let''s look at the important conditions for achieving the goal.',
     NOW() - INTERVAL '2 months', NOW());
    
    INSERT INTO block_states (id, block_id, formatting, created_at, updated_at) VALUES
    (gen_random_uuid(), block_id, '{"format":"text","tags":["пример","тест","мок"]}'::jsonb, 
     NOW() - INTERVAL '2 months', NOW());
    
    block_id := gen_random_uuid();
    INSERT INTO blocks (id, note_id, block_type_id, position, content, created_at, updated_at) VALUES
    (block_id, note_record.id, 1, 2,
     'The goal should be written on paper. An unwritten goal is just a fantasy. When we write a goal on paper, we are sending a signal to our subconscious. From that moment on, the subconscious mind will be busy trying to find the best conditions for us to achieve a goal. This also keeps us motivated.',
     NOW() - INTERVAL '2 months', NOW());
    
    INSERT INTO block_states (id, block_id, formatting, created_at, updated_at) VALUES
    (gen_random_uuid(), block_id, '{"format":"text","tags":["пример","тест","мок"]}'::jsonb, 
     NOW() - INTERVAL '2 months', NOW());

    SELECT id INTO note_record FROM temp_note_ids WHERE title = 'Eagle''s history';
    
    block_id := gen_random_uuid();
    INSERT INTO blocks (id, note_id, block_type_id, position, content, created_at, updated_at) VALUES
    (block_id, note_record.id, 1, 0,
     'The United States of America is a big country in Northern America. Today we will tell you about one of its most famous symbols — a bald eagle. Soon after the USA got its independence from Great Britain, the government decided to use its image on the Great Seal. The picture of a bald eagle is often used as a symbol of courage, strength and power.',
     NOW() - INTERVAL '3 months', NOW());
    
    INSERT INTO block_states (id, block_id, formatting, created_at, updated_at) VALUES
    (gen_random_uuid(), block_id, '{"format":"text","tags":["пример","тест","мок"]}'::jsonb, 
     NOW() - INTERVAL '3 months', NOW());
    
    block_id := gen_random_uuid();
    INSERT INTO blocks (id, note_id, block_type_id, position, content, created_at, updated_at) VALUES
    (block_id, note_record.id, 1, 1,
     'This bird lives only on the territory of Northern America, you won''t find it anywhere else. The eagle is very large: it may grow almost 3 feet high, its wingspan up to 8 feet. To tell the truth, the eagle isn''t really bald. Its head is covered with white feathers, that is why it seems to be bald.',
     NOW() - INTERVAL '3 months', NOW());
    
    INSERT INTO block_states (id, block_id, formatting, created_at, updated_at) VALUES
    (gen_random_uuid(), block_id, '{"format":"text","tags":["пример","тест","мок"]}'::jsonb, 
     NOW() - INTERVAL '3 months', NOW());
    
    block_id := gen_random_uuid();
    INSERT INTO blocks (id, note_id, block_type_id, position, content, created_at, updated_at) VALUES
    (block_id, note_record.id, 1, 2,
     'The bird is a very committed partner in marriage. They choose marriage partners for life, and they take care for their babies together. Males and females look alike, but females are usually larger. Eagles build large nests, and usually they do it together. One of the biggest ones was recorded in the Guinness Book of Records, because it weighed almost 2 tons. These birds are one of a kind, because they can see even with their eyes closed. The thing is, in addition to usual eyelids they have special membranes on their eyes. Those membranes help them better preserve their eyes from the dust.',
     NOW() - INTERVAL '3 months', NOW());
    
    INSERT INTO block_states (id, block_id, formatting, created_at, updated_at) VALUES
    (gen_random_uuid(), block_id, '{"format":"text","tags":["пример","тест","мок"]}'::jsonb, 
     NOW() - INTERVAL '3 months', NOW());

    SELECT id INTO note_record FROM temp_note_ids WHERE title = 'Metro and all about it';
    
    block_id := gen_random_uuid();
    INSERT INTO blocks (id, note_id, block_type_id, position, content, created_at, updated_at) VALUES
    (block_id, note_record.id, 1, 0,
     'Around the world, a lot of big cities have a metro. This mean of transport can carry many people and there are no traffic jams with this vehicle. The first metro was put into service in London in 1863. During the World War II London Underground was a shelter for people.',
     NOW() - INTERVAL '4 months', NOW());
    
    INSERT INTO block_states (id, block_id, formatting, created_at, updated_at) VALUES
    (gen_random_uuid(), block_id, '{"format":"text","tags":["пример","тест","мок"]}'::jsonb, 
     NOW() - INTERVAL '4 months', NOW());
    
    block_id := gen_random_uuid();
    INSERT INTO blocks (id, note_id, block_type_id, position, content, created_at, updated_at) VALUES
    (block_id, note_record.id, 1, 1,
     'When the wartime had been over, the metro began to appear in many cities. The cities grew, and the metro gave a solution to connect the centre of the city and its suburbs. Moreover, people could buy a ticket which had an affordable price. The metro could conquer the world fast, because it has a lot of advantages. It reduces road traffic, and it means that pollution is decreasing too.',
     NOW() - INTERVAL '4 months', NOW());
    
    INSERT INTO block_states (id, block_id, formatting, created_at, updated_at) VALUES
    (gen_random_uuid(), block_id, '{"format":"text","tags":["пример","тест","мок"]}'::jsonb, 
     NOW() - INTERVAL '4 months', NOW());
    
    block_id := gen_random_uuid();
    INSERT INTO blocks (id, note_id, block_type_id, position, content, created_at, updated_at) VALUES
    (block_id, note_record.id, 1, 2,
     'Nowadays, the underground has more comfortable seats. There are escalators that can facilitate access to the platforms. Some stations have beautiful works of art. Their artists are not well-known as a rule, that is why these pictures have no great value. It helps to get rid of stealing. But the main thing is that they provide nice atmosphere. And it is very pleasant to be there. The Stockholm subway in Sweden is considered as the longest museum in the world.',
     NOW() - INTERVAL '4 months', NOW());
    
    INSERT INTO block_states (id, block_id, formatting, created_at, updated_at) VALUES
    (gen_random_uuid(), block_id, '{"format":"text","tags":["пример","тест","мок"]}'::jsonb, 
     NOW() - INTERVAL '4 months', NOW());

    SELECT id INTO note_record FROM temp_note_ids WHERE title = 'Shopping';
    
    block_id := gen_random_uuid();
    INSERT INTO blocks (id, note_id, block_type_id, position, content, created_at, updated_at) VALUES
    (block_id, note_record.id, 1, 0,
     'If we need to buy something, first of all we go to the shop. There are many different shops where you can buy whatever you want - from food to screws, bolts and nuts. It is not difficult to guess what type of store is the most popular. It may be said without exaggeration that these types of shops are supermarkets and grocery stores. A human being eats every day, so passing by such shops is a rather difficult thing.',
     NOW() - INTERVAL '5 months', NOW());
    
    INSERT INTO block_states (id, block_id, formatting, created_at, updated_at) VALUES
    (gen_random_uuid(), block_id, '{"format":"text","tags":["пример","тест","мок"]}'::jsonb, 
     NOW() - INTERVAL '5 months', NOW());
    
    block_id := gen_random_uuid();
    INSERT INTO blocks (id, note_id, block_type_id, position, content, created_at, updated_at) VALUES
    (block_id, note_record.id, 1, 1,
     'In every city you will find such shops as grocery stores, clothing stores, bakeries, butcheries. I love going to the flower shop most of all because flowers are my passion. Every week I go to an antique (curiosity) shop, because I really enjoy the original, ancient things. From time to time I visit the toy store in order to buy toys for my nephews and children. Almost every month I go to the gift shop so that I can buy gifts on birthday for my family and friends.',
     NOW() - INTERVAL '5 months', NOW());
    
    INSERT INTO block_states (id, block_id, formatting, created_at, updated_at) VALUES
    (gen_random_uuid(), block_id, '{"format":"text","tags":["пример","тест","мок"]}'::jsonb, 
     NOW() - INTERVAL '5 months', NOW());
    
    block_id := gen_random_uuid();
    INSERT INTO blocks (id, note_id, block_type_id, position, content, created_at, updated_at) VALUES
    (block_id, note_record.id, 1, 2,
     'I like to spend my time on shopping, preferably I like the self-service shops. You can scrutinize something as long as you like. A nagging seller does not hurry you, you are your own master. After it all, you can calmly go to the cashier, where all purchases will be counted and added up. In our time, it''s not only supermarkets that work in such a way, but also department stores, clothing shops and household goods shops.',
     NOW() - INTERVAL '5 months', NOW());
    
    INSERT INTO block_states (id, block_id, formatting, created_at, updated_at) VALUES
    (gen_random_uuid(), block_id, '{"format":"text","tags":["пример","тест","мок"]}'::jsonb, 
     NOW() - INTERVAL '5 months', NOW());
END $$;

DROP TABLE temp_note_ids;

SELECT setval('block_types_id_seq', (SELECT MAX(id) FROM block_types));