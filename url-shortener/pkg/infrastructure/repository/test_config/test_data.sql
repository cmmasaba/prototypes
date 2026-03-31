-- insert test user
SET search_path TO urlshortener;

INSERT INTO users (public_id, email, password) VALUES ('019d400e-74a2-7e3e-90a9-4761dae23795', 'test@email.com', '$2a$14$6sRnQWiTugnC7PrUWnU/ReyGSY1BtvSEG9yxT7S0Fc1bnZSznaHpe');

INSERT INTO users (public_id, email, oauth_provider, oauth_provider_id) VALUES ('019d4012-8e74-7fe6-a6ea-7f3c2a4cc59f', 'test2@email.com', 'Google', '6aa8ecb7-51ad-4287-aa81-1cc738d9a81a');

INSERT INTO links (user_id, short_code, original_url, ownership_token) VALUES (
	1, 'shortcode', 'http://example.net', 'randomstuff'
);
