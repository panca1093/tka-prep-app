<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed } from 'vue'
import client from '@/api/client'
import type { components } from '@tkaprep/shared-types'

type Stats = components['schemas']['PlatformStatsResponse']
type User  = components['schemas']['UserResponse']

const stats        = ref<Stats | null>(null)
const pendingUsers = ref<User[]>([])
const isLoading    = ref(true)
const actionError  = ref('')
const busyId       = ref<string | null>(null)
const confirmRejectId = ref<string | null>(null)

// Animated KPI counters
const animStudents  = ref(0)
const animContribs  = ref(0)
const animTests     = ref(0)
const animQuestions = ref(0)
const animPending   = ref(0)

function animateCount(setter: (v: number) => void, end: number, duration = 800) {
  const start = performance.now()
  const step = (now: number) => {
    const t = Math.min((now - start) / duration, 1)
    const eased = 1 - Math.pow(1 - t, 3)
    setter(Math.round(eased * end))
    if (t < 1) requestAnimationFrame(step)
  }
  requestAnimationFrame(step)
}

const pendingRatio = computed(() => {
  if (!stats.value) return 0
  const total = stats.value.total_contributors + stats.value.pending_approvals
  return total === 0 ? 0 : Math.min((stats.value.pending_approvals / total) * 100, 100)
})

async function fetchData() {
  const [resStats, resPending] = await Promise.allSettled([
    client.GET('/admin/stats'),
    client.GET('/admin/contributors/pending', { params: { query: { limit: 10 } } }),
  ])
  if (resStats.status === 'fulfilled' && resStats.value.data) {
    stats.value = resStats.value.data
    animateCount(v => animStudents.value  = v, stats.value.total_students)
    animateCount(v => animContribs.value  = v, stats.value.total_contributors)
    animateCount(v => animTests.value     = v, stats.value.total_tests)
    animateCount(v => animQuestions.value = v, stats.value.total_questions)
    animateCount(v => animPending.value   = v, stats.value.pending_approvals, 500)
  }
  if (resPending.status === 'fulfilled' && resPending.value.data) {
    pendingUsers.value = resPending.value.data.data
  }
  isLoading.value = false
}

async function approve(userId: string) {
  busyId.value = userId
  actionError.value = ''
  const { error } = await client.POST('/admin/contributors/{userId}/approve', { params: { path: { userId } } })
  if (error) { actionError.value = 'Gagal menyetujui.'; busyId.value = null; return }
  busyId.value = null
  await fetchData()
}

async function reject(userId: string) {
  busyId.value = userId
  confirmRejectId.value = null
  actionError.value = ''
  const { error } = await client.POST('/admin/contributors/{userId}/reject', { params: { path: { userId } } })
  if (error) { actionError.value = 'Gagal menolak.'; busyId.value = null; return }
  busyId.value = null
  await fetchData()
}

function initials(name: string) {
  return name.split(' ').map(w => w[0]).join('').substring(0, 2).toUpperCase()
}

const now    = ref(new Date())
const ticker = setInterval(() => now.value = new Date(), 1000)
const timeStr = computed(() =>
  now.value.toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit', second: '2-digit' })
)
const dateStr = computed(() =>
  now.value.toLocaleDateString('id-ID', { weekday: 'long', day: 'numeric', month: 'long', year: 'numeric' })
)

onMounted(fetchData)
onUnmounted(() => clearInterval(ticker))
</script>

<template>
  <div class="cp" :class="{ ready: !isLoading }">

    <!-- ── Status bar ──────────────────────────────────────────────────────── -->
    <div class="cp-statusbar">
      <span class="sb-left">
        <span class="live-dot" />
        <span class="sb-status">SISTEM AKTIF</span>
        <span class="sb-sep">·</span>
        <span class="sb-date">{{ dateStr }}</span>
      </span>
      <span class="sb-clock">{{ timeStr }}</span>
    </div>

    <!-- ── Page heading ───────────────────────────────────────────────────── -->
    <div class="cp-head">
      <div>
        <h1 class="cp-title">Control Panel</h1>
        <p class="cp-subtitle">Masukelas · Administrasi Sistem</p>
      </div>
    </div>

    <!-- ── KPI skeleton ───────────────────────────────────────────────────── -->
    <div v-if="isLoading" class="kpi-skeleton">
      <div v-for="i in 5" :key="i" class="kpi-sk-item">
        <div class="sk-block sk-num" />
        <div class="sk-block sk-lbl" />
      </div>
    </div>

    <template v-else>

      <!-- ── KPI Strip ─────────────────────────────────────────────────────── -->
      <div class="kpi-strip">
        <div class="kpi-item">
          <div class="kpi-num">{{ animStudents.toLocaleString('id-ID') }}</div>
          <div class="kpi-label">Siswa</div>
          <div class="kpi-accent kpi-accent--blue" />
        </div>
        <div class="kpi-rule" />
        <div class="kpi-item">
          <div class="kpi-num">{{ animContribs.toLocaleString('id-ID') }}</div>
          <div class="kpi-label">Kontributor</div>
          <div class="kpi-accent kpi-accent--teal" />
        </div>
        <div class="kpi-rule" />
        <div class="kpi-item">
          <div class="kpi-num">{{ animTests.toLocaleString('id-ID') }}</div>
          <div class="kpi-label">Ujian</div>
          <div class="kpi-accent kpi-accent--blue" />
        </div>
        <div class="kpi-rule" />
        <div class="kpi-item">
          <div class="kpi-num">{{ animQuestions.toLocaleString('id-ID') }}</div>
          <div class="kpi-label">Soal</div>
          <div class="kpi-accent kpi-accent--teal" />
        </div>
        <div class="kpi-rule" />
        <div class="kpi-item kpi-item--wide" :class="{ 'kpi-item--alert': animPending > 0 }">
          <div class="kpi-pending-top">
            <div class="kpi-num" :class="animPending > 0 ? 'kpi-num--warn' : ''">
              {{ animPending }}
            </div>
            <span v-if="animPending > 0" class="pending-pill">PERLU TINDAKAN</span>
          </div>
          <div class="kpi-label">Menunggu Persetujuan</div>
          <div class="kpi-pending-track">
            <div class="kpi-pending-fill" :style="`width: ${pendingRatio}%`" />
          </div>
        </div>
      </div>

      <!-- ── Error ─────────────────────────────────────────────────────────── -->
      <div v-if="actionError" class="action-error">
        <span class="error-icon">!</span>{{ actionError }}
      </div>

      <!-- ── Pending Queue ──────────────────────────────────────────────────── -->
      <section class="queue-section">
        <div class="queue-head">
          <div class="queue-head-left">
            <h2 class="queue-title">Antrean Persetujuan Kontributor</h2>
            <span v-if="pendingUsers.length > 0" class="queue-count">{{ pendingUsers.length }}</span>
          </div>
        </div>

        <div v-if="pendingUsers.length === 0" class="queue-empty">
          <div class="queue-empty-check">
            <svg viewBox="0 0 20 20" fill="currentColor" width="20" height="20">
              <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd"/>
            </svg>
          </div>
          <p class="queue-empty-msg">Semua permintaan telah diproses.</p>
        </div>

        <div v-else class="queue-list">
          <div
            v-for="(u, i) in pendingUsers"
            :key="u.id"
            class="queue-row"
            :style="`--i: ${i}`"
          >
            <div class="queue-avatar">{{ initials(u.name) }}</div>
            <div class="queue-info">
              <div class="queue-name">{{ u.name }}</div>
              <div class="queue-email">{{ u.email }}</div>
            </div>

            <div class="queue-actions">
              <template v-if="confirmRejectId === u.id">
                <span class="confirm-label">Tolak permintaan ini?</span>
                <button class="btn btn--danger-solid" @click="reject(u.id)" :disabled="busyId === u.id">
                  {{ busyId === u.id ? '···' : 'Ya, Tolak' }}
                </button>
                <button class="btn btn--ghost" @click="confirmRejectId = null">Batal</button>
              </template>
              <template v-else>
                <button class="btn btn--approve" @click="approve(u.id)" :disabled="!!busyId">
                  <svg viewBox="0 0 16 16" fill="currentColor" width="13" height="13">
                    <path fill-rule="evenodd" d="M13.78 4.22a.75.75 0 010 1.06l-7.25 7.25a.75.75 0 01-1.06 0L2.22 9.28a.75.75 0 011.06-1.06L6 10.94l6.72-6.72a.75.75 0 011.06 0z" clip-rule="evenodd"/>
                  </svg>
                  {{ busyId === u.id ? '···' : 'Setujui' }}
                </button>
                <button class="btn btn--reject" @click="confirmRejectId = u.id" :disabled="!!busyId">Tolak</button>
              </template>
            </div>
          </div>
        </div>
      </section>

    </template>
  </div>
</template>

<style scoped>
@keyframes fade-up {
  from { opacity: 0; transform: translateY(8px); }
  to   { opacity: 1; transform: translateY(0); }
}
@keyframes pulse-dot {
  0%, 100% { opacity: 1; transform: scale(1); }
  50%       { opacity: 0.5; transform: scale(0.8); }
}
@keyframes shimmer {
  0%   { background-position: 200% 0; }
  100% { background-position: -200% 0; }
}

/* ── Container ─────────────────────────────────────────────────────────────── */
.cp {
  display: flex;
  flex-direction: column;
  gap: 1.75rem;
}

/* ── Status bar ────────────────────────────────────────────────────────────── */
.cp-statusbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.45rem 0.875rem;
  border: 1px solid var(--border);
  border-radius: 8px;
  background: var(--bg-surface);
  animation: fade-up 0.3s ease both;
}
.sb-left {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.68rem;
  font-weight: 700;
  letter-spacing: 0.06em;
  color: var(--success);
  font-family: monospace;
}
.sb-sep { color: var(--border); }
.sb-date { color: var(--text-muted); font-weight: 500; letter-spacing: 0; font-family: inherit; }
.sb-clock {
  font-family: monospace;
  font-size: 0.72rem;
  font-weight: 700;
  color: var(--text-muted);
  letter-spacing: 0.06em;
}
.live-dot {
  width: 7px; height: 7px;
  border-radius: 50%;
  background: var(--success);
  flex-shrink: 0;
  animation: pulse-dot 2s ease-in-out infinite;
}
.sb-status { color: var(--success); }

/* ── Heading ───────────────────────────────────────────────────────────────── */
.cp-head {
  animation: fade-up 0.3s 0.05s ease both;
}
.cp-title {
  margin: 0;
  font-size: 2rem;
  font-weight: 800;
  color: var(--text-heading);
  letter-spacing: -0.03em;
  line-height: 1;
}
.cp-subtitle {
  margin: 0.3rem 0 0;
  font-size: 0.7rem;
  color: var(--text-muted);
  letter-spacing: 0.07em;
  text-transform: uppercase;
}

/* ── KPI skeleton ──────────────────────────────────────────────────────────── */
.kpi-skeleton {
  display: flex;
  border: 1px solid var(--border);
  border-radius: 14px;
  overflow: hidden;
  background: var(--bg-surface);
}
.kpi-sk-item {
  flex: 1;
  padding: 1.5rem 1.75rem;
  display: flex;
  flex-direction: column;
  gap: 0.625rem;
  border-right: 1px solid var(--border);
}
.kpi-sk-item:last-child { border-right: none; }
.sk-block {
  border-radius: 4px;
  background: linear-gradient(90deg, var(--border) 25%, var(--bg-input) 50%, var(--border) 75%);
  background-size: 200% 100%;
  animation: shimmer 1.4s infinite;
}
.sk-num { height: 2.25rem; width: 55%; }
.sk-lbl { height: 0.65rem; width: 38%; }

/* ── KPI Strip ─────────────────────────────────────────────────────────────── */
.kpi-strip {
  display: flex;
  align-items: stretch;
  border: 1px solid var(--border);
  border-radius: 14px;
  background: var(--bg-surface);
  overflow: hidden;
  animation: fade-up 0.4s 0.1s ease both;
}
.kpi-item {
  flex: 1;
  padding: 1.5rem 1.75rem 1.375rem;
  display: flex;
  flex-direction: column;
  gap: 0.15rem;
  position: relative;
  transition: background 0.15s;
}
.kpi-item:hover { background: color-mix(in srgb, var(--accent) 3%, transparent); }
.kpi-item--wide { flex: 1.3; }
.kpi-item--alert { background: color-mix(in srgb, var(--warning) 4%, transparent); }
.kpi-item--alert:hover { background: color-mix(in srgb, var(--warning) 7%, transparent); }

.kpi-rule {
  width: 1px;
  background: var(--border);
  flex-shrink: 0;
  margin: 1.25rem 0;
}

.kpi-num {
  font-family: monospace;
  font-size: 2.25rem;
  font-weight: 800;
  color: var(--accent);
  line-height: 1;
  letter-spacing: -0.02em;
  font-variant-numeric: tabular-nums;
}
.kpi-num--warn { color: var(--warning); }
.kpi-label {
  font-size: 0.67rem;
  font-weight: 600;
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 0.08em;
  margin-top: 0.3rem;
  margin-bottom: 0.875rem;
}

.kpi-accent {
  position: absolute;
  bottom: 0; left: 0; right: 0;
  height: 2px;
}
.kpi-accent--blue { background: var(--accent); opacity: 0.4; }
.kpi-accent--teal { background: var(--success); opacity: 0.4; }

.kpi-pending-top {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}
.pending-pill {
  font-size: 0.58rem;
  font-weight: 800;
  letter-spacing: 0.08em;
  padding: 0.2rem 0.5rem;
  border-radius: 99px;
  background: color-mix(in srgb, var(--warning) 18%, transparent);
  color: var(--warning);
  white-space: nowrap;
}
.kpi-pending-track {
  position: absolute;
  bottom: 0; left: 0; right: 0;
  height: 2px;
  background: var(--border);
}
.kpi-pending-fill {
  height: 100%;
  background: var(--warning);
  border-radius: 0 2px 2px 0;
  transition: width 1s cubic-bezier(0.25, 0.8, 0.25, 1);
}

/* ── Error ─────────────────────────────────────────────────────────────────── */
.action-error {
  display: flex;
  align-items: center;
  gap: 0.625rem;
  padding: 0.75rem 1rem;
  background: var(--danger-bg);
  color: var(--danger-text);
  border-radius: 10px;
  font-size: 0.825rem;
  border-left: 3px solid var(--danger);
}
.error-icon {
  width: 18px; height: 18px;
  border-radius: 50%;
  background: var(--danger);
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 0.7rem;
  font-weight: 800;
  flex-shrink: 0;
}

/* ── Queue ─────────────────────────────────────────────────────────────────── */
.queue-section {
  display: flex;
  flex-direction: column;
  gap: 0.875rem;
  animation: fade-up 0.4s 0.2s ease both;
}
.queue-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.queue-head-left {
  display: flex;
  align-items: center;
  gap: 0.625rem;
}
.queue-title {
  margin: 0;
  font-size: 0.7rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  color: var(--text-primary);
}
.queue-count {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 22px;
  height: 22px;
  padding: 0 6px;
  border-radius: 99px;
  background: var(--warning);
  color: #000;
  font-size: 0.68rem;
  font-weight: 800;
  line-height: 1;
}

.queue-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.625rem;
  padding: 3.5rem 2rem;
  border: 1px dashed var(--border);
  border-radius: 12px;
  text-align: center;
}
.queue-empty-check {
  width: 40px; height: 40px;
  border-radius: 50%;
  background: var(--success-bg);
  color: var(--success);
  display: flex;
  align-items: center;
  justify-content: center;
}
.queue-empty-msg {
  margin: 0;
  font-size: 0.875rem;
  color: var(--text-muted);
}

.queue-list {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.queue-row {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 1rem 1.25rem;
  background: var(--bg-surface);
  border: 1px solid var(--border);
  border-left: 3px solid var(--warning);
  border-radius: 10px;
  opacity: 0;
  animation: fade-up 0.35s calc(0.3s + var(--i) * 55ms) ease both;
  transition: box-shadow 0.15s;
}
.cp.ready .queue-row { opacity: 1; }
.queue-row:hover { box-shadow: 0 2px 14px rgba(0,0,0,0.07); }

.queue-avatar {
  width: 40px; height: 40px;
  border-radius: 50%;
  background: linear-gradient(135deg, var(--accent), color-mix(in srgb, var(--accent) 70%, #000 30%));
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 0.72rem;
  font-weight: 800;
  color: #fff;
  flex-shrink: 0;
  letter-spacing: 0.03em;
}
.queue-info { flex: 1; min-width: 0; }
.queue-name  { font-weight: 600; font-size: 0.9rem; color: var(--text-heading); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.queue-email { font-size: 0.78rem; color: var(--text-muted); margin-top: 2px; font-family: monospace; }

.queue-actions {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  flex-shrink: 0;
}
.confirm-label {
  font-size: 0.78rem;
  color: var(--danger);
  font-weight: 600;
  margin-right: 0.25rem;
}

/* ── Buttons ───────────────────────────────────────────────────────────────── */
.btn {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  padding: 0.45rem 0.875rem;
  border-radius: 7px;
  font-size: 0.8rem;
  font-weight: 700;
  cursor: pointer;
  border: none;
  transition: all 0.15s;
  white-space: nowrap;
}
.btn:disabled { opacity: 0.45; cursor: not-allowed; }

.btn--approve {
  background: var(--success);
  color: #fff;
  min-width: 88px;
  justify-content: center;
}
.btn--approve:hover:not(:disabled) { opacity: 0.85; transform: translateY(-1px); }

.btn--reject {
  background: transparent;
  border: 1px solid var(--border);
  color: var(--danger);
}
.btn--reject:hover:not(:disabled) {
  background: color-mix(in srgb, var(--danger) 10%, transparent);
  border-color: var(--danger);
}

.btn--danger-solid {
  background: var(--danger);
  color: #fff;
}
.btn--danger-solid:hover:not(:disabled) { opacity: 0.85; }

.btn--ghost {
  background: transparent;
  border: 1px solid var(--border);
  color: var(--text-muted);
}
.btn--ghost:hover { background: var(--bg-input); }

/* ── Responsive ────────────────────────────────────────────────────────────── */
@media (max-width: 900px) {
  .kpi-strip { flex-wrap: wrap; }
  .kpi-rule { display: none; }
  .kpi-item { flex: 1 1 45%; border-bottom: 1px solid var(--border); }
  .kpi-item:last-child { border-bottom: none; }
}
@media (max-width: 600px) {
  .cp-title { font-size: 1.5rem; }
  .queue-row { flex-wrap: wrap; }
  .queue-actions { width: 100%; justify-content: flex-end; }
  .sb-date { display: none; }
}
</style>
