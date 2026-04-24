<script setup>
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import Player from 'xgplayer'
import 'xgplayer/dist/index.min.css'

const view = ref('home')
const library = ref([])
const selected = ref(null)
const currentEpisode = ref(null)
const currentFile = ref(null)
const playerEl = ref(null)
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
const toast = ref('')
const toastError = ref(false)
let toastTimer = 0
let heroTimer = 0
let xgPlayer = null
const historyStorageKey = 'animex:watch-history'

const featured = computed(() => homeFeed.value[heroIndex.value % Math.max(homeFeed.value.length, 1)] || null)
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
  if (!response.ok || data.ok === false) throw new Error(data.error || response.statusText)
  return data
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
  if (name === 'home') return view.value === 'home'
  if (name === 'library') return view.value === 'library' || view.value === 'detail' || view.value === 'player'
  if (name === 'rank') return view.value === 'discover' && discoverSection.value === 'rank'
  if (name === 'browse') return view.value === 'discover' && discoverSection.value === 'browse'
  if (name === 'schedule') return view.value === 'schedule'
  if (name === 'search') return view.value === 'search'
  if (name === 'history') return view.value === 'history'
  return false
}

function openDetail(item) {
  selected.value = { episodes: [], ...item }
  selectEpisode(selected.value.episodes && selected.value.episodes[0])
  pushHistory(item, currentEpisode.value)
  view.value = 'detail'
  window.scrollTo({ top: 0, behavior: 'smooth' })
}

function openPlayer(episode = currentEpisode.value) {
  if (episode) selectEpisode(episode)
  if (!selected.value || !currentFile.value) return
  pushHistory(selected.value, currentEpisode.value, currentFile.value)
  view.value = 'player'
  window.scrollTo({ top: 0, behavior: 'smooth' })
}

function goHome() { view.value = 'home' }

function openHistory() {
  loadHistory()
  view.value = 'history'
  window.scrollTo({ top: 0, behavior: 'smooth' })
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
  downloading.value = result.torrent_url
  try {
    await api('/api/download', { method: 'POST', body: JSON.stringify(result) })
    notify('已提交：' + (result.bangumi_title || result.title))
    setTimeout(() => loadLibrary(), 1600)
  } catch (error) {
    notify('提交失败：' + error.message, true)
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

function destroyPlayer() {
  if (xgPlayer) {
    xgPlayer.destroy()
    xgPlayer = null
  }
}

async function setupPlayer() {
  if (view.value !== 'player' || !currentFile.value) {
    destroyPlayer()
    return
  }
  await nextTick()
  if (!playerEl.value) return
  destroyPlayer()
  xgPlayer = new Player({
    el: playerEl.value,
    url: currentFile.value.stream_url,
    poster: coverFor(selected.value),
    width: '100%',
    height: '100%',
    autoplay: true,
    playsinline: true,
    pip: true,
    fluid: true,
    lang: 'zh-cn',
    videoInit: true,
    closeVideoClick: false,
  })
}

onMounted(() => {
  loadHistory()
  loadHomeFeed()
})

watch([view, currentFile], setupPlayer)

onUnmounted(() => {
  clearInterval(heroTimer)
  destroyPlayer()
})
</script>

<template>
  <div class="app-frame">
    <aside class="sidebar">
      <div class="logo">Anime<span>X</span><button>≡</button></div>
      <nav class="nav-list">
        <button class="nav-item" :class="{ active: activeNav('home') }" @click="goHome"><span>⌂</span>首页</button>
        <button class="nav-item" :class="{ active: activeNav('library') }" @click="openLibrary"><span>▣</span>媒体库</button>
        <button class="nav-item" :class="{ active: activeNav('rank') }" @click="openDiscover('rank')"><span>♜</span>排行榜</button>
        <button class="nav-item" :class="{ active: activeNav('browse') }" @click="openDiscover('browse', discoverTag)"><span>♧</span>分类浏览</button>
        <button class="nav-item" :class="{ active: activeNav('schedule') }" @click="openSchedule"><span>▤</span>新番时间表</button>
      </nav>
      <nav class="nav-list secondary">
        <button class="nav-item" :class="{ active: activeNav('history') }" @click="openHistory"><span>◴</span>历史记录</button>
      </nav>
    </aside>

    <main class="main-shell">
      <header class="topbar">
        <button class="hamburger" @click="goHome"><span class="desktop-menu">☰</span><span class="mobile-wordmark">Anime<b>X</b></span></button>
        <form class="search" @submit.prevent="search"><span>⌕</span><input v-model="query" placeholder="搜索番剧、角色、制作人员..." /><button :disabled="searching">{{ searching ? '搜索中' : '' }}</button></form>
      </header>

      <section v-if="view === 'home'" class="home-screen">
        <section class="hero-banner">
          <div class="hero-blur" :style="{ backgroundImage: `url(${coverFor(featured)})` }" />
          <div class="hero-art" :style="{ backgroundImage: `url(${coverFor(featured)})` }" />
          <div class="hero-content">
            <span class="badge">Mikan 最近更新</span>
            <h1>{{ featured ? featured.title : '最近更新的新番' }}</h1>
            <p>{{ featured && featured.summary ? featured.summary : '按 Mikan 当前季度更新时间自动推荐，不依赖 PikPak 媒体库。' }}</p>
            <div class="hero-actions"><button class="primary" @click="featured && openDetail(featured)">⊙ 查看详情</button><button class="ghost" @click="featured && scrapeSubject(featured)">⌕ 搜刮下载</button></div>
          </div>
          <div class="slider-dots">
            <button v-for="(_, index) in heroSlides" :key="index" :class="{ active: index === heroIndex }" @click="selectHero(index)" />
          </div>
        </section>

        <section class="rail-section">
          <div class="section-title"><h2>最近播放</h2><button @click="openHistory">查看历史 ›</button></div>
          <div v-if="recentPlayedItems.length" class="continue-grid poster-grid-main">
            <article v-for="item in recentPlayedItems" :key="item.id + '-recent-' + (item.episode_id || item.updated)" class="watch-card" @click="openRecentPlayed(item)">
              <img :src="coverFor(item)" :alt="item.title" />
              <h3>{{ item.title }}</h3><p>{{ item.updated || item.day_label || 'Mikan 时间表' }}</p>
            </article>
          </div>
          <div v-else class="empty-state compact-empty">还没有最近播放记录，进入媒体库选择剧集后会自动显示在这里。</div>
        </section>

        <section class="rail-section">
          <div class="section-title"><h2>{{ homeSchedule.year || '当前' }} {{ homeSchedule.season || '' }}季新番</h2><button @click="openSchedule">查看全部 ›</button></div>
          <div class="poster-row poster-grid-main">
            <article v-for="(item, index) in newItems" :key="item.id + '-n-' + index" class="poster-card" @click="openDetail(item)"><img :src="coverFor(item)" :alt="item.title" /><strong>{{ item.title }}</strong><small>{{ item.day_label || 'Mikan' }} · {{ item.updated || '时间表' }}</small></article>
          </div>
        </section>
      </section>

      
      <section v-else-if="view === 'library'" class="library-screen">
        <div class="crumb"><button @click="goHome">‹</button><span>PikPak</span><b>›</b><span>媒体库</span></div>
        <div class="page-heading"><div><span class="badge">PikPak 缓存媒体库</span><h1>媒体库</h1><p>默认从 MySQL 快照读取，避免频繁触发 PikPak 登录验证码；需要更新时再手动重新扫描。</p></div><button class="primary" @click="loadLibrary(true)" :disabled="loading">{{ loading ? '扫描中' : '重新扫描 PikPak' }}</button></div>
        <div v-if="!library.length" class="empty-state">媒体库为空或还在加载 PikPak 目录。</div>
        <div class="library-grid">
          <article v-for="item in library" :key="item.id" class="library-card" @click="openDetail(item)">
            <img :src="coverFor(item)" :alt="item.title" />
            <strong>{{ item.title }}</strong>
            <small>{{ item.episodes.length }} 个分组 · PikPak</small>
            <p>{{ item.summary || 'Bangumi 简介同步中' }}</p>
          </article>
        </div>
      </section>

      <section v-else-if="view === 'discover'" class="discover-screen">
        <div class="crumb"><button @click="goHome">‹</button><span>Bangumi</span><b>›</b><span>{{ discoverTitle }}</span></div>
        <div class="page-heading discover-heading"><div><span class="badge">Bangumi.tv</span><h1>{{ discoverTitle }}</h1><p>从 Bangumi 获取封面、简介、评分和放送信息；老番入口会跳到 Mikan 搜索发布并一键下载。</p></div><button class="primary" @click="loadDiscover()" :disabled="discovering">{{ discovering ? '加载中' : '刷新' }}</button></div>
        <div v-if="discoverSection === 'browse'" class="tag-strip">
          <button v-for="tag in ['恋爱','奇幻','校园','日常','科幻','百合','漫画改','原创']" :key="tag" :class="{ active: discoverTag === tag }" @click="openDiscover('browse', tag)">{{ tag }}</button>
        </div>
        <div v-if="discovering" class="empty-state">正在连接 Bangumi.tv...</div>
        <div class="discover-grid">
          <article v-for="subject in discover" :key="subject.id" class="subject-card" @click="openDetail({ ...subject, cover_url: subject.cover_url, episodes: [], source: 'bangumi' })">
            <img :src="subject.cover_url || placeholder(subject.title)" :alt="subject.title" />
            <div class="subject-body">
              <strong>{{ subject.title }}</strong>
              <small>Rank #{{ subject.rank || '—' }} · ★ {{ subject.score || '—' }} · {{ subject.air_date || '未知日期' }}</small>
              <p>{{ subject.summary || '暂无简介' }}</p>
              <button class="primary small" @click.stop="scrapeSubject(subject)">搜刮 Mikan 下载</button>
            </div>
          </article>
        </div>
        <div class="load-more-wrap" v-if="discover.length">
          <button class="ghost" :disabled="discovering || !discoverHasMore" @click="loadMoreDiscover">{{ discoverHasMore ? (discovering ? '加载中' : '加载更多') : '没有更多了' }}</button>
        </div>
      </section>

      <section v-else-if="view === 'search'" class="search-screen">
        <div class="crumb"><button @click="goHome">‹</button><span>Mikan</span><b>›</b><span>搜索结果</span></div>
        <div class="page-heading search-heading">
          <div><span class="badge">Mikan Search</span><h1>搜索：{{ searchQuery }}</h1><p>这里显示 Mikan 发布结果，选择合适字幕组/清晰度后可直接提交到 PikPak 网盘。</p></div>
          <button class="primary" @click="search" :disabled="searching">{{ searching ? '搜索中' : '重新搜索' }}</button>
        </div>
        <div v-if="searching" class="empty-state">正在搜索 Mikan 发布...</div>
        <div v-else-if="!results.length" class="empty-state">没有找到发布结果，请换一个关键词。</div>
        <div v-else class="search-results search-page-results">
          <article v-for="result in results" :key="result.torrent_url || result.magnet" class="result-row">
            <img :src="result.cover_url || placeholder(result.bangumi_title || result.title)" :alt="result.bangumi_title || result.title" />
            <div><strong>{{ result.title }}</strong><small>{{ result.bangumi_title || '待解析番剧名' }} · {{ result.episode_label || '未知集数' }} · {{ result.size || '未知大小' }} · {{ result.updated || '未知时间' }}</small><p v-if="result.summary">{{ result.summary }}</p></div>
            <button class="primary small" :disabled="downloading === result.torrent_url" @click="download(result)">{{ downloading === result.torrent_url ? '提交中' : '下载到网盘' }}</button>
          </article>
        </div>
      </section>

      <section v-else-if="view === 'schedule'" class="schedule-screen">
        <div class="crumb"><button @click="goHome">‹</button><span>Mikan</span><b>›</b><span>新番时间表</span></div>
        <div class="page-heading schedule-heading">
          <div>
            <span class="badge">Mikan Project</span>
            <h1>{{ schedule.year || '当前' }} {{ schedule.season || '' }}季新番时间表</h1>
            <p>直接从 Mikan 拉取当前季度按星期分组的新番列表，封面、更新时间和 Mikan 番剧页都使用 Mikan 数据。</p>
          </div>
          <button class="primary" @click="loadSchedule" :disabled="scheduleLoading">{{ scheduleLoading ? '刷新中' : '刷新时间表' }}</button>
        </div>
        <div v-if="scheduleLoading" class="empty-state">正在连接 Mikan Project...</div>
        <div v-else class="schedule-days">
          <section v-for="day in schedule.days" :key="day.weekday" class="schedule-day">
            <div class="section-title"><h2>{{ day.label }}</h2><span>{{ day.items.length }} 部</span></div>
            <div class="schedule-strip">
              <article v-for="item in day.items" :key="item.id || item.title" class="schedule-card" @click="openDetail({ ...item, episodes: [], source: 'mikan' })">
                <img :src="item.cover_url || placeholder(item.title)" :alt="item.title" />
                <div>
                  <strong>{{ item.title }}</strong>
                  <small>{{ item.updated || '暂无更新记录' }}</small>
                  <div class="schedule-actions">
                    <a :href="item.page_url" target="_blank" rel="noreferrer">Mikan 详情</a>
                    <button class="primary small" :disabled="subscribing === String(item.id)" @click.stop="subscribeSubject({ id: item.id, title: item.title, cover_url: item.cover_url, summary: '' })">{{ subscribing === String(item.id) ? '订阅中' : '订阅' }}</button>
                  </div>
                </div>
              </article>
            </div>
          </section>
        </div>
      </section>

      <section v-else-if="view === 'history'" class="history-screen">
        <div class="crumb"><button @click="goHome">‹</button><span>本地</span><b>›</b><span>历史记录</span></div>
        <div class="page-heading history-heading">
          <div><span class="badge">LocalStorage</span><h1>历史记录</h1><p>历史记录只保存在当前浏览器本地，不上传服务器；点击条目可回到媒体库中的番剧详情。</p></div>
          <button class="ghost" @click="clearHistory" :disabled="!historyItems.length">清空历史</button>
        </div>
        <div v-if="!historyItems.length" class="empty-state">还没有历史记录。打开番剧详情或播放剧集后会自动保存。</div>
        <div v-else class="history-list">
          <article v-for="record in historyItems" :key="record.id + '-' + (record.episode_id || 'detail') + '-' + record.watched_at" class="history-row" @click="openHistoryItem(record)">
            <img :src="record.cover_url || placeholder(record.title)" :alt="record.title" />
            <div>
              <strong>{{ record.title }}</strong>
              <small>{{ record.episode_label || '番剧详情' }} · {{ new Date(record.watched_at).toLocaleString() }}</small>
              <p>{{ record.file_name || record.summary || '点击继续浏览' }}</p>
            </div>
            <button class="primary small">打开</button>
          </article>
        </div>
      </section>

      <section v-else-if="view === 'detail'" class="detail-screen">
        <div class="crumb"><button @click="goHome">‹</button><span>媒体库</span><b>›</b><span>番剧详情</span></div>
        <section class="detail-hero" :style="{ backgroundImage: `linear-gradient(90deg, rgba(17,19,27,.96), rgba(17,19,27,.72)), url(${coverFor(selected)})` }">
          <img class="detail-poster" :src="coverFor(selected)" :alt="selected.title" />
          <div class="detail-info"><h1>{{ selected.title }}</h1><div class="tags"><span>{{ selected.source === 'mikan' ? 'Mikan' : 'Bangumi' }}</span><span>{{ selected.day_label || selected.air_date || '动画' }}</span><span>{{ selected.updated || '详情' }}</span></div><p class="rating" v-if="selected.score">☻ {{ selected.score }} ★★★★★ <small>Rank #{{ selected.rank || '—' }}</small></p><p>{{ selected.summary || selected.updated || '暂无简介，可以先搜刮 Mikan 发布并下载到 PikPak。' }}</p><div class="detail-actions"><button v-if="selected.episodes && selected.episodes.length" class="primary" @click="openPlayer()">▶ 播放</button><button class="primary" @click="scrapeSubject(selected)">⌕ 搜刮下载</button><button class="ghost" @click="goHome">返回首页</button></div></div>
        </section>
        <div class="tabs"><button class="active">详情</button><button v-if="selected.episodes && selected.episodes.length">剧集</button></div>
        <section v-if="selected.episodes && selected.episodes.length" class="episodes"><div class="section-title"><h2>剧集</h2><span>全{{ selected.episodes.length }}集</span></div><div class="episode-grid"><button v-for="(episode, index) in selected.episodes" :key="episode.id" :class="{ active: currentEpisode && currentEpisode.id === episode.id }" @click="selectEpisode(episode)"><span>{{ index + 1 }}</span><i v-if="index % 3 === 1">♛</i></button></div></section>
        <section v-else class="episodes"><div class="empty-state">这个条目还不在 PikPak 媒体库中。点击“搜刮下载”从 Mikan 选择发布并下载到网盘。</div></section>
      </section>

      <section v-else class="player-screen">
        <div class="crumb"><button @click="view = 'detail'">‹</button><span>媒体库</span><b>›</b><span>{{ selected.title }}</span><b>›</b><strong>{{ currentEpisode.label }}</strong></div>
        <div class="video-wrap"><div v-if="currentFile" ref="playerEl" class="xgplayer-host"></div></div>
        <div class="player-meta"><div><h1>{{ selected.title }} {{ currentEpisode.label }}</h1><div class="player-links"><button>☆ 收藏</button><button>⌯ 分享</button></div></div><div class="next-actions"><button class="ghost">← 上一集</button><button class="primary">下一集 →</button></div></div>
        <div class="comment-bar"><span>弹幕</span><input placeholder="发送弹幕..." /><button class="primary small">发送</button></div>
      </section>
    </main>

    <nav class="mobile-nav" aria-label="移动端导航">
      <button :class="{ active: activeNav('home') }" @click="goHome"><span>⌂</span><b>首页</b></button>
      <button :class="{ active: activeNav('library') }" @click="openLibrary"><span>▣</span><b>媒体库</b></button>
      <button :class="{ active: activeNav('rank') }" @click="openDiscover('rank')"><span>♜</span><b>排行</b></button>
      <button :class="{ active: activeNav('browse') }" @click="openDiscover('browse', discoverTag)"><span>♧</span><b>分类</b></button>
      <button :class="{ active: activeNav('schedule') }" @click="openSchedule"><span>▤</span><b>新番</b></button>
      <button :class="{ active: activeNav('history') }" @click="openHistory"><span>◴</span><b>历史</b></button>
    </nav>

    <div v-if="toast" class="toast" :class="{ error: toastError }">{{ toast }}</div>
  </div>
</template>
