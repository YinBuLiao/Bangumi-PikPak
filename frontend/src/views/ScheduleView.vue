<script setup>
import { inject } from 'vue'

const ctx = inject('animeX')
if (!ctx) throw new Error('AnimeX context is missing')
const { goHome, schedule, scheduleLoading, loadSchedule, placeholder, openDetail, subscribing, subscribeSubject, isAdmin } = ctx
</script>

<template>
  <section class="schedule-screen">
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
              <button v-if="isAdmin" class="primary small" :disabled="subscribing === String(item.id)" @click.stop="subscribeSubject({ id: item.id, title: item.title, cover_url: item.cover_url, summary: '' })">{{ subscribing === String(item.id) ? '订阅中' : '订阅' }}</button><span v-else class="role-hint">只读</span>
            </div>
          </div>
        </article>
      </div>
    </section>
  </div>
</section>
</template>
