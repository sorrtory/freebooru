import { createRouter, createWebHistory } from 'vue-router'

import AppShell from './components/AppShell.vue'
import CollectionOverviewPage from './pages/CollectionOverviewPage.vue'
import CollectionsPage from './pages/CollectionsPage.vue'
import FilesPage from './pages/FilesPage.vue'
import ImportPage from './pages/ImportPage.vue'
import SearchPage from './pages/SearchPage.vue'
import StatusPage from './pages/StatusPage.vue'

export default createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', component: StatusPage },
    { path: '/collections', component: CollectionsPage },
    { path: '/collections/:collection', component: AppShell, props: true, children: [
      { path: '', component: CollectionOverviewPage, props: true },
      { path: 'files', component: FilesPage, props: true },
      { path: 'search', component: SearchPage, props: true },
      { path: 'import', component: ImportPage, props: true },
    ] },
  ],
})
