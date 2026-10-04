-- ============================================================
-- 000004: 事务性发件箱(sys_outbox_message)
--
-- 从单体 db/migrations/000015 平移。订单域要用它发**延迟取消消息**
-- (下单时写一条 delay:{order_id},由投递器在 TTL 后发出),
-- 与订单表同库才能在一个本地事务里写 —— 这正是 outbox 模式的意义:
-- 事务提交则消息必达,不会出现"订单建了消息没发"。
--
-- message_id 唯一索引是投递幂等的凭据,必须保留。
-- 幂等:CREATE ... IF NOT EXISTS。
-- ============================================================

CREATE TABLE IF NOT EXISTS sys_outbox_message (
    id             BIGSERIAL    PRIMARY KEY,
    message_id     VARCHAR(64)  NOT NULL,
    aggregate_type VARCHAR(32)  NOT NULL,
    aggregate_id   VARCHAR(64)  NOT NULL,
    event_type     VARCHAR(64)  NOT NULL,
    exchange       VARCHAR(64)  NOT NULL,
    routing_key    VARCHAR(64)  NOT NULL,
    payload        TEXT         NOT NULL,
    status         VARCHAR(16)  NOT NULL DEFAULT 'pending',
    retry_count    INT          NOT NULL DEFAULT 0,
    next_retry_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    created_at     TIMESTAMPTZ  NOT NULL DEFAULT now(),
    sent_at        TIMESTAMPTZ,
    CONSTRAINT uk_outbox_message_id UNIQUE (message_id)
);

COMMENT ON COLUMN sys_outbox_message.message_id IS '业务幂等键(如 delay:{order_id})';
COMMENT ON COLUMN sys_outbox_message.aggregate_type IS '聚合根:order/coupon/...';
COMMENT ON COLUMN sys_outbox_message.aggregate_id IS '聚合根ID';
COMMENT ON COLUMN sys_outbox_message.event_type IS '事件类型:OrderDelayCancel/...';
COMMENT ON COLUMN sys_outbox_message.payload IS '事件负载(C4 起为 JSON,不再是单值)';
COMMENT ON COLUMN sys_outbox_message.status IS 'pending/sent';
COMMENT ON COLUMN sys_outbox_message.next_retry_at IS '下次投递时间(失败退避用)';
COMMENT ON COLUMN sys_outbox_message.sent_at IS '投递成功时间';

-- 投递器只扫 pending,部分索引让它不必扫已投递的历史行
CREATE INDEX IF NOT EXISTS idx_outbox_pending
    ON sys_outbox_message (next_retry_at) WHERE status = 'pending';
