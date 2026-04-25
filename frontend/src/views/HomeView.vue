<script setup>
import { inject } from 'vue'

const ctx = inject('animeX')
if (!ctx) throw new Error('AnimeX context is missing')
const { featured, coverFor, openDetail, scrapeSubject, heroSlides, heroIndex, selectHero, recentPlayedItems, openHistory, openRecentPlayed, newItems, isAdmin, homeSchedule, openSchedule } = ctx
</script>

<template>
  <section class="home-screen">
  <section class="hero-banner">
    <div class="hero-blur" :style="{ backgroundImage: `url(${coverFor(featured)})` }" />
    <div class="hero-art" :style="{ backgroundImage: `url(${coverFor(featured)})` }" />
    <div class="hero-content">
      <span class="badge">Mikan 最近更新</span>
      <h1>{{ featured ? featured.title : '最近更新的新番' }}</h1>
      <p>{{ featured && featured.summary ? featured.summary : '按 Mikan 当前季度更新时间自动推荐，不依赖 PikPak 媒体库。' }}</p>
      <div class="hero-actions"><button class="primary" @click="featured && openDetail(featured)">⊙ 查看详情</button><button class="ghost" @click="featured && scrapeSubject(featured)">⌕ {{ isAdmin ? '搜刮下载' : '搜索发布' }}</button></div>
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
</template>
