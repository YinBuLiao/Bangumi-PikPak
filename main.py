import feedparser
import requests
import logging
import re
import os
import time

# 屏蔽ssl警告
requests.packages.urllib3.disable_warnings()

# 全局变量
PIKPAK_API_URL = "https://api-drive.mypikpak.com"
PIKPAK_USER_URL = "https://user.mypikpak.com"

USER = ["用户名"]
PASSWORD = ["密码"]
PATH = ['文件夹ID']
RSS = ["RSS链接"]

pikpak_headers = [None] * len(USER)


def get_rss():
    # 网站种子解析
    rss = feedparser.parse(
        RSS[0])
    # 将pubDate
    for entry in rss['entries']:
        try:
            pubdate = entry['published']
            # 将pubDate转换为yy-mm-dd格式
            pubdate = time.strftime(
                "%Y-%m-%d", time.strptime(pubdate, "%Y-%m-%dT%H:%M:%S.%f"))
        except:
            pubdate = entry['published']
            # 将pubDate转换为yy-mm-dd格式
            pubdate = time.strftime(
                "%Y-%m-%d", time.strptime(pubdate, "%Y-%m-%dT%H:%M:%S"))
    # 整理为JSON数组
    mylist = [{'title': entry['title'], 'link':entry['link'], 'torrent':entry['enclosures']
               [0]['url'], 'pubdate':pubdate}for entry in rss['entries']]
    return mylist


def login(account):
    index = USER.index(account)
    # 登录所需所有信息
    login_admin = account
    login_password = PASSWORD[index]
    login_url = f"{PIKPAK_USER_URL}/v1/auth/signin?client_id=YNxT9w7GMdWvEOKa"
    login_data = {"captcha_token": "",
                  "client_id": "YNxT9w7GMdWvEOKa",
                  "client_secret": "dbw2OtmVEeuUvIptb1Coyg",
                  "password": login_password, "username": login_admin}
    headers = {
        'User-Agent': 'protocolversion/200 clientid/YNxT9w7GMdWvEOKa action_type/ networktype/WIFI sessionid/ '
                      'devicesign/div101.073163586e9858ede866bcc9171ae3dcd067a68cbbee55455ab0b6096ea846a0 sdkversion/1.0.1.101300 '
                      'datetime/1630669401815 appname/android-com.pikcloud.pikpak session_origin/ grant_type/ clientip/ devicemodel/LG '
                      'V30 accesstype/ clientversion/ deviceid/073163586e9858ede866bcc9171ae3dc providername/NONE refresh_token/ '
                      'usrno/null appid/ devicename/Lge_Lg V30 cmd/login osversion/9 platformversion/10 accessmode/',
        'Content-Type': 'application/json; charset=utf-8',
        'Host': 'user.mypikpak.com',
    }
    # 请求登录api
    info = requests.post(url=login_url, json=login_data,
                         headers=headers, timeout=5, verify=False).json()
    # 获得调用其他api所需的headers
    headers['Authorization'] = f"Bearer {info['access_token']}"
    headers['Host'] = 'api-drive.mypikpak.com'
    pikpak_headers[index] = headers.copy()  # 拷贝

    logging.info(f"账号{account}登陆成功！")


def get_date(torrent):
    # 通过torrent获取mylist中的日期
    mylist = get_rss()
    for entry in mylist:
        if entry['torrent'] == torrent:
            return entry['pubdate']

# 获取文件夹id


def get_folder_id(torrent, folder_path, account):
    login_headers = get_headers(account)
    create_url = f"{PIKPAK_API_URL}/drive/v1/files?thumbnail_size=SIZE_MEDIUM&limit=100&parent_id={folder_path}".format(
        folder_path)
    create_result = requests.get(
        url=create_url, headers=login_headers, timeout=5, verify=False).json()
    # 获取files列表
    files = create_result['files']
    # 遍历files列表
    for file in files:
        if file['name'] == get_date(torrent):
            return file['id']
        else:
            folder_id = create_folder(get_date(torrent), PATH[0], USER[0])
            return folder_id

    # 处理请求异常
    if "error" in create_result:
        if create_result['error_code'] == 16:
            logging.info(f"账号{account}登录过期，正在重新登录")
            login(account)
            login_headers = get_headers(account)
            create_result = requests.post(
                url=create_url, headers=login_headers, timeout=5, verify=False).json()
        else:
            logging.error(
                f"账号{account}创建文件夹失败，错误信息：{create_result['error_description']}")
            return None

# 创建文件夹


def create_folder(folder_name, folder_path, account):
    login_headers = get_headers(account)
    create_url = f"{PIKPAK_API_URL}/drive/v1/files"
    create_data = {
        "kind": "drive#folder",
        "name": folder_name,
        "parent_id": folder_path
    }
    create_result = requests.post(
        url=create_url, headers=login_headers, json=create_data, timeout=5, verify=False).json()

    # 处理请求异常
    if "error" in create_result:
        if create_result['error_code'] == 16:
            logging.info(f"账号{account}登录过期，正在重新登录")
            login(account)
            login_headers = get_headers(account)
            create_result = requests.post(
                url=create_url, headers=login_headers, json=create_data, timeout=5, verify=False).json()
        else:
            logging.error(
                f"账号{account}创建文件夹失败，错误信息：{create_result['error_description']}")
            return None
    return create_result['file']['id']


# 获得headers，用于请求api
def get_headers(account):
    index = USER.index(account)

    if not pikpak_headers[index]:  # headers为空则先登录
        login(account)
    return pikpak_headers[index]

# 离线下载磁力


def magnet_upload(file_url, file_path, account):
    # 请求离线下载所需数据
    login_headers = get_headers(account)
    torrent_url = f"{PIKPAK_API_URL}/drive/v1/files"
    torrent_data = {
        "kind": "drive#file",
        "parent_id": file_path,
        "upload_type": "UPLOAD_TYPE_URL",
        "url": {
            "url": file_url
        },
        "params": {
            "with_thumbnail": "true",
            "from": "manual"
        }
    }
    # 请求离线下载
    torrent_result = requests.post(
        url=torrent_url, headers=login_headers, json=torrent_data, timeout=5, verify=False).json()

    # 处理请求异常
    if "error" in torrent_result:
        if torrent_result['error_code'] == 16:
            logging.info(f"账号{account}登录过期，正在重新登录")
            login(account)  # 重新登录该账号
            login_headers = get_headers(account)
            torrent_result = requests.post(
                url=torrent_url, headers=login_headers, json=torrent_data, timeout=5, verify=False).json()

        else:
            # 可以考虑加入删除离线失败任务的逻辑
            logging.error(
                f"账号{account}提交离线下载任务失败，错误信息：{torrent_result['error_description']}")
            return None, None

    # 输出日志
    file_url_part = re.search(r'^(magnet:\?).*(xt=.+?)(&|$)', file_url)
    if file_url_part:
        file_url_simple = ''.join(file_url_part.groups()[:-1])
        logging.info(f"账号{account}添加离线磁力任务:{file_url_simple}")
    else:
        logging.info(f"账号{account}添加离线磁力任务:{file_url}")

    # 返回离线任务id、下载文件名
    return torrent_result['task']['id'], torrent_result['task']['name']


def download_torrent(name, torrent):
    print('Downloading torrent: ' + name)
    torrent_url = torrent
    torrent_data = requests.get(url=torrent_url, timeout=5).content
    with open('torrent/' + name, 'wb') as f:
        f.write(torrent_data)
    print('Finished downloading torrent: ' + name)
    return 'torrent/' + name


def check_torrent(name, torrent):
    # print(torrent)
    # 检查torrent文件是否存在
    if os.path.exists('torrent/' + name):
        pass
    else:
        download_torrent(name, torrent)
        folder_path = get_folder_id(torrent, PATH[0], USER[0])
        magnet_upload(torrent, folder_path, USER[0])
        return True


def main():
    try:
        login(USER[0])
        mylist = get_rss()
        mylist_len = len(mylist)
        for i in range(mylist_len):
            torrent = mylist[i]['torrent']
            name = mylist[i]['torrent'].split('/')[-1]
            check_torrent(name, torrent)
    except requests.exceptions.ReadTimeout:
        print_info = f'下载磁链时网络请求超时！可稍后重试'
        logging.error(print_info)


while True:
    main()
    time.sleep(600)
