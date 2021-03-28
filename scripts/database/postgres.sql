/* SCHEMA */

CREATE TABLE IF NOT EXISTS messages
(
    id SERIAL CONSTRAINT messages_id_pk PRIMARY KEY,
    subsystem character varying[100] NOT NULL,
    message character varying NOT NULL,
    created_at timestamp with time zone NOT NULL,
    sent_at timestamp with time zone
);

CREATE TABLE IF NOT EXISTS recipients
(
	id bigserial CONSTRAINT recipients_id_pk PRIMARY KEY,
	recipient_id bigint UNIQUE NOT NULL,
	created_at timestamp with time zone NOT NULL
);
