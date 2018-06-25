SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET client_min_messages = warning;
SET row_security = off;

\connect hatcher

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET client_min_messages = warning;
SET row_security = off;

CREATE SCHEMA hatcher;

ALTER SCHEMA hatcher OWNER TO postgres;

CREATE EXTENSION IF NOT EXISTS plpgsql WITH SCHEMA pg_catalog;

COMMENT ON EXTENSION plpgsql IS 'PL/pgSQL procedural language';


SET default_tablespace = '';

SET default_with_oids = false;

CREATE TABLE hatcher.happiness (
    user_id text,
    id integer NOT NULL,
    date date DEFAULT CURRENT_DATE NOT NULL,
    results integer DEFAULT '0'::integer
);

ALTER TABLE hatcher.happiness OWNER TO postgres;

CREATE SEQUENCE hatcher.happiness_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER TABLE hatcher.happiness_id_seq OWNER TO postgres;

ALTER SEQUENCE hatcher.happiness_id_seq OWNED BY hatcher.happiness.id;

CREATE TABLE hatcher.users (
    id integer NOT NULL,
    user_id text NOT NULL,
    email text,
    full_name text,
    manager_id text,
    is_manager boolean DEFAULT false,
    displayname text COLLATE pg_catalog."en_US.utf8",
    happiness_schedule TIME
);

ALTER TABLE hatcher.users OWNER TO postgres;

CREATE SEQUENCE hatcher.users_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER TABLE hatcher.users_id_seq OWNER TO postgres;

ALTER SEQUENCE hatcher.users_id_seq OWNED BY hatcher.users.id;

ALTER TABLE ONLY hatcher.happiness ALTER COLUMN id SET DEFAULT nextval('hatcher.happiness_id_seq'::regclass);

ALTER TABLE ONLY hatcher.users ALTER COLUMN id SET DEFAULT nextval('hatcher.users_id_seq'::regclass);

ALTER TABLE ONLY hatcher.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (user_id);
