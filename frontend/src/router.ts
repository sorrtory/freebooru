import { createRouter, createWebHistory } from 'vue-router'

import AppShell from './components/AppShell.vue'
import CollectionOverviewPage from './pages/CollectionOverviewPage.vue'
import CollectionsPage from './pages/CollectionsPage.vue'
import FilesPage from './pages/FilesPage.vue'
import FileDetailPage from './pages/FileDetailPage.vue'
import ImportPage from './pages/ImportPage.vue'
import SettingsPage from './pages/SettingsPage.vue'
import StoragePage from './pages/StoragePage.vue'
import TagsPage from './pages/TagsPage.vue'
import StatusPage from './pages/StatusPage.vue'

export default createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', component: StatusPage },
    { path: '/collections', component: CollectionsPage },
    { path: '/settings', component: SettingsPage },
    { path: '/collections/:collection', component: AppShell, props: true, children: [
      { path: '', component: CollectionOverviewPage, props: true },
      { path: 'files', component: FilesPage, props: true },
      { path: 'files/:sha256', component: FileDetailPage, props: true },
      { path: 'search', redirect: (to) => ({ path: `/collections/${String(to.params.collection)}/files`, query: to.query }) },
      { path: 'tags', component: TagsPage, props: true },
      { path: 'storage', component: StoragePage, props: true },
      { path: 'import', component: ImportPage, props: true },
    ] },
  ],
})
