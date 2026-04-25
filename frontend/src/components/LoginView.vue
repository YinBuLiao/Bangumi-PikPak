<script setup>
import { inject } from 'vue'

const ctx = inject('animeX')
if (!ctx) throw new Error('AnimeX context is missing')

const { login, loginForm, loginLoading, toast, toastError, registerStatus, openRegister } = ctx
</script>

<template>
  <div class="login-frame">
    <form class="login-panel" @submit.prevent="login">
      <div class="install-logo brand-wordmark">Anime<span>X</span></div>
      <p>请输入账号密码进入系统</p>
      <label><span>登录账号</span><input v-model.trim="loginForm.username" autocomplete="username" placeholder="admin" /></label>
      <label><span>登录密码</span><input v-model="loginForm.password" type="password" autocomplete="current-password" placeholder="请输入密码" /></label>
      <button class="primary" type="submit" :disabled="loginLoading">{{ loginLoading ? '登录中...' : '登录' }}</button>
      <div class="login-links">
        <button v-if="registerStatus.enable_registration" type="button" class="text-link" @click="openRegister">没有账号？立即注册</button>
        <small v-else>注册已关闭，请联系管理员创建账号。</small>
      </div>
    </form>
    <div v-if="toast" class="toast" :class="{ error: toastError }">{{ toast }}</div>
  </div>
</template>
