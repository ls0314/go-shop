-- 000002 down:回滚订单三表(先删子表,再删主表)
DROP TABLE IF EXISTS user_order_log;
DROP TABLE IF EXISTS user_order_detail;
DROP TABLE IF EXISTS user_order_master;
