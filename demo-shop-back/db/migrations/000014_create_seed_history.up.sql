-- 种子账本:记录已应用的种子文件版本(与迁移的 schema_migrations 对称)
-- 由 db.RunSeeds 维护:每个种子文件执行成功后写入一行;存在记录即跳过
CREATE TABLE IF NOT EXISTS sys_seed_history (
    version    VARCHAR(100) PRIMARY KEY,
    applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
