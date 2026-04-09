CREATE EXTENSION IF NOT EXISTS pgcrypto;

INSERT INTO profiles (id, username, password, token_version, created_at, updated_at) VALUES
(
    '11111111-1111-1111-1111-111111111111',
    'testuser',
    crypt('Password123', gen_salt('bf'))::bytea,
    1,
    NOW() - INTERVAL '1 month',
    NOW()
),
(
    '22222222-2222-2222-2222-222222222222',
    'testuser2',
    crypt('Password123', gen_salt('bf'))::bytea,
    1,
    NOW() - INTERVAL '1 month',
    NOW()
) ON CONFLICT (id) DO NOTHING;

INSERT INTO notes (id, user_id, title, parent_id, created_at, updated_at) VALUES
(
    'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa',
    '11111111-1111-1111-1111-111111111111',
    'The Forgotten City',
    NULL,
    NOW() - INTERVAL '1 month',
    NOW()
),
(
    'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb',
    '11111111-1111-1111-1111-111111111111',
    'Our goals',
    NULL,
    NOW() - INTERVAL '2 months',
    NOW()
),
(
    'cccccccc-cccc-cccc-cccc-cccccccccccc',
    '11111111-1111-1111-1111-111111111111',
    'Eagle''s history',
    NULL,
    NOW() - INTERVAL '3 months',
    NOW()
),
(
    'dddddddd-dddd-dddd-dddd-dddddddddddd',
    '11111111-1111-1111-1111-111111111111',
    'Metro and all about it',
    NULL,
    NOW() - INTERVAL '4 months',
    NOW()
),
(
    'eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee',
    '11111111-1111-1111-1111-111111111111',
    'Shopping',
    NULL,
    NOW() - INTERVAL '5 months',
    NOW()
);

DO $$
DECLARE
    note_id UUID := 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa';
    block_id UUID;
    block_content TEXT;
BEGIN
    block_id := 'f1111111-1111-1111-1111-111111111111';
    block_content := 'There is no game like Outer Wilds. That doesn''t stop fans from search for the elusive Wilds-like. One game that keeps popping up is The Forgotten City. Being very fond of flying into the sun and eating burned marshmallows, I was intrigued to try another knowledge based game.';
    INSERT INTO blocks (id, note_id, block_type_id, position, content, created_at, updated_at) VALUES
    (block_id, note_id, 1, 0, block_content, NOW() - INTERVAL '1 month', NOW());
    
    block_id := 'f2222222-2222-2222-2222-222222222222';
    block_content := 'I''d like to start by stating my expectations when I opened TFC. Because I feel this review, and my experience with the game at large, are in big part a result of them. When the internets sold me on The Forgotten City, I was painted an image of a knowledge-based time loop mystery with lots of philosophy set in the roman empire. A period drama whodunnit Outer Wilds meets The Talos Principle?! Count me in! My experience was shot through this lens, with expectations you might''ve not had.';
    INSERT INTO blocks (id, note_id, block_type_id, position, content, created_at, updated_at) VALUES
    (block_id, note_id, 1, 1, block_content, NOW() - INTERVAL '1 month', NOW());
    
    INSERT INTO block_formatting (block_id, start_pos, end_pos, bold, italic, underline, text_align) VALUES
    (block_id, 0, 50, TRUE, FALSE, FALSE, 0);
    
    block_id := 'f3333333-3333-3333-3333-333333333333';
    block_content := 'I usually begin reviews by talking about presentation. Alas, the graphics are often wonky and the mechanics are stiff. The NPC models are aggressively Bethesda-esque, and that''s not a compliment. Part of game''s fame comes from it starting life as a Skyrim mod made by 3 people, and unfortunately it shows. I can''t deny it''s an impressive feat. Few could create such an experience, but as I paid more for this than for some of my all-time favourite stories, I can''t see such context as an excuse for rocks with poor clipping and countless opportunities to get soft-locked in invisible walls. I think part of my negative perception in this regard stems from TFC going for a "realistic" look, which beside being subjectively boring is hard to do well on a budget. I genuinely think I would''ve liked the exact same game measurably more if it was styled as well as indies tend to be.';
    INSERT INTO blocks (id, note_id, block_type_id, position, content, created_at, updated_at) VALUES
    (block_id, note_id, 1, 2, block_content, NOW() - INTERVAL '1 month', NOW());
    
    INSERT INTO block_formatting (block_id, start_pos, end_pos, bold, italic, underline, text_align) VALUES
    (block_id, 200, 350, FALSE, TRUE, FALSE, 0);
END $$;

DO $$
DECLARE
    note_id UUID := 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb';
    block_id UUID;
    block_content TEXT;
BEGIN
    block_id := 'a1111111-1111-1111-1111-111111111111';
    block_content := 'People talk a lot these days about the need to have goals. Many books have been written about achieving goals. The speakers give a lot of advice on this topic. Do you have a goal?';
    INSERT INTO blocks (id, note_id, block_type_id, position, content, created_at, updated_at) VALUES
    (block_id, note_id, 1, 0, block_content, NOW() - INTERVAL '2 months', NOW());
    
    INSERT INTO block_formatting (block_id, start_pos, end_pos, bold, italic, underline, text_align) VALUES
    (block_id, 30, 80, FALSE, FALSE, TRUE, 0);
    
    block_id := 'a2222222-2222-2222-2222-222222222222';
    block_content := 'There are special technologies for achieving a goal. Successful people, when they try to achieve their goals, met many obstacles. In order not to stop halfway, they developed their own techniques. Their experience can serve as a good example for others. Let''s look at the important conditions for achieving the goal.';
    INSERT INTO blocks (id, note_id, block_type_id, position, content, created_at, updated_at) VALUES
    (block_id, note_id, 1, 1, block_content, NOW() - INTERVAL '2 months', NOW());
    
    INSERT INTO block_formatting (block_id, start_pos, end_pos, bold, italic, underline, text_align) VALUES
    (block_id, 0, 100, TRUE, TRUE, FALSE, 1);
    
    block_id := 'a3333333-3333-3333-3333-333333333333';
    block_content := 'The goal should be written on paper. An unwritten goal is just a fantasy. When we write a goal on paper, we are sending a signal to our subconscious. From that moment on, the subconscious mind will be busy trying to find the best conditions for us to achieve a goal. This also keeps us motivated.';
    INSERT INTO blocks (id, note_id, block_type_id, position, content, created_at, updated_at) VALUES
    (block_id, note_id, 1, 2, block_content, NOW() - INTERVAL '2 months', NOW());
    
    INSERT INTO block_formatting (block_id, start_pos, end_pos, bold, italic, underline, text_align) VALUES
    (block_id, 0, LENGTH(block_content), FALSE, FALSE, FALSE, 2);
END $$;

DO $$
DECLARE
    note_id UUID := 'cccccccc-cccc-cccc-cccc-cccccccccccc';
    block_id UUID;
    block_content TEXT;
BEGIN
    block_id := 'b1111111-1111-1111-1111-111111111111';
    block_content := 'The United States of America is a big country in Northern America. Today we will tell you about one of its most famous symbols — a bald eagle. Soon after the USA got its independence from Great Britain, the government decided to use its image on the Great Seal. The picture of a bald eagle is often used as a symbol of courage, strength and power.';
    INSERT INTO blocks (id, note_id, block_type_id, position, content, created_at, updated_at) VALUES
    (block_id, note_id, 1, 0, block_content, NOW() - INTERVAL '3 months', NOW());
    
    block_id := 'b2222222-2222-2222-2222-222222222222';
    block_content := 'This bird lives only on the territory of Northern America, you won''t find it anywhere else. The eagle is very large: it may grow almost 3 feet high, its wingspan up to 8 feet. To tell the truth, the eagle isn''t really bald. Its head is covered with white feathers, that is why it seems to be bald.';
    INSERT INTO blocks (id, note_id, block_type_id, position, content, created_at, updated_at) VALUES
    (block_id, note_id, 1, 1, block_content, NOW() - INTERVAL '3 months', NOW());
    
    INSERT INTO block_formatting (block_id, start_pos, end_pos, bold, italic, underline, text_align) VALUES
    (block_id, 50, 120, TRUE, FALSE, TRUE, 0);
    
    block_id := 'b3333333-3333-3333-3333-333333333333';
    block_content := 'The bird is a very committed partner in marriage. They choose marriage partners for life, and they take care for their babies together. Males and females look alike, but females are usually larger. Eagles build large nests, and usually they do it together. One of the biggest ones was recorded in the Guinness Book of Records, because it weighed almost 2 tons. These birds are one of a kind, because they can see even with their eyes closed. The thing is, in addition to usual eyelids they have special membranes on their eyes. Those membranes help them better preserve their eyes from the dust.';
    INSERT INTO blocks (id, note_id, block_type_id, position, content, created_at, updated_at) VALUES
    (block_id, note_id, 1, 2, block_content, NOW() - INTERVAL '3 months', NOW());
    
    INSERT INTO block_formatting (block_id, start_pos, end_pos, bold, italic, underline, text_align) VALUES
    (block_id, 150, 250, FALSE, TRUE, TRUE, 0);
END $$;

DO $$
DECLARE
    note_id UUID := 'dddddddd-dddd-dddd-dddd-dddddddddddd';
    block_id UUID;
    block_content TEXT;
BEGIN
    block_id := 'c1111111-1111-1111-1111-111111111111';
    block_content := 'Around the world, a lot of big cities have a metro. This mean of transport can carry many people and there are no traffic jams with this vehicle. The first metro was put into service in London in 1863. During the World War II London Underground was a shelter for people.';
    INSERT INTO blocks (id, note_id, block_type_id, position, content, created_at, updated_at) VALUES
    (block_id, note_id, 1, 0, block_content, NOW() - INTERVAL '4 months', NOW());
    
    block_id := 'c2222222-2222-2222-2222-222222222222';
    block_content := 'When the wartime had been over, the metro began to appear in many cities. The cities grew, and the metro gave a solution to connect the centre of the city and its suburbs. Moreover, people could buy a ticket which had an affordable price. The metro could conquer the world fast, because it has a lot of advantages. It reduces road traffic, and it means that pollution is decreasing too.';
    INSERT INTO blocks (id, note_id, block_type_id, position, content, created_at, updated_at) VALUES
    (block_id, note_id, 1, 1, block_content, NOW() - INTERVAL '4 months', NOW());
    
    INSERT INTO block_formatting (block_id, start_pos, end_pos, bold, italic, underline, text_align) VALUES
    (block_id, 20, 80, TRUE, FALSE, FALSE, 2);
    
    block_id := 'c3333333-3333-3333-3333-333333333333';
    block_content := 'Nowadays, the underground has more comfortable seats. There are escalators that can facilitate access to the platforms. Some stations have beautiful works of art. Their artists are not well-known as a rule, that is why these pictures have no great value. It helps to get rid of stealing. But the main thing is that they provide nice atmosphere. And it is very pleasant to be there. The Stockholm subway in Sweden is considered as the longest museum in the world.';
    INSERT INTO blocks (id, note_id, block_type_id, position, content, created_at, updated_at) VALUES
    (block_id, note_id, 1, 2, block_content, NOW() - INTERVAL '4 months', NOW());
    
    INSERT INTO block_formatting (block_id, start_pos, end_pos, bold, italic, underline, text_align) VALUES
    (block_id, 0, LENGTH(block_content), FALSE, TRUE, TRUE, 0);
END $$;

DO $$
DECLARE
    note_id UUID := 'eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee';
    block_id UUID;
    block_content TEXT;
    block_content2 TEXT;
BEGIN
    block_id := 'd1111111-1111-1111-1111-111111111111';
    block_content := 'If we need to buy something, first of all we go to the shop. There are many different shops where you can buy whatever you want - from food to screws, bolts and nuts. It is not difficult to guess what type of store is the most popular. It may be said without exaggeration that these types of shops are supermarkets and grocery stores. A human being eats every day, so passing by such shops is a rather difficult thing.';
    INSERT INTO blocks (id, note_id, block_type_id, position, content, created_at, updated_at) VALUES
    (block_id, note_id, 1, 0, block_content, NOW() - INTERVAL '5 months', NOW());
    
    block_id := 'd2222222-2222-2222-2222-222222222222';
    block_content2 := 'In every city you will find such shops as grocery stores, clothing stores, bakeries, butcheries. I love going to the flower shop most of all because flowers are my passion. Every week I go to an antique (curiosity) shop, because I really enjoy the original, ancient things. From time to time I visit the toy store in order to buy toys for my nephews and children. Almost every month I go to the gift shop so that I can buy gifts on birthday for my family and friends.';
    INSERT INTO blocks (id, note_id, block_type_id, position, content, created_at, updated_at) VALUES
    (block_id, note_id, 1, 1, block_content2, NOW() - INTERVAL '5 months', NOW());
    
    INSERT INTO block_formatting (block_id, start_pos, end_pos, bold, italic, underline, text_align) VALUES
    (block_id, 0, 50, TRUE, TRUE, FALSE, 0);
    INSERT INTO block_formatting (block_id, start_pos, end_pos, bold, italic, underline, text_align) VALUES
    (block_id, 100, 180, FALSE, FALSE, TRUE, 0);
    INSERT INTO block_formatting (block_id, start_pos, end_pos, bold, italic, underline, text_align) VALUES
    (block_id, 200, 280, FALSE, FALSE, FALSE, 1);
    
    block_id := 'd3333333-3333-3333-3333-333333333333';
    block_content := 'I like to spend my time on shopping, preferably I like the self-service shops. You can scrutinize something as long as you like. A nagging seller does not hurry you, you are your own master. After it all, you can calmly go to the cashier, where all purchases will be counted and added up. In our time, it''s not only supermarkets that work in such a way, but also department stores, clothing shops and household goods shops.';
    INSERT INTO blocks (id, note_id, block_type_id, position, content, created_at, updated_at) VALUES
    (block_id, note_id, 1, 2, block_content, NOW() - INTERVAL '5 months', NOW());
    
    INSERT INTO block_formatting (block_id, start_pos, end_pos, bold, italic, underline, text_align) VALUES
    (block_id, 50, 150, FALSE, TRUE, FALSE, 2);
END $$;

SELECT setval('block_types_id_seq', (SELECT MAX(id) FROM block_types));