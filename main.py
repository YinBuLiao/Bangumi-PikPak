import asyncio
import feedparser
import logging
from logging.handlers import RotatingFileHandler
import os
import signal
import sys
import time
import httpx
import json
from pikpakapi import PikPakApi  # requirement: python >= 3.10


CONFIG_FILE = "config.json"     # 配置文件（保存基本配置）
CLIENT_STATE_FILE = "pikpak.json"    # 客户端状态文件（保存 PikPakApi 登录状态及 token 等信息）

# 全局变量（由配置文件或手动填写）
USER = [""]
PASSWORD = [""]
PATH = [""]
RSS = [""]
INTERVAL_TIME_RSS = 600  # rss 检查间隔
INTERVAL_TIME_REFRESH = 21600  # token 刷新间隔
PIKPAK_CLIENTS = [""]
last_refresh_time = 0


# 加载基本配置文件，并更新全局变量
def load_config():
    if os.path.exists(CONFIG_FILE):
        try:
            with open(CONFIG_FILE, "r", encoding="utf-8") as f:
                config = json.load(f)
            if config.get("username") and config.get("password") and config.get("path") and config.get("rss"):
                USER[0] = config.get("username")
                PASSWORD[0] = config.get("password")
                PATH[0] = config.get("path")
                RSS[0] = config.get("rss")
            logging.info("配置文件加载成功！")
        except Exception as e:
            logging.error(f"加载配置文件失败: {str(e)}")
    else:
        logging.info("配置文件不存在，使用默认设置。")


# 如果存在保存的客户端状态，则优先从 CLIENT_STATE_FILE 中加载token
# 否则根据用户名和密码新建客户端对象
# 此外，检验客户端是否是当前用户的，若不是则重新登录
def init_clients():
    global last_refresh_time
    client = None
    if os.path.exists(CLIENT_STATE_FILE):
        try:
            with open(CLIENT_STATE_FILE, "r", encoding="utf-8") as f:
                config = json.load(f)
            last_refresh_time = config.get("last_refresh_time", 0)
            client_token = config.get("client_token", {})
            if client_token and client_token.get("username") == USER[0]:
                client = PikPakApi.from_dict(client_token)
                logging.info("成功从客户端状态文件加载登录状态！")
            else:
                client = PikPakApi(username=USER[0], password=PASSWORD[0])
        except Exception as e:
            logging.warning(f"加载客户端状态失败: {str(e)}，将重新创建客户端。")
            client = PikPakApi(username=USER[0], password=PASSWORD[0])
    else:
        client = PikPakApi(username=USER[0], password=PASSWORD[0])
    PIKPAK_CLIENTS[0] = client


# 保存基本配置到 CONFIG_FILE
def update_config():
    config = {
        "username": USER[0],
        "password": PASSWORD[0],
        "path": PATH[0],
        "rss": RSS[0],
    }
    try:
        with open(CONFIG_FILE, "w", encoding="utf-8") as f:
            json.dump(config, f, indent=4, ensure_ascii=False)
        logging.info("配置文件更新成功！")
    except Exception as e:
        logging.error(f"配置文件更新失败: {str(e)}")


# 保存token到 CLIENT_STATE_FILE
def save_client():
    config = {
        "last_refresh_time": last_refresh_time,
        "client_token": PIKPAK_CLIENTS[0].to_dict(),
    }
    try:
        with open(CLIENT_STATE_FILE, "w", encoding="utf-8") as f:
            json.dump(config, f, indent=4, ensure_ascii=False)
        logging.info("客户端状态保存成功！")
    except Exception as e:
        logging.error(f"客户端状态保存失败: {str(e)}")


# 1. 先尝试调用 file_list() 检查 token 是否有效；
# 2. 若调用失败，则重新使用用户名密码登录；
async def login(account_index):
    client = PIKPAK_CLIENTS[account_index]
    try:
        # 尝试用 token 调用 file_list() 检查 token 是否有效
        await client.file_list(parent_id=PATH[account_index])
        logging.info(f"账号 {USER[account_index]} Token 有效")
    except Exception as e:
        logging.warning(f"使用 token 读取文件列表失败: {str(e)}，将重新登录。")
        try:
            await client.login()
        except Exception as e:
            logging.error(f"账号 {USER[account_index]} 登录失败: {str(e)}")
            return

    logging.info(f"账号 {USER[account_index]} 登录成功！")

    await auto_refresh_token()


# 定时刷新 token
async def auto_refresh_token():
    global last_refresh_time
    current_time = time.time()
    if current_time - last_refresh_time >= INTERVAL_TIME_REFRESH:
        try:
            client = PIKPAK_CLIENTS[0]
            await client.refresh_access_token()
            logging.info("token 刷新成功！")
            last_refresh_time = current_time
            save_client()
        except Exception as e:
            logging.error(f"token 刷新失败: {str(e)}")
            last_refresh_time = 0


# 解析 RSS 并返回种子列表
async def get_rss():
    rss = feedparser.parse(RSS[0])
    return [
        {
            'title': entry['title'],
            'link': entry['link'],
            'torrent': entry['enclosures'][0]['url'],
            'pubdate': entry['published'].split("T")[0]
        }
        for entry in rss['entries']
    ]


# 根据种子对应的发布时间获取或创建存放该种子的文件夹
async def get_folder_id(account_index, torrent):
    client = PIKPAK_CLIENTS[account_index]
    folder_path = PATH[account_index]
    pubdate = await get_date(torrent)
    # 获取文件夹列表
    folder_list = await client.file_list(parent_id=folder_path)
    for file in folder_list.get('files', []):
        if file['name'] == pubdate:
            return file['id']
    # 未找到则创建新文件夹
    folder_info = await client.create_folder(name=pubdate, parent_id=folder_path)
    return folder_info['file']['id']


# 通过解析 RSS 查找 torrent 对应的发布时间
async def get_date(torrent):
    mylist = await get_rss()
    for entry in mylist:
        if entry['torrent'] == torrent:
            logging.info(f"种子标题: {entry['title']}")
            logging.info(f"发布时间: {entry['pubdate']}")
            return entry['pubdate']
    return None


# 提交离线磁力任务至 PikPak
async def magnet_upload(account_index, file_url, folder_id):
    client = PIKPAK_CLIENTS[account_index]
    try:
        result = await client.offline_download(file_url=file_url, parent_id=folder_id)
    except Exception as e:
        logging.error(
            f"账号 {USER[account_index]} 添加离线磁力任务失败: {e}")
        return None, None
    logging.info(f"账号 {USER[account_index]} 添加离线磁力任务: {file_url}")
    return result['task']['id'], result['task']['name']


# 下载 torrent 文件并保存到本地
async def download_torrent(name, torrent):
    async with httpx.AsyncClient() as client:
        response = await client.get(torrent)
    os.makedirs('torrent', exist_ok=True)
    with open(f'torrent/{name}', 'wb') as f:
        f.write(response.content)
    logging.info(f"Finished downloading torrent: {name}")
    return f'torrent/{name}'


# 检查本地是否存在种子文件；若不存在则下载并提交离线任务
async def check_torrent(account_index, name, torrent, check_mode: str):
    if not os.path.exists(f'torrent/{name}'):
        if check_mode == "local":
            return True
        else:
            await download_torrent(name, torrent)
            folder_id = await get_folder_id(account_index, torrent)
            
            #遍历该文件夹下的文件，若已存在该种子则不再创建
            info_hash = name.rsplit('.', 1)[0]
            magnet_link = f"magnet:?xt=urn:btih:{info_hash}"
            client = PIKPAK_CLIENTS[account_index]
            sub_folder_list = await client.file_list(parent_id=folder_id)
            for sub_file in sub_folder_list.get('files', []):
                if sub_file['params']['url'] == magnet_link:
                    return False
            
            await magnet_upload(account_index, torrent, folder_id)
            return True
    else:
        return False


async def main():
    # 刷新 token
    await auto_refresh_token()
    # 获取 RSS 种子列表
    mylist = await get_rss()
    # 先检查本地文件是否存在，减少重复请求次数
    needLogin = False
    for entry in mylist:
        name = entry['torrent'].split('/')[-1]
        torrent = entry['torrent']
        needLogin = await check_torrent(0, name, torrent, "local")
        if needLogin:
            break

    # 如果需要下载文件，则登录（若有token，实际上是复用之前的连接状态）
    if needLogin:
        await asyncio.gather(*[login(i) for i in range(len(USER))])
        # 遍历所有账号和 RSS 列表，串行处理避免文件夹创建冲突
        for i in range(len(USER)):
            for entry in mylist:
                name = entry['torrent'].split('/')[-1]
                torrent = entry['torrent']
                await check_torrent(i, name, torrent, "network")
    else:
        logging.info("RSS源没有新的更新")


def setup_logging(
    log_file="rss-pikpak.log",
    log_level=logging.INFO,
    max_bytes=10*1024*1024,  # 10MB
    backup_count=5
):
    """配置日志系统
    
    Args:
        log_file: 日志文件路径
        log_level: 日志级别
        max_bytes: 单个日志文件最大大小
        backup_count: 保留的日志文件数量
    """
    try:
        # 创建logger对象
        logger = logging.getLogger()
        logger.setLevel(log_level)

        # 日志格式
        formatter = logging.Formatter(
            fmt="%(asctime)s [%(levelname)s] %(filename)s:%(lineno)d - %(message)s",
            datefmt="%Y-%m-%d %H:%M:%S"
        )

        # 文件处理器(启用日志轮转)
        file_handler = RotatingFileHandler(
            filename=log_file,
            maxBytes=max_bytes,
            backupCount=backup_count,
            encoding='utf-8'
        )
        file_handler.setFormatter(formatter)
        logger.addHandler(file_handler)

        # 控制台处理器
        console_handler = logging.StreamHandler(sys.stdout)
        console_handler.setFormatter(formatter)
        logger.addHandler(console_handler)

        logging.info("日志系统初始化成功")
        return logger

    except Exception as e:
        print(f"日志系统初始化失败: {str(e)}")
        sys.exit(1)

if __name__ == "__main__":
    setup_logging()
    load_config()  
    init_clients()
    update_config()  # 将当前基本配置写入文件（用户将配置写在main.py内的情况）

    # 处理退出情况
    def signal_handler(sig, frame):
        logging.info("正在保存状态并退出...")
        save_client()  # 保存客户端状态
        update_config()  # 保存配置
        sys.exit(0)
    signal.signal(signal.SIGINT, signal_handler)
    signal.signal(signal.SIGTERM, signal_handler)

    while True:
        asyncio.run(main())
        time.sleep(INTERVAL_TIME_RSS)
