"""
从 TMDB 批量下载 35 部真实电影海报
流程: 搜索 -> 提取电影ID -> 进详情页 -> 抓海报(w500) -> 下载
"""
import requests
import re
import os
import time
import zipfile

MOVIE_DIR = r"D:\code\go\repository\learn-Go\GoFilm\.workbuddy\temp\real_posters"
IMG_DIR = r"D:\code\go\repository\learn-Go\GoFilm\views\static\img\movies"
ZIP_PATH = r"D:\code\go\repository\learn-Go\GoFilm\.workbuddy\temp\real_posters.zip"

os.makedirs(MOVIE_DIR, exist_ok=True)
os.makedirs(IMG_DIR, exist_ok=True)

movies = [
    ("01_战狼2",              "Wolf Warrior 2 战狼"),
    ("02_红海行动",            "Operation Red Sea 红海行动"),
    ("03_碟中谍7",            "Mission Impossible Dead Reckoning Part One"),
    ("04_速度与激情10",        "Fast X"),
    ("05_你好_李焕英",         "Hi Mom 你好李焕英"),
    ("06_飞驰人生2",           "Pegasus 2 飞驰人生"),
    ("07_小偷家族",            "Shoplifters 万引き家族"),
    ("08_这么多年",            "All These Years 这么多年"),
    ("09_你的名字",            "Your Name 君の名は"),
    ("10_泰坦尼克号",          "Titanic"),
    ("11_哪吒",               "Ne Zha 哪吒之魔童降世"),
    ("12_长安三万里",          "Chang An 长安三万里"),
    ("13_千与千寻",            "Spirited Away 千と千尋"),
    ("14_疯狂动物城",          "Zootopia"),
    ("15_我不是药神",          "Dying to Survive 我不是药神"),
    ("16_活着",               "To Live 1994 活着"),
    ("17_寄生虫",             "Parasite 기생충"),
    ("18_消失的她",            "Lost in the Stars 消失的她"),
    ("19_盗梦空间",            "Inception"),
    ("20_看不见的客人",        "Invisible Guest Contratiempo"),
    ("21_沉默的羔羊",          "Silence of the Lambs"),
    ("22_恐怖直播",            "The Terror Live 恐怖直播"),
    ("23_流浪地球2",           "Wandering Earth 2"),
    ("24_星际穿越",            "Interstellar"),
    ("25_阿丽塔",              "Alita Battle Angel"),
    ("26_长津湖",              "Battle at Lake Changjin 长津湖"),
    ("27_血战钢锯岭",          "Hacksaw Ridge"),
    ("28_少年派",              "Life of Pi"),
    ("29_阿凡达2",             "Avatar Way of Water"),
    ("30_无间道",              "Infernal Affairs 无间道"),
    ("31_追龙",               "Chasing the Dragon 追龙"),
    ("32_海蒂和爷爷",          "Heidi 2015"),
    ("33_爆裂鼓手",            "Whiplash 2014"),
    ("34_封神第一部",          "Creation of the Gods 封神"),
    ("35_铃芽之旅",            "Suzume 鈴芽のいる"),
]

HEADERS = {
    "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Safari/537.36",
    "Accept-Language": "zh-CN,zh;q=0.9,en;q=0.8",
}

def fetch_poster(name, query, index):
    """搜索+详情页取海报"""
    fname = os.path.join(MOVIE_DIR, f"{name}.jpg")
    imgname = os.path.join(IMG_DIR, f"{name}.jpg")
    
    if os.path.exists(fname) and os.path.getsize(fname) > 2000:
        print(f"[{index:2d}/35] {name}: cached")
        return True
    
    print(f"[{index:2d}/35] {name}: ", end="", flush=True)
    
    try:
        # 搜索
        s = requests.get(
            f"https://www.themoviedb.org/search?query={requests.utils.quote(query)}&language=zh-CN",
            headers=HEADERS, timeout=12
        )
        if s.status_code != 200:
            print(f"search HTTP {s.status_code}")
            return False
        
        ids = re.findall(r'/movie/(\d+)-', s.text)
        if not ids:
            print("no movie ID")
            return False
        
        # 取第一个 ID（去重保持顺序）
        seen = set()
        unique_ids = []
        for i in ids:
            if i not in seen:
                seen.add(i)
                unique_ids.append(i)
        movie_id = unique_ids[0]
        
        # 详情页
        d = requests.get(
            f"https://www.themoviedb.org/movie/{movie_id}?language=zh-CN",
            headers=HEADERS, timeout=12
        )
        if d.status_code != 200:
            print(f"detail HTTP {d.status_code}")
            return False
        
        # 提取 w500 海报（第一个）
        m = re.search(r'/t/p/w500/([a-zA-Z0-9]+)\.(jpg|webp)', d.text)
        if not m:
            print("no poster in detail")
            return False
        
        poster_url = f"https://image.tmdb.org/t/p/w500/{m.group(1)}.{m.group(2)}"
        
        # 下载图片
        img = requests.get(poster_url, headers=HEADERS, timeout=15)
        if img.status_code == 200 and len(img.content) > 2000:
            with open(fname, "wb") as f:
                f.write(img.content)
            with open(imgname, "wb") as f:
                f.write(img.content)
            size_kb = len(img.content) // 1024
            print(f"OK ({size_kb}KB)")
            return True
        else:
            print(f"img fail {img.status_code}")
            return False
    
    except Exception as e:
        print(f"err: {e}")
        return False


# ===== Main =====
print("=" * 60)
print("TMDB 真实电影海报批量下载 (35 部)")
print("=" * 60)

success = 0
for i, (name, query) in enumerate(movies, 1):
    if fetch_poster(name, query, i):
        success += 1
    time.sleep(1.8)  # 避免被限流

print(f"\n成功: {success}/35")

# 打包
files = sorted([f for f in os.listdir(MOVIE_DIR) if f.endswith(".jpg")])
if files:
    with zipfile.ZipFile(ZIP_PATH, "w", zipfile.ZIP_DEFLATED) as zf:
        for f in files:
            zf.write(os.path.join(MOVIE_DIR, f), f)
    print(f"ZIP: {ZIP_PATH} ({os.path.getsize(ZIP_PATH)//1024} KB)")
