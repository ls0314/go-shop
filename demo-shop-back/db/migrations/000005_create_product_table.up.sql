CREATE TABLE IF NOT EXISTS sys_product_spu (
    spu_id BIGSERIAL PRIMARY KEY,
    spu_name VARCHAR(200) NOT NULL ,
    category_id BIGINT NOT NULL ,
    brand VARCHAR(200) DEFAULT NULL,
    description TEXT DEFAULT NULL,
    main_image VARCHAR(500) DEFAULT NULL,
    spec_template JSONB DEFAULT NULL,
    spu_status VARCHAR(20) NOT NULL DEFAULT 'draft',
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
    priority INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    create_by BIGINT DEFAULT NULL,
    update_by BIGINT DEFAULT NULL
);

COMMENT ON COLUMN sys_product_spu.spu_id IS '主键，商品SPU唯一标识';
COMMENT ON COLUMN sys_product_spu.spu_name IS '商品名称';
COMMENT ON COLUMN sys_product_spu.category_id IS '所属类目ID，关联sys_category';
COMMENT ON COLUMN sys_product_spu.brand IS '品牌';
COMMENT ON COLUMN sys_product_spu.description IS '商品描述（富文本HTML）';
COMMENT ON COLUMN sys_product_spu.main_image IS '商品主图URL';
COMMENT ON COLUMN sys_product_spu.spec_template IS '规格模板';
COMMENT ON COLUMN sys_product_spu.spu_status IS '商品状态';
COMMENT ON COLUMN sys_product_spu.is_deleted IS '软删除标记';
COMMENT ON COLUMN sys_product_spu.priority IS '排序权重';
COMMENT ON COLUMN sys_product_spu.created_at IS '创建时间';
COMMENT ON COLUMN sys_product_spu.updated_at IS '更新时间';
COMMENT ON COLUMN sys_product_spu.create_by IS '创建人ID';
COMMENT ON COLUMN sys_product_spu.update_by IS '更新人ID';

-- 按类目查询商品
CREATE INDEX IF NOT EXISTS idx_spu_category_id ON sys_product_spu (category_id);
-- 按状态筛选
CREATE INDEX IF NOT EXISTS idx_spu_status ON sys_product_spu (spu_status);
-- 排除软删除数据
CREATE INDEX IF NOT EXISTS idx_spu_is_deleted ON sys_product_spu (is_deleted);
-- 管理后台列表排序（联合索引）
CREATE INDEX IF NOT EXISTS idx_spu_status_priority ON sys_product_spu (spu_status, priority DESC, updated_at DESC);

-- 商品状态枚举约束：仅允许指定状态值
ALTER TABLE sys_product_spu ADD CONSTRAINT ck_spu_status CHECK (spu_status IN ('draft', 'published', 'withdrawn'));

-- 商品名称非空校验：禁止空字符串（配合 NOT NULL 同时禁止 NULL 和空串）
ALTER TABLE sys_product_spu ADD CONSTRAINT ck_spu_name_not_empty CHECK (char_length(spu_name) > 0);

-- 外键约束：禁止级联删除（类目下有商品时，禁止删除类目）
-- 创建外键，设置级联策略
ALTER TABLE sys_product_spu ADD CONSTRAINT fk_spu_category_spu FOREIGN KEY (category_id) REFERENCES sys_category (category_id) ON DELETE RESTRICT  ON UPDATE CASCADE;

CREATE TABLE IF NOT EXISTS sys_product_sku(
    sku_id BIGSERIAL PRIMARY KEY ,
    spu_id BIGINT NOT NULL ,
    sku_name VARCHAR(300) ,
    spec_values JSONB NOT NULL DEFAULT '{}'::jsonb,
    price DECIMAL(10, 2) NOT NULL ,
    cost_price DECIMAL(10, 2) ,
    stock INT NOT NULL DEFAULT 0,
    lock_stock INT NOT NULL DEFAULT 0,
    sold_count INT NOT NULL DEFAULT 0,
    sku_code VARCHAR(100) ,
    sku_image VARCHAR(500) ,
    sku_status VARCHAR(20) NOT NULL  DEFAULT 'active',
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    create_by BIGINT,
    update_by BIGINT
);

COMMENT ON COLUMN sys_product_sku.sku_id      IS '主键，SKU唯一标识';
COMMENT ON COLUMN sys_product_sku.spu_id      IS '所属SPU ';
COMMENT ON COLUMN sys_product_sku.sku_name    IS 'SKU名称';
COMMENT ON COLUMN sys_product_sku.spec_values IS '规格键值对';
COMMENT ON COLUMN sys_product_sku.price       IS '销售价格（元）';
COMMENT ON COLUMN sys_product_sku.cost_price  IS '成本价格（元）';
COMMENT ON COLUMN sys_product_sku.stock       IS '当前可用库存';
COMMENT ON COLUMN sys_product_sku.lock_stock  IS '锁定库存（下单未支付锁定数）';
COMMENT ON COLUMN sys_product_sku.sold_count  IS '累计销量';
COMMENT ON COLUMN sys_product_sku.sku_code    IS '商家自定义SKU编码';
COMMENT ON COLUMN sys_product_sku.sku_image   IS 'SKU配图URL，可覆盖SPU主图';
COMMENT ON COLUMN sys_product_sku.sku_status  IS 'SKU状态：active/inactive';
COMMENT ON COLUMN sys_product_sku.is_deleted  IS '软删除标记';
COMMENT ON COLUMN sys_product_sku.created_at  IS '创建时间';
COMMENT ON COLUMN sys_product_sku.updated_at  IS '更新时间';
COMMENT ON COLUMN sys_product_sku.create_by   IS '创建人用户ID';
COMMENT ON COLUMN sys_product_sku.update_by   IS '更新人用户ID';

-- 按SPU查所有SKU
CREATE INDEX IF NOT EXISTS idx_sku_spu_id ON sys_product_sku (spu_id);
-- 状态筛选（联合索引）
CREATE INDEX IF NOT EXISTS idx_sku_status ON sys_product_sku (sku_status, is_deleted);

-- SKU状态枚举约束
ALTER TABLE sys_product_sku ADD CONSTRAINT ck_sku_status CHECK (sku_status IN ('active', 'inactive'));

-- 销售价格必须为正数
ALTER TABLE sys_product_sku ADD CONSTRAINT ck_sku_price_positive CHECK (price > 0);

-- 库存全字段非负约束
ALTER TABLE sys_product_sku ADD CONSTRAINT ck_sku_stock_non_negative CHECK (stock >= 0 AND lock_stock >= 0 AND sold_count >= 0);

-- SKU编码唯一（非空时唯一）：通过部分唯一索引实现
CREATE UNIQUE INDEX IF NOT EXISTS idx_sku_code ON sys_product_sku (sku_code) WHERE is_deleted = FALSE;

-- 外键约束：禁止级联删除（SPU下有SKU时，禁止删除SPU）
ALTER TABLE sys_product_sku ADD CONSTRAINT fk_sku_spu_sku FOREIGN KEY (spu_id) REFERENCES sys_product_spu (spu_id) ON DELETE RESTRICT  ON UPDATE CASCADE;

CREATE TABLE  IF NOT EXISTS sys_product_spu_image (
    image_id BIGSERIAL PRIMARY KEY ,
    spu_id BIGINT NOT NULL ,
    image_url VARCHAR(500) NOT NULL ,
    sort_order INT NOT NULL DEFAULT 0,
    is_main BOOLEAN NOT NULL  DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_sys_product_spu_image_spu FOREIGN KEY (spu_id)
        REFERENCES sys_product_spu (spu_id)
        ON DELETE CASCADE
        ON UPDATE CASCADE
);

COMMENT ON COLUMN sys_product_spu_image.image_id   IS '主键，图片唯一标识';
COMMENT ON COLUMN sys_product_spu_image.spu_id     IS '所属SPU ID';
COMMENT ON COLUMN sys_product_spu_image.image_url  IS '图片URL';
COMMENT ON COLUMN sys_product_spu_image.sort_order IS '展示排序，越小越靠前';
COMMENT ON COLUMN sys_product_spu_image.is_main    IS '是否主图';
COMMENT ON COLUMN sys_product_spu_image.created_at IS '创建时间';

-- 按SPU查图片列表
CREATE INDEX IF NOT EXISTS idx_spu_image_spu_id ON sys_product_spu_image (spu_id);
-- 按排序展示（联合索引）
CREATE INDEX IF NOT EXISTS idx_spu_image_sort ON sys_product_spu_image (spu_id, sort_order);

CREATE TABLE IF NOT EXISTS sys_product_stock_log (
   log_id       BIGSERIAL       PRIMARY KEY ,
   sku_id       BIGINT          NOT NULL,
   change_type  VARCHAR(30)     NOT NULL,
   change_qty   INT             NOT NULL,
   before_stock INT             NOT NULL,
   after_stock  INT             NOT NULL,
   before_lock  INT,
   after_lock   INT,
   order_id     BIGINT,
   remark       VARCHAR(500),
   created_at   TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP,
   create_by    INT ,

   CONSTRAINT fk_sys_product_stock_log_sku FOREIGN KEY (sku_id)
       REFERENCES sys_product_sku (sku_id)
       ON DELETE RESTRICT
       ON UPDATE CASCADE
);

COMMENT ON COLUMN sys_product_stock_log.log_id       IS '主键，日志唯一标识';
COMMENT ON COLUMN sys_product_stock_log.sku_id       IS '关联SKU ID';
COMMENT ON COLUMN sys_product_stock_log.change_type  IS '变更类型：order_lock/pay_deduct/order_release/refund_release/manual_adjust';
COMMENT ON COLUMN sys_product_stock_log.change_qty   IS '变更数量（正为增加，负为减少）';
COMMENT ON COLUMN sys_product_stock_log.before_stock IS '变更前可用库存';
COMMENT ON COLUMN sys_product_stock_log.after_stock  IS '变更后可用库存';
COMMENT ON COLUMN sys_product_stock_log.before_lock  IS '变更前锁定库存';
COMMENT ON COLUMN sys_product_stock_log.after_lock   IS '变更后锁定库存';
COMMENT ON COLUMN sys_product_stock_log.order_id     IS '关联订单ID（手动调整时为NULL）';
COMMENT ON COLUMN sys_product_stock_log.remark       IS '备注说明';
COMMENT ON COLUMN sys_product_stock_log.created_at   IS '创建时间';
COMMENT ON COLUMN sys_product_stock_log.create_by    IS '创建人用户ID';

-- 按SKU查日志
CREATE INDEX IF NOT EXISTS idx_stock_log_sku_id ON sys_product_stock_log (sku_id);
-- 按订单查日志
CREATE INDEX IF NOT EXISTS idx_stock_log_order_id ON sys_product_stock_log (order_id);
-- 按时间范围查询
CREATE INDEX IF NOT EXISTS idx_stock_log_created_at ON sys_product_stock_log (created_at);
-- 三元组库存流水唯一索引
CREATE UNIQUE INDEX IF NOT EXISTS uk_stock_log_order ON sys_product_stock_log (order_id, sku_id, change_type) WHERE order_id IS NOT NULL;

-- 库存变更类型枚举约束
ALTER TABLE sys_product_stock_log ADD CONSTRAINT ck_stock_log_change_type CHECK (change_type IN ('order_lock','pay_deduct','order_release','refund_release','manual_adjust'));


WITH
platform_menu AS(
    SELECT menu_id
    FROM sys_menu
    WHERE route_path = '/platform'
    LIMIT 1
),
product_menu AS (
    INSERT  INTO sys_menu(
        parent_id,
        menu_name,
        menu_type,
        icon,
        route_path,
        component,
        is_visible,
        is_cache,
        sort_order,
        meta_info
    )
    SELECT
        p.menu_id,
        '商品管理',
        'M',
        'product',
        '/platform/product',
        'ParentView',
        TRUE,
        TRUE,
        2,
        '{"title":"商品管理","icon":"product","noCache":false}'::jsonb
    FROM platform_menu p
    WHERE NOT EXISTS (SELECT 1 FROM sys_menu WHERE route_path = '/platform/product')
    RETURNING menu_id
),
product_menu_id AS (
    SELECT menu_id FROM product_menu
    UNION ALL
    SELECT menu_id FROM sys_menu WHERE route_path = '/platform/product' LIMIT 1
)
INSERT INTO sys_menu (
    parent_id,
    menu_name,
    menu_type,
    icon,
    route_path,
    component,
    is_visible,
    is_cache,
    sort_order,
    meta_info
)
SELECT
    pr.menu_id,
    '商品列表',
    'M',
    NULL,
    '/platform/product/list',
    'platform/product/list/index',
    TRUE,
    TRUE,
    1,
    '{"title":"商品列表","icon":"","noCache":false}'::jsonb
FROM product_menu pr
WHERE NOT EXISTS (
    SELECT 1
    FROM sys_menu
    WHERE route_path = '/platform/product/list'
);

-- 创建系统权限
INSERT INTO sys_permission (
    permission_code,
    permission_name,
    permission_type,
    request_method,
    api_path,
    description,
    is_system
) VALUES
      (
          'platform:product:create',
          '创建商品',
          'api',
          'POST',
          '/api/v1/platform/products',
          '所属模块：平台商品管理',
          TRUE
      ),
      (
          'platform:product:update',
          '更新商品',
          'api',
          'PUT',
          '/api/v1/platform/product/:id',
          '所属模块：平台商品管理',
          TRUE
      ),
      (
          'platform:product:delete',
          '删除商品',
          'api',
          'DELETE',
          '/api/v1/platform/product/:id',
          '所属模块：平台商品管理',
          TRUE
      ),
      (
          'platform:product:view',
          '查看商品',
          'api',
          'GET',
          '/api/v1/platform/product/:id',
          '所属模块：平台商品管理',
          TRUE
      ),
      (
          'platform:product:publish',
          '上架商品',
          'api',
          'POST',
          '/api/v1/platform/product/publish/:id',
          '所属模块：平台商品管理',
          TRUE
      ),
    (
        'platform:product:withdraw',
        '下架商品',
        'api',
        'POST',
        '/api/v1/platform/product/withdraw/:id',
        '所属模块：平台商品管理',
        TRUE
    )
ON CONFLICT (permission_code, request_method, api_path) DO NOTHING;

-- 菜单权限绑定数据
INSERT INTO sys_menu_permission (
    menu_id,
    permission_id
)
SELECT
    m.menu_id,
    p.permission_id
FROM sys_menu m
         JOIN sys_permission p
              ON p.permission_code = 'platform:product:view'
WHERE m.route_path IN (
                       '/platform/product',
                       '/platform/product/list'
    )
ON CONFLICT (menu_id, permission_id) DO NOTHING;

-- 平台超级管理员：拥有 platform:category:* 对应的全部类目权限
INSERT INTO sys_role_permission (
    role_id,
    permission_id
)
SELECT
    r.role_id,
    p.permission_id
FROM sys_role r
         JOIN sys_permission p
              ON p.permission_code IN (
                                       'platform:product:create',
                                       'platform:product:update',
                                       'platform:product:delete',
                                       'platform:product:view',
                                       'platform:product:publish',
                                       'platform:product:withdraw'
                  )
WHERE r.role_name = '平台超级管理员'
ON CONFLICT (role_id, permission_id) DO NOTHING;


-- 平台运营人员：查看、类目树、子类目、更新
INSERT INTO sys_role_permission (
    role_id,
    permission_id
)
SELECT
    r.role_id,
    p.permission_id
FROM sys_role r
         JOIN sys_permission p
              ON p.permission_code IN (
                                       'platform:product:view',
                                       'platform:product:publish',
                                       'platform:product:withdraw',
                                       'platform:product:update'
                  )
WHERE r.role_name = '平台运营人员'
ON CONFLICT (role_id, permission_id) DO NOTHING;


-- 平台审核人员：查看、类目树、子类目
INSERT INTO sys_role_permission (
    role_id,
    permission_id
)
SELECT
    r.role_id,
    p.permission_id
FROM sys_role r
         JOIN sys_permission p
              ON p.permission_code IN (
                                       'platform:product:view'
                  )
WHERE r.role_name = '平台审核人员'
ON CONFLICT (role_id, permission_id) DO NOTHING;


-- 绑定按钮权限
WITH category_list_menu AS (
    SELECT menu_id
    FROM sys_menu
    WHERE route_path = '/platform/product/list'
    LIMIT 1
),
     button_data AS (
         SELECT *
         FROM (
                  VALUES
                      ('新增商品', 'F', 1, '{"title":"新增商品"}'::jsonb),
                      ('编辑',     'F', 2, '{"title":"编辑"}'::jsonb),
                      ('删除',     'F', 3, '{"title":"删除"}'::jsonb),
                      ('查看详情', 'F', 4, '{"title":"查看详情"}'::jsonb),
                      ('上架',     'F', 5, '{"title":"上架"}'::jsonb),
                      ('下架',     'F', 6, '{"title":"下架"}'::jsonb)
              ) AS t(menu_name, menu_type, sort_order, meta_info)
     )
INSERT INTO sys_menu (
    parent_id,
    menu_name,
    menu_type,
    icon,
    route_path,
    component,
    is_visible,
    is_cache,
    sort_order,
    meta_info
)
SELECT
    c.menu_id,
    b.menu_name,
    b.menu_type,
    NULL,
    NULL,
    NULL,
    FALSE,
    FALSE,
    b.sort_order,
    b.meta_info
FROM category_list_menu c
         CROSS JOIN button_data b
WHERE NOT EXISTS (
    SELECT 1
    FROM sys_menu m
    WHERE m.parent_id = c.menu_id
      AND m.menu_name = b.menu_name
      AND m.menu_type = 'F'
);


-- 绑定按钮菜单和权限编码
WITH category_list_menu AS (
    SELECT menu_id
    FROM sys_menu
    WHERE route_path = '/platform/product/list'
    LIMIT 1
),
     button_permission_map AS (
         SELECT *
         FROM (
                  VALUES
                      ('新增类目', 'platform:product:create'),
                      ('编辑',     'platform:product:update'),
                      ('删除',     'platform:product:delete'),
                      ('查看详情', 'platform:product:view'),
                      ('上架', 'platform:product:withdraw'),
                      ('下架', 'platform:product:update')
              ) AS t(menu_name, permission_code)
     )
INSERT INTO sys_menu_permission (
    menu_id,
    permission_id
)
SELECT
    m.menu_id,
    p.permission_id
FROM category_list_menu c
         JOIN sys_menu m
              ON m.parent_id = c.menu_id
         JOIN button_permission_map bpm
              ON bpm.menu_name = m.menu_name
         JOIN sys_permission p
              ON p.permission_code = bpm.permission_code
WHERE m.menu_type = 'F'
ON CONFLICT (menu_id, permission_id) DO NOTHING;

-- ====================== 角色-菜单关联 ======================
-- 平台超级管理员：拥有所有平台管理相关菜单
WITH platform_menus AS (
    SELECT menu_id FROM sys_menu WHERE route_path LIKE '/platform%'
)
INSERT INTO sys_role_menu (role_id, menu_id)
SELECT r.role_id, m.menu_id
FROM sys_role r
         CROSS JOIN platform_menus m
WHERE r.role_name = '平台超级管理员'
ON CONFLICT (role_id, menu_id) DO NOTHING;

-- 平台运营人员：拥有平台管理和商品管理菜单
WITH allowed_menus AS (
    SELECT menu_id FROM sys_menu
    WHERE route_path IN ('/platform', '/platform/product', '/platform/product/list')
)
INSERT INTO sys_role_menu (role_id, menu_id)
SELECT r.role_id, m.menu_id
FROM sys_role r
         CROSS JOIN allowed_menus m
WHERE r.role_name = '平台运营人员'
ON CONFLICT (role_id, menu_id) DO NOTHING;

-- 平台审核人员：拥有平台管理和商品管理菜单（和运营人员相同）
WITH allowed_menus AS (
    SELECT menu_id FROM sys_menu
    WHERE route_path IN ('/platform', '/platform/product', '/platform/product/list')
)
INSERT INTO sys_role_menu (role_id, menu_id)
SELECT r.role_id, m.menu_id
FROM sys_role r
         CROSS JOIN allowed_menus m
WHERE r.role_name = '平台审核人员'
ON CONFLICT (role_id, menu_id) DO NOTHING;





