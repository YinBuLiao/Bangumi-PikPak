<script setup>
import { computed, inject, onMounted, ref } from 'vue'

const ctx = inject('animeX')
if (!ctx) throw new Error('AnimeX context is missing')

const {
  adminTab, adminUserTab, adminAnimeTab, adminTabTitle, adminLoading, loadAdminPanel, switchAdminTab, switchAdminUserTab, switchAdminAnimeTab,
  adminOverview, formatStorage, adminUsers, adminSavingUser, adminUserForm, saveAdminUser,
  adminInvites, adminInviteLoading, generateInviteCodes, deleteInviteCodes,
  adminAnime, adminDownloadRequests, approveDownloadRequest, placeholder, adminLogs, adminConfigForm, saveAdminConfig, adminConfigSaving, deleteAdminAnime, deleteAdminEpisode,
  adminMonitor, statusText, goHome, authUser, logout, openPasswordModal, scanLibrary, loading,
} = ctx

const inviteModalOpen = ref(false)
const inviteForm = ref({ count: 1, expires_at: '' })
const selectedInvites = ref([])
const selectedAnime = ref([])
const animeDeleteModalOpen = ref(false)
const animeDeleteFiles = ref(false)
const episodeModalOpen = ref(false)
const episodeDeleteFiles = ref(false)
const episodeTargetAnime = ref(null)
const selectedEpisodeID = ref('')
const selectedEpisode = computed(() => episodeTargetAnime.value?.episodes?.find((ep) => ep.id === selectedEpisodeID.value) || null)
const adminUserMenuOpen = ref(false)

const overviewCards = computed(() => {
  if (adminOverview.value?.cards?.length) return adminOverview.value.cards
  return [
    { label: '总用户数', value: adminUsers.value?.length || 0, trend: '本地账号', icon: '♟', tone: '' },
    { label: '番剧总数', value: adminOverview.value?.anime_count || 0, trend: `${adminOverview.value?.storage_provider || 'storage'} 快照`, icon: '📺', tone: 'orange' },
    { label: '剧集总数', value: adminOverview.value?.episode_count || 0, trend: '已索引', icon: '🧩', tone: 'blue' },
    { label: '文件总数', value: adminOverview.value?.file_count || 0, trend: formatStorage(adminOverview.value?.storage_bytes || 0), icon: '▶', tone: 'green' },
    { label: '运行时间', value: adminMonitor.value?.uptime || '—', trend: '当前进程', icon: '⏱', tone: 'pink' },
  ]
})

onMounted(() => {
  loadAdminPanel()
})

function formatChinaTime(value) {
  if (!value) return '永久'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return new Intl.DateTimeFormat('zh-CN', {
    timeZone: 'Asia/Shanghai',
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
    hour12: false,
  }).format(date).replace(/\//g, '-')
}

function inviteStatus(code) {
  if (code.used_by) return '已使用'
  if (code.expires_at && new Date(code.expires_at).getTime() <= Date.now()) return '已过期'
  return '未使用'
}

function inviteStatusClass(code) {
  if (code.used_by) return 'used'
  if (code.expires_at && new Date(code.expires_at).getTime() <= Date.now()) return 'expired'
  return 'active'
}

function toggleAllInvites(event) {
  selectedInvites.value = event.target.checked ? adminInvites.value.map((item) => item.code) : []
}

async function submitInviteGenerate() {
  const payload = {
    count: Number(inviteForm.value.count || 1),
    expires_at: inviteForm.value.expires_at ? new Date(inviteForm.value.expires_at).toISOString() : '',
  }
  await generateInviteCodes(payload)
  inviteModalOpen.value = false
}

async function deleteSelectedInvites() {
  await deleteInviteCodes(selectedInvites.value)
  selectedInvites.value = []
}

function toggleAllAnime(event) {
  selectedAnime.value = event.target.checked ? adminAnime.value.map((item) => item.id || item.title) : []
}

async function deleteSelectedAnime() {
  if (!selectedAnime.value.length) return
  animeDeleteFiles.value = false
  animeDeleteModalOpen.value = true
}

async function confirmDeleteSelectedAnime() {
  await deleteAdminAnime(selectedAnime.value, { delete_files: animeDeleteFiles.value })
  selectedAnime.value = []
  animeDeleteModalOpen.value = false
}

function shortSummary(text) {
  const clean = String(text || '暂无简介').replace(/\s+/g, ' ').trim()
  return clean.length > 34 ? clean.slice(0, 34) + '…' : clean
}

function openEpisodeModal(item) {
  episodeTargetAnime.value = item
  selectedEpisodeID.value = item.episodes?.[0]?.id || ''
  episodeDeleteFiles.value = false
  episodeModalOpen.value = true
}

function downloadRequestStatusText(status) {
  return ({ pending: '待处理', downloading: '提交中', approved: '已同意', rejected: '已拒绝', failed: '失败' })[status] || status || '待处理'
}

function downloadRequestStatusClass(status) {
  return status || 'pending'
}

async function confirmDeleteEpisode() {
  const item = episodeTargetAnime.value
  const episode = selectedEpisode.value
  if (!item || !episode) return
  await deleteAdminEpisode(item, episode, { delete_files: episodeDeleteFiles.value && !episode.direct })
  episodeModalOpen.value = false
  selectedEpisodeID.value = ''
  episodeTargetAnime.value = null
}
</script>

<template>
  <section class="admin-screen admin-console-screen">
  <aside class="admin-sidebar-panel">
    <div class="admin-mobile-head">
      <button type="button" class="admin-mobile-home" @click="goHome">‹</button>
      <strong class="brand-wordmark">Anime<span>X</span></strong>
      <button type="button" class="admin-mobile-user" :class="{ open: adminUserMenuOpen }" @click="adminUserMenuOpen = !adminUserMenuOpen">
        <span class="sidebar-avatar">{{ (authUser || 'A').slice(0, 1).toUpperCase() }}</span>
        <span>{{ authUser || 'admin' }}</span>
        <b>⌄</b>
      </button>
      <div v-if="adminUserMenuOpen" class="admin-user-menu admin-mobile-user-menu" @click.stop>
        <button type="button" @click="openPasswordModal(); adminUserMenuOpen = false">修改密码</button>
        <button type="button" class="danger" @click="logout">退出登录</button>
      </div>
    </div>
    <div class="admin-brand">
        <strong class="brand-wordmark">Anime<span>X</span></strong>
    </div>
    <nav class="admin-side-tabs">
      <button :class="{ active: adminTab === 'overview' }" @click="switchAdminTab('overview')"><span>⌂</span>数据概览<i>⌄</i></button>
      <button :class="{ active: adminTab === 'users' }" @click="switchAdminTab('users')"><span>♙</span>用户管理<i>⌄</i></button>
      <div v-if="adminTab === 'users'" class="admin-side-subtabs">
        <button :class="{ active: adminUserTab === 'list' }" @click="switchAdminUserTab('list')">用户列表</button>
        <button :class="{ active: adminUserTab === 'invites' }" @click="switchAdminUserTab('invites')">邀请码管理</button>
      </div>
      <button :class="{ active: adminTab === 'storage' }" @click="switchAdminTab('storage')"><span>▱</span>储存桶配置<i>⌄</i></button>
      <button :class="{ active: adminTab === 'anime' }" @click="switchAdminTab('anime')"><span>▣</span>番剧管理<i>⌄</i></button>
      <div v-if="adminTab === 'anime'" class="admin-side-subtabs">
        <button :class="{ active: adminAnimeTab === 'list' }" @click="switchAdminAnimeTab('list')">番剧列表</button>
        <button :class="{ active: adminAnimeTab === 'requests' }" @click="switchAdminAnimeTab('requests')">下载申请</button>
      </div>
      <button :class="{ active: adminTab === 'logs' }" @click="switchAdminTab('logs')"><span>▤</span>日志管理<i>⌄</i></button>
      <button :class="{ active: adminTab === 'monitor' }" @click="switchAdminTab('monitor')"><span>◈</span>系统监控<i>⌄</i></button>
      <button :class="{ active: adminTab === 'config' }" @click="switchAdminTab('config')"><span>⚙</span>系统设置<i>⌄</i></button>
    </nav>
    <div class="admin-side-bottom">
      <button class="admin-return-home" @click="goHome"><span>«</span>返回主页</button>
      <div class="admin-user-card" :class="{ open: adminUserMenuOpen }" @click="adminUserMenuOpen = !adminUserMenuOpen">
        <span class="sidebar-avatar">{{ (authUser || 'A').slice(0, 1).toUpperCase() }}</span>
        <div class="admin-user-meta">
          <strong>{{ authUser || 'admin' }}</strong>
          <small>超级管理员</small>
        </div>
        <span class="admin-user-arrow">⌄</span>
        <div v-if="adminUserMenuOpen" class="admin-user-menu" @click.stop>
          <button type="button" @click="openPasswordModal(); adminUserMenuOpen = false">修改密码</button>
          <button type="button" class="danger" @click="logout">退出登录</button>
        </div>
      </div>
    </div>
  </aside>
  <div class="admin-content-shell">
  <div class="admin-topline">
    <div>
      <div class="admin-breadcrumb"><span>首页</span><b>/</b><span>管理员面板</span><b>/</b><strong>{{ adminTabTitle }}</strong></div>
      <h1>管理控制台</h1>
      <p>集中管理用户、番剧索引、运行日志、系统监控和运行配置。</p>
    </div>
    <button class="primary" :disabled="adminLoading" @click="loadAdminPanel()">{{ adminLoading ? '刷新中' : '刷新' }}</button>
  </div>

  <div v-if="adminTab === 'overview'" class="admin-overview">
    <article v-for="card in overviewCards" :key="card.label" class="admin-stat" :class="card.tone">
      <div><span>{{ card.label }}</span><strong>{{ card.value }}</strong><small>{{ card.trend }}</small></div>
      <i>{{ card.icon }}</i>
    </article>
    <section class="admin-chart-panel">
      <div class="section-title"><h2>播放量趋势</h2><span>最近 30 天</span></div>
      <div class="fake-chart" aria-hidden="true">
        <span v-for="h in [38,28,52,58,44,31,40,36,49,43,54,47,60,50,64]" :key="h" :style="{ height: h + '%' }"></span>
      </div>
    </section>
    <section class="admin-ring-panel">
      <div class="section-title"><h2>资源分布</h2><span>{{ formatStorage(adminOverview.storage_bytes) }}</span></div>
      <div class="admin-ring"><b>{{ adminOverview.anime_count || 0 }}</b><small>番剧总数</small></div>
      <ul>
        <li><span></span>剧集 <b>{{ adminOverview.episode_count || 0 }}</b></li>
        <li><span></span>文件 <b>{{ adminOverview.file_count || 0 }}</b></li>
        <li><span></span>存储 <b>{{ formatStorage(adminOverview.storage_bytes) }}</b></li>
      </ul>
    </section>
    <section class="admin-quick-panel">
      <div class="section-title"><h2>快捷操作</h2></div>
      <div class="admin-quick-grid">
        <button @click="switchAdminTab('users')"><b>＋</b>添加用户</button>
        <button @click="switchAdminTab('storage')"><b>▱</b>储存桶配置</button>
        <button @click="switchAdminTab('anime')"><b>▦</b>番剧管理</button>
        <button @click="switchAdminTab('logs')"><b>▤</b>查看日志</button>
        <button @click="switchAdminTab('monitor')"><b>⚙</b>系统监控</button>
        <button @click="switchAdminTab('config')"><b>☰</b>系统配置</button>
      </div>
    </section>
  </div>

  <section v-else-if="adminTab === 'users'" class="admin-panel admin-users-panel">
    <div class="section-title"><h2>用户管理</h2><span>{{ adminUserTab === 'invites' ? adminInvites.length + ' 个邀请码' : adminUsers.length + ' 个账号' }}</span></div>
    <div class="admin-subtab-row">
      <button :class="{ active: adminUserTab === 'list' }" @click="switchAdminUserTab('list')">用户列表</button>
      <button :class="{ active: adminUserTab === 'invites' }" @click="switchAdminUserTab('invites')">邀请码管理</button>
    </div>
    <template v-if="adminUserTab === 'list'">
      <form class="admin-user-form" @submit.prevent="saveAdminUser">
        <input v-model.trim="adminUserForm.username" placeholder="用户名" required />
        <input v-model="adminUserForm.password" type="password" placeholder="密码" required />
        <select v-model="adminUserForm.role"><option value="user">普通用户</option><option value="admin">管理员</option></select>
        <button class="primary" :disabled="adminSavingUser">{{ adminSavingUser ? '保存中' : '保存用户' }}</button>
      </form>
      <div class="admin-table">
        <div class="admin-table-head"><span>用户</span><span>角色</span><span>状态</span><span>权限</span></div>
        <div v-for="user in adminUsers" :key="user.username" class="admin-table-row">
          <span><b class="admin-avatar">{{ user.username.slice(0, 1).toUpperCase() }}</b>{{ user.username }}</span>
          <span><i :class="['role-pill', user.role]">{{ user.role === 'admin' ? '管理员' : '普通用户' }}</i></span>
          <span class="ok-text">在线</span>
          <span>{{ user.role === 'admin' ? '全部管理权限' : '浏览和播放' }}</span>
        </div>
      </div>
    </template>
    <section v-else class="invite-admin-panel">
      <div class="section-title">
        <h2>邀请码管理</h2>
        <div class="invite-actions">
          <button class="ghost small" type="button" :disabled="!selectedInvites.length || adminInviteLoading" @click="deleteSelectedInvites">批量删除</button>
          <button class="primary small invite-generate-btn" type="button" :disabled="adminInviteLoading" @click="inviteModalOpen = true">{{ adminInviteLoading ? '处理中' : '生成邀请码' }}</button>
        </div>
      </div>
      <div class="invite-table">
        <div class="invite-table-head">
          <label><input type="checkbox" :checked="adminInvites.length && selectedInvites.length === adminInvites.length" @change="toggleAllInvites" /></label>
          <span>邀请码</span><span>状态</span><span>使用人</span><span>过期时间</span><span>创建时间</span>
        </div>
        <div v-for="code in adminInvites" :key="code.code" class="invite-table-row">
          <label><input v-model="selectedInvites" type="checkbox" :value="code.code" /></label>
          <strong>{{ code.code }}</strong>
          <span :class="['invite-status', inviteStatusClass(code)]">{{ inviteStatus(code) }}</span>
          <span>{{ code.used_by || '—' }}</span>
          <span>{{ formatChinaTime(code.expires_at) }}</span>
          <span>{{ formatChinaTime(code.created_at) }}</span>
        </div>
        <p v-if="!adminInvites.length" class="empty-tip">暂无邀请码，点击右上角生成。</p>
      </div>
    </section>
  </section>

  <section v-else-if="adminTab === 'anime'" class="admin-panel anime-admin-panel">
    <div class="section-title anime-admin-title">
      <div>
        <h2>番剧管理</h2>
        <p>{{ adminAnimeTab === 'requests' ? '审核普通用户提交的剧集下载申请，同意后会自动提交到当前储存桶。' : '管理已索引的番剧、剧集资源和储存桶文件。' }}</p>
      </div>
      <div class="invite-actions">
        <span class="anime-admin-count">{{ adminAnimeTab === 'requests' ? adminDownloadRequests.length + ' 条待处理申请' : adminAnime.length + ' 部番剧' }}</span>
        <button v-if="adminAnimeTab === 'list'" class="primary small" type="button" :disabled="loading" @click="scanLibrary">{{ loading ? '扫描中' : '扫描媒体库' }}</button>
        <button v-if="adminAnimeTab === 'list'" class="ghost small" type="button" :disabled="!selectedAnime.length || adminLoading" @click="deleteSelectedAnime">删除选中</button>
      </div>
    </div>
    <div class="anime-admin-toolbar">
      <div class="admin-subtab-row anime-subtabs">
        <button :class="{ active: adminAnimeTab === 'list' }" @click="switchAdminAnimeTab('list')">番剧列表</button>
        <button :class="{ active: adminAnimeTab === 'requests' }" @click="switchAdminAnimeTab('requests')">下载申请</button>
      </div>
      <small>{{ adminAnimeTab === 'requests' ? '仅展示待处理申请' : selectedAnime.length ? '已选择 ' + selectedAnime.length + ' 项' : '支持批量删除番剧' }}</small>
    </div>
    <div v-if="adminAnimeTab === 'list'" class="admin-table anime-table">
      <div class="admin-table-head"><label><input type="checkbox" :checked="adminAnime.length && selectedAnime.length === adminAnime.length" @change="toggleAllAnime" /></label><span>封面</span><span>番剧名称</span><span>剧集</span><span>文件</span><span>状态</span><span>更新时间</span><span>操作</span></div>
      <div v-for="item in adminAnime" :key="item.id || item.title" class="admin-table-row">
        <label><input v-model="selectedAnime" type="checkbox" :value="item.id || item.title" /></label>
        <span><img :src="item.cover_url || placeholder(item.title)" :alt="item.title" /></span>
        <span class="anime-title-cell"><strong :title="item.title">{{ item.title }}</strong><small :title="item.summary || '暂无简介'">{{ shortSummary(item.summary) }}</small></span>
        <span>{{ item.episode_count }}</span>
        <span>{{ item.file_count }}</span>
        <span><i class="status-pill">{{ item.status }}</i></span>
        <span>{{ item.updated_at }}</span>
        <span><button class="ghost small" type="button" :disabled="!item.episodes?.length" @click="openEpisodeModal(item)">剧集管理</button></span>
      </div>
      <p v-if="!adminAnime.length" class="empty-tip">暂无番剧数据。</p>
    </div>
    <template v-else>
      <div v-if="adminDownloadRequests.length" class="admin-table request-table">
        <div class="admin-table-head"><span>申请用户</span><span>番剧 / 剧集</span><span>来源</span><span>状态</span><span>申请时间</span><span>操作</span></div>
        <div v-for="item in adminDownloadRequests" :key="item.id" class="admin-table-row">
          <span><b class="admin-avatar">{{ item.username.slice(0, 1).toUpperCase() }}</b><em>{{ item.username }}</em></span>
          <span class="request-title-cell"><strong :title="item.title">{{ item.bangumi_title || item.title }}</strong><small>{{ item.episode_label || item.title }}</small></span>
          <span><i class="source-pill">{{ item.magnet ? 'Magnet' : 'Torrent' }}</i></span>
          <span><i :class="['request-status', downloadRequestStatusClass(item.status)]">{{ downloadRequestStatusText(item.status) }}</i></span>
          <span>{{ formatChinaTime(item.created_at) }}</span>
          <span class="request-actions">
            <button class="primary small" type="button" :disabled="adminLoading || item.status !== 'pending'" @click="approveDownloadRequest(item.id, 'approve')">同意下载</button>
            <button class="ghost small" type="button" :disabled="adminLoading || item.status !== 'pending'" @click="approveDownloadRequest(item.id, 'reject')">拒绝</button>
          </span>
        </div>
      </div>
      <div v-else class="request-empty-state">
        <div class="request-empty-icon">↯</div>
        <div>
          <strong>暂无下载申请</strong>
          <small>普通用户点击“申请下载”后会出现在这里；管理员同意后系统会自动提交下载到当前储存桶。</small>
        </div>
        <button class="ghost small" type="button" @click="switchAdminAnimeTab('list')">查看番剧列表</button>
      </div>
    </template>
  </section>

  <section v-else-if="adminTab === 'logs'" class="admin-panel">
    <div class="section-title"><h2>日志管理</h2><span>运行事件</span></div>
    <div class="log-list">
      <article v-for="log in adminLogs" :key="log.time + log.module" class="log-row">
        <b>{{ log.level }}</b>
        <span>{{ log.time }}</span>
        <i>{{ log.module }}</i>
        <p>{{ log.message }}</p>
      </article>
    </div>
  </section>

  <section v-else-if="adminTab === 'storage'" class="admin-panel admin-config-panel">
    <div class="section-title"><h2>储存桶配置</h2><span>PikPak / 115 / 本地 Aria2 / NAS 与 Mikan 订阅配置</span></div>
    <form class="admin-config-form" @submit.prevent="saveAdminConfig">
      <section>
        <h3>储存桶类型</h3>
        <div class="storage-provider-picker">
          <button type="button" :class="{ active: adminConfigForm.storage_provider === 'pikpak' }" @click="adminConfigForm.storage_provider = 'pikpak'"><b>☁</b><span>PikPak</span><small>保留原有网盘离线下载</small></button>
          <button type="button" :class="{ active: adminConfigForm.storage_provider === 'drive115' }" @click="adminConfigForm.storage_provider = 'drive115'"><b>115</b><span>115 网盘</span><small>Cookie 离线任务</small></button>
          <button type="button" :class="{ active: adminConfigForm.storage_provider === 'local' }" @click="adminConfigForm.storage_provider = 'local'"><b>↯</b><span>本地存储</span><small>Aria2 下载到本机目录</small></button>
          <button type="button" :class="{ active: adminConfigForm.storage_provider === 'nas' }" @click="adminConfigForm.storage_provider = 'nas'"><b>▦</b><span>NAS 存储</span><small>Aria2 下载到挂载路径</small></button>
        </div>
      </section>
      <section v-if="adminConfigForm.storage_provider === 'pikpak'">
        <h3>PikPak 登录配置</h3>
        <div class="config-grid">
          <label><span>登录方式</span><select v-model="adminConfigForm.pikpak_auth_mode"><option value="password">账号密码</option><option value="token">Token</option></select></label>
          <label><span>PikPak 目录 ID</span><input v-model.trim="adminConfigForm.path" placeholder="PikPak 目标目录 ID" /></label>
          <label><span>PikPak 账号</span><input v-model.trim="adminConfigForm.username" autocomplete="username" placeholder="email / phone" /></label>
          <label><span>PikPak 密码</span><input v-model="adminConfigForm.password" type="password" autocomplete="current-password" placeholder="PikPak 密码" /></label>
          <label><span>Access Token</span><input v-model.trim="adminConfigForm.pikpak_access_token" placeholder="token 模式可填" /></label>
          <label><span>Refresh Token</span><input v-model.trim="adminConfigForm.pikpak_refresh_token" placeholder="token 模式可填" /></label>
          <label class="wide"><span>Encoded Token</span><input v-model.trim="adminConfigForm.pikpak_encoded_token" placeholder="如果使用 encoded token 可填写这里" /></label>
        </div>
      </section>
      <section v-else-if="adminConfigForm.storage_provider === 'drive115'">
        <h3>115 网盘配置</h3>
        <div class="config-grid">
          <label class="wide"><span>115 Cookie</span><input v-model.trim="adminConfigForm.drive115_cookie" placeholder="UID=...; CID=...; SEID=..." /></label>
          <label><span>根目录 CID</span><input v-model.trim="adminConfigForm.drive115_root_cid" placeholder="默认 0" /></label>
        </div>
      </section>
      <section v-else-if="adminConfigForm.storage_provider === 'local'">
        <h3>本地 Aria2 配置</h3>
        <div class="config-grid">
          <label><span>Aria2 RPC 地址</span><input v-model.trim="adminConfigForm.aria2_rpc_url" placeholder="http://127.0.0.1:6800/jsonrpc" /></label>
          <label><span>Aria2 Secret（可选）</span><input v-model.trim="adminConfigForm.aria2_rpc_secret" placeholder="未设置可留空" /></label>
          <label class="wide"><span>本地保存路径</span><input v-model.trim="adminConfigForm.local_storage_path" placeholder="D:\\AnimeX\\downloads" /></label>
        </div>
      </section>
      <section v-else>
        <h3>NAS 挂载配置</h3>
        <div class="config-grid">
          <label><span>Aria2 RPC 地址</span><input v-model.trim="adminConfigForm.aria2_rpc_url" placeholder="http://127.0.0.1:6800/jsonrpc" /></label>
          <label><span>Aria2 Secret（可选）</span><input v-model.trim="adminConfigForm.aria2_rpc_secret" placeholder="未设置可留空" /></label>
          <label class="wide"><span>NAS 挂载路径</span><input v-model.trim="adminConfigForm.nas_storage_path" placeholder="Z:\\AnimeX 或 /mnt/nas/anime" /></label>
        </div>
      </section>
      <section>
        <h3>Mikan 订阅配置</h3>
        <div class="config-grid">
          <label class="wide"><span>RSS 地址</span><input v-model.trim="adminConfigForm.rss" placeholder="https://mikanani.me/RSS/..." /></label>
          <label><span>Mikan 用户名（可选）</span><input v-model.trim="adminConfigForm.mikan_username" autocomplete="username" placeholder="用于一键订阅" /></label>
          <label><span>Mikan 密码（可选）</span><input v-model="adminConfigForm.mikan_password" type="password" autocomplete="current-password" placeholder="可留空" /></label>
        </div>
      </section>
      <footer>
        <button class="primary" :disabled="adminConfigSaving">{{ adminConfigSaving ? '保存中...' : '保存储存桶配置' }}</button>
      </footer>
    </form>
  </section>


  <section v-else-if="adminTab === 'config'" class="admin-panel admin-config-panel">
    <div class="section-title"><h2>系统配置</h2></div>
    <form class="admin-config-form" @submit.prevent="saveAdminConfig">
      <section>
        <h3>访问权限</h3>
        <label class="config-switch-row">
          <span><strong>强制要求登录可见</strong><small>开启后访问首页、媒体库等页面前必须登录；关闭后游客可直接浏览，管理员面板仍需管理员登录。</small></span>
          <input v-model="adminConfigForm.require_login" type="checkbox" />
        </label>
        <label class="config-switch-row">
          <span><strong>开启用户注册</strong><small>关闭后注册页面会提示不可注册，只允许管理员在后台创建账号。</small></span>
          <input v-model="adminConfigForm.enable_registration" type="checkbox" />
        </label>
        <label class="config-switch-row">
          <span><strong>开启邀请注册</strong><small>开启后普通用户注册必须填写后台生成的邀请码，邀请码使用后自动失效。</small></span>
          <input v-model="adminConfigForm.require_invite" type="checkbox" />
        </label>
        <div class="config-grid">
          <label class="wide">
            <span>普通用户每日申请下载上限</span>
            <input v-model.number="adminConfigForm.user_daily_download_limit" type="number" min="0" step="1" placeholder="3" />
            <small>限制每个普通用户每天最多申请几个剧集；填 0 表示不限制。游客仍会显示“请登录后下载”。</small>
          </label>
        </div>
      </section>
      <footer>
        <button class="primary" :disabled="adminConfigSaving">{{ adminConfigSaving ? '保存中...' : '保存配置' }}</button>
      </footer>
    </form>
  </section>

  <section v-else class="admin-panel monitor-panel">
    <div class="section-title"><h2>系统监控</h2><span>{{ adminMonitor.checked_at || '待刷新' }}</span></div>
    <div class="monitor-grid">
      <article><span>运行时间</span><strong>{{ adminMonitor.uptime || '-' }}</strong></article>
      <article><span>Goroutines</span><strong>{{ adminMonitor.goroutines || 0 }}</strong></article>
      <article><span>内存占用</span><strong>{{ formatStorage(adminMonitor.memory_alloc) }}</strong></article>
      <article><span>系统内存</span><strong>{{ formatStorage(adminMonitor.memory_sys) }}</strong></article>
      <article><span>MySQL</span><strong :class="{ good: adminMonitor.mysql_ready }">{{ statusText(adminMonitor.mysql_ready) }}</strong></article>
      <article><span>Redis</span><strong :class="{ good: adminMonitor.redis_ready }">{{ statusText(adminMonitor.redis_ready) }}</strong></article>
      <article><span>储存桶</span><strong :class="{ good: adminMonitor.storage_ready }">{{ adminMonitor.storage_provider || '未配置' }} · {{ statusText(adminMonitor.storage_ready) }}</strong></article>
      <article><span>代理</span><strong :class="{ good: adminMonitor.proxy_enabled }">{{ adminMonitor.proxy_enabled ? '已启用' : '未启用' }}</strong></article>
    </div>
  </section>
  </div>
  <div v-if="inviteModalOpen" class="modal-mask" @click.self="inviteModalOpen = false">
    <form class="admin-modal" @submit.prevent="submitInviteGenerate">
      <header>
        <h2>生成邀请码</h2>
        <button type="button" @click="inviteModalOpen = false">×</button>
      </header>
      <label>
        <span>生成数量</span>
        <input v-model.number="inviteForm.count" type="number" min="1" max="100" required />
        <small>一次最多生成 100 个。</small>
      </label>
      <label>
        <span>到期时间（中国时间）</span>
        <input v-model="inviteForm.expires_at" type="datetime-local" />
        <small>不选择到期时间则为永久有效。</small>
      </label>
      <footer>
        <button type="button" class="ghost" @click="inviteModalOpen = false">取消</button>
        <button class="primary invite-generate-btn" :disabled="adminInviteLoading">{{ adminInviteLoading ? '生成中...' : '确认生成' }}</button>
      </footer>
    </form>
  </div>
  <div v-if="animeDeleteModalOpen" class="modal-mask" @click.self="animeDeleteModalOpen = false">
    <form class="admin-modal" @submit.prevent="confirmDeleteSelectedAnime">
      <header>
        <h2>删除番剧</h2>
        <button type="button" @click="animeDeleteModalOpen = false">×</button>
      </header>
      <p class="modal-warning">
        即将删除 <b>{{ selectedAnime.length }}</b> 部番剧的后台索引与媒体库记录。
      </p>
      <label class="config-switch-row modal-switch">
        <span>
          <strong>同时删除储存桶内文件</strong>
          <small>开启后会一起删除 PikPak / 115 目录；本地存储和 NAS 会删除对应本地目录。</small>
        </span>
        <input v-model="animeDeleteFiles" type="checkbox" />
      </label>
      <footer>
        <button type="button" class="ghost" @click="animeDeleteModalOpen = false">取消</button>
        <button class="primary danger-primary" :disabled="adminLoading">{{ adminLoading ? '删除中...' : '确认删除' }}</button>
      </footer>
    </form>
  </div>
  <div v-if="episodeModalOpen" class="modal-mask" @click.self="episodeModalOpen = false">
    <form class="admin-modal episode-delete-modal" @submit.prevent="confirmDeleteEpisode">
      <header>
        <h2>剧集资源管理</h2>
        <button type="button" @click="episodeModalOpen = false">×</button>
      </header>
      <p class="modal-warning">
        {{ episodeTargetAnime?.title }}：选择要删除的某一集资源。
      </p>
      <div class="episode-admin-list">
        <label v-for="episode in episodeTargetAnime?.episodes || []" :key="episode.id" :class="{ active: selectedEpisodeID === episode.id, disabled: episode.direct }">
          <input v-model="selectedEpisodeID" type="radio" :value="episode.id" />
          <span><strong>{{ episode.label }}</strong><small>{{ episode.file_count }} 个文件 · {{ formatStorage(episode.storage_bytes || 0) }}<template v-if="episode.direct"> · 直接文件</template></small></span>
        </label>
      </div>
      <label class="config-switch-row modal-switch">
        <span>
          <strong>同时删除储存桶内文件</strong>
          <small>开启后删除该集目录；本地存储和 NAS 会删除对应剧集目录。直接文件不支持按集删除目录。</small>
        </span>
        <input v-model="episodeDeleteFiles" type="checkbox" :disabled="selectedEpisode?.direct" />
      </label>
      <footer>
        <button type="button" class="ghost" @click="episodeModalOpen = false">取消</button>
        <button class="primary danger-primary" :disabled="adminLoading || !selectedEpisodeID">{{ adminLoading ? '删除中...' : '删除这一集' }}</button>
      </footer>
    </form>
  </div>
</section>
</template>

