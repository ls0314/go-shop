-- ============================================================
-- 000001: 类目表
-- 幂等:CREATE ... IF NOT EXISTS / 索引 IF NOT EXISTS
-- ============================================================

CREATE TABLE IF NOT EXISTS sys_category (
    category_id BIGSERIAL PRIMARY KEY,
    parent_id BIGINT NOT NULL DEFAULT 0,
    category_name VARCHAR(100) NOT NULL,
    category_level SMALLINT NOT NULL DEFAULT 1,
    category_path VARCHAR(500) NOT NULL,
    sort_order INT NOT NULL DEFAULT 0,
    icon_url VARCHAR(500),
    is_leaf BOOLEAN NOT NULL DEFAULT FALSE,
    is_visible BOOLEAN NOT NULL DEFAULT TRUE,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    create_by BIGINT,
    update_by BIGINT,

    CONSTRAINT chk_category_level CHECK (category_level BETWEEN 1 AND 4),
    CONSTRAINT chk_category_status CHECK (status IN ('active', 'disabled')),
    CONSTRAINT chk_category_path CHECK (category_path ~ '^0(,[0-9]+)*$')
);

COMMENT ON COLUMN sys_category.category_id IS '主键，类目唯一标识';
COMMENT ON COLUMN sys_category.parent_id IS '父类目ID，0表示根节点';
COMMENT ON COLUMN sys_category.category_name IS '类目名称';
COMMENT ON COLUMN sys_category.category_level IS '类目层级（1-4）';
COMMENT ON COLUMN sys_category.category_path IS '路径枚举，如0,1,5';
COMMENT ON COLUMN sys_category.sort_order IS '排序权重，越小越靠前';
COMMENT ON COLUMN sys_category.icon_url IS '类目图标URL';
COMMENT ON COLUMN sys_category.is_leaf IS '是否叶子节点';
COMMENT ON COLUMN sys_category.is_visible IS '是否可见';
COMMENT ON COLUMN sys_category.status IS '状态：active/disabled';
COMMENT ON COLUMN sys_category.created_at IS '创建时间';
COMMENT ON COLUMN sys_category.updated_at IS '更新时间';
COMMENT ON COLUMN sys_category.create_by IS '创建人ID';
COMMENT ON COLUMN sys_category.update_by IS '更新人ID';

-- 查询子类目
CREATE INDEX IF NOT EXISTS idx_parent_id ON sys_category (parent_id);
-- 树形递归查询
CREATE INDEX IF NOT EXISTS idx_category_path ON sys_category (category_path);
-- 按层级排序展示
CREATE INDEX IF NOT EXISTS idx_level_sort ON sys_category (category_level, sort_order);
