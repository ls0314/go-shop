CREATE TABLE IF NOT EXISTS user_cart_item (
    cart_item_id BIGSERIAL NOT NULL,
    user_id BIGINT NOT NULL ,
    sku_id BIGINT NOT NULL ,
    quantity INT NOT NULL DEFAULT 1,
    is_selected BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL  DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    -- 主键约束
    CONSTRAINT pk_cart_item_id PRIMARY KEY (cart_item_id),
    -- 数值范围校验：购买数量最小为 1
    CONSTRAINT ck_cart_quantity CHECK (quantity >= 1),
    -- 外键约束：关联用户表
    CONSTRAINT fk_cart_user FOREIGN KEY (user_id) REFERENCES sys_user(user_id) ON DELETE CASCADE,
    -- 外键约束：关联商品 SKU 表
    CONSTRAINT fk_cart_sku FOREIGN KEY (sku_id) REFERENCES sys_product_sku(sku_id) ON DELETE CASCADE
);

COMMENT ON COLUMN user_cart_item.cart_item_id IS '主键，购物车项唯一标识';
COMMENT ON COLUMN user_cart_item.user_id IS '所属用户ID，关联sys_user';
COMMENT ON COLUMN user_cart_item.sku_id IS '关联SKU ID，关联sys_product_sku';
COMMENT ON COLUMN user_cart_item.quantity IS '购买数量，最小1';
COMMENT ON COLUMN user_cart_item.is_selected IS '是否选中（用于结算）';
COMMENT ON COLUMN user_cart_item.created_at IS '加入时间';
COMMENT ON COLUMN user_cart_item.updated_at IS '最后修改时间';

-- 用户筛选（唯一索引）
CREATE UNIQUE INDEX IF NOT EXISTS uk_user_sku ON user_cart_item (user_id, sku_id);

-- 联合索引：优化按用户查询购物车列表的性能，支持按更新时间倒序排序
CREATE INDEX idx_cart_user_updated ON user_cart_item (user_id, updated_at DESC);

