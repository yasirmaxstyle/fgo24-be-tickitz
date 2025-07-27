DROP TRIGGER IF EXISTS update_users_updated_at ON users;

DROP TRIGGER IF EXISTS update_profile_updated_at ON profile;

DROP FUNCTION IF EXISTS update_updated_at_column ();

DROP TABLE IF EXISTS tickets;

DROP TABLE IF EXISTS transactions;

DROP TABLE IF EXISTS users;

DROP TABLE IF EXISTS profile;

DROP TABLE IF EXISTS payment_method;

DROP TABLE IF EXISTS showtimes;

DROP TABLE IF EXISTS cinemas;

DROP TABLE IF EXISTS movies_genres;

DROP TABLE IF EXISTS movies_cast;

DROP TABLE IF EXISTS movies;

DROP TABLE IF EXISTS genres;

DROP TABLE IF EXISTS actors;

DROP TABLE IF EXISTS directors;