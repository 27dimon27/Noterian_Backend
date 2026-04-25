UPDATE profiles SET role_id = 1 WHERE id = '11111111-1111-1111-1111-111111111111';
UPDATE profiles SET role_id = 1 WHERE id = '22222222-2222-2222-2222-222222222222';
UPDATE profiles SET role_id = 2 WHERE id = '33333333-3333-3333-3333-333333333333';

-- Тикет 1: open (саппорт не назначен)
INSERT INTO support_tickets (id, user_id, category_id, status_id, assigned_to, title, description, priority, created_at, updated_at) VALUES
(
    '77777777-7777-7777-7777-777777777777',
    '11111111-1111-1111-1111-111111111111',
    (SELECT id FROM support_categories WHERE name = 'bug'),
    (SELECT id FROM support_statuses WHERE name = 'open'),
    NULL,
    'Не сохраняются изменения в блоке текста',
    'При редактировании текстового блока изменения исчезают после обновления страницы. Пробовал в разных браузерах (Chrome, Firefox) - проблема повторяется. В консоли ошибок нет. Версия приложения последняя.',
    2,
    NOW() - INTERVAL '2 days',
    NOW() - INTERVAL '2 days'
);

-- Сообщение к тикету 1
INSERT INTO support_messages (id, ticket_id, user_id, message, is_internal, created_at, updated_at) VALUES
(
    '88888888-8888-8888-8888-888888888888',
    '77777777-7777-7777-7777-777777777777',
    '11111111-1111-1111-1111-111111111111',
    'Здравствуйте! Уже несколько дней не могу нормально работать с заметками. Пожалуйста, помогите разобраться.',
    FALSE,
    NOW() - INTERVAL '2 days',
    NOW() - INTERVAL '2 days'
);

-- Тикет 2: in_progress (саппорт назначен)
INSERT INTO support_tickets (id, user_id, category_id, status_id, assigned_to, title, description, priority, created_at, updated_at) VALUES
(
    '77778888-7777-7888-7777-777777777778',
    '22222222-2222-2222-2222-222222222222',
    (SELECT id FROM support_categories WHERE name = 'bug'),
    (SELECT id FROM support_statuses WHERE name = 'in_progress'),
    '33333333-3333-3333-3333-333333333333',
    'Не загружаются изображения в блоке image',
    'При попытке вставить изображение в блок типа "image" появляется ошибка "Failed to upload". Размер файла около 2 МБ, формат JPG. Другие форматы тоже не работают.',
    1,
    NOW() - INTERVAL '5 days',
    NOW() - INTERVAL '1 day'
);

-- Сообщения к тикету 2
INSERT INTO support_messages (id, ticket_id, user_id, message, is_internal, created_at, updated_at) VALUES
(
    '88888888-8888-8888-8888-888888888889',
    '77778888-7777-7888-7777-777777777778',
    '22222222-2222-2222-2222-222222222222',
    'Добрый день! Уже третий день не могу вставить изображения в заметки. Очень мешает работе.',
    FALSE,
    NOW() - INTERVAL '5 days',
    NOW() - INTERVAL '5 days'
),
(
    '88888888-8888-8888-8888-888888888890',
    '77778888-7777-7888-7777-777777777778',
    '33333333-3333-3333-3333-333333333333',
    'Здравствуйте! Я назначен ответственным по вашему вопросу. Проверяю проблему с загрузкой изображений. Какая у вас операционная система?',
    FALSE,
    NOW() - INTERVAL '3 days',
    NOW() - INTERVAL '3 days'
),
(
    '88888888-8888-8888-8888-888888888891',
    '77778888-7777-7888-7777-777777777778',
    '22222222-2222-2222-2222-222222222222',
    'Windows 11, последняя версия. Проблема проявляется и в Edge, и в Chrome.',
    FALSE,
    NOW() - INTERVAL '2 days',
    NOW() - INTERVAL '2 days'
),
(
    '88888888-8888-8888-8888-888888888892',
    '77778888-7777-7888-7777-777777777778',
    '33333333-3333-3333-3333-333333333333',
    'Внутренняя заметка: проверить настройки CORS и лимиты загрузки на MinIO',
    TRUE,
    NOW() - INTERVAL '1 day',
    NOW() - INTERVAL '1 day'
);

-- Тикет 3: waiting_user (саппорт ждёт ответа пользователя)
INSERT INTO support_tickets (id, user_id, category_id, status_id, assigned_to, title, description, priority, created_at, updated_at) VALUES
(
    '77778888-7777-7888-7777-777777777779',
    '11111111-1111-1111-1111-111111111111',
    (SELECT id FROM support_categories WHERE name = 'suggestion'),
    (SELECT id FROM support_statuses WHERE name = 'waiting_user'),
    '33333333-3333-3333-3333-333333333333',
    'Предложение: добавить тёмную тему',
    'Было бы здорово добавить тёмную тему в приложение. Сейчас при работе по вечерам глаза сильно устают от светлого фона. Многие пользователи поддержат эту идею!',
    3,
    NOW() - INTERVAL '10 days',
    NOW() - INTERVAL '2 days'
);

-- Сообщения к тикету 3
INSERT INTO support_messages (id, ticket_id, user_id, message, is_internal, created_at, updated_at) VALUES
(
    '88888888-8888-8888-8888-888888888893',
    '77778888-7777-7888-7777-777777777779',
    '11111111-1111-1111-1111-111111111111',
    'Привет! Есть предложение по улучшению продукта. Работаю в основном вечером, светлая тема режет глаза.',
    FALSE,
    NOW() - INTERVAL '10 days',
    NOW() - INTERVAL '10 days'
),
(
    '88888888-8888-8888-8888-888888888894',
    '77778888-7777-7888-7777-777777777779',
    '33333333-3333-3333-3333-333333333333',
    'Здравствуйте! Спасибо за предложение. Тёмная тема действительно в планах развития. Можете уточнить, какие элементы интерфейса для вас наиболее критичны? Рабочая область редактора, боковая панель или всё вместе?',
    FALSE,
    NOW() - INTERVAL '5 days',
    NOW() - INTERVAL '5 days'
),
(
    '88888888-8888-8888-8888-888888888895',
    '77778888-7777-7888-7777-777777777779',
    '33333333-3333-3333-3333-333333333333',
    'Внутренняя заметка: предложение добавить в бэклог на Q2',
    TRUE,
    NOW() - INTERVAL '2 days',
    NOW() - INTERVAL '2 days'
);

-- Тикет 4: closed (закрыт)
INSERT INTO support_tickets (id, user_id, category_id, status_id, assigned_to, title, description, priority, created_at, updated_at, resolved_at) VALUES
(
    '77778888-7777-7888-7777-777777777780',
    '22222222-2222-2222-2222-222222222222',
    (SELECT id FROM support_categories WHERE name = 'bug'),
    (SELECT id FROM support_statuses WHERE name = 'closed'),
    '33333333-3333-3333-3333-333333333333',
    'Ошибка: дублирование записей при быстром сохранении',
    'При быстром нажатии Ctrl+S несколько раз подряд создаются дубликаты блоков. Воспроизводится всегда. Заметка может разрастаться до десятков копий одного блока.',
    1,
    NOW() - INTERVAL '20 days',
    NOW() - INTERVAL '1 day',
    NOW() - INTERVAL '1 day'
);

-- Сообщения к тикету 4
INSERT INTO support_messages (id, ticket_id, user_id, message, is_internal, created_at, updated_at) VALUES
(
    '88888888-8888-8888-8888-888888888896',
    '77778888-7777-7888-7777-777777777780',
    '22222222-2222-2222-2222-222222222222',
    'Обнаружил странный баг с дублированием контента.',
    FALSE,
    NOW() - INTERVAL '20 days',
    NOW() - INTERVAL '20 days'
),
(
    '88888888-8888-8888-8888-888888888897',
    '77778888-7777-7888-7777-777777777780',
    '33333333-3333-3333-3333-333333333333',
    'Принято в работу. Проблема воспроизводится. Ищем решение.',
    FALSE,
    NOW() - INTERVAL '18 days',
    NOW() - INTERVAL '18 days'
),
(
    '88888888-8888-8888-8888-888888888898',
    '77778888-7777-7888-7777-777777777780',
    '33333333-3333-3333-3333-333333333333',
    'Проблема исправлена в версии 1.2.4. Обновите приложение и проверьте.',
    FALSE,
    NOW() - INTERVAL '2 days',
    NOW() - INTERVAL '2 days'
),
(
    '88888888-8888-8888-8888-888888888899',
    '77778888-7777-7888-7777-777777777780',
    '22222222-2222-2222-2222-222222222222',
    'Обновил, всё работает! Спасибо за быструю помощь!',
    FALSE,
    NOW() - INTERVAL '1 day',
    NOW() - INTERVAL '1 day'
);

-- Оценка для закрытого тикета
INSERT INTO support_ratings (id, ticket_id, user_id, rating, comment, created_at, updated_at) VALUES
(
    '99999999-9999-9999-9999-999999999999',
    '77778888-7777-7888-7777-777777777780',
    '22222222-2222-2222-2222-222222222222',
    5,
    'Отличная работа поддержки! Быстро нашли и исправили проблему.',
    NOW() - INTERVAL '1 day',
    NOW() - INTERVAL '1 day'
);

-- Тикет 5: reopened (пользователь открыл заново)
INSERT INTO support_tickets (id, user_id, category_id, status_id, assigned_to, title, description, priority, created_at, updated_at) VALUES
(
    '77778888-7777-7888-7777-777777777781',
    '11111111-1111-1111-1111-111111111111',
    (SELECT id FROM support_categories WHERE name = 'bug'),
    (SELECT id FROM support_statuses WHERE name = 'reopened'),
    '33333333-3333-3333-3333-333333333333',
    'Проблема с сохранением вернулась',
    'После последнего обновления снова перестали сохраняться изменения. Думал проблема была исправлена, но нет. Та же ситуация - изменения пропадают при перезагрузке.',
    1,
    NOW() - INTERVAL '3 days',
    NOW() - INTERVAL '1 day'
);

-- Сообщения к тикету 5
INSERT INTO support_messages (id, ticket_id, user_id, message, is_internal, created_at, updated_at) VALUES
(
    '88888888-8888-8888-8888-888888888900',
    '77778888-7777-7888-7777-777777777781',
    '11111111-1111-1111-1111-111111111111',
    'Проблема с сохранением вернулась в новой версии. Очень расстраивает.',
    FALSE,
    NOW() - INTERVAL '3 days',
    NOW() - INTERVAL '3 days'
),
(
    '88888888-8888-8888-8888-888888888901',
    '77778888-7777-7888-7777-777777777781',
    '33333333-3333-3333-3333-333333333333',
    'Спасибо, что сообщили. Проверяем регрессию. Приносим извинения за неудобства.',
    FALSE,
    NOW() - INTERVAL '1 day',
    NOW() - INTERVAL '1 day'
);

-- Тикет 6: еще один тикет для user1 (other категория)
INSERT INTO support_tickets (id, user_id, category_id, status_id, assigned_to, title, description, priority, created_at, updated_at) VALUES
(
    '77778888-7777-7888-7777-777777777782',
    '11111111-1111-1111-1111-111111111111',
    (SELECT id FROM support_categories WHERE name = 'other'),
    (SELECT id FROM support_statuses WHERE name = 'open'),
    NULL,
    'Вопрос по API интеграции',
    'Планирую интегрировать ваше приложение со своим сервисом. Есть ли публичное API для работы с заметками? Если да, то где можно посмотреть документацию?',
    2,
    NOW() - INTERVAL '1 day',
    NOW() - INTERVAL '1 day'
);

-- Сообщение к тикету 6
INSERT INTO support_messages (id, ticket_id, user_id, message, is_internal, created_at, updated_at) VALUES
(
    '88888888-8888-8888-8888-888888888902',
    '77778888-7777-7888-7777-777777777782',
    '11111111-1111-1111-1111-111111111111',
    'Интересует возможность программного создания и редактирования заметок.',
    FALSE,
    NOW() - INTERVAL '1 day',
    NOW() - INTERVAL '1 day'
);

SELECT setval('support_categories_id_seq', (SELECT MAX(id) FROM support_categories));
SELECT setval('support_statuses_id_seq', (SELECT MAX(id) FROM support_statuses));
SELECT setval('user_roles_id_seq', (SELECT MAX(id) FROM user_roles));