-- 类目表
CREATE TABLE "category" (
  "id" uuid DEFAULT uuid_generate_v4() PRIMARY KEY,
  "name" VARCHAR(50) NOT NULL,
  "user_id" UUID REFERENCES "user"(id),
  "type" VARCHAR(50) CHECK (type IN ('income', 'expense', 'transfer')),
  "is_system" BOOLEAN NOT NULL DEFAULT false,
  "created_at" TIMESTAMPTZ DEFAULT NOW(),
  "updated_at" TIMESTAMPTZ DEFAULT NOW(),
  "deleted_at" TIMESTAMP NULL,
  UNIQUE(name, user_id, type) -- 同一用户下类目名称唯一
);

COMMENT ON TABLE "category" IS '类目表';
COMMENT ON COLUMN "category"."id" IS '主键ID';
COMMENT ON COLUMN "category"."name" IS '类目名称';
COMMENT ON COLUMN "category"."user_id" IS '创建者ID（为空表示系统默认类目）';
COMMENT ON COLUMN "category"."type" IS '类型：收入、支出、转账';
COMMENT ON COLUMN "category"."is_system" IS '是否为系统默认类目';
COMMENT ON COLUMN "category"."created_at" IS '创建时间';
COMMENT ON COLUMN "category"."updated_at" IS '修改时间';
COMMENT ON COLUMN "category"."deleted_at" IS '删除时间';

-- 子类目表
CREATE TABLE "subcategory" (
  "id" uuid DEFAULT uuid_generate_v4() PRIMARY KEY,
  "category_id" UUID REFERENCES "category"(id) ON DELETE CASCADE,
  "name" VARCHAR(50) NOT NULL,
  "user_id" UUID REFERENCES "user"(id),
  "is_system" BOOLEAN NOT NULL DEFAULT false,
  "created_at" TIMESTAMPTZ DEFAULT NOW(),
  "updated_at" TIMESTAMPTZ DEFAULT NOW(),
  "deleted_at" TIMESTAMP NULL,
  UNIQUE(name, category_id, user_id) -- 同一用户同一类目下子类目名称唯一
);

COMMENT ON TABLE "subcategory" IS '子类目表';
COMMENT ON COLUMN "subcategory"."id" IS '主键ID';
COMMENT ON COLUMN "subcategory"."category_id" IS '类目ID';
COMMENT ON COLUMN "subcategory"."name" IS '子类目名称';
COMMENT ON COLUMN "subcategory"."user_id" IS '创建者ID（为空表示系统默认子类目）';
COMMENT ON COLUMN "subcategory"."is_system" IS '是否为系统默认子类目';
COMMENT ON COLUMN "subcategory"."created_at" IS '创建时间';
COMMENT ON COLUMN "subcategory"."updated_at" IS '修改时间';
COMMENT ON COLUMN "subcategory"."deleted_at" IS '删除时间';


-- 创建类目，子类目数据

-- 支出类目
-- 餐饮：三餐，蔬菜，外卖，酸奶
-- 购物：日用品，服饰，电子产品，书籍
-- 零食：零食，饮料，小吃，奶茶
-- 水果：
-- 服饰：衣服，裤子，鞋子，包包，饰品
-- 日用：
-- 数码：
-- 美妆：面霜，口红，洗面奶，粉底，头饰，眼影，护肤品，面膜，香水，SPA
-- 护肤：
-- 应用软件：
-- 住房：房租，物业，水电，燃气，宽带
-- 交通：公交，地铁，火车，高铁，飞机，打车，共享单车，停车，加油
-- 娱乐：电影，音乐，游戏，旅游，健身，聚会，KTV
-- 医疗：门诊，住院，疫苗，体检，药品
-- 通讯：话费，流量，宽带，电视
-- 汽车：加油，保养，维修，保险，停车
-- 学习：书籍，课程，教育
-- 办公：
-- 运动：
-- 恋爱：
-- 社交：亲友，宠物，孩子，老人
-- 人情：家人，礼金，礼物，捐赠
-- 育儿：
-- 宠物：
-- 旅行：
-- 度假：
-- 烟酒：
-- 彩票：
-- 其他

-- 收入类目
-- 工资：
-- 奖金：
-- 理财：
-- 投资：
-- 报销：
-- 副业：
-- 红包：
-- 退税：
-- 兼职：
-- 其他：


INSERT INTO "category" ("name", "user_id", "type", "is_system") VALUES 
    ('餐饮', NULL, 'expense', true),
    ('购物', NULL, 'expense', true),
    ('零食', NULL, 'expense', true),
    ('水果', NULL, 'expense', true),
    ('服饰', NULL, 'expense', true),
    ('日用', NULL, 'expense', true),
    ('数码', NULL, 'expense', true),
    ('美妆', NULL, 'expense', true),
    ('护肤', NULL, 'expense', true),
    ('应用软件', NULL, 'expense', true),
    ('住房', NULL, 'expense', true),
    ('交通', NULL, 'expense', true),
    ('娱乐', NULL, 'expense', true),
    ('医疗', NULL, 'expense', true),
    ('通讯', NULL, 'expense', true),
    ('汽车', NULL, 'expense', true),
    ('学习', NULL, 'expense', true),
    ('办公', NULL, 'expense', true),
    ('运动', NULL, 'expense', true),
    ('恋爱', NULL, 'expense', true),
    ('社交', NULL, 'expense', true),
    ('人情', NULL, 'expense', true),
    ('育儿', NULL, 'expense', true),
    ('宠物', NULL, 'expense', true),
    ('旅行', NULL, 'expense', true),
    ('度假', NULL, 'expense', true),
    ('烟酒', NULL, 'expense', true),
    ('彩票', NULL, 'expense', true),
    ('其他', NULL, 'expense', true);
INSERT INTO "subcategory" ("category_id","name", "user_id", "is_system") VALUES 
    ((SELECT id FROM "category" WHERE name = '餐饮' AND user_id IS NULL), '三餐', NULL, true),
    ((SELECT id FROM "category" WHERE name = '餐饮' AND user_id IS NULL), '蔬菜', NULL, true),
    ((SELECT id FROM "category" WHERE name = '餐饮' AND user_id IS NULL), '饮料', NULL, true),
    ((SELECT id FROM "category" WHERE name = '餐饮' AND user_id IS NULL), '酸奶', NULL, true),
    ((SELECT id FROM "category" WHERE name = '餐饮' AND user_id IS NULL), '外卖', NULL, true),

    ((SELECT id FROM "category" WHERE name = '购物' AND user_id IS NULL), '日用品', NULL, true),
    ((SELECT id FROM "category" WHERE name = '购物' AND user_id IS NULL), '服饰', NULL, true),
    ((SELECT id FROM "category" WHERE name = '购物' AND user_id IS NULL), '电子产品', NULL, true),
    ((SELECT id FROM "category" WHERE name = '购物' AND user_id IS NULL), '书籍', NULL, true),

    ((SELECT id FROM "category" WHERE name = '零食' AND user_id IS NULL), '零食', NULL, true),
    ((SELECT id FROM "category" WHERE name = '零食' AND user_id IS NULL), '饮料', NULL, true),
    ((SELECT id FROM "category" WHERE name = '零食' AND user_id IS NULL), '小吃', NULL, true),
    ((SELECT id FROM "category" WHERE name = '零食' AND user_id IS NULL), '奶茶', NULL, true),

    ((SELECT id FROM "category" WHERE name = '水果' AND user_id IS NULL), '苹果', NULL, true),
    ((SELECT id FROM "category" WHERE name = '水果' AND user_id IS NULL), '香蕉', NULL, true),
    ((SELECT id FROM "category" WHERE name = '水果' AND user_id IS NULL), '橙子', NULL, true),
    ((SELECT id FROM "category" WHERE name = '水果' AND user_id IS NULL), '橘子', NULL, true),

    ((SELECT id FROM "category" WHERE name = '服饰' AND user_id IS NULL), '衣服', NULL, true),
    ((SELECT id FROM "category" WHERE name = '服饰' AND user_id IS NULL), '裤子', NULL, true),
    ((SELECT id FROM "category" WHERE name = '服饰' AND user_id IS NULL), '鞋子', NULL, true),
    ((SELECT id FROM "category" WHERE name = '服饰' AND user_id IS NULL), '包包', NULL, true),
    ((SELECT id FROM "category" WHERE name = '服饰' AND user_id IS NULL), '饰品', NULL, true),

    ((SELECT id FROM "category" WHERE name = '日用' AND user_id IS NULL), '洗衣液', NULL, true),
    ((SELECT id FROM "category" WHERE name = '日用' AND user_id IS NULL), '洗发水', NULL, true),
    ((SELECT id FROM "category" WHERE name = '日用' AND user_id IS NULL), '沐浴露', NULL, true),
    ((SELECT id FROM "category" WHERE name = '日用' AND user_id IS NULL), '牙膏', NULL, true),
    ((SELECT id FROM "category" WHERE name = '日用' AND user_id IS NULL), '纸巾', NULL, true),

    ((SELECT id FROM "category" WHERE name = '数码' AND user_id IS NULL), '手机', NULL, true),
    ((SELECT id FROM "category" WHERE name = '数码' AND user_id IS NULL), '电脑', NULL, true),
    ((SELECT id FROM "category" WHERE name = '数码' AND user_id IS NULL), '平板', NULL, true),
    ((SELECT id FROM "category" WHERE name = '数码' AND user_id IS NULL), '耳机', NULL, true),

    ((SELECT id FROM "category" WHERE name = '美妆' AND user_id IS NULL), '面霜', NULL, true),
    ((SELECT id FROM "category" WHERE name = '美妆' AND user_id IS NULL), '口红', NULL, true),
    ((SELECT id FROM "category" WHERE name = '美妆' AND user_id IS NULL), '洗面奶', NULL, true),
    ((SELECT id FROM "category" WHERE name = '美妆' AND user_id IS NULL), '粉底', NULL, true),
    ((SELECT id FROM "category" WHERE name = '美妆' AND user_id IS NULL), '头饰', NULL, true),
    ((SELECT id FROM "category" WHERE name = '美妆' AND user_id IS NULL), '眼影', NULL, true),
    ((SELECT id FROM "category" WHERE name = '美妆' AND user_id IS NULL), '护肤品', NULL, true),
    ((SELECT id FROM "category" WHERE name = '美妆' AND user_id IS NULL), '面膜', NULL, true),
    ((SELECT id FROM "category" WHERE name = '美妆' AND user_id IS NULL), '香水', NULL, true),
    ((SELECT id FROM "category" WHERE name = '美妆' AND user_id IS NULL), 'SPA', NULL, true),


    ((SELECT id FROM "category" WHERE name = '住房' AND user_id IS NULL), '房租', NULL, true),
    ((SELECT id FROM "category" WHERE name = '住房' AND user_id IS NULL), '物业', NULL, true),
    ((SELECT id FROM "category" WHERE name = '住房' AND user_id IS NULL), '水电', NULL, true),
    ((SELECT id FROM "category" WHERE name = '住房' AND user_id IS NULL), '燃气', NULL, true),
    ((SELECT id FROM "category" WHERE name = '住房' AND user_id IS NULL), '宽带', NULL, true),

    ((SELECT id FROM "category" WHERE name = '交通' AND user_id IS NULL), '公交', NULL, true),
    ((SELECT id FROM "category" WHERE name = '交通' AND user_id IS NULL), '地铁', NULL, true),
    ((SELECT id FROM "category" WHERE name = '交通' AND user_id IS NULL), '火车', NULL, true),
    ((SELECT id FROM "category" WHERE name = '交通' AND user_id IS NULL), '高铁', NULL, true),
    ((SELECT id FROM "category" WHERE name = '交通' AND user_id IS NULL), '飞机', NULL, true),
    ((SELECT id FROM "category" WHERE name = '交通' AND user_id IS NULL), '打车', NULL, true),
    ((SELECT id FROM "category" WHERE name = '交通' AND user_id IS NULL), '共享单车', NULL, true),
    ((SELECT id FROM "category" WHERE name = '交通' AND user_id IS NULL), '停车', NULL, true),
    ((SELECT id FROM "category" WHERE name = '交通' AND user_id IS NULL), '加油', NULL, true),


    ((SELECT id FROM "category" WHERE name = '娱乐' AND user_id IS NULL), '电影', NULL, true),
    ((SELECT id FROM "category" WHERE name = '娱乐' AND user_id IS NULL), '音乐', NULL, true),
    ((SELECT id FROM "category" WHERE name = '娱乐' AND user_id IS NULL), '游戏', NULL, true),
    ((SELECT id FROM "category" WHERE name = '娱乐' AND user_id IS NULL), '旅游', NULL, true),
    ((SELECT id FROM "category" WHERE name = '娱乐' AND user_id IS NULL), '健身', NULL, true),
    ((SELECT id FROM "category" WHERE name = '娱乐' AND user_id IS NULL), '聚会', NULL, true),
    ((SELECT id FROM "category" WHERE name = '娱乐' AND user_id IS NULL), 'KTV', NULL, true),

    ((SELECT id FROM "category" WHERE name = '医疗' AND user_id IS NULL), '门诊', NULL, true),
    ((SELECT id FROM "category" WHERE name = '医疗' AND user_id IS NULL), '住院', NULL, true),
    ((SELECT id FROM "category" WHERE name = '医疗' AND user_id IS NULL), '疫苗', NULL, true),
    ((SELECT id FROM "category" WHERE name = '医疗' AND user_id IS NULL), '体检', NULL, true),
    ((SELECT id FROM "category" WHERE name = '医疗' AND user_id IS NULL), '药品', NULL, true),

    ((SELECT id FROM "category" WHERE name = '通讯' AND user_id IS NULL), '话费', NULL, true),
    ((SELECT id FROM "category" WHERE name = '通讯' AND user_id IS NULL), '流量', NULL, true),
    ((SELECT id FROM "category" WHERE name = '通讯' AND user_id IS NULL), '宽带', NULL, true),
    ((SELECT id FROM "category" WHERE name = '通讯' AND user_id IS NULL), '电视', NULL, true),

    ((SELECT id FROM "category" WHERE name = '汽车' AND user_id IS NULL), '加油', NULL, true),
    ((SELECT id FROM "category" WHERE name = '汽车' AND user_id IS NULL), '保养', NULL, true),
    ((SELECT id FROM "category" WHERE name = '汽车' AND user_id IS NULL), '维修', NULL, true),
    ((SELECT id FROM "category" WHERE name = '汽车' AND user_id IS NULL), '保险', NULL, true),
    ((SELECT id FROM "category" WHERE name = '汽车' AND user_id IS NULL), '停车', NULL, true),

    ((SELECT id FROM "category" WHERE name = '学习' AND user_id IS NULL), '书籍', NULL, true),
    ((SELECT id FROM "category" WHERE name = '学习' AND user_id IS NULL), '课程', NULL, true),
    ((SELECT id FROM "category" WHERE name = '学习' AND user_id IS NULL), '软件', NULL, true),

    ((SELECT id FROM "category" WHERE name = '社交' AND user_id IS NULL), '亲友', NULL, true),
    ((SELECT id FROM "category" WHERE name = '社交' AND user_id IS NULL), '宠物', NULL, true),
    ((SELECT id FROM "category" WHERE name = '社交' AND user_id IS NULL), '孩子', NULL, true),
    ((SELECT id FROM "category" WHERE name = '社交' AND user_id IS NULL), '老人', NULL, true),

    ((SELECT id FROM "category" WHERE name = '人情' AND user_id IS NULL), '亲友', NULL, true),
    ((SELECT id FROM "category" WHERE name = '人情' AND user_id IS NULL), '礼金', NULL, true),
    ((SELECT id FROM "category" WHERE name = '人情' AND user_id IS NULL), '礼物', NULL, true),
    ((SELECT id FROM "category" WHERE name = '人情' AND user_id IS NULL), '捐赠', NULL, true);


-- 收入类目
INSERT INTO "category" ("name", "user_id", "type", "is_system") VALUES
    ('工资', NULL, 'income', true),
    ('奖金', NULL, 'income', true),
    ('理财', NULL, 'income', true),
    ('投资', NULL, 'income', true),
    ('报销', NULL, 'income', true),
    ('副业', NULL, 'income', true),
    ('红包', NULL, 'income', true),
    ('退税', NULL, 'income', true),
    ('兼职', NULL, 'income', true),
    ('其他', NULL, 'income', true);
