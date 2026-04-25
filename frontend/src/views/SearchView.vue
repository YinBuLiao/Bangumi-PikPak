<script setup>
import { inject } from 'vue'

const ctx = inject('animeX')
if (!ctx) throw new Error('AnimeX context is missing')
const { goHome, searchQuery, search, searching, results, placeholder, downloading, downloadStatuses, download, isAdmin, authenticated } = ctx

function resultKey(result) {
  return result.torrent_url || result.magnet || result.title
}

function resultStatus(result) {
  return downloadStatuses.value[resultKey(result)] || ''
}

function resultButtonText(result) {
  const status = resultStatus(result)
  if (status === 'submitting') return isAdmin.value ? '提交中' : '申请中'
  if (status === 'downloaded') return '已下载'
  if (status === 'failed') return '下载失败'
  if (status === 'requested') return '已申请'
  return isAdmin.value ? '下载到网盘' : '申请下载'
}

function resultButtonDisabled(result) {
  const status = resultStatus(result)
  return status === 'submitting' || status === 'downloaded' || status === 'requested'
}
</script>

<template>
  <section class="search-screen">
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
      <button v-if="isAdmin" class="primary small download-state-btn" :class="resultStatus(result)" :disabled="downloading === resultKey(result) || resultButtonDisabled(result)" @click="download(result)">{{ resultButtonText(result) }}</button>
      <button v-else-if="authenticated" class="primary small download-state-btn" :class="resultStatus(result)" :disabled="downloading === resultKey(result) || resultButtonDisabled(result)" @click="download(result)">{{ resultButtonText(result) }}</button>
      <span v-else class="role-hint">请登录后下载</span>
    </article>
  </div>
</section>
</template>
