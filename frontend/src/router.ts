import { createRouter, createWebHistory } from 'vue-router'

import ImportPage from './pages/ImportPage.vue'
import StatusPage from './pages/StatusPage.vue'

export default createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', component: StatusPage },
    {
      path: '/collections/:collection/import',
      component: ImportPage,
      props: (route) => ({ collection: String(route.params.collection) }),
    },
  ],
})
