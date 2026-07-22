<script setup lang="ts">
defineProps<{ collection: string }>()
</script>
<template><main class="page"><header class="page-heading"><p>{{ collection }}</p><h1>All files</h1></header><p class="empty-copy">Files will appear here.</p></main></template>
