<script setup>
import { inject } from 'vue'

const ctx = inject('animeX')
if (!ctx) throw new Error('AnimeX context is missing')
const { goHome, historyItems, clearHistory, placeholder, openHistoryItem } = ctx
</script>

<template>
  <section class="history-screen">
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
</template>
