-- Table: public.coins

-- DROP TABLE IF EXISTS public.coins;

CREATE TABLE IF NOT EXISTS public.coins
(
    id BIGSERIAL PRIMARY KEY,
    symbol character varying(32) COLLATE pg_catalog."default" NOT NULL,
    symbol2 character varying(32) COLLATE pg_catalog."default" DEFAULT ''::character varying,
    name character varying(255) COLLATE pg_catalog."default" NOT NULL DEFAULT ''::character varying,
    algo character varying(255) COLLATE pg_catalog."default" NOT NULL DEFAULT ''::character varying,
    image character varying(255) COLLATE pg_catalog."default" NOT NULL DEFAULT ''::character varying,
    min_withdraw numeric(10,6) NOT NULL DEFAULT '0'::numeric,
    transactions_explorer character varying(255) COLLATE pg_catalog."default" NOT NULL DEFAULT ''::character varying,
    block_explorer character varying(255) COLLATE pg_catalog."default" NOT NULL DEFAULT ''::character varying,
    is_active boolean NOT NULL DEFAULT false,
    params text COLLATE pg_catalog."default" NOT NULL DEFAULT ''::text,
    average_round_diff numeric(25,10) NOT NULL DEFAULT '0'::numeric,
    average_solo_round_diff numeric(25,10) NOT NULL DEFAULT '0'::numeric,
    current_effort numeric(10,2) NOT NULL DEFAULT '0'::numeric,
    coins_in_block numeric(25,10) NOT NULL DEFAULT '0'::numeric,
    average_pps_round_diff numeric(25,10) NOT NULL DEFAULT '0'::numeric,
    average_effort numeric(5,2) NOT NULL DEFAULT '0'::numeric,
    average_last_effort numeric(5,2) NOT NULL DEFAULT '0'::numeric,
    seo_title text COLLATE pg_catalog."default" NOT NULL DEFAULT ''::text,
    last_reward_processed_id bigint NOT NULL DEFAULT '0'::bigint

)

TABLESPACE pg_default;

-- Index: coins_is_active_index

-- DROP INDEX IF EXISTS public.coins_is_active_index;

CREATE INDEX IF NOT EXISTS coins_is_active_index
    ON public.coins USING btree
        (is_active ASC NULLS LAST)
    TABLESPACE pg_default;
-- Index: coins_symbol_index

-- DROP INDEX IF EXISTS public.coins_symbol_index;

CREATE INDEX IF NOT EXISTS coins_symbol_index
    ON public.coins USING btree
        (symbol COLLATE pg_catalog."default" ASC NULLS LAST)
    TABLESPACE pg_default;
-- Index: lower_symbol_field_index

-- DROP INDEX IF EXISTS public.lower_symbol_field_index;

CREATE INDEX IF NOT EXISTS lower_symbol_field_index
    ON public.coins USING btree
        (lower(symbol::text) COLLATE pg_catalog."default" ASC NULLS LAST)
    TABLESPACE pg_default;
