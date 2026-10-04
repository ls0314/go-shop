-- ============================================================
-- 000001: 购物车表(user_cart_item)
--
-- 从单体 db/migrations/000007 平移,两处**刻意的删减**:
--   ① 去掉 fk_cart_user → sys_user 外键:sys_user 在 user_db,跨库外键建不出来。
--      归属改由应用层保证(读写都校验 cart_item.user_id == 调用方传入的 user_id)。
--   ② 去掉 fk_cart_sku → sys_product_sku 外键:那张表在 product_db,同理。
--
-- **保留 uk_user_sku**:同一个用户对同一 SKU 只允许一行 —— 加购逻辑靠它
-- 做 upsert(存在则累加数量),少了它并发加购会产生两行,合计就算重了。
--
-- 幂等:CREATE ... IF NOT EXISTS;约束包在 DO 块里(重跑不报重复)。
-- ============================================================

CREATE TABLE IF NOT EXISTS user_cart_item (
    cart_item_id BIGSERIAL NOT NULL,
    user_id BIGINT NOT NULL,
    sku_id BIGINT NOT NULL,
    quantity INT NOT NULL DEFAULT 1,
    is_selected BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT pk_cart_item_id PRIMARY KEY (cart_item_id),
    CONSTRAINT ck_cart_quantity CHECK (quantity >= 1)
);

COMMENT ON COLUMN user_cart_item.cart_item_id IS '主键，购物车项唯一标识';
COMMENT ON COLUMN user_cart_item.user_id IS '所属用户ID(跨库引用 user_db.sys_user,无外键)';
COMMENT ON COLUMN user_cart_item.sku_id IS '关联SKU ID(跨库引用 product_db.sys_product_sku,无外键)';
COMMENT ON COLUMN user_cart_item.quantity IS '购买数量，最小1';
COMMENT ON COLUMN user_cart_item.is_selected IS '是否选中（用于结算）';
COMMENT ON COLUMN user_cart_item.created_at IS '加入时间';
COMMENT ON COLUMN user_cart_item.updated_at IS '最后修改时间';

-- 一个用户对一个 SKU 只有一行:加购走 upsert,并发加购不会产生重复行
CREATE UNIQUE INDEX IF NOT EXISTS uk_user_sku ON user_cart_item (user_id, sku_id);

-- 购物车列表:按用户查、按更新时间倒序
CREATE INDEX IF NOT EXISTS idx_cart_user_updated ON user_cart_item (user_id, updated_at DESC);
