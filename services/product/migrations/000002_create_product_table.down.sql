-- ============================================================
-- 000002 down:回滚商品域四张表
-- 按外键依赖逆序删除(先子表后父表)
-- ============================================================

BEGIN;

-- 库存流水 → 依赖 sys_product_sku
DROP TABLE IF EXISTS sys_product_stock_log;

-- SPU 图片 → 依赖 sys_product_spu
DROP TABLE IF EXISTS sys_product_spu_image;

-- SKU → 依赖 sys_product_spu
DROP TABLE IF EXISTS sys_product_sku;

-- SPU → 依赖 sys_category
DROP TABLE IF EXISTS sys_product_spu;

COMMIT;
