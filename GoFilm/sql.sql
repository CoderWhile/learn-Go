-- ==========================================
-- GoFilm 影院管理系统 - 数据库初始化脚本
-- ==========================================

-- 创建数据库
CREATE DATABASE IF NOT EXISTS theater_data DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci;
USE theater_data;

-- ==========================================
-- 用户表
-- ==========================================
CREATE TABLE IF NOT EXISTS `users` (
    `id` int NOT NULL AUTO_INCREMENT,
    `username` varchar(100) NOT NULL,
    `password` varchar(100) NOT NULL,
    `identity` varchar(100) DEFAULT NULL COMMENT '角色: customer/admin',
    PRIMARY KEY (`id`),
    UNIQUE KEY `username` (`username`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- ==========================================
-- 会话表 (JWT 模式下可逐步废弃)
-- ==========================================
CREATE TABLE IF NOT EXISTS `sessions` (
    `session_id` varchar(100) NOT NULL,
    `username` varchar(100) NOT NULL,
    `user_id` int NOT NULL,
    `user_identity` varchar(100) DEFAULT NULL,
    PRIMARY KEY (`session_id`),
    KEY `user_id` (`user_id`),
    FOREIGN KEY (`user_id`) REFERENCES `users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- ==========================================
-- 电影表
-- ==========================================
CREATE TABLE IF NOT EXISTS `movies` (
    `id` int NOT NULL AUTO_INCREMENT,
    `title` varchar(200) NOT NULL COMMENT '电影名称',
    `genre` varchar(100) NOT NULL COMMENT '类型',
    `area` varchar(50) NOT NULL COMMENT '地区',
    `intro` text COMMENT '简介',
    `image_path` varchar(500) DEFAULT NULL COMMENT '海报地址',
    `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    INDEX `idx_genre` (`genre`),
    INDEX `idx_area` (`area`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- 补全电影表的评分、时长、状态字段（已有表执行，新建表跳过）
ALTER TABLE `movies`
    ADD COLUMN IF NOT EXISTS `rating`      DOUBLE(3,1) DEFAULT NULL COMMENT '评分',
    ADD COLUMN IF NOT EXISTS `duration`    INT          DEFAULT NULL COMMENT '时长(分钟)',
    ADD COLUMN IF NOT EXISTS `status`      VARCHAR(20)  DEFAULT '上映中' COMMENT '未上映/上映中/下架',
    ADD COLUMN IF NOT EXISTS `category_id` INT          DEFAULT NULL COMMENT '分类ID';

-- ==========================================
-- 分类表
-- ==========================================
CREATE TABLE IF NOT EXISTS `categories` (
    `id`   INT AUTO_INCREMENT PRIMARY KEY,
    `name` VARCHAR(50) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

INSERT INTO categories (id, name) VALUES
(1,'动作'),(2,'喜剧'),(3,'爱情'),(4,'动画'),(5,'剧情'),(6,'悬疑'),(7,'惊悚'),(8,'科幻'),(9,'战争'),(10,'冒险'),(11,'警匪'),(12,'家庭'),(13,'音乐')
ON DUPLICATE KEY UPDATE name=VALUES(name);

-- ==========================================
-- 标签表
-- ==========================================
CREATE TABLE IF NOT EXISTS `tags` (
    `id`   INT AUTO_INCREMENT PRIMARY KEY,
    `name` VARCHAR(50) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

INSERT INTO tags (id, name) VALUES
(1,'热门'),(2,'高分'),(3,'经典'),(4,'院线'),(5,'国产'),(6,'进口'),(7,'IMAX'),(8,'3D')
ON DUPLICATE KEY UPDATE name=VALUES(name);

-- ==========================================
-- 电影-标签关联表
-- ==========================================
CREATE TABLE IF NOT EXISTS `movie_tags` (
    `movie_id` INT NOT NULL,
    `tag_id`   INT NOT NULL,
    PRIMARY KEY (`movie_id`, `tag_id`),
    FOREIGN KEY (`movie_id`) REFERENCES `movies`(`id`),
    FOREIGN KEY (`tag_id`)   REFERENCES `tags`(`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

INSERT INTO movie_tags (movie_id, tag_id) VALUES
(1,1),(1,5), (2,2),(2,5), (10,2),(10,3),(10,6),
(11,1),(11,2),(11,5), (13,2),(13,3),(13,6),
(15,1),(15,2),(15,5), (19,2),(19,3),(19,6),
(21,2),(21,3),(21,6), (23,1),(23,5),(23,7),
(24,2),(24,3),(24,6),(24,7), (28,2),(28,3),(28,6),
(30,1),(30,2),(30,3),(30,5)
ON DUPLICATE KEY UPDATE tag_id=VALUES(tag_id);



-- ==========================================
-- 插入 35 部电影素材
-- 海报使用本地真实电影海报 (static/img/movies/)
-- 已通过 TMDB 批量下载并保存
-- ==========================================
INSERT INTO `movies` (`id`, `title`, `genre`, `area`, `intro`, `image_path`) VALUES
-- 动作类 (4部)
(1,  '战狼2',        '动作', 'china',  '退伍军人冷锋卷入非洲国家叛乱，为同胞而战。',     '/static/img/movies/01_战狼2.jpg'),
(2,  '红海行动',     '动作', 'china',  '中国海军蛟龙突击队深入伊维亚共和国撤侨。',         '/static/img/movies/02_红海行动.jpg'),
(3,  '碟中谍7',      '动作', 'america','伊森·亨特面对失控AI，完成不可能的任务。',           '/static/img/movies/03_碟中谍7.jpg'),
(4,  '速度与激情10',  '动作', 'america','但丁复仇家族，飞车家族再次集结。',                 '/static/img/movies/04_速度与激情10.jpg'),

-- 喜剧类 (3部)
(5,  '你好，李焕英',  '喜剧', 'china',  '女儿穿越回1981年，与年轻时的母亲相遇。',           '/static/img/movies/05_你好_李焕英.jpg'),
(6,  '飞驰人生2',     '喜剧', 'china',  '昔日冠军车手重返巴音布鲁克赛场。',                 '/static/img/movies/06_飞驰人生2.jpg'),
(7,  '小偷家族',      '喜剧', 'japan',  '一个靠偷窃维生的家庭收养了一名被遗弃女孩。',       '/static/img/movies/07_小偷家族.jpg'),

-- 爱情类 (3部)
(8,  '这么多年',      '爱情', 'china',  '小镇女孩与叛逆少年跨越十年的青春恋曲。',           '/static/img/movies/08_这么多年.jpg'),
(9,  '你的名字。',    '爱情', 'japan',  '东京男孩与乡下女孩身体互换的奇幻爱情。',           '/static/img/movies/09_你的名字.jpg'),
(10, '泰坦尼克号',    '爱情', 'america','穷小子与富家女在巨轮上的生死之恋。',               '/static/img/movies/10_泰坦尼克号.jpg'),

-- 动画类 (4部)
(11, '哪吒之魔童降世', '动画', 'china',  '生而为魔的哪吒逆天改命，我命由我不由天。',         '/static/img/movies/11_哪吒.jpg'),
(12, '长安三万里',    '动画', 'china',  '以高适视角回顾李白的一生与大唐盛世。',             '/static/img/movies/12_长安三万里.jpg'),
(13, '千与千寻',      '动画', 'japan',  '少女千寻误入神灵世界，为救父母勇敢成长。',         '/static/img/movies/13_千与千寻.jpg'),
(14, '疯狂动物城',    '动画', 'america','兔子朱迪与狐狸尼克联手揭开动物城阴谋。',           '/static/img/movies/14_疯狂动物城.jpg'),

-- 剧情类 (3部)
(15, '我不是药神',    '剧情', 'china',  '小商贩从印度走私抗癌药，拯救病人引发法律与良知的对决。', '/static/img/movies/15_我不是药神.jpg'),
(16, '活着',          '剧情', 'china',  '福贵一家历经战争、运动与苦难，顽强生存。',         '/static/img/movies/16_活着.jpg'),
(17, '寄生虫',        '剧情', 'korea',  '贫家四口寄生富家，欲望终将吞噬一切。',             '/static/img/movies/17_寄生虫.jpg'),

-- 悬疑类 (3部)
(18, '消失的她',      '悬疑', 'china',  '妻子度假失踪，丈夫陷入惊天迷局。',                 '/static/img/movies/18_消失的她.jpg'),
(19, '盗梦空间',      '悬疑', 'america','盗梦者潜入多层梦境植入想法，现实与梦境的边界模糊。', '/static/img/movies/19_盗梦空间.jpg'),
(20, '看不见的客人',  '悬疑', 'europe',  '企业家被控谋杀情人，一场精心策划的完美辩护。',     '/static/img/movies/20_看不见的客人.jpg'),

-- 惊悚类 (2部)
(21, '沉默的羔羊',    '惊悚', 'america','实习探员向食人魔医生求助追捕连环杀手。',           '/static/img/movies/21_沉默的羔羊.jpg'),
(22, '恐怖直播',      '惊悚', 'korea',  '电台主持直播中遭遇恐怖袭击，演播室沦为战场。',     '/static/img/movies/22_恐怖直播.jpg'),

-- 科幻类 (3部)
(23, '流浪地球2',     '科幻', 'china',  '太阳即将毁灭，人类启动流浪地球计划。',             '/static/img/movies/23_流浪地球2.jpg'),
(24, '星际穿越',      '科幻', 'america','宇航员穿越虫洞为人类寻找新家园。',                 '/static/img/movies/24_星际穿越.jpg'),
(25, '阿丽塔：战斗天使','科幻','america','改造人少女在废铁城中追寻自己的身世。',             '/static/img/movies/25_阿丽塔.jpg'),

-- 战争类 (2部)
(26, '长津湖',        '战争', 'china',  '志愿军第九兵团在极寒中血战长津湖。',               '/static/img/movies/26_长津湖.jpg'),
(27, '血战钢锯岭',    '战争', 'america','军医戴斯蒙德在冲绳战役中不持一枪救下75人。',       '/static/img/movies/27_血战钢锯岭.jpg'),

-- 冒险类 (3部)
(28, '少年派的奇幻漂流','冒险','america','少年与猛虎在太平洋上227天漂流求生。',             '/static/img/movies/28_少年派.jpg'),
(29, '阿凡达2：水之道','冒险','america','杰克一家为躲避追捕迁居潘多拉海洋部落。',           '/static/img/movies/29_阿凡达2.jpg'),
(34, '封神第一部',    '冒险', 'china',  '殷寿暴政引发人仙妖三界大战，少年英雄觉醒。',       '/static/img/movies/34_封神第一部.jpg'),

-- 警匪类 (2部)
(30, '无间道',        '警匪', 'china',  '警察卧底与黑帮卧底的双面人生，终极对决。',         '/static/img/movies/30_无间道.jpg'),
(31, '追龙',          '警匪', 'china',  '毒枭跛豪与探长雷洛联手统治香港地下世界。',         '/static/img/movies/31_追龙.jpg'),

-- 家庭/音乐/补充 (4部)
(32, '海蒂和爷爷',    '家庭', 'europe',  '孤儿海蒂被送到阿尔卑斯山与孤僻爷爷生活。',         '/static/img/movies/32_海蒂和爷爷.jpg'),
(33, '爆裂鼓手',      '音乐', 'america','爵士鼓学生对完美苛求到极致的疯狂之路。',           '/static/img/movies/33_爆裂鼓手.jpg'),
(35, '铃芽之旅',      '动画', 'japan',  '少女铃芽与闭门师草太踏上关闭灾难之门之旅。',       '/static/img/movies/35_铃芽之旅.jpg');

-- ==========================================
-- 给所有电影补充评分、时长、状态
-- ==========================================
UPDATE movies SET rating=7.2, duration=123, status='上映中' WHERE id=1;
UPDATE movies SET rating=8.3, duration=138, status='上映中' WHERE id=2;
UPDATE movies SET rating=7.8, duration=163, status='上映中' WHERE id=3;
UPDATE movies SET rating=6.2, duration=141, status='上映中' WHERE id=4;
UPDATE movies SET rating=7.8, duration=128, status='下架'   WHERE id=5;
UPDATE movies SET rating=7.3, duration=121, status='上映中' WHERE id=6;
UPDATE movies SET rating=8.1, duration=117, status='下架'   WHERE id=7;
UPDATE movies SET rating=7.1, duration=114, status='下架'   WHERE id=8;
UPDATE movies SET rating=8.4, duration=106, status='下架'   WHERE id=9;
UPDATE movies SET rating=9.4, duration=194, status='下架'   WHERE id=10;
UPDATE movies SET rating=8.5, duration=110, status='上映中' WHERE id=11;
UPDATE movies SET rating=8.0, duration=168, status='上映中' WHERE id=12;
UPDATE movies SET rating=9.3, duration=125, status='下架'   WHERE id=13;
UPDATE movies SET rating=9.2, duration=108, status='下架'   WHERE id=14;
UPDATE movies SET rating=9.0, duration=116, status='上映中' WHERE id=15;
UPDATE movies SET rating=9.3, duration=132, status='下架'   WHERE id=16;
UPDATE movies SET rating=8.7, duration=132, status='上映中' WHERE id=17;
UPDATE movies SET rating=6.7, duration=121, status='下架'   WHERE id=18;
UPDATE movies SET rating=9.3, duration=148, status='下架'   WHERE id=19;
UPDATE movies SET rating=8.6, duration=106, status='下架'   WHERE id=20;
UPDATE movies SET rating=8.8, duration=118, status='下架'   WHERE id=21;
UPDATE movies SET rating=8.1, duration=98,  status='下架'   WHERE id=22;
UPDATE movies SET rating=8.3, duration=173, status='上映中' WHERE id=23;
UPDATE movies SET rating=9.3, duration=169, status='上映中' WHERE id=24;
UPDATE movies SET rating=7.3, duration=122, status='下架'   WHERE id=25;
UPDATE movies SET rating=7.4, duration=176, status='上映中' WHERE id=26;
UPDATE movies SET rating=8.7, duration=139, status='上映中' WHERE id=27;
UPDATE movies SET rating=9.1, duration=127, status='上映中' WHERE id=28;
UPDATE movies SET rating=7.6, duration=192, status='上映中' WHERE id=29;
UPDATE movies SET rating=9.2, duration=101, status='上映中' WHERE id=30;
UPDATE movies SET rating=7.3, duration=132, status='下架'   WHERE id=31;
UPDATE movies SET rating=9.2, duration=111, status='下架'   WHERE id=32;
UPDATE movies SET rating=8.5, duration=106, status='下架'   WHERE id=33;
UPDATE movies SET rating=7.2, duration=148, status='上映中' WHERE id=34;
UPDATE movies SET rating=7.7, duration=122, status='上映中' WHERE id=35;

-- ==========================================
-- 电影票表（防并发：UNIQUE(showtime_id, seat_id)）
-- ==========================================
CREATE TABLE IF NOT EXISTS `tickets` (
    `id`          INT AUTO_INCREMENT PRIMARY KEY,
    `showtime_id` INT NOT NULL,
    `user_id`     INT NOT NULL,
    `seat_id`     BIGINT NOT NULL,
    `price`       DOUBLE(11,2) NOT NULL DEFAULT 0,
    `status`      VARCHAR(20) NOT NULL DEFAULT 'paid' COMMENT 'locked/paid',
    `lock_time`   DATETIME,
    `created_at`  DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY `uk_showtime_seat` (`showtime_id`, `seat_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
