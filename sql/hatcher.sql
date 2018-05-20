--
-- PostgreSQL database dump
--

-- Dumped from database version 10.4 (Debian 10.4-1.pgdg90+1)
-- Dumped by pg_dump version 10.3

-- Started on 2018-05-19 10:04:53 PDT

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET client_min_messages = warning;
SET row_security = off;

--
-- TOC entry 2856 (class 1262 OID 16384)
-- Name: hatcher; Type: DATABASE; Schema: -; Owner: postgres
--

CREATE DATABASE hatcher WITH TEMPLATE = template0 ENCODING = 'UTF8' LC_COLLATE = 'en_US.utf8' LC_CTYPE = 'en_US.utf8';


ALTER DATABASE hatcher OWNER TO postgres;

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

--
-- TOC entry 7 (class 2615 OID 16385)
-- Name: hatcher; Type: SCHEMA; Schema: -; Owner: postgres
--

CREATE SCHEMA hatcher;


ALTER SCHEMA hatcher OWNER TO postgres;

--
-- TOC entry 1 (class 3079 OID 12980)
-- Name: plpgsql; Type: EXTENSION; Schema: -; Owner: 
--

CREATE EXTENSION IF NOT EXISTS plpgsql WITH SCHEMA pg_catalog;


--
-- TOC entry 2858 (class 0 OID 0)
-- Dependencies: 1
-- Name: EXTENSION plpgsql; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION plpgsql IS 'PL/pgSQL procedural language';


SET default_tablespace = '';

SET default_with_oids = false;

--
-- TOC entry 198 (class 1259 OID 16396)
-- Name: users; Type: TABLE; Schema: hatcher; Owner: postgres
--

CREATE TABLE hatcher.users (
    id integer NOT NULL,
    user_id text NOT NULL,
    email text,
    full_name text
);


ALTER TABLE hatcher.users OWNER TO postgres;

--
-- TOC entry 197 (class 1259 OID 16394)
-- Name: users_id_seq; Type: SEQUENCE; Schema: hatcher; Owner: postgres
--

CREATE SEQUENCE hatcher.users_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE hatcher.users_id_seq OWNER TO postgres;

--
-- TOC entry 2859 (class 0 OID 0)
-- Dependencies: 197
-- Name: users_id_seq; Type: SEQUENCE OWNED BY; Schema: hatcher; Owner: postgres
--

ALTER SEQUENCE hatcher.users_id_seq OWNED BY hatcher.users.id;


--
-- TOC entry 2727 (class 2604 OID 16399)
-- Name: users id; Type: DEFAULT; Schema: hatcher; Owner: postgres
--

ALTER TABLE ONLY hatcher.users ALTER COLUMN id SET DEFAULT nextval('hatcher.users_id_seq'::regclass);


--
-- TOC entry 2729 (class 2606 OID 16404)
-- Name: users users_pkey; Type: CONSTRAINT; Schema: hatcher; Owner: postgres
--

ALTER TABLE ONLY hatcher.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (user_id);


-- Completed on 2018-05-19 10:04:54 PDT

--
-- PostgreSQL database dump complete
--

