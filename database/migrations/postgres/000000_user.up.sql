-- 添加 uuid-ossp 扩展
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- 用户表
-- postgres
CREATE TABLE "user" (
  "id" uuid DEFAULT uuid_generate_v4() PRIMARY KEY,
  "username" varchar(255) NULL DEFAULT NULL,
  "password" varchar(255) NULL DEFAULT NULL,
  "email" VARCHAR(255) NULL DEFAULT NULL,
  "nickname" varchar(255) NULL DEFAULT NULL,
  "gender" VARCHAR(50) NULL DEFAULT NULL,
  "avatar" varchar(255) NULL DEFAULT NULL,
  "status" INT DEFAULT 1,
  "created_at" TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP,
  "deleted_at" TIMESTAMP NULL
);


COMMENT ON TABLE "user" IS '用户表';
COMMENT ON COLUMN "user"."id" IS '主键ID';
COMMENT ON COLUMN "user"."username" IS '用户名';
COMMENT ON COLUMN "user"."password" IS '密码';
COMMENT ON COLUMN "user"."email" IS '邮箱';
COMMENT ON COLUMN "user"."nickname" IS '昵称';
COMMENT ON COLUMN "user"."gender" IS '性别';
COMMENT ON COLUMN "user"."avatar" IS '头像';
COMMENT ON COLUMN "user"."status" IS '状态, 1: 正常, 2: 禁用';
COMMENT ON COLUMN "user"."created_at" IS '创建时间';
COMMENT ON COLUMN "user"."updated_at" IS '修改时间';
COMMENT ON COLUMN "user"."deleted_at" IS '删除时间';

-- 账本表
CREATE TABLE "ledger" (
  "id" uuid DEFAULT uuid_generate_v4() PRIMARY KEY,
  "name" VARCHAR(100) NOT NULL,
  "description" TEXT,
  "owner_id" UUID REFERENCES users(id) ON DELETE CASCADE,
  "created_at" TIMESTAMPTZ DEFAULT NOW(),
  "updated_at" TIMESTAMPTZ DEFAULT NOW(),
  "deleted_at" TIMESTAMP NULL
);

COMMENT ON TABLE "ledger" IS '账本表';
COMMENT ON COLUMN "ledger"."id" IS '主键ID';
COMMENT ON COLUMN "ledger"."name" IS '账本名称';
COMMENT ON COLUMN "ledger"."description" IS '账本描述';
COMMENT ON COLUMN "ledger"."owner_id" IS '账本拥有者';
COMMENT ON COLUMN "ledger"."created_at" IS '创建时间';
COMMENT ON COLUMN "ledger"."updated_at" IS '修改时间';
COMMENT ON COLUMN "ledger"."deleted_at" IS '删除时间';

-- 账单表
CREATE TABLE "transaction" (
  "id" uuid DEFAULT uuid_generate_v4() PRIMARY KEY,
  "ledger_id" UUID REFERENCES ledger(id) ON DELETE CASCADE,
  "user_id" UUID REFERENCES user(id) ON DELETE SET NULL,
  "category_id" UUID REFERENCES category(id) ON DELETE SET NULL,
  "subcategory_id" UUID REFERENCES subcategory(id) ON DELETE SET NULL,
  "amount" NUMERIC(10, 2) NOT NULL,
  "type" TEXT CHECK (type IN ('income', 'expense')),
  "asset_id" UUID REFERENCES asset(id) ON DELETE SET NULL,
  "note" TEXT,
  "transaction_date" TIMESTAMPTZ NOT NULL,
  "created_at" TIMESTAMPTZ DEFAULT NOW(),
  "updated_at" TIMESTAMPTZ DEFAULT NOW(),
  "deleted_at" TIMESTAMP NULL
);

COMMENT ON TABLE "transaction" IS '账单表';
COMMENT ON COLUMN "transaction"."id" IS '主键ID';
COMMENT ON COLUMN "transaction"."ledger_id" IS '账本ID';
COMMENT ON COLUMN "transaction"."user_id" IS '用户ID';
COMMENT ON COLUMN "transaction"."category_id" IS '类目ID';
COMMENT ON COLUMN "transaction"."subcategory_id" IS '子类目ID';
COMMENT ON COLUMN "transaction"."amount" IS '金额';

-- 资产表
CREATE TABLE "asset" (
  "id" uuid DEFAULT uuid_generate_v4() PRIMARY KEY,
  "user_id" UUID REFERENCES user(id) ON DELETE CASCADE,
  "name" VARCHAR(50) NOT NULL,
  "type" TEXT CHECK (type IN ('cash', 'bank', 'wallet')),
  "balance" NUMERIC(10, 2) DEFAULT 0,
  "created_at" TIMESTAMPTZ DEFAULT NOW(),
  "updated_at" TIMESTAMPTZ DEFAULT NOW(),
  "deleted_at" TIMESTAMP NULL
);

COMMENT ON TABLE "asset" IS '资产表';
COMMENT ON COLUMN "asset"."id" IS '主键ID';
COMMENT ON COLUMN "asset"."user_id" IS '用户ID';
COMMENT ON COLUMN "asset"."name" IS '资产名称';
COMMENT ON COLUMN "asset"."type" IS '资产类型';
COMMENT ON COLUMN "asset"."balance" IS '余额';