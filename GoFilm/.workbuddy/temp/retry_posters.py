"""
补全失败的 16 张海报 - 用更精确的搜索词
"""
import requests
import re
import os

MOVIE_DIR = r"D:\code\go\repository\learn-Go\GoFilm\.workbuddy\temp\real_posters"
IMG_DIR = r"D:\code\go\repository\learn-Go\GoFilm\views\static\img\movies"

HEADERS = {
    "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Safari/537.36",
    "Accept-Language": "zh-CN,zh;q=0.9,en;q=0.8",
}

# 已存在的图片
existing = {f.split(".")[0] for f in os.listdir(MOVIE_DIR) if f.endswith(".jpg")}
print("Existing:", len(existing))

# 失败的电影用更直接的搜索
retry = [
    ("02_红海行动",      "Operation Red Sea 2018"),
    ("05_你好_李焕英",   "Hi Mom 2021"),
    ("07_小偷家族",      "Shoplifters 2018"),
    ("08_这么多年",      "All These Years 这么多年 2023"),
    ("09_你的名字",      "Your Name 2016"),
    ("11_哪吒",          "Ne Zha 哪吒之魔童降世"),
    ("12_长安三万里",    "Chang An"),
    ("13_千与千寻",      "Spirited Away 2001"),
    ("15_我不是药神",    "Dying to Survive 2018"),
    ("16_活着",          "To Live 1994 活着"),
    ("18_消失的她",      "Lost in the Stars 2023"),
    ("22_恐怖直播",      "Terror Live 2013"),
    ("26_长津湖",        "Battle Lake Changjin"),
    ("32_海蒂和爷爷",    "Heidi 2015"),
    ("34_封神第一部",    "Creation Gods 封神"),
    ("35_铃芽之旅",      "Suzume"),
]

import time
success = 0
for name, query in retry:
    if name in existing:
        print(f"[skip] {name}: already exists")
        continue
    
    print(f"[retry] {name}: {query} ... ", end="", flush=True)
    
    try:
        s = requests.get(
            f"https://www.themoviedb.org/search?query={requests.utils.quote(query)}",
            headers=HEADERS, timeout=12
        )
        # 更宽松的正则
        ids = re.findall(r'/movie/(\d+)\b', s.text)
        if not ids:
            # Try alt
            ids = re.findall(r'"id":(\d+),.*?"media_type":"movie"', s.text)
        if not ids:
            print("no ID")
            time.sleep(1.5)
            continue
        
        # 去重保序
        seen = set()
        uniq = []
        for i in ids:
            if i not in seen:
                seen.add(i)
                uniq.append(i)
        movie_id = uniq[0]
        
        # 详情页
        d = requests.get(
            f"https://www.themoviedb.org/movie/{movie_id}",
            headers=HEADERS, timeout=12
        )
        if d.status_code != 200:
            print(f"detail {d.status_code}")
            time.sleep(1.5)
            continue
        
        m = re.search(r'/t/p/w500/([a-zA-Z0-9]+)\.(jpg|webp)', d.text)
        if not m:
            print("no poster")
            time.sleep(1.5)
            continue
        
        url = f"https://image.tmdb.org/t/p/w500/{m.group(1)}.{m.group(2)}"
        img = requests.get(url, headers=HEADERS, timeout=15)
        if img.status_code == 200 and len(img.content) > 2000:
            with open(os.path.join(MOVIE_DIR, f"{name}.jpg"), "wb") as f:
                f.write(img.content)
            with open(os.path.join(IMG_DIR, f"{name}.jpg"), "wb") as f:
                f.write(img.content)
            print(f"OK ({len(img.content)//1024}KB)")
            success += 1
        else:
            print("img fail")
    
    except Exception as e:
        print(f"err: {e}")
    
    time.sleep(1.5)

print(f"\n重试成功: {success}")
print(f"当前总数: {len([f for f in os.listdir(MOVIE_DIR) if f.endswith('.jpg')])}")
