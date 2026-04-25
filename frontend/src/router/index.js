import { createRouter, createWebHistory } from 'vue-router'
import MainLayout from '../layouts/MainLayout.vue'
import InstallWizard from '../components/InstallWizard.vue'
import LoginView from '../components/LoginView.vue'
import RegisterView from '../components/RegisterView.vue'
import AdminPanel from '../components/AdminPanel.vue'
import HomeView from '../views/HomeView.vue'
import LibraryView from '../views/LibraryView.vue'
import DiscoverView from '../views/DiscoverView.vue'
import SearchView from '../views/SearchView.vue'
import ScheduleView from '../views/ScheduleView.vue'
import HistoryView from '../views/HistoryView.vue'
import DetailView from '../views/DetailView.vue'
import PlayerView from '../views/PlayerView.vue'

const routes = [
  {
    path: '/',
    component: MainLayout,
    children: [
      { path: '', name: 'home', component: HomeView },
      { path: 'install', name: 'install', component: InstallWizard },
      { path: 'login', name: 'login', component: LoginView },
      { path: 'register', name: 'register', component: RegisterView },
      { path: 'admin', name: 'admin', component: AdminPanel },
      { path: 'library', name: 'library', component: LibraryView },
      { path: 'discover', name: 'discover', component: DiscoverView },
      { path: 'search', name: 'search', component: SearchView },
      { path: 'schedule', name: 'schedule', component: ScheduleView },
      { path: 'history', name: 'history', component: HistoryView },
      { path: 'detail', name: 'detail', component: DetailView },
      { path: 'player', name: 'player', component: PlayerView },
    ],
  },
  { path: '/:pathMatch(.*)*', redirect: '/' },
]

export default createRouter({
  history: createWebHistory(),
  routes,
})
