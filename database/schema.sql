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

CREATE EXTENSION IF NOT EXISTS plpgsql WITH SCHEMA pg_catalog;

COMMENT ON EXTENSION plpgsql IS 'PL/pgSQL procedural language';

SET default_with_oids = false;

CREATE TABLE hatcher.happiness (
    userid text,
    id integer NOT NULL,
    date text,
    time text,
    results integer DEFAULT 0
);

CREATE SEQUENCE hatcher.happinessid_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE hatcher.happinessid_seq OWNED BY hatcher.happiness.id;

CREATE TABLE hatcher.standupyesterday (
    response text,
    timestamp text,
    date text,
    userid text,
    time text,
    uuid text,
    id integer NOT NULL
);

CREATE SEQUENCE hatcher.standupyesterday_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE hatcher.standupyesterday_seq OWNED BY hatcher.standupyesterday.id;

CREATE TABLE hatcher.standuptoday (
    response text,
    timestamp text,
    date text,
    userid text,
    time text,
    uuid text,
    id integer NOT NULL
);

CREATE SEQUENCE hatcher.standuptoday_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE hatcher.standuptoday_seq OWNED BY hatcher.standuptoday.id;

CREATE TABLE hatcher.standupblocker (
    response text,
    timestamp text,
    date text,
    userid text,
    time text,
    uuid text,
    id integer NOT NULL
);

CREATE SEQUENCE hatcher.standupblocker_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE hatcher.standupblocker_seq OWNED BY hatcher.standupblocker.id;

CREATE TABLE hatcher.users (
    id integer NOT NULL,
    userid text NOT NULL,
    email text,
    full_name text,
    managerid text,
    ismanager boolean DEFAULT false,
    displayname text,
    happiness_schedule time without time zone,
    standup_schedule time without time zone,
    standup_channel text
);

CREATE SEQUENCE hatcher.usersid_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE hatcher.usersid_seq OWNED BY hatcher.users.id;

CREATE TABLE hatcher.standupresults (
    id integer NOT NULL,
    date text,
    time text,
    uuid text,
    timestamp text
);

CREATE SEQUENCE hatcher.standupresults_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE hatcher.standupresults_seq OWNED BY hatcher.standupresults.id;

ALTER TABLE ONLY hatcher.happiness ALTER COLUMN id SET DEFAULT nextval('hatcher.happinessid_seq'::regclass);

ALTER TABLE ONLY hatcher.users ALTER COLUMN id SET DEFAULT nextval('hatcher.usersid_seq'::regclass);

ALTER TABLE ONLY hatcher.standupyesterday ALTER COLUMN id SET DEFAULT nextval('hatcher.standupyesterday_seq'::regclass);

ALTER TABLE ONLY hatcher.standuptoday ALTER COLUMN id SET DEFAULT nextval('hatcher.standuptoday_seq'::regclass);

ALTER TABLE ONLY hatcher.standupblocker ALTER COLUMN id SET DEFAULT nextval('hatcher.standupblocker_seq'::regclass);

ALTER TABLE ONLY hatcher.standupresults ALTER COLUMN id SET DEFAULT nextval('hatcher.standupresults_seq'::regclass);

ALTER TABLE ONLY hatcher.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (userid);

