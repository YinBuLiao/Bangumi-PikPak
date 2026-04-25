<script setup>
import { inject, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import Player from 'xgplayer'
import 'xgplayer/dist/index.min.css'

const ctx = inject('animeX')
if (!ctx) throw new Error('AnimeX context is missing')
const { view, selected, currentEpisode, currentFile, coverFor } = ctx

const playerEl = ref(null)
let xgPlayer = null

function destroyPlayer() {
  if (xgPlayer) {
    xgPlayer.destroy()
    xgPlayer = null
  }
}

async function setupPlayer() {
  if (view.value !== 'player' || !currentFile.value || !currentFile.value.stream_url) {
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

onMounted(setupPlayer)
watch([view, currentFile, playerEl], setupPlayer, { flush: 'post' })
onUnmounted(destroyPlayer)
</script>

<template>
  <section v-if="selected && currentEpisode && currentFile" class="player-screen">
    <div class="crumb"><button @click="view = 'detail'">‹</button><span>媒体库</span><b>›</b><span>{{ selected.title || '番剧' }}</span><b>›</b><strong>{{ currentEpisode.label || '剧集' }}</strong></div>
    <div class="video-wrap"><div ref="playerEl" class="xgplayer-host"></div></div>
    <div class="player-meta"><div><h1>{{ selected.title || '番剧' }} {{ currentEpisode.label || '' }}</h1><div class="player-links"><button>☆ 收藏</button><button>⌯ 分享</button></div></div><div class="next-actions"><button class="ghost">← 上一集</button><button class="primary">下一集 →</button></div></div>
    <div class="comment-bar"><span>弹幕</span><input placeholder="发送弹幕..." /><button class="primary small">发送</button></div>
  </section>
  <section v-else class="player-screen">
    <div class="empty-state"><h2>没有可播放内容</h2><p>请先从详情页选择一个包含视频文件的剧集。</p><button class="primary" @click="view = selected ? 'detail' : 'home'">返回</button></div>
  </section>
</template>
