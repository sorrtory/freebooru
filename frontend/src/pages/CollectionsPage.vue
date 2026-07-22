<script setup lang="ts">
import { computed, onMounted, shallowRef } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import ThemeMenu from '../components/ThemeMenu.vue'
import { useCollections } from '../useCollections'

const route = useRoute()
const router = useRouter()
const { collections, loading, errorMessage, load, create } = useCollections()
const name = shallowRef('')
const creating = shallowRef(false)
const createError = shallowRef('')
const showCreate = computed(() => route.query.create === '1' || collections.value.length === 0)

async function submit() {
  creating.value = true
  createError.value = ''
  try {
    const result = await create(name.value.trim())
    localStorage.setItem('freebooru.lastCollection', result.name)
    await router.push(`/collections/${encodeURIComponent(result.name)}`)
  } catch (error) {
    createError.value = error instanceof Error ? error.message : 'Unable to create collection'
  } finally {
    creating.value = false
  }
}

onMounted(() => void load())
</script>

<template>
  <main class="collections-page">
    <header><RouterLink class="brand" to="/">FreeBooru</RouterLink><ThemeMenu /></header>
    <div class="collections-content">
      <section v-if="showCreate" class="create-panel">
        <p>New collection</p><h1>Create collection</h1>
        <form @submit.prevent="submit">
          <label for="collection-name">Name</label>
          <input id="collection-name" v-model="name" required autocomplete="off" placeholder="illustrations" autofocus>
          <small>Letters, numbers, and underscores.</small>
          <p v-if="createError" role="alert">{{ createError }}</p>
          <div><button class="button button--primary" type="submit" :disabled="creating">{{ creating ? 'Creating…' : 'Create' }}</button><RouterLink v-if="collections.length" class="button" to="/collections">Cancel</RouterLink></div>
        </form>
      </section>
      <section v-else>
        <div class="list-heading"><div><p>Library</p><h1>Collections</h1></div><RouterLink class="button button--primary" to="/collections?create=1">Add collection</RouterLink></div>
        <p v-if="loading" role="status">Loading…</p><p v-else-if="errorMessage" role="alert">{{ errorMessage }}</p>
        <div class="collection-list"><RouterLink v-for="item in collections" :key="item.name" :to="`/collections/${encodeURIComponent(item.name)}`"><strong>{{ item.name }}</strong><span>{{ item.is_default ? 'Default' : '→' }}</span></RouterLink></div>
      </section>
    </div>
  </main>
</template>

<style scoped>
.collections-page { min-height: 100vh; padding: 1rem; background: radial-gradient(circle at 15% 15%, var(--primary-soft), transparent 32rem); }
header { display: flex; max-width: 72rem; align-items: center; justify-content: space-between; margin: 0 auto; }
.brand { color: var(--primary); font: 800 1.2rem var(--display); text-decoration: none; }
.collections-content { width: min(100%, 52rem); margin: clamp(3rem, 10vh, 8rem) auto 0; }
h1 { margin: .2rem 0 1.5rem; font: clamp(2.5rem, 8vw, 5rem)/1 var(--display); letter-spacing: -.05em; }
section > p, .list-heading p { margin: 0; color: var(--text-muted); font-size: .78rem; font-weight: 800; letter-spacing: .08em; text-transform: uppercase; }
.create-panel { width: min(100%, 34rem); }
form { display: grid; gap: .6rem; padding: 1.25rem; border: 1px solid var(--border); border-radius: var(--radius-lg); background: var(--surface); }
form label { font-weight: 800; }
form input { padding: .7rem .8rem; border: 1px solid var(--border); border-radius: var(--radius); background: var(--canvas); }
form small { color: var(--text-muted); }
form p { margin: .25rem 0; color: var(--danger); }
form div { display: flex; flex-wrap: wrap; gap: .6rem; margin-top: .5rem; }
.list-heading { display: flex; align-items: end; justify-content: space-between; gap: 1rem; }
.collection-list { display: grid; gap: .55rem; }
.collection-list a { display: flex; min-height: 4rem; align-items: center; justify-content: space-between; padding: .8rem 1rem; border: 1px solid var(--border); border-radius: var(--radius-lg); color: var(--text); background: var(--surface); text-decoration: none; }
.collection-list a:hover { border-color: var(--primary); }
.collection-list span { color: var(--primary); }
</style>
