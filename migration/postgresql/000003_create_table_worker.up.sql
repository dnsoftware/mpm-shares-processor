-- Table: public.workers

-- DROP TABLE IF EXISTS public.workers;

CREATE TABLE IF NOT EXISTS public.workers
(
    id BIGSERIAL PRIMARY KEY,
    coin_id bigint NOT NULL,
    workerfull character varying(255) COLLATE pg_catalog."default" NOT NULL,
    wallet character varying(255) COLLATE pg_catalog."default" NOT NULL,
    worker character varying(255) COLLATE pg_catalog."default" NOT NULL,
    server_id character varying(32) COLLATE pg_catalog."default" NOT NULL,
    ip character varying(32) COLLATE pg_catalog."default",
    current_hashrate bigint NOT NULL DEFAULT '0'::bigint,
    average_hashrate bigint NOT NULL DEFAULT '0'::bigint,
    last_share_date timestamp(0) without time zone,
    is_connect boolean NOT NULL DEFAULT false,
    current_diff numeric(25,10) NOT NULL DEFAULT '0'::numeric,
    created_at timestamp(0) without time zone,
    updated_at timestamp(0) without time zone,
    reported_hashrate bigint,
    reported_hashrate_date timestamp(0) without time zone,
    threshold_change text COLLATE pg_catalog."default" NOT NULL DEFAULT ''::text,
    is_solo boolean NOT NULL DEFAULT false,
    miner_client character varying(255) COLLATE pg_catalog."default" NOT NULL DEFAULT ''::character varying,
    reward_method character varying(16) COLLATE pg_catalog."default" NOT NULL DEFAULT ''::character varying,
    CONSTRAINT workers_coin_id_foreign FOREIGN KEY (coin_id)
        REFERENCES public.coins (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
)

    TABLESPACE pg_default;

-- Index: workers_coin_id_index

-- DROP INDEX IF EXISTS public.workers_coin_id_index;

CREATE INDEX IF NOT EXISTS workers_coin_id_index
    ON public.workers USING btree
        (coin_id ASC NULLS LAST)
    TABLESPACE pg_default;
-- Index: workers_created_at_index

-- DROP INDEX IF EXISTS public.workers_created_at_index;

CREATE INDEX IF NOT EXISTS workers_created_at_index
    ON public.workers USING btree
        (created_at ASC NULLS LAST)
    TABLESPACE pg_default;
-- Index: workers_is_solo_index

-- DROP INDEX IF EXISTS public.workers_is_solo_index;

CREATE INDEX IF NOT EXISTS workers_is_solo_index
    ON public.workers USING btree
        (is_solo ASC NULLS LAST)
    TABLESPACE pg_default;
-- Index: workers_reward_method_index

-- DROP INDEX IF EXISTS public.workers_reward_method_index;

CREATE INDEX IF NOT EXISTS workers_reward_method_index
    ON public.workers USING btree
        (reward_method COLLATE pg_catalog."default" ASC NULLS LAST)
    TABLESPACE pg_default;
-- Index: workers_server_id_index

-- DROP INDEX IF EXISTS public.workers_server_id_index;

CREATE INDEX IF NOT EXISTS workers_server_id_index
    ON public.workers USING btree
        (server_id COLLATE pg_catalog."default" ASC NULLS LAST)
    TABLESPACE pg_default;
-- Index: workers_updated_at_index

-- DROP INDEX IF EXISTS public.workers_updated_at_index;

CREATE INDEX IF NOT EXISTS workers_updated_at_index
    ON public.workers USING btree
        (updated_at ASC NULLS LAST)
    TABLESPACE pg_default;
-- Index: workers_wallet_index

-- DROP INDEX IF EXISTS public.workers_wallet_index;

CREATE INDEX IF NOT EXISTS workers_wallet_index
    ON public.workers USING btree
        (wallet COLLATE pg_catalog."default" ASC NULLS LAST)
    TABLESPACE pg_default;
-- Index: workers_worker_index

-- DROP INDEX IF EXISTS public.workers_worker_index;

CREATE INDEX IF NOT EXISTS workers_worker_index
    ON public.workers USING btree
        (worker COLLATE pg_catalog."default" ASC NULLS LAST)
    TABLESPACE pg_default;
-- Index: workers_workerfull_index

-- DROP INDEX IF EXISTS public.workers_workerfull_index;

CREATE INDEX IF NOT EXISTS workers_workerfull_index
    ON public.workers USING btree
        (workerfull COLLATE pg_catalog."default" ASC NULLS LAST)
    TABLESPACE pg_default;