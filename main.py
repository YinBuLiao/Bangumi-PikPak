import asyncio
import feedparser
import logging
import os
import time
import httpx
from pikpakapi import PikPakApi # requirement: python >= 3.10

# 全局变量
USER = [""]
PASSWORD = [""]     # 密码中包含双引号时，需要转义，如：PASSWORD = ["123\"456"]
PATH = ['']
RSS = [""]
INTERVAL_TIME = 600  # 检查间隔时间，单位秒

# 创建 PikPakApi 对象
pikpak_clients = [PikPakApi(username=USER[i], password=PASSWORD[i]) for i in range(len(USER))]

async def login(account_index):
    # 使用 PikPakApi 登录
    client = pikpak_clients[account_index]
    await client.login()
    logging.info(f"账号 {USER[account_index]} 登录成功！")

async def get_rss():
    # 获取 RSS 种子列表
    rss = feedparser.parse(RSS[0])
    return [
        {
            'title': entry['title'],
            'link': entry['link'],
            'torrent': entry['enclosures'][0]['url'],
            'pubdate': entry['published'].split("T")[0]  # 只取日期部分
        }
        for entry in rss['entries']
    ]

async def get_folder_id(account_index, torrent):
    # 获取或创建存放种子的文件夹
    client = pikpak_clients[account_index]
    folder_path = PATH[account_index]
    pubdate = await get_date(torrent)

    # 获取文件夹列表
    folder_list = await client.file_list(parent_id=folder_path)

    for file in folder_list.get('files', []):
        if file['name'] == pubdate:
            return file['id']  # 找到已有文件夹，返回 ID

    # 没有找到则创建
    folder_info = await client.create_folder(name=pubdate, parent_id=folder_path)
    return folder_info['file']['id']

async def get_date(torrent):
    # 通过torrent获取mylist中的日期
    mylist = await get_rss()
    for entry in mylist:
        if entry['torrent'] == torrent:
            print(entry['pubdate'])
            return entry['pubdate']
    return None

async def magnet_upload(account_index, file_url, folder_id):
    # 请求离线下载所需数据
    client = pikpak_clients[account_index]
    result = await client.offline_download(file_url=file_url, parent_id=folder_id)

    if "error" in result:
        logging.error(f"账号 {USER[account_index]} 添加离线磁力任务失败: {result['error_description']}")
        return None, None

    logging.info(f"账号 {USER[account_index]} 添加离线磁力任务: {file_url}")
    return result['task']['id'], result['task']['name']

async def download_torrent(name, torrent):
    #下载torrent文件
    async with httpx.AsyncClient() as client:
        response = await client.get(torrent)
    
    os.makedirs('torrent', exist_ok=True)
    with open(f'torrent/{name}', 'wb') as f:
        f.write(response.content)

    logging.info(f"Finished downloading torrent: {name}")
    return f'torrent/{name}'

async def check_torrent(account_index, name, torrent):
    # print(torrent)
    # 检查torrent文件是否存在
    if not os.path.exists(f'torrent/{name}'):
        await download_torrent(name, torrent)
        folder_id = await get_folder_id(account_index, torrent)
        await magnet_upload(account_index, torrent, folder_id)
        return True

async def main():
    try:
        await asyncio.gather(*[login(i) for i in range(len(USER))])  # 并发登录
        mylist = await get_rss()

        for i in range(len(USER)):
            for entry in mylist:
                name = entry['torrent'].split('/')[-1]
                torrent = entry['torrent']
                await check_torrent(i, name, torrent)  # 串行检查种子并上传，避免并行时文件夹的创建冲突
    except Exception as e:
        logging.error(f"发生错误: {str(e)}")

if __name__ == "__main__":
    logging.basicConfig(level=logging.INFO, format="%(asctime)s - %(levelname)s - %(message)s")
    
    while True:
        asyncio.run(main())
        time.sleep(INTERVAL_TIME)