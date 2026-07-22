<script setup lang="ts">
import { onBeforeUnmount, onMounted } from 'vue'

import FileResults from '../components/FileResults.vue'
import { useFileSearch } from '../useFileSearch'

const props = defineProps<{ collection: string }>()
const { page, loading, errorMessage, load, cancel } = useFileSearch(props.collection)
function loadPage(offset = 0) { void load([], offset) }
onMounted(() => loadPage())
onBeforeUnmount(cancel)
</script>

<template>
  <main class="page">
    <header class="page-heading"><p>{{ collection }}</p><h1>All files</h1></header>
    <FileResults :page="page" :collection="collection" :loading="loading" :error-message="errorMessage" empty-message="This collection has no files yet." :return-to="`/collections/${encodeURIComponent(collection)}/files`" @page="loadPage" @retry="loadPage(page?.offset ?? 0)" />
  </main>
</template>
