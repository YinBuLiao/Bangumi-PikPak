<script setup>
import { inject } from 'vue'

const ctx = inject('animeX')
if (!ctx) throw new Error('AnimeX context is missing')

const {
  installSteps, installActiveStep, installDone, installLoading, installPage, installForm,
  installTesting, installTestResults, testInstallConnection, installWriteStage, installWriteSteps,
  installProgress, installError, installPath, installSaving, saveInstall, prevInstallPage,
  nextInstallPage, goHome, toast, toastError, installDockerMode,
} = ctx
</script>

<template>
  <div class="install-frame">
    <aside class="install-sidebar">
      <div>
      <div class="install-logo brand-wordmark">Anime<span>X</span></div>
        <p class="install-subtitle">番剧在线搭建系统</p>
      </div>
      <div class="install-stepper">
        <div v-for="step in installSteps" :key="step.page" class="install-step" :class="{ active: step.page === installActiveStep, done: step.page < installActiveStep }">
          <b>{{ step.n }}</b>
          <span><strong>{{ step.title }}</strong><small>{{ step.desc }}</small></span>
        </div>
      </div>
      <div class="install-mascot">
        <div class="install-cube"><span>▶</span></div>
        <strong>AnimeX 安装向导</strong>
        <small>简单 · 快速 · 安全</small>
      </div>
      <em>v0.1.0</em>
    </aside>

    <main class="install-main">
      <section class="install-head">
        <h1>安装 <span>AnimeX</span></h1>
        <p>欢迎使用 AnimeX 番剧在线搭建系统安装向导</p>
      </section>

      <div class="install-tip">
        <b>ⓘ 提示</b>
        <span>{{ installDockerMode ? '检测到 Docker 部署，数据库与 Redis 已使用容器预设，直接创建管理员账号即可。' : '请先确保您已创建好数据库和Redis，并拥有相应的访问权限。' }}</span>
      </div>

      <form class="install-panel" @submit.prevent="saveInstall">
        <section v-if="installPage === 1" class="install-card install-page-card">
          <div class="install-card-head">
            <h2>
              <span class="mysql-icon" aria-hidden="true">
                <svg viewBox="0 0 24 24"><ellipse cx="12" cy="5" rx="8" ry="3.2"/><path d="M4 5v6c0 1.8 3.6 3.2 8 3.2s8-1.4 8-3.2V5"/><path d="M4 11v6c0 1.8 3.6 3.2 8 3.2s8-1.4 8-3.2v-6"/></svg>
              </span>
              MySQL 配置信息
            </h2>
            <button class="install-test-btn" type="button" :disabled="installTesting === 'mysql'" @click="testInstallConnection('mysql')">{{ installTesting === 'mysql' ? '测试中...' : '测试 MySQL 连接' }}</button>
          </div>
          <div class="install-grid">
            <label><span>数据库主机</span><input v-model.trim="installForm.mysql_host" placeholder="127.0.0.1" /><small>通常为 127.0.0.1</small></label>
            <label><span>端口</span><input v-model.number="installForm.mysql_port" type="number" min="1" placeholder="3306" /><small>MySQL 服务端口，默认 3306</small></label>
            <label><span>数据库名</span><input v-model.trim="installForm.mysql_database" placeholder="anime" /><small>用于保存 AnimeX 数据</small></label>
            <label><span>用户名</span><input v-model.trim="installForm.mysql_username" placeholder="anime" /><small>拥有该数据库权限的用户名</small></label>
            <label><span>密码</span><input v-model="installForm.mysql_password" type="password" placeholder="数据库密码" /><small>该用户的密码</small></label>
            <label><span>DSN（可选）</span><input v-model.trim="installForm.mysql_dsn" placeholder="留空时自动根据上方生成" /><small>高级连接串，填写后优先使用</small></label>
          </div>
          <div v-if="installTestResults.mysql" class="install-test-result" :class="{ ok: installTestResults.mysql.ok, error: !installTestResults.mysql.ok }">
            {{ installTestResults.mysql.message }}<span v-if="installTestResults.mysql.ok && installTestResults.mysql.duration !== undefined"> · {{ installTestResults.mysql.duration }}ms</span>
          </div>
        </section>

        <section v-else-if="installPage === 2" class="install-card install-page-card">
          <div class="install-card-head">
            <h2>
              <span class="redis-proxy-icon" aria-hidden="true">
                <svg viewBox="0 0 24 24"><path d="M12 3 3.5 7.5 12 12l8.5-4.5L12 3Z"/><path d="m3.5 12 8.5 4.5 8.5-4.5"/><path d="m3.5 16.5 8.5 4.5 8.5-4.5"/></svg>
              </span>
              Redis 与代理配置
            </h2>
            <button class="install-test-btn" type="button" :disabled="installTesting === 'redis'" @click="testInstallConnection('redis')">{{ installTesting === 'redis' ? '测试中...' : '测试 Redis 连接' }}</button>
          </div>
          <div class="install-grid">
            <label><span>Redis 地址</span><input v-model.trim="installForm.redis_addr" placeholder="127.0.0.1:6379" /><small>Redis 服务地址</small></label>
            <label><span>Redis 密码（可选）</span><input v-model="installForm.redis_password" type="password" placeholder="未设置请留空" /><small>Redis 访问密码</small></label>
            <label><span>数据库索引</span><input v-model.number="installForm.redis_db" type="number" min="0" placeholder="0" /><small>Redis DB，默认 0</small></label>
            <label class="install-switch"><span>启用代理</span><input v-model="installForm.enable_proxy" type="checkbox" /><small>打开后使用下方代理地址</small></label>
            <label><span>HTTP_PROXY</span><input v-model.trim="installForm.http_proxy" placeholder="http://127.0.0.1:7890" /><small>HTTP 请求代理</small></label>
            <label><span>HTTPS_PROXY</span><input v-model.trim="installForm.https_proxy" placeholder="http://127.0.0.1:7890" /><small>HTTPS 请求代理</small></label>
            <label><span>SOCKS_PROXY</span><input v-model.trim="installForm.socks_proxy" placeholder="socks5://127.0.0.1:7890" /><small>SOCKS5 请求代理</small></label>
          </div>
          <div v-if="installTestResults.redis" class="install-test-result" :class="{ ok: installTestResults.redis.ok, error: !installTestResults.redis.ok }">
            {{ installTestResults.redis.message }}<span v-if="installTestResults.redis.ok && installTestResults.redis.duration !== undefined"> · {{ installTestResults.redis.duration }}ms</span>
          </div>
        </section>

        <section v-else-if="installPage === 3" class="install-card install-page-card">
          <div class="install-card-head">
            <h2>
              <span class="admin-icon" aria-hidden="true">
                <svg viewBox="0 0 24 24"><path d="M12 12a4 4 0 1 0-4-4 4 4 0 0 0 4 4Z"/><path d="M4.5 21a7.5 7.5 0 0 1 15 0"/><path d="m16.5 13.5 1.5 1.5 2.5-2.5"/></svg>
              </span>
              管理员账号
            </h2>
          </div>
          <div class="install-grid">
            <label><span>管理员账号</span><input v-model.trim="installForm.admin_username" autocomplete="username" placeholder="admin" /><small>用于登录 AnimeX 后台</small></label>
            <label><span>登录密码</span><input v-model="installForm.admin_password" type="password" autocomplete="new-password" placeholder="请输入管理员密码" /><small>请设置一个不容易被猜到的密码</small></label>
          </div>
        </section>

        <section v-else-if="installPage === 4" class="install-write-flow install-page-card">
          <h2>程序安装中...</h2>
          <div class="install-progress">
            <span :style="{ width: installProgress + '%' }"></span>
          </div>
          <div class="install-progress-meta">
            <strong>{{ installProgress }}%</strong>
            <small>{{ installWriteStage || '正在准备安装' }}</small>
          </div>
          <div class="sql-flow-list">
            <article v-for="(stage, index) in installWriteSteps" :key="stage.title" class="sql-flow-row" :class="{ active: installWriteStage === stage.title, done: installDone || installWriteSteps.findIndex((item) => item.title === installWriteStage) > index }">
              <i>{{ index + 1 }}</i>
              <div>
                <strong>{{ stage.title }}</strong>
                <code>{{ stage.sql }}</code>
              </div>
            </article>
          </div>
        </section>

        <section v-else class="install-complete-page install-page-card">
          <div class="complete-mark">✓</div>
          <h2>安装配置已写入</h2>
          <p>配置已写入本地数据库 <b>{{ installPath }}</b>，当前程序已热加载新配置，可以直接进入系统登录。</p>
        </section>

        <div v-if="installError" class="install-error">{{ installError }}</div>
        <div v-if="installDone && installPage !== 5" class="install-success">
          配置已写入本地数据库 <b>{{ installPath }}</b>，当前程序已热加载新配置。
        </div>

        <footer class="install-actions">
          <button v-if="installPage > (installDockerMode ? 3 : 1) && installPage < 5" class="ghost" type="button" :disabled="installSaving" @click="prevInstallPage">上一步</button>
          <button v-if="installPage < 4" class="primary" type="button" @click="nextInstallPage">下一步 →</button>
          <button v-else-if="installPage === 4" class="primary" type="button" disabled>{{ installSaving ? '安装中...' : '正在安装...' }}</button>
          <button v-else class="primary" type="button" @click="goHome">完成</button>
        </footer>
      </form>

      <p class="install-copy">© 2024 AnimeX. All rights reserved.</p>
      <div v-if="toast" class="toast" :class="{ error: toastError }">{{ toast }}</div>
    </main>
  </div>
</template>
