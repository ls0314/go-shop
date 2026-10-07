-- 建出四个服务各自的库。
--
-- ============================================================
-- 为什么需要这个文件
-- ============================================================
--
-- 各服务的 migrations/ 里**只建表,不建库** —— 原先是单体 db/migrate.go
-- 负责建库,拆分后没有接替者,四个库一直是开发期手工建的。
--
-- 于是全新环境(空数据卷)下,迁移容器会因"库不存在"失败,而它又是
-- 各服务 depends_on: service_completed_successfully 的前置 —— 整条启动链
-- 会卡住,且报错出现在 migrate 容器里,不指向"缺库"这个真实原因。
--
-- ============================================================
-- 执行时机(重要)
-- ============================================================
--
-- 本目录挂到 postgres 容器的 /docker-entrypoint-initdb.d/,**仅在数据卷为空时
-- 执行一次**。所以:
--   - 全新环境:自动建库,启动链能跑通;
--   - 已有数据卷:本脚本**不会执行** —— 那种情况库早已存在(手工建的),
--     若缺库请手工 CREATE DATABASE 或删卷重建。
--
-- ============================================================
-- 为什么用 \gexec 而不是 CREATE DATABASE IF NOT EXISTS
-- ============================================================
--
-- PostgreSQL 的 CREATE DATABASE **没有** IF NOT EXISTS 语法,重复执行会报错。
-- 这里先 SELECT 出一个建库语句、再用 \gexec 执行它 —— 库已存在时 SELECT
-- 返回 0 行,\gexec 无事可做,于是整个脚本幂等。
--
-- \gexec 是 psql 元命令,而这个目录里的脚本由 postgres 镜像用 psql 执行,
-- 所以可用。

\set ON_ERROR_STOP on

SELECT 'CREATE DATABASE user_db'
 WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'user_db')\gexec

SELECT 'CREATE DATABASE product_db'
 WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'product_db')\gexec

SELECT 'CREATE DATABASE trade_db'
 WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'trade_db')\gexec

SELECT 'CREATE DATABASE marketing_db'
 WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'marketing_db')\gexec
