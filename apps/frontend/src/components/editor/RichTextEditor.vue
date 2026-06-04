<script setup lang="ts">
import { useEditor, EditorContent } from '@tiptap/vue-3'
import StarterKit from '@tiptap/starter-kit'
import TipTapImage from '@tiptap/extension-image'
import Underline from '@tiptap/extension-underline'
import { Node } from '@tiptap/core'
import katex from 'katex'
import { getAccessToken } from '@/api/client'


const props = defineProps<{ modelValue: string }>()
const emit = defineEmits<{ 'update:modelValue': [value: string] }>()

// ─── Custom Formula Node ────────────────────────────────────────────────────

const Formula = Node.create({
  name: 'formula',
  group: 'inline',
  inline: true,
  atom: true,

  addAttributes() {
    return {
      formula: { default: '' },
    }
  },

  parseHTML() {
    return [{ tag: 'span[data-formula]' }]
  },

  renderHTML({ HTMLAttributes }) {
    const latex = HTMLAttributes.formula || ''
    let rendered = latex
    try {
      rendered = katex.renderToString(latex, { throwOnError: false, output: 'html' })
    } catch { /* fallback to raw latex */ }
    return ['span', { 'data-formula': latex, class: 'formula-render' }, ...(rendered ? [] : [latex])]
  },

  addNodeView() {
    return ({ node }) => {
      const dom = document.createElement('span')
      dom.classList.add('formula-render')
      const latex = node.attrs.formula || ''
      try {
        katex.render(latex, dom, { throwOnError: false })
      } catch {
        dom.textContent = latex
      }
      return { dom }
    }
  },
})

// ─── Editor ─────────────────────────────────────────────────────────────────

const editor = useEditor({
  content: props.modelValue,
  extensions: [
    StarterKit,
    Underline,
    TipTapImage.configure({
      allowBase64: false,
      HTMLAttributes: { class: 'editor-image' },
    }),
    Formula,
  ],
  onUpdate: ({ editor }) => {
    emit('update:modelValue', editor.getHTML())
  },
})

// ─── Formula Dialog ─────────────────────────────────────────────────────────

// Standard max rendered width for question images.
const MAX_IMAGE_WIDTH = 800

const showFormulaDialog = ref(false)
const formulaInput = ref('')

function openFormulaDialog() {
  formulaInput.value = ''
  showFormulaDialog.value = true
}

function insertFormula() {
  if (!formulaInput.value.trim() || !editor.value) return
  editor.value.chain().focus().insertContent({
    type: 'formula',
    attrs: { formula: formulaInput.value.trim() },
  }).run()
  showFormulaDialog.value = false
}

// ─── Image Upload ───────────────────────────────────────────────────────────

const UPLOAD_URL = `${import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:8080/api/v1'}/upload`
const MAX_SIZE = 2 * 1024 * 1024

function uploadImage(file: File): Promise<string> {
  return new Promise((resolve, reject) => {
    const formData = new FormData()
    formData.append('file', file)
    fetch(UPLOAD_URL, { method: 'POST', body: formData, headers: { Authorization: `Bearer ${getAccessToken()}` } })
      .then(async res => {
        if (!res.ok) {
          const err = await res.json().catch(() => ({}))
          throw new Error(err.error || 'Upload failed')
        }
        return res.json()
      })
      .then((data: { url: string }) => resolve(data.url))
      .catch(reject)
  })
}

// Resize image on a canvas so it's no wider than maxWidth, keeping aspect ratio.
function resizeImage(file: File, maxWidth: number): Promise<Blob> {
  return new Promise((resolve, reject) => {
    const img = new Image()
    img.onload = () => {
      if (img.width <= maxWidth) {
        resolve(file)
        return
      }
      const ratio = maxWidth / img.width
      const canvas = document.createElement('canvas')
      canvas.width = maxWidth
      canvas.height = Math.round(img.height * ratio)
      const ctx = canvas.getContext('2d')!
      ctx.drawImage(img, 0, 0, canvas.width, canvas.height)
      canvas.toBlob((blob) => {
        if (blob) resolve(blob)
        else reject(new Error('Resize failed'))
      }, file.type, 0.85)
    }
    img.onerror = () => reject(new Error('Failed to load image'))
    img.src = URL.createObjectURL(file)
  })
}

function handleImageUpload() {
  const input = document.createElement('input')
  input.type = 'file'
  input.accept = 'image/png,image/jpeg,image/gif,image/webp'
  input.onchange = async () => {
    const file = input.files?.[0]
    if (!file) return
    if (file.size > MAX_SIZE) {
      alert('Image must be under 2 MB')
      return
    }
    if (!['image/png', 'image/jpeg', 'image/gif', 'image/webp'].includes(file.type)) {
      alert('Only PNG, JPEG, GIF, WebP images are allowed')
      return
    }
    try {
      const resized = await resizeImage(file, MAX_IMAGE_WIDTH)
      const url = await uploadImage(resized instanceof File ? resized : new File([resized], file.name, { type: file.type }))
      // Use relative path so the backend sanitizer accepts it.
      // RichTextViewer will rewrite to absolute at render time.
      editor.value?.chain().focus().setImage({ src: url }).run()
    } catch (e: any) {
      alert(e.message || 'Upload failed')
    }
  }
  input.click()
}

// Handle clipboard paste of images
function handlePaste(_view: any, event: ClipboardEvent) {
  const items = event.clipboardData?.items
  if (!items) return false
  for (const item of items) {
    if (item.type.startsWith('image/')) {
      event.preventDefault()
      const file = item.getAsFile()
      if (!file) continue
      if (file.size > MAX_SIZE) {
        alert('Image must be under 2 MB')
        return true
      }
      resizeImage(file, MAX_IMAGE_WIDTH).then(resized => {
        const f = resized instanceof File ? resized : new File([resized], file.name, { type: file.type })
        return uploadImage(f)
      }).then(url => {
        // Use relative path so the backend sanitizer accepts it.
        // RichTextViewer will rewrite to absolute at render time.
        editor.value?.chain().focus().setImage({ src: url }).run()
      }).catch(e => alert(e.message || 'Upload failed'))
      return true
    }
  }
  return false
}

// Wire paste handler directly on the ProseMirror view
import { ref, watch, onBeforeUnmount } from 'vue'

// Sync editor content when modelValue changes externally (e.g. switching edit targets)
watch(() => props.modelValue, (val) => {
  if (!editor.value) return
  if (editor.value.getHTML() !== val) {
    editor.value.commands.setContent(val ?? '')
  }
})

watch(() => editor.value, (ed) => {
  if (!ed) return
  const pmView = ed.view as any
  const dom = pmView?.dom as HTMLElement | undefined
  if (!dom) return
  dom.addEventListener('paste', (e: ClipboardEvent) => handlePaste(pmView, e))
})

onBeforeUnmount(() => editor.value?.destroy())
</script>

<template>
  <div class="rich-editor-wrapper">
    <div class="toolbar">
      <button type="button" class="tb-btn" title="Bold (Ctrl+B)" @click="editor?.chain().focus().toggleBold().run()"
        :class="{ active: editor?.isActive('bold') }"><b>B</b></button>
      <button type="button" class="tb-btn" title="Italic (Ctrl+I)" @click="editor?.chain().focus().toggleItalic().run()"
        :class="{ active: editor?.isActive('italic') }"><i>I</i></button>
      <button type="button" class="tb-btn" title="Underline (Ctrl+U)" @click="editor?.chain().focus().toggleUnderline().run()"
        :class="{ active: editor?.isActive('underline') }"><u>U</u></button>
      <span class="tb-sep"></span>
      <button type="button" class="tb-btn" title="Bullet List" @click="editor?.chain().focus().toggleBulletList().run()"
        :class="{ active: editor?.isActive('bulletList') }">•</button>
      <button type="button" class="tb-btn" title="Numbered List" @click="editor?.chain().focus().toggleOrderedList().run()"
        :class="{ active: editor?.isActive('orderedList') }">1.</button>
      <span class="tb-sep"></span>
      <button type="button" class="tb-btn tb-formula" title="Insert Formula" @click="openFormulaDialog">fx</button>
      <button type="button" class="tb-btn" title="Insert Image" @click="handleImageUpload">🖼</button>
    </div>
    <EditorContent :editor="editor" class="editor-content" />

    <!-- Formula Input Dialog -->
    <div v-if="showFormulaDialog" class="formula-overlay" @click.self="showFormulaDialog = false">
      <div class="formula-dialog">
        <h3>Insert Formula</h3>
        <p class="formula-hint">Enter LaTeX, e.g. <code>F(x) = x^2 - 2x + 1</code></p>
        <input
          v-model="formulaInput"
          class="formula-input"
          placeholder="F(x) = x - 1"
          @keydown.enter.prevent="insertFormula"
          @keydown.escape="showFormulaDialog = false"
          ref="formulaInputRef"
        />
        <div class="formula-preview" v-if="formulaInput.trim()">
          <span class="formula-render" v-html="
            (() => {
              try { return katex.renderToString(formulaInput.trim(), { throwOnError: false }) }
              catch { return '' }
            })()
          "></span>
        </div>
        <div class="formula-actions">
          <button type="button" class="btn-cancel" @click="showFormulaDialog = false">Cancel</button>
          <button type="button" class="btn-primary" @click="insertFormula">Insert</button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.rich-editor-wrapper {
  border: 1px solid var(--border);
  border-radius: 8px;
  background: var(--bg-surface);
  overflow: hidden;
}

.toolbar {
  display: flex;
  align-items: center;
  gap: 2px;
  padding: 0.35rem 0.5rem;
  border-bottom: 1px solid var(--border);
  background: var(--bg-input);
  flex-wrap: wrap;
}

.tb-btn {
  width: 28px;
  height: 28px;
  border: none;
  border-radius: 4px;
  background: transparent;
  color: var(--text-muted);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 0.875rem;
}
.tb-btn:hover { background: var(--border); color: var(--text-primary); }
.tb-btn.active { background: color-mix(in srgb, var(--accent) 13%, transparent); color: var(--accent); }
.tb-formula { font-family: 'Times New Roman', serif; font-style: italic; font-weight: 600; }
.tb-sep { width: 1px; height: 18px; background: var(--border); margin: 0 2px; }

.editor-content {
  padding: 0.6rem 0.75rem;
  min-height: 80px;
}

.editor-content :deep(.ProseMirror) {
  outline: none;
  min-height: 60px;
  color: var(--text-primary);
  font-size: 0.9rem;
  line-height: 1.6;
}

.editor-content :deep(.ProseMirror p) { margin: 0 0 0.5rem; }
.editor-content :deep(.ProseMirror p:last-child) { margin-bottom: 0; }
.editor-content :deep(.ProseMirror ul),
.editor-content :deep(.ProseMirror ol) { margin: 0.25rem 0; padding-left: 1.5rem; }
.editor-content :deep(.ProseMirror img.editor-image) { max-width: 100%; border-radius: 4px; margin: 0.25rem 0; }
.editor-content :deep(.formula-render) { display: inline; }

/* Formula Dialog */
.formula-overlay {
  position: fixed; inset: 0; background: rgba(0,0,0,0.5);
  display: flex; align-items: center; justify-content: center;
  z-index: 1000;
}
.formula-dialog {
  background: var(--bg-surface); border: 1px solid var(--border); border-radius: 12px;
  padding: 1.25rem; width: 420px; max-width: 90vw;
}
.formula-dialog h3 { margin: 0 0 0.25rem; font-size: 1rem; color: var(--text-primary); }
.formula-hint { font-size: 0.8rem; color: var(--text-muted); margin: 0 0 0.6rem; }
.formula-hint code { background: var(--bg-input); padding: 1px 4px; border-radius: 3px; font-size: 0.8rem; }
.formula-input {
  width: 100%; padding: 0.5rem; border-radius: 8px;
  border: 1px solid var(--border); background: var(--bg-surface); color: var(--text-primary);
}
.formula-preview { margin-top: 0.5rem; padding: 0.5rem; background: var(--bg-surface); border-radius: 6px; min-height: 30px; color: var(--text-primary); font-size: 1.1rem; }
.formula-actions { display: flex; justify-content: flex-end; gap: 0.5rem; margin-top: 0.75rem; }
</style>
