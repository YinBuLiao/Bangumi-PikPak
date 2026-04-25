<script setup>
import { inject, onMounted } from 'vue'

const ctx = inject('animeX')
if (!ctx) throw new Error('AnimeX context is missing')

const {
  register, registerForm, registerLoading, registerStatus, loadRegisterStatus,
  loginForm, toast, toastError, view,
} = ctx

function backToLogin() {
  loginForm.value.username = registerForm.value.username
  view.value = 'login'
}

onMounted(loadRegisterStatus)
</script>

<template>
  <div class="login-frame">
    <form class="login-panel register-panel" @submit.prevent="register">
      <div class="install-logo brand-wordmark">Anime<span>X</span></div>
      <p>创建普通用户账号</p>

      <div v-if="!registerStatus.enable_registration" class="install-test-result fail">
        当前系统已关闭用户注册，请联系管理员在后台开启。
      </div>

      <template v-else>
        <label><span>用户名</span><input v-model.trim="registerForm.username" autocomplete="username" placeholder="请输入用户名" /></label>
        <label><span>密码</span><input v-model="registerForm.password" type="password" autocomplete="new-password" placeholder="至少 6 位密码" /></label>
        <label v-if="registerStatus.require_invite"><span>邀请码</span><input v-model.trim="registerForm.invite_code" placeholder="XXXX-XXXX-XXXX-XXXX" /></label>
        <button class="primary" type="submit" :disabled="registerLoading">{{ registerLoading ? '注册中...' : '注册并登录' }}</button>
      </template>

      <div class="login-links">
        <button type="button" class="text-link" @click="backToLogin">已有账号？返回登录</button>
      </div>
    </form>
    <div v-if="toast" class="toast" :class="{ error: toastError }">{{ toast }}</div>
  </div>
</template>
