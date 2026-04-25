<script setup>
import { computed, onMounted, onUnmounted, provide, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

const route = useRoute()
const router = useRouter()

const view = ref('home')
const library = ref([])
const selected = ref(null)
const currentEpisode = ref(null)
const currentFile = ref(null)
const query = ref('')
const results = ref([])
const searchQuery = ref('')
const homeFeed = ref([])
const homeSchedule = ref({ year: 0, season: '', days: [], items: [] })
const homeLoading = ref(false)
const heroIndex = ref(0)
const discover = ref([])
const discoverOffset = ref(0)
const discoverLimit = 24
const discoverHasMore = ref(false)
const discoverSection = ref('rank')
const discoverTag = ref('恋爱')
const discoverTitle = ref('Bangumi 排行榜')
const discovering = ref(false)
const subscribing = ref('')
const schedule = ref({ year: 0, season: '', days: [], items: [] })
const scheduleLoading = ref(false)
const historyItems = ref([])
const loading = ref(false)
const syncing = ref(false)
const searching = ref(false)
const downloading = ref('')
const downloadStatuses = ref({})
const authChecked = ref(false)
const authenticated = ref(false)
const authUser = ref('')
const authRole = ref('')
const isAdmin = computed(() => authRole.value === 'admin')
const userMenuOpen = ref(false)
const passwordModalOpen = ref(false)
const passwordSaving = ref(false)
const passwordForm = ref({ old_password: '', new_password: '', confirm_password: '' })
const requireLogin = ref(true)
const loginLoading = ref(false)
const loginForm = ref({ username: '', password: '' })
const registerLoading = ref(false)
const registerStatus = ref({ enable_registration: true, require_invite: false })
const registerForm = ref({ username: '', password: '', invite_code: '' })
const adminTab = ref('overview')
const adminUserTab = ref('list')
const adminAnimeTab = ref('list')
const adminLoading = ref(false)
const adminSavingUser = ref(false)
const adminInviteLoading = ref(false)
const adminOverview = ref({ cards: [], anime_count: 0, episode_count: 0, file_count: 0, storage_bytes: 0 })
const adminUsers = ref([])
const adminInvites = ref([])
const adminAnime = ref([])
const adminDownloadRequests = ref([])
const adminLogs = ref([])
const adminMonitor = ref({})
const adminConfigSaving = ref(false)
const adminConfigForm = ref({
  username: '',
  password: '',
  pikpak_auth_mode: 'password',
  pikpak_access_token: '',
  pikpak_refresh_token: '',
  pikpak_encoded_token: '',
  path: '',
  rss: '',
  mikan_username: '',
  mikan_password: '',
  require_login: true,
  enable_registration: true,
  require_invite: false,
  user_daily_download_limit: 3,
  storage_provider: 'pikpak',
  drive115_cookie: '',
  drive115_root_cid: '0',
  aria2_rpc_url: 'http://127.0.0.1:6800/jsonrpc',
  aria2_rpc_secret: '',
  local_storage_path: 'downloads',
  nas_storage_path: '',
})
const adminTabTitle = computed(() => ({
  overview: '数据概览',
  users: adminUserTab.value === 'invites' ? '邀请码管理' : '用户列表',
  storage: '储存桶配置',
  anime: adminAnimeTab.value === 'requests' ? '下载申请' : '番剧列表',
  logs: '日志管理',
  monitor: '系统监控',
  config: '系统配置',
}[adminTab.value] || '管理员面板'))
const adminUserForm = ref({ username: '', password: '', role: 'user' })
const toast = ref('')
const toastError = ref(false)
let toastTimer = 0
let heroTimer = 0
const historyStorageKey = 'animex:watch-history'
const installLoading = ref(false)
const installSaving = ref(false)
const installDone = ref(false)
const installError = ref('')
const installPath = ref('本地配置数据库')
const installRequired = ref(true)
const installOnly = ref(false)
const installDockerMode = ref(false)
const installTesting = ref('')
const installTestResults = ref({ mysql: null, redis: null })
const installWriteStage = ref('')
const installPage = ref(1)
const installSteps = computed(() => {
  if (installDockerMode.value) {
    return [
      { n: 1, page: 3, title: '管理员账号', desc: '配置后台登录' },
      { n: 2, page: 4, title: '数据写入', desc: '自动写入配置' },
      { n: 3, page: 5, title: '安装完成', desc: '直接进入系统' },
    ]
  }
  return [
    { n: 1, page: 1, title: '配置数据库', desc: '填写 MySQL 连接' },
    { n: 2, page: 2, title: '缓存代理', desc: 'Redis / Proxy 配置' },
    { n: 3, page: 3, title: '管理员账号', desc: '配置后台登录' },
    { n: 4, page: 4, title: '数据写入', desc: '可视化保存过程' },
    { n: 5, page: 5, title: '安装完成', desc: '直接进入系统' },
  ]
})
const installWriteSteps = [
  { title: '创建管理员表', sql: 'CREATE TABLE IF NOT EXISTS admin_users (...)' },
  { title: '写入管理员账号', sql: 'INSERT INTO admin_users (username, password_hash) VALUES (...)' },
  { title: '写入系统配置', sql: 'UPDATE system_config SET value = ... WHERE name = ...' },
  { title: '创建安装锁', sql: 'INSERT INTO install_state (installed_at, version) VALUES (...)' },
]
const installForm = ref({
  username: '',
  password: '',
  pikpak_auth_mode: 'password',
  pikpak_access_token: '',
  pikpak_refresh_token: '',
  pikpak_encoded_token: '',
  path: '',
  rss: '',
  http_proxy: 'http://127.0.0.1:7890',
  https_proxy: 'http://127.0.0.1:7890',
  socks_proxy: 'socks5://127.0.0.1:7890',
  enable_proxy: false,
  mikan_username: '',
  mikan_password: '',
  storage_provider: 'pikpak',
  drive115_cookie: '',
  drive115_root_cid: '0',
  aria2_rpc_url: 'http://127.0.0.1:6800/jsonrpc',
  aria2_rpc_secret: '',
  local_storage_path: 'downloads',
  nas_storage_path: '',
  mysql_host: '127.0.0.1',
  mysql_port: 3306,
  mysql_database: 'anime',
  mysql_username: '',
  mysql_password: '',
  mysql_dsn: '',
  redis_addr: '127.0.0.1:6379',
  redis_password: '',
  redis_db: 0,
  admin_username: 'admin',
  admin_password: '',
  require_login: true,
})

const featured = computed(() => homeFeed.value[heroIndex.value % Math.max(homeFeed.value.length, 1)] || null)
const routeViewMap = {
  '/': 'home',
  '/install': 'install',
  '/login': 'login',
  '/register': 'register',
  '/admin': 'admin',
  '/library': 'library',
  '/discover': 'discover',
  '/search': 'search',
  '/schedule': 'schedule',
  '/history': 'history',
  '/detail': 'detail',
  '/player': 'player',
}
const viewPathMap = {
  home: '/',
  install: '/install',
  login: '/login',
  register: '/register',
  admin: '/admin',
  library: '/library',
  discover: '/discover',
  search: '/search',
  schedule: '/schedule',
  history: '/history',
  detail: '/detail',
  player: '/player',
}
function syncViewFromRoute(path = route.path) {
  const nextView = routeViewMap[path] || 'home'
  if (view.value !== nextView) view.value = nextView
}
function syncRouteFromView(nextView) {
  const nextPath = viewPathMap[nextView] || '/'
  if (route.path !== nextPath) router.replace(nextPath)
}
const installActiveStep = computed(() => {
  if (installDone.value) return 5
  if (installSaving.value || installWriteStage.value) return 4
  return installPage.value
})
const installProgress = computed(() => {
  if (installDone.value) return 100
  const index = installWriteSteps.findIndex((step) => step.title === installWriteStage.value)
  if (index < 0) return 0
  return Math.round(((index + 1) / installWriteSteps.length) * 100)
})
const heroSlides = computed(() => homeFeed.value.slice(0, 4))
const recentPlayedItems = computed(() => historyItems.value
  .filter((record) => record.file_id || record.episode_id)
  .slice(0, 5)
  .map((record) => ({
    id: record.id,
    title: record.title,
    cover_url: record.cover_url,
    summary: record.summary,
    episode_id: record.episode_id,
    updated: [record.episode_label, formatHistoryTime(record.watched_at)].filter(Boolean).join(' · '),
    source: 'history',
    historyRecord: record,
    episodes: [],
  })))
const newItems = computed(() => homeFeed.value.slice(5, 15))

async function api(path, options = {}) {
  const response = await fetch(path, { headers: { 'Content-Type': 'application/json' }, ...options })
  const data = await response.json().catch(() => ({}))
  if (!response.ok || data.ok === false) {
    const error = new Error(data.error || response.statusText)
    error.status = response.status
    error.payload = data
    throw error
  }
  return data
}

function isAuthError(error) {
  const message = String(error?.message || '').toLowerCase()
  return error?.status === 401 || error?.status === 403 || message.includes('not authenticated') || message.includes('permission')
}

function leaveAdminAfterAuthFailure(message = '管理员登录已失效，已返回主页') {
  authenticated.value = false
  authUser.value = ''
  authRole.value = ''
  adminTab.value = 'overview'
  view.value = 'home'
  notify(message, true)
  if (!requireLogin.value) loadHomeFeed()
}

function updatedTimestamp(item) {
  const text = (item && item.updated) || ''
  const match = text.match(/(\d{4})\/(\d{2})\/(\d{2})/)
  if (!match) return 0
  return new Date(Number(match[1]), Number(match[2]) - 1, Number(match[3])).getTime()
}

function normalizeHomeItem(item) {
  return {
    id: item.id,
    title: item.title,
    cover_url: item.cover_url,
    summary: item.updated ? `${item.day_label || '本季'} · ${item.updated}` : '来自 Mikan 当前季度时间表',
    updated: item.updated,
    day_label: item.day_label,
    page_url: item.page_url,
    cover_from: item.cover_from,
    source: 'mikan',
    episodes: [],
  }
}

async function loadHomeFeed() {
  homeLoading.value = true
  try {
    const data = await api('/api/mikan/schedule?cover=bangumi&limit=30')
    homeSchedule.value = data
    homeFeed.value = (data.items || [])
      .filter((item) => item.cover_from === 'bangumi')
      .map(normalizeHomeItem)
      .sort((a, b) => updatedTimestamp(b) - updatedTimestamp(a))
      .slice(0, 30)
    heroIndex.value = 0
    startHeroAutoPlay()
  } catch (error) {
    notify('首页新番加载失败：' + error.message, true)
  } finally {
    homeLoading.value = false
  }
}

function selectHero(index) {
  heroIndex.value = index
  startHeroAutoPlay()
}

function startHeroAutoPlay() {
  clearInterval(heroTimer)
  if (heroSlides.value.length <= 1) return
  heroTimer = setInterval(() => {
    heroIndex.value = (heroIndex.value + 1) % heroSlides.value.length
  }, 5200)
}

async function loadLibrary(refresh = false) {
  loading.value = true
  try {
    const data = await api('/api/library' + (refresh ? '?refresh=1' : ''))
    library.value = data.bangumi || []
    if (selected.value) selected.value = library.value.find((item) => item.id === selected.value.id) || selected.value
  } catch (error) {
    notify('媒体库加载失败：' + error.message, true)
  } finally {
    loading.value = false
  }
}


async function loadDiscover(section = discoverSection.value, tag = discoverTag.value, append = false) {
  discoverSection.value = section
  if (tag !== undefined) discoverTag.value = tag
  discovering.value = true
  const titles = { rank: 'Bangumi 排行榜', browse: 'Bangumi 分类浏览' }
  discoverTitle.value = titles[section] || 'Bangumi'
  const offset = append ? discoverOffset.value : 0
  try {
    const params = new URLSearchParams({ section, tag: discoverTag.value || '', limit: String(discoverLimit), offset: String(offset) })
    const data = await api('/api/bangumi/discover?' + params.toString())
    const items = data.subjects || []
    discover.value = append ? [...discover.value, ...items] : items
    discoverOffset.value = offset + items.length
    discoverHasMore.value = Boolean(data.has_more)
  } catch (error) {
    notify('Bangumi 加载失败：' + error.message, true)
  } finally {
    discovering.value = false
  }
}

async function loadMoreDiscover() {
  await loadDiscover(discoverSection.value, discoverTag.value, true)
}

async function openDiscover(section, tag = '') {
  view.value = 'discover'
  await loadDiscover(section, tag)
  window.scrollTo({ top: 0, behavior: 'smooth' })
}

async function loadSchedule() {
  scheduleLoading.value = true
  try {
    schedule.value = await api('/api/mikan/schedule')
  } catch (error) {
    notify('Mikan 新番时间表加载失败：' + error.message, true)
  } finally {
    scheduleLoading.value = false
  }
}

async function openSchedule() {
  view.value = 'schedule'
  await loadSchedule()
  window.scrollTo({ top: 0, behavior: 'smooth' })
}

function openLibrary() {
  view.value = 'library'
  loadLibrary()
  window.scrollTo({ top: 0, behavior: 'smooth' })
}

async function subscribeSubject(subject) {
  subscribing.value = String(subject.id)
  try {
    const data = await api('/api/mikan/subscribe', { method: 'POST', body: JSON.stringify({ subject_id: subject.id, title: subject.title, cover_url: subject.cover_url, summary: subject.summary, language: 0 }) })
    notify(data.message || '订阅成功：' + subject.title)
  } catch (error) {
    notify('订阅失败：' + error.message, true)
  } finally {
    subscribing.value = ''
  }
}

async function scrapeSubject(subject) {
  query.value = subject.title
  await search()
}

function activeNav(name) {
  if (name === 'install') return view.value === 'install'
  if (name === 'home') return view.value === 'home'
  if (name === 'library') return view.value === 'library' || view.value === 'detail' || view.value === 'player'
  if (name === 'rank') return view.value === 'discover' && discoverSection.value === 'rank'
  if (name === 'browse') return view.value === 'discover' && discoverSection.value === 'browse'
  if (name === 'schedule') return view.value === 'schedule'
  if (name === 'search') return view.value === 'search'
  if (name === 'history') return view.value === 'history'
  if (name === 'admin') return view.value === 'admin'
  return false
}

function fillInstallForm(config = {}) {
  installForm.value = {
    ...installForm.value,
    ...config,
    pikpak_auth_mode: config.pikpak_auth_mode || installForm.value.pikpak_auth_mode || 'password',
    storage_provider: config.storage_provider || installForm.value.storage_provider || 'pikpak',
    drive115_root_cid: config.drive115_root_cid || installForm.value.drive115_root_cid || '0',
    aria2_rpc_url: config.aria2_rpc_url || installForm.value.aria2_rpc_url || 'http://127.0.0.1:6800/jsonrpc',
    local_storage_path: config.local_storage_path || installForm.value.local_storage_path || 'downloads',
    mysql_port: Number(config.mysql_port || installForm.value.mysql_port || 3306),
    redis_db: Number(config.redis_db || installForm.value.redis_db || 0),
    enable_proxy: Boolean(config.enable_proxy),
  }
}

function isDockerInstallConfig(config = {}) {
  const mysqlHost = String(config.mysql_host || '').trim().toLowerCase()
  const redisAddr = String(config.redis_addr || '').trim().toLowerCase()
  const configPath = String(installPath.value || '').trim().toLowerCase().replaceAll('\\', '/')
  return mysqlHost === 'mysql' || redisAddr === 'redis:6379' || configPath.startsWith('/app/data/')
}

async function loadInstallStatus() {
  installLoading.value = true
  try {
    const data = await api('/api/install/status')
    installPath.value = data.config_path || '本地配置数据库'
    // 安装过程中不要把全量配置校验错误直接显示在当前步骤；
    // 比如停留在 MySQL 页面时，管理员账号还没填写是正常状态。
    installError.value = ''
    installOnly.value = Boolean(data.install_only)
    installDockerMode.value = Boolean(data.docker_mode) || isDockerInstallConfig(data.config || {})
    installRequired.value = !data.installed || installOnly.value
    requireLogin.value = data.config && data.config.require_login !== undefined ? Boolean(data.config.require_login) : true
    registerStatus.value = {
      enable_registration: data.config && data.config.enable_registration !== undefined ? Boolean(data.config.enable_registration) : true,
      require_invite: data.config && data.config.require_invite !== undefined ? Boolean(data.config.require_invite) : false,
    }
    fillInstallForm(data.config || {})
    if (installRequired.value) {
      if (installDockerMode.value && installPage.value < 3) installPage.value = 3
      view.value = 'install'
    }
    return Boolean(data.installed) && !installOnly.value
  } catch (error) {
    installError.value = error.message
    installRequired.value = true
    installOnly.value = true
    view.value = 'install'
    return false
  } finally {
    installLoading.value = false
  }
}

async function loadAuthStatus() {
  try {
    const data = await api('/api/auth/me')
    authenticated.value = true
    authUser.value = data.username || ''
    authRole.value = data.role || ''
    return true
  } catch {
    authenticated.value = false
    authUser.value = ''
    authRole.value = ''
    if (view.value === 'register') return false
    if (requireLogin.value) {
      view.value = 'login'
      return false
    }
    if (view.value === 'login') view.value = 'home'
    return true
  } finally {
    authChecked.value = true
  }
}

async function scanLibrary() {
  await loadLibrary(true)
  notify('媒体库扫描完成，已从储存桶重新读取现有文件')
  if (adminTab.value === 'anime') {
    await loadAdminAnime()
    await loadAdminOverview()
  }
}

async function loadRegisterStatus() {
  try {
    const data = await api('/api/auth/register')
    registerStatus.value = {
      enable_registration: Boolean(data.enable_registration),
      require_invite: Boolean(data.require_invite),
    }
    return registerStatus.value
  } catch {
    registerStatus.value = { enable_registration: false, require_invite: false }
    return registerStatus.value
  }
}

async function login() {
  loginLoading.value = true
  try {
    const data = await api('/api/auth/login', { method: 'POST', body: JSON.stringify(loginForm.value) })
    authenticated.value = true
    authUser.value = data.username || loginForm.value.username
    authRole.value = data.role || ''
    loginForm.value.password = ''
    view.value = 'home'
    notify('登录成功')
    await loadHomeFeed()
  } catch (error) {
    notify('登录失败：' + error.message, true)
  } finally {
    loginLoading.value = false
  }
}

async function register() {
  registerLoading.value = true
  try {
    const data = await api('/api/auth/register', { method: 'POST', body: JSON.stringify(registerForm.value) })
    authenticated.value = true
    authUser.value = data.username || registerForm.value.username
    authRole.value = data.role || 'user'
    registerForm.value = { username: '', password: '', invite_code: '' }
    view.value = 'home'
    notify('注册成功，已自动登录')
    await loadHomeFeed()
  } catch (error) {
    notify('注册失败：' + error.message, true)
  } finally {
    registerLoading.value = false
  }
}

async function openRegister() {
  await loadRegisterStatus()
  view.value = 'register'
}

async function logout() {
  try {
    await api('/api/auth/logout', { method: 'POST', body: '{}' })
  } catch {
    // ignore logout network errors
  }
  authenticated.value = false
  authUser.value = ''
  authRole.value = ''
  userMenuOpen.value = false
  passwordModalOpen.value = false
  view.value = requireLogin.value ? 'login' : 'home'
}

function openPasswordModal() {
  if (!authenticated.value) {
    view.value = 'login'
    return
  }
  userMenuOpen.value = false
  passwordForm.value = { old_password: '', new_password: '', confirm_password: '' }
  passwordModalOpen.value = true
}

function closePasswordModal() {
  if (passwordSaving.value) return
  passwordModalOpen.value = false
  passwordForm.value = { old_password: '', new_password: '', confirm_password: '' }
}

async function changePassword() {
  passwordSaving.value = true
  try {
    const data = await api('/api/auth/password', { method: 'POST', body: JSON.stringify(passwordForm.value) })
    notify(data.message || '密码已修改，请重新登录')
    authenticated.value = false
    authUser.value = ''
    authRole.value = ''
    passwordModalOpen.value = false
    passwordForm.value = { old_password: '', new_password: '', confirm_password: '' }
    view.value = 'login'
  } catch (error) {
    notify('修改密码失败：' + error.message, true)
  } finally {
    passwordSaving.value = false
  }
}

function openInstall() {
  installDone.value = false
  if (installDockerMode.value && installPage.value < 3) installPage.value = 3
  view.value = 'install'
  loadInstallStatus()
  window.scrollTo({ top: 0, behavior: 'smooth' })
}

function installPayload() {
  return {
    ...installForm.value,
    mysql_port: Number(installForm.value.mysql_port || 0),
    redis_db: Number(installForm.value.redis_db || 0),
    enable_proxy: Boolean(installForm.value.enable_proxy),
  }
}

function nextInstallPage() {
  installError.value = ''
  if (installPage.value < 4) {
    installPage.value += 1
    if (installPage.value === 4) {
      setTimeout(() => {
        if (!installSaving.value && !installDone.value) saveInstall()
      }, 120)
    }
  }
  window.scrollTo({ top: 0, behavior: 'smooth' })
}

function prevInstallPage() {
  installError.value = ''
  const minPage = installDockerMode.value ? 3 : 1
  if (installPage.value > minPage) installPage.value -= 1
  window.scrollTo({ top: 0, behavior: 'smooth' })
}

function sleep(ms) {
  return new Promise((resolve) => setTimeout(resolve, ms))
}

async function saveInstall() {
  installPage.value = 4
  installSaving.value = true
  installDone.value = false
  installWriteStage.value = installWriteSteps[0].title
  try {
    for (const step of installWriteSteps.slice(0, 3)) {
      installWriteStage.value = step.title
      await sleep(180)
    }
    const data = await api('/api/install', { method: 'POST', body: JSON.stringify(installPayload()) })
    installWriteStage.value = installWriteSteps[3].title
    await sleep(180)
    installDone.value = true
    installError.value = ''
    installOnly.value = Boolean(data.install_only)
    installRequired.value = Boolean(data.install_only)
    installPage.value = 5
    notify(data.message || '配置已写入本地数据库，安装已生效')
  } catch (error) {
    installError.value = error.message
    notify('安装配置保存失败：' + error.message, true)
  } finally {
    installSaving.value = false
  }
}

async function testInstallConnection(kind) {
  installTesting.value = kind
  installTestResults.value = { ...installTestResults.value, [kind]: null }
  try {
    const endpoint = kind === 'mysql' ? '/api/install/test/mysql' : '/api/install/test/redis'
    const data = await api(endpoint, { method: 'POST', body: JSON.stringify(installPayload()) })
    installTestResults.value = {
      ...installTestResults.value,
      [kind]: { ok: true, message: data.message || '连接成功', duration: data.duration_ms },
    }
  } catch (error) {
    installTestResults.value = {
      ...installTestResults.value,
      [kind]: { ok: false, message: error.message },
    }
  } finally {
    installTesting.value = ''
  }
}

function openDetail(item) {
  selected.value = { episodes: [], ...item }
  selectEpisode(selected.value.episodes && selected.value.episodes[0])
  pushHistory(item, currentEpisode.value)
  view.value = 'detail'
  window.scrollTo({ top: 0, behavior: 'smooth' })
}

function firstPlayableEpisode(item) {
  return (item?.episodes || []).find((entry) => entry && entry.files && entry.files.length)
}

async function openPlayer(episode = currentEpisode.value) {
  const selectedID = selected.value && selected.value.id
  const selectedTitle = selected.value && selected.value.title
  const wantedEpisodeID = episode?.id || currentEpisode.value?.id
  if (episode) selectEpisode(episode)
  if (selected.value && !currentFile.value) {
    const playableEpisode = firstPlayableEpisode(selected.value)
    if (playableEpisode) selectEpisode(playableEpisode)
  }
  if (selected.value && !currentFile.value && isAdmin.value) {
    await loadLibrary(true)
    const refreshed = library.value.find((item) => item.id === selectedID || item.title === selectedTitle)
    if (refreshed) {
      selected.value = refreshed
      const refreshedEpisode = (refreshed.episodes || []).find((item) => item.id === wantedEpisodeID) || firstPlayableEpisode(refreshed)
      if (refreshedEpisode) selectEpisode(refreshedEpisode)
    }
  }
  if (!selected.value || !currentFile.value) {
    notify('该剧集还没有可播放视频文件，请等待储存桶离线下载完成后刷新媒体库', true)
    return
  }
  pushHistory(selected.value, currentEpisode.value, currentFile.value)
  view.value = 'player'
  window.scrollTo({ top: 0, behavior: 'smooth' })
}

function goHome() {
  if (installRequired.value) {
    view.value = 'install'
    notify(installOnly.value ? '安装已完成，请继续登录' : '请先完成安装')
    return
  }
  if (!authenticated.value && requireLogin.value) {
    view.value = 'login'
    return
  }
  view.value = 'home'
}

function openHistory() {
  loadHistory()
  view.value = 'history'
  window.scrollTo({ top: 0, behavior: 'smooth' })
}

function openAdmin(tab = 'overview') {
  if (!isAdmin.value) {
    notify('只有管理员可以进入管理员面板', true)
    return
  }
  adminTab.value = tab
  view.value = 'admin'
  loadAdminPanel(tab)
  window.scrollTo({ top: 0, behavior: 'smooth' })
}

async function loadAdminPanel(tab = adminTab.value) {
  adminLoading.value = true
  try {
    if (tab === 'overview') await loadAdminOverview()
    if (tab === 'users') {
      await loadAdminUsers()
      await loadAdminInvites()
    }
    if (tab === 'anime') {
      if (adminAnimeTab.value === 'requests') await loadAdminDownloadRequests()
      else await loadAdminAnime()
    }
    if (tab === 'logs') await loadAdminLogs()
    if (tab === 'monitor') await loadAdminMonitor()
    if (tab === 'config' || tab === 'storage') await loadAdminConfig()
  } catch (error) {
    if (isAuthError(error)) {
      leaveAdminAfterAuthFailure(error.status === 403 ? '当前账号不是管理员，已返回主页' : '管理员登录已失效，已返回主页')
      return
    }
    notify('管理员面板加载失败：' + error.message, true)
  } finally {
    adminLoading.value = false
  }
}

async function switchAdminTab(tab) {
  adminTab.value = tab
  await loadAdminPanel(tab)
}

async function switchAdminUserTab(tab) {
  adminUserTab.value = tab
  if (adminTab.value !== 'users') {
    adminTab.value = 'users'
  }
  await loadAdminPanel('users')
}

async function switchAdminAnimeTab(tab) {
  adminAnimeTab.value = tab
  if (adminTab.value !== 'anime') {
    adminTab.value = 'anime'
  }
  await loadAdminPanel('anime')
}

async function loadAdminOverview() {
  const data = await api('/api/admin/overview')
  adminOverview.value = data
}

async function loadAdminUsers() {
  const data = await api('/api/users')
  adminUsers.value = data.users || []
}

async function loadAdminInvites() {
  const data = await api('/api/admin/invite-codes')
  adminInvites.value = data.codes || []
}

async function generateInviteCodes(payload = {}) {
  adminInviteLoading.value = true
  try {
    const data = await api('/api/admin/invite-codes', { method: 'POST', body: JSON.stringify(payload) })
    adminInvites.value = data.codes || []
    notify('已生成 ' + ((data.codes_created && data.codes_created.length) || 0) + ' 个邀请码')
  } catch (error) {
    notify('生成邀请码失败：' + error.message, true)
  } finally {
    adminInviteLoading.value = false
  }
}

async function deleteInviteCodes(codes = []) {
  if (!codes.length) return
  adminInviteLoading.value = true
  try {
    const data = await api('/api/admin/invite-codes', { method: 'DELETE', body: JSON.stringify({ codes }) })
    adminInvites.value = data.codes || []
    notify('已删除 ' + (data.deleted || 0) + ' 个邀请码')
  } catch (error) {
    notify('删除邀请码失败：' + error.message, true)
  } finally {
    adminInviteLoading.value = false
  }
}

async function saveAdminUser() {
  adminSavingUser.value = true
  try {
    const data = await api('/api/users', { method: 'POST', body: JSON.stringify(adminUserForm.value) })
    notify('用户已保存：' + data.username)
    adminUserForm.value = { username: '', password: '', role: 'user' }
    await loadAdminUsers()
  } catch (error) {
    notify('保存用户失败：' + error.message, true)
  } finally {
    adminSavingUser.value = false
  }
}

async function loadAdminAnime() {
  const data = await api('/api/admin/anime')
  adminAnime.value = data.items || []
}

async function loadAdminDownloadRequests() {
  const data = await api('/api/admin/download-requests')
  adminDownloadRequests.value = data.items || []
}

async function approveDownloadRequest(id, action = 'approve') {
  adminLoading.value = true
  try {
    const data = await api('/api/admin/download-requests', { method: 'POST', body: JSON.stringify({ id, action }) })
    adminDownloadRequests.value = data.items || []
    notify(data.message || '下载申请已处理')
    if (action === 'approve') setTimeout(() => loadLibrary(), 1600)
  } catch (error) {
    notify('处理下载申请失败：' + error.message, true)
  } finally {
    adminLoading.value = false
  }
}

async function deleteAdminAnime(ids = [], options = {}) {
  if (!ids.length) return
  adminLoading.value = true
  try {
    const selectedIDs = new Set(ids)
    const titles = adminAnime.value
      .filter((item) => selectedIDs.has(item.id || item.title))
      .map((item) => item.title)
      .filter(Boolean)
    const data = await api('/api/admin/anime', { method: 'DELETE', body: JSON.stringify({ ids, titles, delete_files: Boolean(options.delete_files) }) })
    notify('已删除 ' + (data.deleted || 0) + ' 部番剧' + (data.deleted_files ? '，并已删除储存桶文件' : ''))
    await loadAdminAnime()
    await loadAdminOverview()
    library.value = library.value.filter((item) => !selectedIDs.has(item.id || item.title))
    if (selected.value && (ids.includes(selected.value.id) || titles.includes(selected.value.title))) selected.value = null
  } catch (error) {
    notify('删除番剧失败：' + error.message, true)
  } finally {
    adminLoading.value = false
  }
}

async function deleteAdminEpisode(item, episode, options = {}) {
  if (!item || !episode) return
  adminLoading.value = true
  try {
    const data = await api('/api/admin/anime/episode', {
      method: 'DELETE',
      body: JSON.stringify({
        bangumi_id: item.id,
        bangumi_title: item.title,
        episode_id: episode.id,
        episode_label: episode.label,
        delete_files: Boolean(options.delete_files),
      }),
    })
    notify('已删除 ' + (data.episode_label || episode.label || '该集') + (data.deleted_files ? '，并已删除储存桶文件' : ''))
    await loadAdminAnime()
    await loadAdminOverview()
    await loadLibrary()
  } catch (error) {
    notify('删除剧集失败：' + error.message, true)
    throw error
  } finally {
    adminLoading.value = false
  }
}

async function loadAdminLogs() {
  const data = await api('/api/admin/logs')
  adminLogs.value = data.logs || []
}

async function loadAdminMonitor() {
  const data = await api('/api/admin/monitor')
  adminMonitor.value = data
}

async function loadAdminConfig() {
  const data = await api('/api/admin/config')
  adminConfigForm.value = { ...adminConfigForm.value, ...(data.config || {}) }
  adminConfigForm.value.storage_provider = adminConfigForm.value.storage_provider || 'pikpak'
  adminConfigForm.value.drive115_root_cid = adminConfigForm.value.drive115_root_cid || '0'
  adminConfigForm.value.aria2_rpc_url = adminConfigForm.value.aria2_rpc_url || 'http://127.0.0.1:6800/jsonrpc'
  adminConfigForm.value.local_storage_path = adminConfigForm.value.local_storage_path || 'downloads'
  adminConfigForm.value.user_daily_download_limit = Number(adminConfigForm.value.user_daily_download_limit ?? 3)
  requireLogin.value = adminConfigForm.value.require_login !== undefined ? Boolean(adminConfigForm.value.require_login) : true
  registerStatus.value = {
    enable_registration: adminConfigForm.value.enable_registration !== undefined ? Boolean(adminConfigForm.value.enable_registration) : true,
    require_invite: adminConfigForm.value.require_invite !== undefined ? Boolean(adminConfigForm.value.require_invite) : false,
  }
}

async function saveAdminConfig() {
  adminConfigSaving.value = true
  try {
    const data = await api('/api/admin/config', { method: 'POST', body: JSON.stringify(adminConfigForm.value) })
    adminConfigForm.value = { ...adminConfigForm.value, ...(data.config || {}) }
    adminConfigForm.value.storage_provider = adminConfigForm.value.storage_provider || 'pikpak'
    adminConfigForm.value.drive115_root_cid = adminConfigForm.value.drive115_root_cid || '0'
    adminConfigForm.value.aria2_rpc_url = adminConfigForm.value.aria2_rpc_url || 'http://127.0.0.1:6800/jsonrpc'
    adminConfigForm.value.local_storage_path = adminConfigForm.value.local_storage_path || 'downloads'
    adminConfigForm.value.user_daily_download_limit = Number(adminConfigForm.value.user_daily_download_limit ?? 3)
    requireLogin.value = adminConfigForm.value.require_login !== undefined ? Boolean(adminConfigForm.value.require_login) : true
    registerStatus.value = {
      enable_registration: adminConfigForm.value.enable_registration !== undefined ? Boolean(adminConfigForm.value.enable_registration) : true,
      require_invite: adminConfigForm.value.require_invite !== undefined ? Boolean(adminConfigForm.value.require_invite) : false,
    }
    notify(data.message || '配置已保存')
  } catch (error) {
    notify('保存配置失败：' + error.message, true)
  } finally {
    adminConfigSaving.value = false
  }
}

function formatStorage(bytes = 0) {
  const n = Number(bytes || 0)
  if (n < 1024) return n + ' B'
  const units = ['KiB', 'MiB', 'GiB', 'TiB']
  let value = n / 1024
  let index = 0
  while (value >= 1024 && index < units.length - 1) {
    value /= 1024
    index += 1
  }
  return value.toFixed(1) + ' ' + units[index]
}

function statusText(ok) {
  return ok ? '正常' : '未就绪'
}


function loadHistory() {
  try {
    historyItems.value = JSON.parse(localStorage.getItem(historyStorageKey) || '[]')
  } catch {
    historyItems.value = []
  }
}

function saveHistory() {
  localStorage.setItem(historyStorageKey, JSON.stringify(historyItems.value.slice(0, 60)))
}

function pushHistory(item, episode = null, file = null) {
  if (!item || !item.id) return
  loadHistory()
  const record = {
    id: item.id,
    title: item.title,
    cover_url: item.cover_url,
    summary: item.summary,
    episode_id: episode && episode.id,
    episode_label: episode && episode.label,
    file_id: file && file.id,
    file_name: file && file.name,
    watched_at: new Date().toISOString()
  }
  historyItems.value = [record, ...historyItems.value.filter((old) => !(old.id === record.id && old.episode_id === record.episode_id))].slice(0, 60)
  saveHistory()
}

function clearHistory() {
  historyItems.value = []
  localStorage.removeItem(historyStorageKey)
  notify('历史记录已清空')
}

async function openHistoryItem(record) {
  if (!library.value.length) await loadLibrary()
  const item = library.value.find((entry) => entry.id === record.id || entry.title === record.title)
  if (!item) {
    notify('这个番剧暂时不在当前媒体库里，请先刷新媒体库', true)
    return
  }
  selected.value = item
  const episode = item.episodes && item.episodes.find((ep) => ep.id === record.episode_id || ep.label === record.episode_label)
  selectEpisode(episode || (item.episodes && item.episodes[0]))
  view.value = 'detail'
  window.scrollTo({ top: 0, behavior: 'smooth' })
}

function selectEpisode(episode) {
  currentEpisode.value = episode || null
  currentFile.value = episode && episode.files && episode.files.length ? episode.files[0] : null
}

async function search() {
  if (!query.value.trim()) return
  searching.value = true
  searchQuery.value = query.value.trim()
  view.value = 'search'
  window.scrollTo({ top: 0, behavior: 'smooth' })
  try {
    const data = await api('/api/search?q=' + encodeURIComponent(searchQuery.value))
    results.value = data.results || []
    notify(results.value.length ? '找到 ' + results.value.length + ' 条发布' : '没有找到结果')
  } catch (error) {
    notify('搜索失败：' + error.message, true)
  } finally {
    searching.value = false
  }
}

async function download(result) {
  if (!authenticated.value) {
    notify('请登录后下载', true)
    return
  }
  const key = result.torrent_url || result.magnet || result.title
  downloading.value = key
  downloadStatuses.value = { ...downloadStatuses.value, [key]: 'submitting' }
  try {
    const endpoint = isAdmin.value ? '/api/download' : '/api/download/request'
    const data = await api(endpoint, { method: 'POST', body: JSON.stringify(result) })
    notify(data.message || (isAdmin.value ? '已提交：' : '申请已提交：') + (result.bangumi_title || result.title))
    downloadStatuses.value = { ...downloadStatuses.value, [key]: isAdmin.value ? 'downloaded' : 'requested' }
    if (isAdmin.value) setTimeout(() => loadLibrary(), 1600)
  } catch (error) {
    downloadStatuses.value = { ...downloadStatuses.value, [key]: 'failed' }
    notify((isAdmin.value ? '提交失败：' : '申请失败：') + error.message, true)
  } finally {
    downloading.value = ''
  }
}

async function syncRSS() {
  syncing.value = true
  try {
    const data = await api('/api/sync', { method: 'POST', body: '{}' })
    notify(data.message || '同步完成')
    await loadLibrary()
  } catch (error) {
    notify('RSS 搜刮失败：' + error.message, true)
  } finally {
    syncing.value = false
  }
}

function notify(message, isError = false) {
  toast.value = message
  toastError.value = isError
  clearTimeout(toastTimer)
  toastTimer = setTimeout(() => { toast.value = ''; toastError.value = false }, 3600)
}

function coverFor(item) {
  return (item && item.cover_url) || placeholder((item && item.title) || 'Bangumi')
}

function placeholder(text) {
  return 'https://api.dicebear.com/8.x/shapes/svg?seed=' + encodeURIComponent(text || 'Bangumi') + '&backgroundColor=181a22,2a2146,6f4bd8'
}

function formatHistoryTime(value) {
  if (!value) return ''
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return ''
  return date.toLocaleString()
}

function openRecentPlayed(item) {
  if (item && item.historyRecord) {
    openHistoryItem(item.historyRecord)
    return
  }
  openDetail(item)
}

provide('animeX', {
  adminTab, adminUserTab, adminAnimeTab, adminTabTitle, adminLoading, loadAdminPanel, switchAdminTab, switchAdminUserTab, switchAdminAnimeTab,
  adminOverview, formatStorage, adminUsers, adminSavingUser, adminUserForm, saveAdminUser,
  adminInvites, adminInviteLoading, generateInviteCodes, deleteInviteCodes,
  adminAnime, adminDownloadRequests, approveDownloadRequest, placeholder, adminLogs, adminConfigForm, saveAdminConfig, adminConfigSaving, deleteAdminAnime, deleteAdminEpisode,
  adminMonitor, statusText, goHome, authUser, logout, openPasswordModal,
  installSteps, installActiveStep, installDone, installLoading, installPage, installForm,
  installTesting, installTestResults, testInstallConnection, installWriteStage, installWriteSteps,
  installProgress, installError, installPath, installSaving, saveInstall, prevInstallPage,
  nextInstallPage, toast, toastError, installDockerMode,
  login, loginForm, loginLoading, register, registerForm, registerLoading, registerStatus, loadRegisterStatus, openRegister,
  view, featured, coverFor, openDetail, scrapeSubject, heroSlides, heroIndex, selectHero,
  recentPlayedItems, openHistory, openRecentPlayed, newItems, homeSchedule, openSchedule,
  loadLibrary, scanLibrary, loading, library, isAdmin, authenticated,
  discoverTitle, loadDiscover, discovering, discoverSection, discoverTag, openDiscover, discover,
  loadMoreDiscover, discoverHasMore,
  searchQuery, search, searching, results, downloading, downloadStatuses, download,
  schedule, scheduleLoading, loadSchedule, subscribing, subscribeSubject,
  historyItems, clearHistory, openHistoryItem,
  selected, openPlayer, currentEpisode, selectEpisode,
  currentFile,
})

watch(() => route.path, (path) => syncViewFromRoute(path))
watch(view, (nextView) => syncRouteFromView(nextView))

onMounted(() => {
  syncViewFromRoute()
  loadHistory()
  loadInstallStatus().then((installed) => {
    if (installed) {
      loadRegisterStatus()
      loadAuthStatus().then((ok) => {
        if (ok || !requireLogin.value) loadHomeFeed()
      })
    }
  })
})

watch(installPage, (page) => {
  if (page === 4 && !installSaving.value && !installDone.value) {
    saveInstall()
  }
})

onUnmounted(() => {
  clearInterval(heroTimer)
})
</script>

<template>
  <RouterView v-if="view === 'install' || view === 'login' || view === 'register'" />

  <div v-else class="app-frame" :class="{ 'admin-only-frame': view === 'admin' }">
    <aside v-if="view !== 'admin'" class="sidebar">
      <div class="logo"><strong class="brand-wordmark">Anime<span>X</span></strong><button>≡</button></div>
      <nav class="nav-list">
        <button class="nav-item" :class="{ active: activeNav('home') }" @click="goHome"><span>⌂</span>首页</button>
        <button class="nav-item" :class="{ active: activeNav('library') }" @click="openLibrary"><span>▣</span>媒体库</button>
        <button class="nav-item" :class="{ active: activeNav('rank') }" @click="openDiscover('rank')"><span>♜</span>排行榜</button>
        <button class="nav-item" :class="{ active: activeNav('browse') }" @click="openDiscover('browse', discoverTag)"><span>♧</span>分类浏览</button>
        <button class="nav-item" :class="{ active: activeNav('schedule') }" @click="openSchedule"><span>▤</span>新番时间表</button>
        <button v-if="isAdmin" class="nav-item" :class="{ active: activeNav('admin') }" @click="openAdmin()"><span>⚙</span>管理员面板</button>
      </nav>
      <nav class="nav-list secondary">
        <button class="nav-item" :class="{ active: activeNav('history') }" @click="openHistory"><span>◴</span>历史记录</button>
      </nav>
      <div class="sidebar-user" :class="{ open: userMenuOpen }" title="当前登录用户" @click="userMenuOpen = !userMenuOpen">
        <span class="sidebar-avatar">{{ (authUser || 'A').slice(0, 1).toUpperCase() }}</span>
        <span class="sidebar-user-meta">
          <strong>{{ authUser || '游客' }} <i>{{ isAdmin ? 'LV4' : (authenticated ? 'LV1' : 'GUEST') }}</i></strong>
          <small>{{ isAdmin ? '管理员' : (authenticated ? '普通用户' : '未登录') }}</small>
        </span>
        <span class="sidebar-user-arrow">⌄</span>
        <div v-if="userMenuOpen" class="sidebar-user-menu" @click.stop>
          <button v-if="authenticated" type="button" @click="openPasswordModal">修改密码</button>
          <button v-if="authenticated" type="button" class="danger" @click="logout">退出登录</button>
          <button v-else type="button" @click="view = 'login'; userMenuOpen = false">登录账号</button>
        </div>
      </div>
    </aside>

    <main class="main-shell">
      <header v-if="view !== 'admin'" class="topbar">
        <button class="hamburger" @click="goHome"><span class="desktop-menu">☰</span><span class="mobile-wordmark brand-wordmark">Anime<span>X</span></span></button>
        <form class="search" @submit.prevent="search"><span>⌕</span><input v-model="query" placeholder="搜索番剧、角色、制作人员..." /><button :disabled="searching">{{ searching ? '搜索中' : '' }}</button></form>
        <div class="topbar-actions">
          <button v-if="!authenticated" class="ghost small-login" @click="view = 'login'">登录</button>
        </div>
      </header>

      <RouterView />
    </main>

    <nav v-if="view !== 'admin'" class="mobile-nav" aria-label="移动端导航">
      <button :class="{ active: activeNav('home') }" @click="goHome"><span>⌂</span><b>首页</b></button>
      <button :class="{ active: activeNav('library') }" @click="openLibrary"><span>▣</span><b>媒体库</b></button>
      <button :class="{ active: activeNav('rank') }" @click="openDiscover('rank')"><span>♜</span><b>排行</b></button>
      <button :class="{ active: activeNav('browse') }" @click="openDiscover('browse', discoverTag)"><span>♧</span><b>分类</b></button>
      <button :class="{ active: activeNav('schedule') }" @click="openSchedule"><span>▤</span><b>新番</b></button>
      <button :class="{ active: activeNav('history') }" @click="openHistory"><span>◴</span><b>历史</b></button>
      <button v-if="isAdmin" :class="{ active: activeNav('admin') }" @click="openAdmin()"><span>⚙</span><b>管理</b></button>
    </nav>

    <div v-if="toast" class="toast" :class="{ error: toastError }">{{ toast }}</div>

    <div v-if="passwordModalOpen" class="modal-backdrop" @click.self="closePasswordModal">
      <form class="account-modal" @submit.prevent="changePassword">
        <div class="modal-head">
          <div>
            <span class="modal-kicker">Account</span>
            <h3>修改密码</h3>
          </div>
          <button type="button" class="modal-close" @click="closePasswordModal">×</button>
        </div>
        <label><span>当前密码</span><input v-model="passwordForm.old_password" type="password" autocomplete="current-password" placeholder="请输入当前密码" /></label>
        <label><span>新密码</span><input v-model="passwordForm.new_password" type="password" autocomplete="new-password" placeholder="至少 6 位密码" /></label>
        <label><span>确认新密码</span><input v-model="passwordForm.confirm_password" type="password" autocomplete="new-password" placeholder="再次输入新密码" /></label>
        <p>修改成功后会自动退出登录，请使用新密码重新登录。</p>
        <div class="modal-actions">
          <button type="button" class="ghost" :disabled="passwordSaving" @click="closePasswordModal">取消</button>
          <button type="submit" class="primary" :disabled="passwordSaving">{{ passwordSaving ? '保存中...' : '保存新密码' }}</button>
        </div>
      </form>
    </div>
  </div>
</template>
