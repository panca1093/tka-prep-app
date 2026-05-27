<script setup lang="ts">
import { onMounted, onUnmounted } from 'vue'
import type { components } from '@tkaprep/shared-types'
import RichTextViewer from '@/components/editor/RichTextViewer.vue'

type Question = components['schemas']['QuestionDetailResponse']

defineProps<{ question: Question; topicName: string }>()
const emit = defineEmits<{ close: [] }>()

function onKey(e: KeyboardEvent) {
  if (e.key === 'Escape') emit('close')
}
onMounted(() => document.addEventListener('keydown', onKey))
onUnmounted(() => document.removeEventListener('keydown', onKey))

const typeLabel: Record<string, string> = { mcq: 'Pilihan Ganda', multi_correct: 'PGK', true_false: 'Benar / Salah' }
const diffLabel: Record<string, string> = { easy: 'Mudah', medium: 'Sedang', hard: 'Sulit' }
</script>

<template>
  <Teleport to="body">
    <div class="preview-backdrop" @click.self="emit('close')">
      <div class="preview-dialog" role="dialog" aria-modal="true">

        <div class="preview-header">
          <div class="preview-meta">
            <span class="badge badge-type">{{ typeLabel[question.question_type] ?? question.question_type }}</span>
            <span class="badge badge-topic">{{ topicName }}</span>
            <span class="badge" :class="'badge-diff--' + question.difficulty">{{ diffLabel[question.difficulty] }}</span>
            <span v-if="(question as any).education_level" class="badge badge-edu">{{ ((question as any).education_level as string).toUpperCase() }}</span>
          </div>
          <button class="preview-close" @click="emit('close')" aria-label="Tutup">×</button>
        </div>

        <div class="preview-body">
          <div class="q-number">Soal</div>
          <div class="q-text">
            <RichTextViewer :html="question.text" />
          </div>

          <!-- MCQ / PGK options — no correct-answer highlighting -->
          <div v-if="question.question_type !== 'true_false'" class="options-list">
            <div v-for="opt in question.options" :key="opt.id" class="option-row">
              <span class="opt-label">{{ opt.label }}</span>
              <span class="opt-text"><RichTextViewer :html="opt.text" /></span>
            </div>
          </div>

          <!-- B/S statements — shown as neutral buttons -->
          <div v-else class="stmts-list">
            <div v-for="(stmt, i) in question.statements" :key="stmt.id" class="stmt-row">
              <span class="stmt-num">{{ i + 1 }}</span>
              <span class="stmt-text"><RichTextViewer :html="stmt.text" /></span>
              <span class="stmt-toggle stmt-toggle--neutral">Benar</span>
              <span class="stmt-toggle stmt-toggle--neutral">Salah</span>
            </div>
          </div>
        </div>

        <div class="preview-footer">
          <span class="preview-note">Tampilan ini seperti yang dilihat oleh siswa saat mengerjakan tes.</span>
          <button class="btn-close-preview" @click="emit('close')">Tutup</button>
        </div>

      </div>
    </div>
  </Teleport>
</template>

<style scoped>
@keyframes modal-in {
  from { opacity: 0; transform: scale(0.96) translateY(10px); }
  to   { opacity: 1; transform: scale(1) translateY(0); }
}

.preview-backdrop {
  position: fixed; inset: 0; z-index: 300;
  background: rgba(0,0,0,0.6); backdrop-filter: blur(4px);
  display: flex; align-items: center; justify-content: center; padding: 1rem;
}

.preview-dialog {
  width: 100%; max-width: 640px; max-height: 88vh;
  background: var(--bg-surface);
  border-radius: 18px;
  box-shadow: 0 24px 64px rgba(0,0,0,0.45), 0 4px 16px rgba(0,0,0,0.25);
  display: flex; flex-direction: column;
  animation: modal-in 0.22s cubic-bezier(0.22, 1, 0.36, 1) both;
  overflow: hidden;
}

.preview-header {
  display: flex; align-items: center; justify-content: space-between; gap: 0.75rem;
  padding: 1rem 1.25rem;
  border-bottom: 1px solid var(--border);
  flex-shrink: 0;
}

.preview-meta { display: flex; align-items: center; gap: 0.4rem; flex-wrap: wrap; }

.badge {
  font-size: 0.6rem; font-weight: 700; text-transform: uppercase; letter-spacing: 0.04em;
  padding: 0.18rem 0.45rem; border-radius: 4px;
}
.badge-type  { background: color-mix(in srgb, var(--accent) 12%, transparent); color: var(--accent); border: 1px solid color-mix(in srgb, var(--accent) 25%, transparent); }
.badge-topic { background: var(--bg-input); color: var(--text-muted); border: 1px solid var(--border); }
.badge-edu   { background: color-mix(in srgb, var(--accent) 8%, transparent); color: var(--accent); border: 1px solid color-mix(in srgb, var(--accent) 18%, transparent); }
.badge-diff--easy   { background: rgba(0,210,160,0.08); color: var(--success); border: 1px solid rgba(0,210,160,0.2); }
.badge-diff--medium { background: rgba(245,166,35,0.08); color: var(--warning); border: 1px solid rgba(245,166,35,0.2); }
.badge-diff--hard   { background: rgba(255,82,82,0.08); color: var(--danger); border: 1px solid rgba(255,82,82,0.2); }

.preview-close {
  width: 2rem; height: 2rem; border-radius: 8px; flex-shrink: 0;
  border: 1px solid var(--border); background: transparent;
  color: var(--text-muted); font-size: 1.2rem; line-height: 1;
  cursor: pointer; display: flex; align-items: center; justify-content: center; transition: all 0.15s;
}
.preview-close:hover { border-color: var(--danger); color: var(--danger); background: color-mix(in srgb, var(--danger) 10%, transparent); }

.preview-body {
  flex: 1; overflow-y: auto; padding: 1.5rem;
  display: flex; flex-direction: column; gap: 1.25rem;
}

.q-number { font-size: 0.65rem; font-weight: 700; text-transform: uppercase; letter-spacing: 0.08em; color: var(--text-muted); }
.q-text { font-size: 0.925rem; line-height: 1.65; color: var(--text-primary); }

/* Options */
.options-list { display: flex; flex-direction: column; gap: 0.5rem; }
.option-row {
  display: flex; align-items: flex-start; gap: 0.75rem;
  padding: 0.7rem 0.9rem; border: 1.5px solid var(--border);
  border-radius: 10px; background: var(--bg-input);
  cursor: default; transition: border-color 0.12s;
}
.option-row:hover { border-color: color-mix(in srgb, var(--accent) 30%, var(--border)); }
.opt-label {
  width: 1.75rem; height: 1.75rem; border-radius: 50%; flex-shrink: 0;
  background: var(--bg-surface); border: 1.5px solid var(--border);
  display: flex; align-items: center; justify-content: center;
  font-size: 0.72rem; font-weight: 700; color: var(--text-muted);
}
.opt-text { font-size: 0.875rem; line-height: 1.5; color: var(--text-primary); flex: 1; padding-top: 0.1rem; }

/* Statements */
.stmts-list { display: flex; flex-direction: column; gap: 0.5rem; }
.stmt-row {
  display: flex; align-items: center; gap: 0.6rem;
  padding: 0.65rem 0.9rem; border: 1.5px solid var(--border);
  border-radius: 10px; background: var(--bg-input);
}
.stmt-num {
  width: 1.5rem; height: 1.5rem; border-radius: 50%; flex-shrink: 0;
  background: var(--bg-surface); border: 1px solid var(--border);
  display: flex; align-items: center; justify-content: center;
  font-size: 0.68rem; font-weight: 700; color: var(--text-muted);
}
.stmt-text { flex: 1; font-size: 0.875rem; line-height: 1.5; color: var(--text-primary); }
.stmt-toggle--neutral {
  padding: 0.3rem 0.55rem; border-radius: 6px; font-size: 0.72rem; font-weight: 600; flex-shrink: 0;
  border: 1px solid var(--border); background: transparent; color: var(--text-muted); cursor: default;
}

.preview-footer {
  display: flex; align-items: center; justify-content: space-between; gap: 1rem;
  padding: 0.875rem 1.25rem;
  border-top: 1px solid var(--border);
  flex-shrink: 0;
}
.preview-note { font-size: 0.7rem; color: var(--text-muted); font-style: italic; }

.btn-close-preview {
  padding: 0.5rem 1.1rem; border-radius: 8px; border: 1px solid var(--border);
  background: transparent; color: var(--text-muted); cursor: pointer;
  font-size: 0.85rem; font-weight: 600; transition: all 0.15s; font-family: inherit;
}
.btn-close-preview:hover { border-color: var(--accent); color: var(--accent); }

@media (max-width: 640px) {
  .preview-dialog { max-height: 95vh; border-radius: 12px; }
}
</style>
