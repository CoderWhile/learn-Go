import re

with open(r'D:\code\go\repository\learn-Go\GoFilm\sql.sql', 'r', encoding='utf-8') as f:
    content = f.read()

movies_order = [
    (1, '01_战狼2'), (2, '02_红海行动'), (3, '03_碟中谍7'),
    (4, '04_速度与激情10'), (5, '05_你好_李焕英'), (6, '06_飞驰人生2'),
    (7, '07_小偷家族'), (8, '08_这么多年'), (9, '09_你的名字'),
    (10, '10_泰坦尼克号'), (11, '11_哪吒'), (12, '12_长安三万里'),
    (13, '13_千与千寻'), (14, '14_疯狂动物城'), (15, '15_我不是药神'),
    (16, '16_活着'), (17, '17_寄生虫'), (18, '18_消失的她'),
    (19, '19_盗梦空间'), (20, '20_看不见的客人'), (21, '21_沉默的羔羊'),
    (22, '22_恐怖直播'), (23, '23_流浪地球2'), (24, '24_星际穿越'),
    (25, '25_阿丽塔'), (26, '26_长津湖'), (27, '27_血战钢锯岭'),
    (28, '28_少年派'), (29, '29_阿凡达2'), (30, '30_无间道'),
    (31, '31_追龙'), (32, '32_海蒂和爷爷'), (33, '33_爆裂鼓手'),
    (34, '34_封神第一部'), (35, '35_铃芽之旅'),
]

count = 0
for mid, fn in movies_order:
    # 找以 (mid, 开头的行（允许前置空白）
    pattern = re.compile(r"^\s*\(\s*" + str(mid) + r"\s*,[^)]+\)\s*,?\s*$", re.MULTILINE)
    m = pattern.search(content)
    if m:
        old_line = m.group(0)
        new_path = f"/static/img/movies/{fn}.jpg"
        new_line = re.sub(r"https://api\.lorem\.space/image/movie\?w=300&h=450&hash=[a-zA-Z0-9]+", new_path, old_line)
        content = content.replace(old_line, new_line)
        count += 1
    else:
        print(f"ID {mid} ({fn}): not found")

# 更新注释
content = content.replace(
    "海报使用 Lorem.space 电影占位图，每部不同图片",
    "海报使用本地真实电影海报 (static/img/movies/)"
).replace(
    "换正式海报只需修改 image_path 字段",
    "已通过 TMDB 批量下载并保存到 views/static/img/movies/"
)

with open(r'D:\code\go\repository\learn-Go\GoFilm\sql.sql', 'w', encoding='utf-8') as f:
    f.write(content)

print(f"\nReplaced {count} lines")
print("Remaining Lorem URLs:", content.count('lorem.space'))
