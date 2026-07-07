CREATE TABLE IF NOT EXISTS user_address (
    address_id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL ,
    receiver_name VARCHAR(50) NOT NULL ,
    receiver_phone VARCHAR(20) NOT NULL ,
    province VARCHAR(50) NOT NULL ,
    city VARCHAR(50) NOT NULL ,
    district VARCHAR(50) NOT NULL ,
    detail_address VARCHAR(200) NOT NULL ,
    postal_code VARCHAR(10) DEFAULT NULL,
    is_default BOOLEAN NOT NULL DEFAULT FALSE,
    address_tag VARCHAR(20) DEFAULT NULL,
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL  DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL  DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON COLUMN user_address.address_id IS '主键，地址唯一标识';
COMMENT ON COLUMN user_address.user_id IS '所属用户ID，关联sys_user';
COMMENT ON COLUMN user_address.receiver_name IS '收货人姓名';
COMMENT ON COLUMN user_address.receiver_phone IS '收货人手机号';
COMMENT ON COLUMN user_address.province IS '省份';
COMMENT ON COLUMN user_address.city IS '城市';
COMMENT ON COLUMN user_address.district IS '区/县';
COMMENT ON COLUMN user_address.detail_address IS '详细地址（街道、门牌号等）';
COMMENT ON COLUMN user_address.postal_code IS '邮政编码';
COMMENT ON COLUMN user_address.is_default IS '是否默认地址';
COMMENT ON COLUMN user_address.address_tag IS '地址标签（家/公司/学校）';
COMMENT ON COLUMN user_address.is_deleted IS '软删除标记';
COMMENT ON COLUMN user_address.created_at IS '创建时间';
COMMENT ON COLUMN user_address.updated_at IS '更新时间';

-- 联合索引
-- 按用户查询有效地址
CREATE INDEX IF NOT EXISTS idx_address_user_id ON user_address(user_id, is_deleted);
-- 快速定位默认地址
CREATE INDEX IF NOT EXISTS idx_address_default ON user_address(user_id,  is_default);

-- 约束说明
-- 手机号格式
ALTER TABLE user_address ADD CONSTRAINT ck_receiver_phone CHECK (receiver_phone ~ '^1[3-9]\d{9}$');
-- 收货人非空
ALTER TABLE user_address ADD CONSTRAINT ck_receiver_name CHECK (char_length(receiver_name) > 0);

-- 外键约束：禁止级联删除（用户有地址时禁止删用户）
-- 创建外键，设置级联策略
ALTER TABLE user_address ADD CONSTRAINT fk_address_user_id FOREIGN KEY (user_id) REFERENCES sys_user (user_id) ON DELETE RESTRICT  ON UPDATE CASCADE;




