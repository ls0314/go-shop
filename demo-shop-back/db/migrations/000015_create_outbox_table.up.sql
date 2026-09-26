CREATE TABLE IF NOT EXISTS sys_outbox_message (
    id             BIGSERIAL    PRIMARY KEY,
    message_id     VARCHAR(64)  NOT NULL,      -- 业务幂等键(如 delay:{order_id})
    aggregate_type VARCHAR(32)  NOT NULL,      -- 聚合根:order/coupon/...(C4 Saga 复用)
    aggregate_id   VARCHAR(64)  NOT NULL,
    event_type     VARCHAR(64)  NOT NULL,      -- 事件类型:OrderDelayCancel/...
    exchange       VARCHAR(64)  NOT NULL,
    routing_key    VARCHAR(64)  NOT NULL,
    payload        TEXT         NOT NULL,
    status         VARCHAR(16)  NOT NULL DEFAULT 'pending',   -- pending/sent
    retry_count    INT          NOT NULL DEFAULT 0,
    next_retry_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    created_at     TIMESTAMPTZ  NOT NULL DEFAULT now(),
    sent_at        TIMESTAMPTZ,
    CONSTRAINT uk_outbox_message_id UNIQUE (message_id)
);
CREATE INDEX idx_outbox_pending ON sys_outbox_message (next_retry_at) WHERE status = 'pending';