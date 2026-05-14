<script setup lang="ts">
import { computed, onMounted, onUpdated, useTemplateRef } from 'vue'
import DOMPurify from 'dompurify'
import katex from 'katex'
import { resolveUrl } from '@/utils/url'

const props = defineProps<{ html: string }>()

const sanitized = computed(() => {
  if (!props.html) return ''
  const clean = DOMPurify.sanitize(props.html, {
    ALLOWED_TAGS: ['b', 'i', 'u', 'strong', 'em', 'p', 'br', 'ul', 'ol', 'li', 'img', 'span', 'sup', 'sub'],
    ALLOWED_ATTR: ['src', 'alt', 'data-formula', 'class'],
    ALLOW_DATA_ATTR: true,
    ALLOWED_URI_REGEXP: /^(https?:\/\/|\/uploads\/)/i,
  })
  // Rewrite relative /uploads/ src to absolute URL so the broken Vite proxy is bypassed
  return clean.replace(/src="(\/uploads\/[^"]+)"/g, (_, path) => `src="${resolveUrl(path)}"`)
})

const viewerRef = useTemplateRef<HTMLElement>('viewerRef')

function renderFormulas() {
  if (!viewerRef.value) return
  viewerRef.value.querySelectorAll<HTMLElement>('span[data-formula]').forEach((el) => {
    const latex = el.getAttribute('data-formula') || ''
    try {
      katex.render(latex, el, { throwOnError: false })
    } catch {
      el.textContent = latex
    }
  })
}

onMounted(renderFormulas)
onUpdated(renderFormulas)
</script>

<template>
  <div ref="viewerRef" class="richtext-viewer" v-html="sanitized"></div>
</template>

<style scoped>
.richtext-viewer {
  color: #f1f5f9;
  font-size: 0.9rem;
  line-height: 1.6;
}

.richtext-viewer :deep(p) { margin: 0 0 0.5rem; }
.richtext-viewer :deep(p:last-child) { margin-bottom: 0; }
.richtext-viewer :deep(ul),
.richtext-viewer :deep(ol) { margin: 0.25rem 0; padding-left: 1.5rem; }
.richtext-viewer :deep(b),
.richtext-viewer :deep(strong) { font-weight: 700; }
.richtext-viewer :deep(i),
.richtext-viewer :deep(em) { font-style: italic; }
.richtext-viewer :deep(u) { text-decoration: underline; }
.richtext-viewer :deep(img) { max-width: 100%; border-radius: 4px; margin: 0.25rem 0; }
.richtext-viewer :deep(.formula-render) { display: inline; }
</style>
