-- 创建用户认证表 (高频查询，需要分表)
CREATE TABLE IF NOT EXISTS sys_user (
    user_id uuid PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    email VARCHAR(100) UNIQUE,
    phone VARCHAR(20) UNIQUE,
    status VARCHAR(20) DEFAULT 'active',
    last_login_time TIMESTAMP NULL,
    last_login_ip VARCHAR(45),
    failed_attempts INT DEFAULT 0,
    lock_until TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 添加列注释
COMMENT ON COLUMN sys_user.user_id IS '用户ID';
COMMENT ON COLUMN sys_user.username IS '用户名';
COMMENT ON COLUMN sys_user.password_hash IS '用户密码';
COMMENT ON COLUMN sys_user.email IS '邮箱';
COMMENT ON COLUMN sys_user.phone IS '手机号';
COMMENT ON COLUMN sys_user.status IS '状态';
COMMENT ON COLUMN sys_user.last_login_time IS '最后登录时间';
COMMENT ON COLUMN sys_user.last_login_ip IS '最后登录IP';
COMMENT ON COLUMN sys_user.failed_attempts IS '连续失败次数';
COMMENT ON COLUMN sys_user.lock_until IS '锁定截止时间';
COMMENT ON COLUMN sys_user.created_at IS '创建时间';
COMMENT ON COLUMN sys_user.updated_at IS '更新时间';

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_username ON sys_user(username);
CREATE INDEX IF NOT EXISTS idx_email ON sys_user(email);
CREATE INDEX IF NOT EXISTS idx_phone ON sys_user(phone);
CREATE INDEX IF NOT EXISTS idx_status_created ON sys_user(status, created_at);

-- 创建用户信息表
CREATE TABLE IF NOT EXISTS user_profile (
    user_info_id uuid PRIMARY KEY,
    nickname VARCHAR(50),
    real_name VARCHAR(50),
    avatar_url VARCHAR(500),
    gender VARCHAR(20) DEFAULT 'unknown',
    birthdate DATE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 添加列注释
COMMENT ON COLUMN user_profile.user_info_id IS '用户ID';
COMMENT ON COLUMN user_profile.nickname IS '昵称';
COMMENT ON COLUMN user_profile.real_name IS '真实姓名';
COMMENT ON COLUMN user_profile.avatar_url IS '头像URL';
COMMENT ON COLUMN user_profile.gender IS '性别';
COMMENT ON COLUMN user_profile.birthdate IS '出生日期';
COMMENT ON COLUMN user_profile.created_at IS '创建时间';
COMMENT ON COLUMN user_profile.updated_at IS '更新时间';

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_real_name ON user_profile(real_name);
CREATE INDEX IF NOT EXISTS idx_created ON user_profile(created_at);

-- 创建登录日志表
CREATE TABLE IF NOT EXISTS user_login_log (
    id BIGSERIAL PRIMARY KEY,
    user_id uuid NOT NULL,
    login_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    login_ip VARCHAR(45),
    login_device VARCHAR(200),
    login_type VARCHAR(20) DEFAULT 'password',
    login_status VARCHAR(20) DEFAULT 'success',
    failure_reason VARCHAR(200),
    location VARCHAR(200),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 添加列注释
COMMENT ON COLUMN user_login_log.id IS '日志ID';
COMMENT ON COLUMN user_login_log.user_id IS '用户ID';
COMMENT ON COLUMN user_login_log.login_time IS '登录时间';
COMMENT ON COLUMN user_login_log.login_ip IS '登录IP';
COMMENT ON COLUMN user_login_log.login_device IS '登录设备';
COMMENT ON COLUMN user_login_log.login_type IS '登录类型';
COMMENT ON COLUMN user_login_log.login_status IS '登录状态';
COMMENT ON COLUMN user_login_log.failure_reason IS '失败原因';
COMMENT ON COLUMN user_login_log.location IS '登录地点';
COMMENT ON COLUMN user_login_log.created_at IS '创建时间';

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_user_id ON user_login_log(user_id);
CREATE INDEX IF NOT EXISTS idx_login_time ON user_login_log(login_time);
CREATE INDEX IF NOT EXISTS idx_login_status ON user_login_log(login_status);
CREATE INDEX IF NOT EXISTS idx_user_time ON user_login_log(user_id, login_time);
