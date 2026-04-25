<script setup>
import { inject } from 'vue'

const ctx = inject('animeX')
if (!ctx) throw new Error('AnimeX context is missing')
const { goHome, scanLibrary, loading, library, coverFor, openDetail, isAdmin } = ctx
</script>

<template>
  <section class="library-screen">
  <div class="crumb"><button @click="goHome">‹</button><span>PikPak</span><b>›</b><span>媒体库</span></div>
  <div class="page-heading"><div><span class="badge">PikPak 缓存媒体库</span><h1>媒体库</h1><p>默认从 MySQL 快照读取；需要更新时点击扫描，只读取储存桶现有文件，不会重新下载。</p></div><button v-if="isAdmin" class="primary" @click="scanLibrary" :disabled="loading">{{ loading ? '扫描中' : '扫描媒体库（不下载）' }}</button></div>
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
</template>
