CREATE USER noterian_app_user WITH PASSWORD 'noterian_app_password';

GRANT CONNECT ON DATABASE noterian_db TO noterian_app_user;

GRANT USAGE ON SCHEMA public TO noterian_app_user;

GRANT SELECT, INSERT, UPDATE, DELETE ON profiles, notes, blocks, block_formatting, attachments, avatars, headers TO noterian_app_user;

GRANT USAGE ON SEQUENCE block_types_id_seq TO noterian_app_user;