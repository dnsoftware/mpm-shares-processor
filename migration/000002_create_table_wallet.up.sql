-- Table: public.wallets

-- DROP TABLE IF EXISTS public.wallets;

CREATE TABLE IF NOT EXISTS public.wallets
(
    id BIGSERIAL PRIMARY KEY,
    coin_id bigint NOT NULL,
    name character varying(255) COLLATE pg_catalog."default" NOT NULL,
    current_hashrate bigint NOT NULL DEFAULT '0'::bigint,
    average_hashrate bigint NOT NULL DEFAULT '0'::bigint,
    payment_threshold numeric(20,6) NOT NULL DEFAULT '0'::numeric,
    is_solo boolean NOT NULL DEFAULT false,
    reward_method character varying(16) COLLATE pg_catalog."default" NOT NULL DEFAULT ''::character varying,
    CONSTRAINT wallets_coin_id_foreign FOREIGN KEY (coin_id)
        REFERENCES public.coins (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
)

    TABLESPACE pg_default;

-- Index: wallets_coin_id_index

-- DROP INDEX IF EXISTS public.wallets_coin_id_index;

CREATE INDEX IF NOT EXISTS wallets_coin_id_index
    ON public.wallets USING btree
        (coin_id ASC NULLS LAST)
    TABLESPACE pg_default;
-- Index: wallets_index_3

-- DROP INDEX IF EXISTS public.wallets_index_3;

CREATE INDEX IF NOT EXISTS wallets_index_3
    ON public.wallets USING btree
        (name COLLATE pg_catalog."default" ASC NULLS LAST)
    TABLESPACE pg_default;
-- Index: wallets_is_solo_index

-- DROP INDEX IF EXISTS public.wallets_is_solo_index;

CREATE INDEX IF NOT EXISTS wallets_is_solo_index
    ON public.wallets USING btree
        (is_solo ASC NULLS LAST)
    TABLESPACE pg_default;
-- Index: wallets_reward_method_index

-- DROP INDEX IF EXISTS public.wallets_reward_method_index;

CREATE INDEX IF NOT EXISTS wallets_reward_method_index
    ON public.wallets USING btree
        (reward_method COLLATE pg_catalog."default" ASC NULLS LAST)
    TABLESPACE pg_default;