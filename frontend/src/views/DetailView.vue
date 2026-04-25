<script setup>
import { inject } from 'vue'

const ctx = inject('animeX')
if (!ctx) throw new Error('AnimeX context is missing')
const { goHome, selected, coverFor, openPlayer, scrapeSubject, isAdmin, currentEpisode, selectEpisode } = ctx

function episodeText(episode) {
  const label = String(episode?.label || '').trim()
  const match = label.match(/\d+/)
  return match ? String(Number(match[0])) : (label || '剧集')
}
</script>

<template>
  <section v-if="selected" class="detail-screen">
    <div class="crumb"><button @click="goHome">‹</button><span>媒体库</span><b>›</b><span>番剧详情</span></div>
    <section class="detail-hero" :style="{ backgroundImage: `linear-gradient(90deg, rgba(17,19,27,.96), rgba(17,19,27,.72)), url(${coverFor(selected)})` }">
      <img class="detail-poster" :src="coverFor(selected)" :alt="selected.title || '番剧详情'" />
      <div class="detail-info"><h1>{{ selected.title || '未选择番剧' }}</h1><div class="tags"><span>{{ selected.source === 'mikan' ? 'Mikan' : 'Bangumi' }}</span><span>{{ selected.day_label || selected.air_date || '动画' }}</span><span>{{ selected.updated || '详情' }}</span></div><p class="rating" v-if="selected.score">☻ {{ selected.score }} ★★★★★ <small>Rank #{{ selected.rank || '—' }}</small></p><p>{{ selected.summary || selected.updated || '暂无简介，可以先搜刮 Mikan 发布并下载到储存桶。' }}</p><div class="detail-actions"><button v-if="selected.episodes && selected.episodes.length" class="primary" @click="openPlayer()">▶ 播放</button><button class="primary" @click="scrapeSubject(selected)">⌕ {{ isAdmin ? '搜刮下载' : '搜索发布' }}</button><button class="ghost" @click="goHome">返回首页</button></div></div>
    </section>
    <section v-if="selected.episodes && selected.episodes.length" class="episodes"><div class="section-title"><h2>剧集</h2><span>共{{ selected.episodes.length }}集</span></div><div class="episode-grid"><button v-for="episode in selected.episodes" :key="episode.id" :class="{ active: currentEpisode && currentEpisode.id === episode.id }" @click="selectEpisode(episode)"><span>{{ episodeText(episode) }}</span></button></div></section>
    <section v-else class="episodes"><div class="empty-state">这个条目还不在媒体库中。点击“搜刮下载”从 Mikan 选择发布并下载到储存桶。</div></section>
  </section>
  <section v-else class="detail-screen">
    <div class="crumb"><button @click="goHome">‹</button><span>媒体库</span><b>›</b><span>番剧详情</span></div>
    <div class="empty-state"><h2>未选择番剧</h2><p>请从首页、媒体库或番剧列表进入详情页。</p><button class="primary" @click="goHome">返回首页</button></div>
  </section>
</template>
