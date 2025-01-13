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