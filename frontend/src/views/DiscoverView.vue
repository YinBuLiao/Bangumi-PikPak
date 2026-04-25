<script setup>
import { inject } from 'vue'

const ctx = inject('animeX')
if (!ctx) throw new Error('AnimeX context is missing')
const { goHome, discoverTitle, loadDiscover, discovering, discoverSection, discoverTag, openDiscover, discover, placeholder, openDetail, scrapeSubject, loadMoreDiscover, discoverHasMore, isAdmin } = ctx
</script>

<template>
  <section class="discover-screen">
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
        <button class="primary small" @click.stop="scrapeSubject(subject)">{{ isAdmin ? '搜刮 Mikan 下载' : '搜索 Mikan 发布' }}</button>
      </div>
    </article>
  </div>
  <div class="load-more-wrap" v-if="discover.length">
    <button class="ghost" :disabled="discovering || !discoverHasMore" @click="loadMoreDiscover">{{ discoverHasMore ? (discovering ? '加载中' : '加载更多') : '没有更多了' }}</button>
  </div>
</section>
</template>
