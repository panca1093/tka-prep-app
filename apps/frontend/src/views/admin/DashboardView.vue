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

// Animated KPI counters (user growth only)
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

const maxAttemptCount = computed(() => {
  if (!stats.value) return 1
  let m = 1
  for (const r of stats.value.attempts_by_topic ?? []) {
    m = Math.max(m, r.sd, r.smp, r.sma, r.smk)
  }
  return m
})

// Filter out topic rows with zero attempts across all education levels
const activeAttemptRows = computed(() => {
  return (stats.value?.attempts_by_topic ?? []).filter(r => topicSum(r) > 0)
})

const maxPerfScore = computed(() => {
  if (!stats.value) return 100
  let m = 100
  for (const r of stats.value.topic_performance ?? []) {
    m = Math.max(m, r.best_score)
  }
  return m
})

const edLevelLabels: Record<string, string> = {
  sd: 'SD', smp: 'SMP', sma: 'SMA', smk: 'SMK',
}

async function fetchData() {
  const [resStats, resPending] = await Promise.allSettled([
    client.GET('/admin/stats'),
    client.GET('/admin/contributors/pending', { params: { query: { limit: 10 } } }),
  ])
  if (resStats.status === 'fulfilled' && resStats.value.data) {
    stats.value = resStats.value.data
    animateCount(v => animStudents.value  = v, stats.value!.total_students)
    animateCount(v => animContribs.value  = v, stats.value!.total_contributors)
    animateCount(v => animTests.value     = v, stats.value!.total_tests)
    animateCount(v => animQuestions.value = v, stats.value!.total_questions)
    animateCount(v => animPending.value   = v, stats.value!.pending_approvals, 500)
  }
  if (resPending.status === 'fulfilled' && resPending.value.data) {
    pendingUsers.value = resPending.value.data.data
  }
  isLoading.value = false
}

async function approve(userId: string) {
  busyId.value = userId; actionError.value = ''
  const { error } = await client.POST('/admin/contributors/{userId}/approve', { params: { path: { userId } } })
  if (error) { actionError.value = 'Gagal menyetujui.'; busyId.value = null; return }
  busyId.value = null; await fetchData()
}

async function reject(userId: string) {
  busyId.value = userId; confirmRejectId.value = null; actionError.value = ''
  const { error } = await client.POST('/admin/contributors/{userId}/reject', { params: { path: { userId } } })
  if (error) { actionError.value = 'Gagal menolak.'; busyId.value = null; return }
  busyId.value = null; await fetchData()
}

function initials(name: string) {
  return name.split(' ').map(w => w[0]).join('').substring(0, 2).toUpperCase()
}

function topicSum(row: any): number {
  return (row.sd ?? 0) + (row.smp ?? 0) + (row.sma ?? 0) + (row.smk ?? 0)
}

function relativeTime(ts: string): string {
  const diff = Date.now() - new Date(ts).getTime()
  const mins = Math.floor(diff / 60000)
  if (mins < 1) return 'Baru saja'
  if (mins < 60) return `${mins} menit lalu`
  const hours = Math.floor(mins / 60)
  if (hours < 24) return `${hours} jam lalu`
  const days = Math.floor(hours / 24)
  return `${days} hari lalu`
}

function eventIcon(et: string): string {
  if (et === 'registration') return '👤'
  if (et === 'publication') return '📋'
  return '✅'
}

function completionClass(pct: number): string {
  if (pct >= 80) return 'pill--good'
  if (pct >= 50) return 'pill--warn'
  return 'pill--bad'
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

    <!-- ── Status bar ── -->
    <div class="cp-statusbar">
      <span class="sb-left">
        <span class="live-dot" />
        <span class="sb-status">SISTEM AKTIF</span>
        <span class="sb-sep">·</span>
        <span class="sb-date">{{ dateStr }}</span>
      </span>
      <span class="sb-clock">{{ timeStr }}</span>
    </div>

    <!-- ── Page heading ── -->
    <div class="cp-head">
      <h1 class="cp-title">Control Panel</h1>
      <p class="cp-subtitle">Masukelas · Administrasi Sistem</p>
    </div>

    <!-- ── KPI skeleton ── -->
    <div v-if="isLoading" class="kpi-skeleton">
      <div v-for="i in 5" :key="i" class="kpi-sk-item">
        <div class="sk-block sk-num" />
        <div class="sk-block sk-lbl" />
      </div>
    </div>

    <template v-else>

      <!-- ── KPI strip ── -->
      <div class="kpi-strip">
        <div class="kpi-cell">
          <div class="kpi-val">{{ animStudents.toLocaleString('id-ID') }}</div>
          <div class="kpi-lbl">Siswa</div>
          <span v-if="stats?.students_this_week" class="kpi-badge kpi-badge--blue">+{{ stats.students_this_week }} baru 7 hari</span>
          <div class="kpi-accent kpi-accent--blue" />
        </div>
        <div class="kpi-divider" />
        <div class="kpi-cell">
          <div class="kpi-val">{{ animContribs.toLocaleString('id-ID') }}</div>
          <div class="kpi-lbl">Kontributor</div>
          <span v-if="stats?.contributors_this_week" class="kpi-badge kpi-badge--teal">+{{ stats.contributors_this_week }} baru 7 hari</span>
          <div class="kpi-accent kpi-accent--teal" />
        </div>
        <div class="kpi-divider" />
        <div class="kpi-cell">
          <div class="kpi-val">{{ animTests.toLocaleString('id-ID') }}</div>
          <div class="kpi-lbl">Ujian</div>
          <div class="kpi-accent kpi-accent--blue" />
        </div>
        <div class="kpi-divider" />
        <div class="kpi-cell">
          <div class="kpi-val">{{ animQuestions.toLocaleString('id-ID') }}</div>
          <div class="kpi-lbl">Soal</div>
          <div class="kpi-accent kpi-accent--teal" />
        </div>
        <div class="kpi-divider" />
        <div class="kpi-cell kpi-cell--wide" :class="{ 'kpi-cell--alert': animPending > 0 }">
          <div class="kpi-pend-top">
            <div class="kpi-val" :class="animPending > 0 ? 'kpi-val--warn' : ''">{{ animPending }}</div>
            <span v-if="animPending > 0" class="kpi-badge kpi-badge--warn">PERLU TINDAKAN</span>
          </div>
          <div class="kpi-lbl">Menunggu Persetujuan</div>
          <div class="kpi-pend-bar"><div class="kpi-pend-fill" :style="`width: ${pendingRatio}%`" /></div>
          <div class="kpi-accent kpi-accent--warn" />
        </div>
      </div>

      <!-- ── Bento row 1: Heatmap (2/3) + Contributors (1/3) ── -->
      <div class="bento">
        <div class="card span2">
          <div class="card-lbl">Percobaan per Topik &amp; Jenjang</div>
          <div v-if="activeAttemptRows.length === 0" class="card-empty">
            <p>Belum ada percobaan ujian tercatat.</p>
          </div>
          <table v-else class="heat-table">
            <thead>
              <tr>
                <th class="heat-th-topic">Topik</th>
                <th v-for="ed in ['sd','smp','sma','smk']" :key="ed" class="heat-th-ed">{{ edLevelLabels[ed] }}</th>
              </tr>
            </thead>
            <tbody>
              <tr class="heat-tr" v-for="row in activeAttemptRows" :key="row.topic">
                <td class="heat-td-topic">{{ row.topic }}</td>
                <td class="heat-td" v-for="ed in ['sd','smp','sma','smk']" :key="ed">
                  <div class="heat-cell" :style="{ '--fill': Math.max(((row as any)[ed] ?? 0) / maxAttemptCount, 0.02) }">
                    <span class="heat-bg" />
                    <span class="heat-num">{{ (row as any)[ed] ?? 0 }}</span>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <div class="card span1">
          <div class="card-lbl">Top Kontributor</div>
          <div v-if="(stats?.contributor_productivity?.length ?? 0) === 0" class="card-empty">
            <p>Belum ada kontributor dengan konten.</p>
          </div>
          <div v-else class="contrib-list">
            <div
              class="contrib-row"
              v-for="(cp, i) in stats!.contributor_productivity"
              :key="cp.email"
            >
              <span class="rank-badge" :class="i===0?'rank--1':i===1?'rank--2':i===2?'rank--3':'rank--n'">{{ i + 1 }}</span>
              <span class="contrib-name">{{ cp.name }}</span>
              <span class="contrib-stat">
                <span class="contrib-num">{{ cp.question_count }}</span>
                <span class="contrib-hint">soal</span>
              </span>
              <span class="contrib-stat contrib-stat--teal">
                <span class="contrib-num contrib-num--teal">{{ cp.test_count }}</span>
                <span class="contrib-hint">ujian</span>
              </span>
            </div>
          </div>
        </div>
      </div>

      <!-- ── Bento row 2: Performance (1/3) + Test Completion (2/3) ── -->
      <div class="bento">
        <div class="card span1">
          <div class="card-lbl">Performa per Topik</div>
          <div v-if="(stats?.topic_performance?.length ?? 0) === 0" class="card-empty">
            <p>Belum ada data performa.</p>
          </div>
          <div v-else class="perf-list">
            <div class="perf-row" v-for="tp in stats!.topic_performance" :key="tp.topic">
              <span class="perf-topic">{{ tp.topic }}</span>
              <span class="perf-avg">{{ tp.avg_score.toFixed(1) }}</span>
              <span class="perf-track"><span class="perf-fill" :style="{ width: Math.min((tp.avg_score / maxPerfScore) * 100, 100) + '%' }" /></span>
              <span class="perf-best">{{ tp.best_score.toFixed(0) }}</span>
            </div>
          </div>
        </div>

        <div class="card span2">
          <div class="card-lbl">Tingkat Penyelesaian Ujian</div>
          <div v-if="(stats?.test_completion?.length ?? 0) === 0" class="card-empty">
            <p>Belum ada ujian dengan sesi tercatat.</p>
          </div>
          <div v-else class="compl-list">
            <div class="compl-row" v-for="tc in stats!.test_completion" :key="tc.test_title">
              <span class="compl-title">{{ tc.test_title }}</span>
              <span class="compl-bar">
                <span class="compl-done" :style="{ flex: tc.submitted }" />
                <span class="compl-exp"  :style="{ flex: tc.expired }" />
                <span class="compl-prog" :style="{ flex: (tc.started - tc.submitted - tc.expired) }" />
              </span>
              <span class="pill" :class="completionClass(tc.completion_pct)">{{ tc.completion_pct.toFixed(0) }}%</span>
              <span class="compl-nums">{{ tc.submitted }}/{{ tc.submitted + tc.expired }}</span>
            </div>
          </div>
        </div>
      </div>

      <!-- ── Bento row 3: Activity Feed (2/3) + Question Bank (1/3) ── -->
      <div class="bento">
        <div class="card span2">
          <div class="card-lbl">Aktivitas Terbaru</div>
          <div v-if="(stats?.activity_feed?.length ?? 0) === 0" class="card-empty">
            <p>Belum ada aktivitas dalam 7 hari terakhir.</p>
          </div>
          <div v-else class="feed-list">
            <div class="feed-row" v-for="af in stats!.activity_feed" :key="af.timestamp + af.actor_name">
              <span class="feed-icon">{{ eventIcon(af.event_type) }}</span>
              <span class="feed-body">
                <span class="feed-actor">{{ af.actor_name }}</span>
                <span v-if="af.detail" class="feed-detail">{{ af.detail }}</span>
              </span>
              <span class="feed-time">{{ relativeTime(af.timestamp) }}</span>
            </div>
          </div>
        </div>

        <div class="card span1">
          <div class="card-lbl">Bank Soal per Jenjang</div>
          <div v-if="(stats?.question_counts?.length ?? 0) === 0" class="card-empty">
            <p>Belum ada data bank soal.</p>
          </div>
          <div v-else class="qc-groups">
            <div v-for="grp in stats!.question_counts" :key="grp.education_level" class="qc-group">
              <div class="qc-group-hd">
                <span class="qc-level">{{ edLevelLabels[grp.education_level] ?? grp.education_level }}</span>
                <span class="qc-total">{{ grp.topics.length }} topik</span>
              </div>
              <div class="qc-rows">
                <div class="qc-row" v-for="t in grp.topics" :key="t.topic">
                  <span class="qc-topic">{{ t.topic }}</span>
                  <span class="qc-bar">
                    <span class="qc-bar-used"   :style="{ flex: t.used }" />
                    <span class="qc-bar-unused" :style="{ flex: t.unused }" />
                  </span>
                  <span class="qc-nums"><span class="qc-u">{{ t.used }}</span>/{{ t.used + t.unused }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- ── Error ── -->
      <div v-if="actionError" class="action-error">
        <span class="error-icon">!</span>{{ actionError }}
      </div>

      <!-- ── Approval queue ── -->
      <section class="queue-section">
        <div class="queue-hd">
          <h2 class="queue-title">Antrean Persetujuan Kontributor</h2>
          <span v-if="pendingUsers.length > 0" class="queue-badge">{{ pendingUsers.length }}</span>
        </div>

        <div v-if="pendingUsers.length === 0" class="queue-empty">
          <div class="queue-check">
            <svg viewBox="0 0 20 20" fill="currentColor" width="20" height="20"><path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd"/></svg>
          </div>
          <p class="queue-empty-msg">Semua permintaan telah diproses.</p>
        </div>

        <div v-else class="queue-list">
          <div v-for="(u, i) in pendingUsers" :key="u.id" class="queue-row" :style="`--i: ${i}`">
            <div class="q-avatar">{{ initials(u.name) }}</div>
            <div class="q-info">
              <div class="q-name">{{ u.name }}</div>
              <div class="q-email">{{ u.email }}</div>
            </div>
            <div class="q-actions">
              <template v-if="confirmRejectId === u.id">
                <span class="confirm-label">Tolak permintaan ini?</span>
                <button class="btn btn--danger" @click="reject(u.id)" :disabled="busyId === u.id">{{ busyId === u.id ? '···' : 'Ya, Tolak' }}</button>
                <button class="btn btn--ghost" @click="confirmRejectId = null">Batal</button>
              </template>
              <template v-else>
                <button class="btn btn--approve" @click="approve(u.id)" :disabled="!!busyId">
                  <svg viewBox="0 0 16 16" fill="currentColor" width="12" height="12"><path fill-rule="evenodd" d="M13.78 4.22a.75.75 0 010 1.06l-7.25 7.25a.75.75 0 01-1.06 0L2.22 9.28a.75.75 0 011.06-1.06L6 10.94l6.72-6.72a.75.75 0 011.06 0z" clip-rule="evenodd"/></svg>
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
/* ── Keyframes ── */
@keyframes fadeUp {
  from { opacity: 0; transform: translateY(10px); }
  to   { opacity: 1; transform: translateY(0); }
}
@keyframes pulseDot {
  0%, 100% { box-shadow: 0 0 6px var(--success); }
  50%       { box-shadow: 0 0 2px var(--success); opacity: 0.5; }
}
@keyframes pulseWarn {
  0%, 100% { opacity: 1; }
  50%       { opacity: 0.6; }
}
@keyframes shimmer {
  0%   { background-position: 200% 0; }
  100% { background-position: -200% 0; }
}
@keyframes growBar {
  from { width: 0; }
}

/* ── Container ── */
.cp { display: flex; flex-direction: column; gap: 1.5rem; }

/* ── Status bar ── */
.cp-statusbar {
  display: flex; align-items: center; justify-content: space-between;
  padding: .5rem .875rem;
  background: var(--bg-surface); border: 1px solid var(--border); border-radius: 9px;
  animation: fadeUp .4s ease both;
}
.sb-left { display: flex; align-items: center; gap: .5rem; font-size: .68rem; font-weight: 600; letter-spacing: .06em; text-transform: uppercase; font-family: var(--font-mono); }
.sb-status { color: var(--success); }
.sb-sep { color: var(--border); }
.sb-date { color: var(--text-muted); font-weight: 400; font-family: inherit; letter-spacing: 0; }
.sb-clock { font-family: var(--font-mono); font-size: .72rem; color: var(--text-muted); letter-spacing: .05em; }
.live-dot {
  width: 7px; height: 7px; border-radius: 50%; background: var(--success); flex-shrink: 0;
  box-shadow: 0 0 6px var(--success);
  animation: pulseDot 2s ease-in-out infinite;
}

/* ── Page heading ── */
.cp-head { animation: fadeUp .4s .05s ease both; }
.cp-title { font-size: 2.25rem; font-weight: 800; color: var(--text-heading); letter-spacing: -.03em; line-height: 1; }
.cp-subtitle { font-size: .68rem; color: var(--text-muted); letter-spacing: .08em; text-transform: uppercase; margin-top: .35rem; }

/* ── KPI skeleton ── */
.kpi-skeleton { display: flex; border: 1px solid var(--border); border-radius: 14px; overflow: hidden; background: var(--bg-surface); }
.kpi-sk-item { flex: 1; padding: 1.5rem 1.75rem; display: flex; flex-direction: column; gap: .625rem; border-right: 1px solid var(--border); }
.kpi-sk-item:last-child { border-right: none; }
.sk-block { border-radius: 4px; background: linear-gradient(90deg, var(--border) 25%, var(--bg-input) 50%, var(--border) 75%); background-size: 200% 100%; animation: shimmer 1.4s infinite; }
.sk-num { height: 2.25rem; width: 55%; }
.sk-lbl { height: .65rem; width: 38%; }

/* ── KPI strip ── */
.kpi-strip {
  display: flex; align-items: stretch;
  background: var(--bg-surface); border: 1px solid var(--border); border-radius: 14px;
  overflow: hidden; animation: fadeUp .4s .1s ease both;
}
.kpi-cell {
  flex: 1; padding: 1.375rem 1.625rem 1.25rem;
  position: relative; transition: background .15s; cursor: default;
}
.kpi-cell:hover { background: color-mix(in srgb, var(--accent) 3%, var(--bg-surface)); }
.kpi-cell--wide { flex: 1.35; }
.kpi-cell--alert { background: color-mix(in srgb, var(--warning) 4%, transparent); }
.kpi-cell--alert:hover { background: color-mix(in srgb, var(--warning) 7%, transparent); }
.kpi-divider { width: 1px; background: var(--border); flex-shrink: 0; margin: 1.25rem 0; }
.kpi-val {
  font-family: var(--font-mono); font-size: 2.4rem; font-weight: 700;
  color: var(--accent); line-height: 1; letter-spacing: -.03em; font-variant-numeric: tabular-nums;
}
.kpi-val--warn { color: var(--warning); }
.kpi-lbl {
  font-size: .63rem; font-weight: 600; text-transform: uppercase;
  letter-spacing: .09em; color: var(--text-muted); margin-top: .4rem; margin-bottom: .6rem;
}
.kpi-badge {
  display: inline-flex; align-items: center; font-size: .6rem; font-weight: 700;
  padding: .18rem .5rem; border-radius: 99px; letter-spacing: .02em;
}
.kpi-badge--blue { background: color-mix(in srgb, var(--accent) 12%, transparent); color: var(--accent); }
.kpi-badge--teal { background: color-mix(in srgb, var(--success) 12%, transparent); color: var(--success); }
.kpi-badge--warn {
  background: color-mix(in srgb, var(--warning) 12%, transparent); color: var(--warning);
  animation: pulseWarn 2s ease-in-out infinite;
}
.kpi-accent { position: absolute; bottom: 0; left: 0; right: 0; height: 2px; }
.kpi-accent--blue { background: var(--accent); opacity: .4; }
.kpi-accent--teal { background: var(--success); opacity: .4; }
.kpi-accent--warn { background: var(--warning); }
.kpi-pend-top { display: flex; align-items: center; gap: .625rem; }
.kpi-pend-bar { margin-top: .75rem; height: 3px; background: var(--border); border-radius: 99px; overflow: hidden; }
.kpi-pend-fill { height: 100%; background: var(--warning); border-radius: 99px; transition: width 1s cubic-bezier(.25,.8,.25,1); }

/* ── Cards ── */
.card {
  background: var(--bg-surface); border: 1px solid var(--border); border-radius: 14px;
  padding: 1.375rem 1.5rem;
  transition: border-color .2s, box-shadow .2s; position: relative; overflow: hidden;
}
.card:hover {
  border-color: color-mix(in srgb, var(--accent) 40%, var(--border));
  box-shadow: 0 0 0 1px color-mix(in srgb, var(--accent) 20%, transparent), 0 8px 32px rgba(0,0,0,.12);
}
.card-lbl {
  font-size: .62rem; font-weight: 700; text-transform: uppercase;
  letter-spacing: .1em; color: var(--text-muted); margin-bottom: .875rem;
}
.card-empty { padding: 1.5rem .5rem; }
.card-empty p { font-size: .8rem; color: var(--text-muted); }

/* ── Bento grid ── */
.bento {
  display: grid; grid-template-columns: 1fr 1fr 1fr; gap: 1rem;
  animation: fadeUp .4s .15s ease both;
}
.span2 { grid-column: span 2; }
.span1 { grid-column: span 1; }

/* ── Heatmap ── */
.heat-table { width: 100%; border-collapse: collapse; }
.heat-th-topic { text-align: left; font-size: .58rem; font-weight: 700; text-transform: uppercase; letter-spacing: .08em; color: var(--text-muted); padding: 0 .5rem .75rem 0; }
.heat-th-ed { font-size: .62rem; font-weight: 700; color: var(--text-muted); padding: 0 0 .75rem; text-align: center; width: 64px; }
.heat-tr { border-top: 1px solid var(--border); transition: background .12s; }
.heat-tr:hover { background: color-mix(in srgb, var(--accent) 3%, transparent); }
.heat-td-topic { font-size: .8rem; font-weight: 600; color: var(--text-heading); padding: .5rem .5rem .5rem 0; white-space: nowrap; }
.heat-td { padding: .375rem; text-align: center; }
.heat-cell {
  width: 52px; height: 34px; border-radius: 6px;
  display: flex; align-items: center; justify-content: center;
  margin: 0 auto; position: relative; overflow: hidden;
}
.heat-bg {
  position: absolute; inset: 0; border-radius: 6px;
  background: var(--accent);
  opacity: calc(var(--fill, 0) * 0.65);
}
.heat-num {
  font-family: var(--font-mono); font-size: .78rem; font-weight: 600;
  font-variant-numeric: tabular-nums; position: relative; z-index: 1;
  color: var(--text-heading);
}

/* ── Contributors ── */
.contrib-list { display: flex; flex-direction: column; }
.contrib-row {
  display: flex; align-items: center; gap: .625rem;
  padding: .55rem 0; border-bottom: 1px solid var(--border);
  transition: all .12s;
}
.contrib-row:first-child { border-top: 1px solid var(--border); }
.contrib-row:last-child { border-bottom: none; }
.contrib-row:hover { background: color-mix(in srgb, var(--accent) 3%, transparent); margin: 0 -.5rem; padding: .55rem .5rem; border-radius: 7px; border-color: transparent; }
.rank-badge {
  width: 24px; height: 24px; border-radius: 50%;
  display: flex; align-items: center; justify-content: center;
  font-size: .62rem; font-weight: 800; flex-shrink: 0; font-family: var(--font-mono);
}
.rank--1 { background: linear-gradient(135deg, #f59e0b, #d97706); color: #fff; box-shadow: 0 0 8px rgba(245,158,11,.4); }
.rank--2 { background: linear-gradient(135deg, #94a3b8, #64748b); color: #fff; }
.rank--3 { background: linear-gradient(135deg, #b87333, #92400e); color: #fff; }
.rank--n { background: var(--bg-input); color: var(--text-muted); }
.contrib-name { flex: 1; font-size: .82rem; font-weight: 600; color: var(--text-heading); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.contrib-stat { display: flex; flex-direction: column; align-items: flex-end; flex-shrink: 0; min-width: 40px; }
.contrib-stat--teal { margin-left: .375rem; }
.contrib-num { font-family: var(--font-mono); font-size: .88rem; font-weight: 700; color: var(--accent); font-variant-numeric: tabular-nums; }
.contrib-num--teal { color: var(--success); }
.contrib-hint { font-size: .56rem; color: var(--text-muted); text-transform: uppercase; letter-spacing: .05em; }

/* ── Performance ── */
.perf-list { display: flex; flex-direction: column; }
.perf-row {
  display: grid; grid-template-columns: 1fr 44px 1fr 40px;
  align-items: center; gap: .75rem;
  padding: .5rem 0; border-bottom: 1px solid var(--border);
}
.perf-row:last-child { border-bottom: none; }
.perf-topic { font-size: .78rem; font-weight: 600; color: var(--text-heading); }
.perf-avg { font-family: var(--font-mono); font-size: .85rem; font-weight: 700; color: var(--accent); font-variant-numeric: tabular-nums; text-align: right; }
.perf-track { height: 4px; background: var(--bg-input); border-radius: 2px; overflow: hidden; }
.perf-fill { display: block; height: 100%; border-radius: 2px; background: linear-gradient(90deg, var(--accent), color-mix(in srgb, var(--accent) 60%, var(--success))); animation: growBar 1s ease both; }
.perf-best { font-family: var(--font-mono); font-size: .78rem; font-weight: 600; color: var(--success); text-align: right; font-variant-numeric: tabular-nums; }

/* ── Completion ── */
.compl-list { display: flex; flex-direction: column; }
.compl-row {
  display: flex; align-items: center; gap: .75rem;
  padding: .5rem 0; border-bottom: 1px solid var(--border);
}
.compl-row:last-child { border-bottom: none; }
.compl-title { font-size: .78rem; font-weight: 600; color: var(--text-heading); flex: 1; min-width: 0; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.compl-bar {
  width: 100px; height: 6px; background: var(--bg-input); border-radius: 3px;
  overflow: hidden; display: flex; flex-shrink: 0;
}
.compl-done { background: var(--success); min-width: 0; }
.compl-exp  { background: #ff8080; min-width: 0; }
.compl-prog { background: var(--border); min-width: 0; }
.pill {
  padding: .16rem .5rem; border-radius: 99px;
  font-size: .62rem; font-weight: 700; font-family: var(--font-mono); flex-shrink: 0;
}
.pill--good { background: color-mix(in srgb, var(--success) 15%, transparent); color: var(--success); }
.pill--warn { background: color-mix(in srgb, var(--warning) 15%, transparent); color: var(--warning); }
.pill--bad  { background: color-mix(in srgb, var(--danger) 12%, transparent); color: var(--danger); }
.compl-nums { font-size: .62rem; color: var(--text-muted); white-space: nowrap; flex-shrink: 0; font-family: var(--font-mono); }

/* ── Activity feed ── */
.feed-list { display: flex; flex-direction: column; }
.feed-row {
  display: flex; align-items: flex-start; gap: .75rem;
  padding: .625rem 0; border-bottom: 1px solid var(--border);
}
.feed-row:last-child { border-bottom: none; }
.feed-icon {
  width: 28px; height: 28px; border-radius: 50%;
  background: var(--bg-input); border: 1px solid var(--border);
  display: flex; align-items: center; justify-content: center;
  font-size: .75rem; flex-shrink: 0; margin-top: 1px;
}
.feed-body { flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 1px; }
.feed-actor { font-size: .8rem; font-weight: 600; color: var(--text-heading); }
.feed-detail { font-size: .7rem; color: var(--text-muted); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.feed-time { font-size: .62rem; color: var(--text-muted); white-space: nowrap; flex-shrink: 0; margin-top: 2px; font-family: var(--font-mono); }

/* ── Question bank ── */
.qc-groups { display: flex; flex-direction: column; gap: .875rem; }
.qc-group-hd { display: flex; align-items: center; justify-content: space-between; margin-bottom: .5rem; }
.qc-level { font-size: .7rem; font-weight: 700; text-transform: uppercase; letter-spacing: .07em; color: var(--text-heading); }
.qc-total { font-size: .6rem; color: var(--text-muted); }
.qc-rows { display: flex; flex-direction: column; }
.qc-row { display: flex; align-items: center; gap: .625rem; padding: .25rem 0; font-size: .72rem; }
.qc-topic { width: 110px; font-weight: 600; color: var(--text-heading); flex-shrink: 0; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.qc-bar { flex: 1; height: 7px; background: var(--bg-input); border-radius: 4px; overflow: hidden; display: flex; }
.qc-bar-used   { background: var(--success); }
.qc-bar-unused { background: var(--border); }
.qc-nums { font-family: var(--font-mono); font-size: .68rem; color: var(--text-muted); flex-shrink: 0; width: 52px; text-align: right; }
.qc-u { color: var(--success); }

/* ── Error ── */
.action-error {
  display: flex; align-items: center; gap: .625rem;
  padding: .75rem 1rem; background: var(--danger-bg); color: var(--danger-text);
  border-radius: 10px; font-size: .825rem; border-left: 3px solid var(--danger);
}
.error-icon {
  width: 18px; height: 18px; border-radius: 50%;
  background: var(--danger); color: #fff;
  display: flex; align-items: center; justify-content: center;
  font-size: .7rem; font-weight: 800; flex-shrink: 0;
}

/* ── Approval queue ── */
.queue-section { display: flex; flex-direction: column; gap: .875rem; animation: fadeUp .4s .2s ease both; }
.queue-hd { display: flex; align-items: center; gap: .625rem; }
.queue-title { margin: 0; font-size: .68rem; font-weight: 700; text-transform: uppercase; letter-spacing: .09em; color: var(--text-muted); }
.queue-badge {
  min-width: 22px; height: 22px; padding: 0 6px; border-radius: 99px;
  background: var(--warning); color: #000;
  font-size: .65rem; font-weight: 800; display: inline-flex; align-items: center; justify-content: center;
  font-family: var(--font-mono);
}
.queue-empty {
  display: flex; flex-direction: column; align-items: center; gap: .625rem;
  padding: 3rem 2rem; border: 1px dashed var(--border); border-radius: 12px; text-align: center;
}
.queue-check { width: 40px; height: 40px; border-radius: 50%; background: var(--success-bg); color: var(--success); display: flex; align-items: center; justify-content: center; }
.queue-empty-msg { margin: 0; font-size: .875rem; color: var(--text-muted); }
.queue-list { display: flex; flex-direction: column; gap: .5rem; }
.queue-row {
  display: flex; align-items: center; gap: 1rem;
  padding: .875rem 1.125rem;
  background: var(--bg-surface); border: 1px solid var(--border); border-left: 3px solid var(--warning);
  border-radius: 10px; opacity: 0;
  animation: fadeUp .35s calc(.25s + var(--i) * 60ms) ease both;
  transition: box-shadow .15s, border-color .15s;
}
.cp.ready .queue-row { opacity: 1; }
.queue-row:hover { box-shadow: 0 4px 20px rgba(0,0,0,.12); border-color: var(--warning); }
.q-avatar {
  width: 40px; height: 40px; border-radius: 50%;
  background: linear-gradient(135deg, var(--accent), color-mix(in srgb, var(--accent) 70%, #000 30%));
  display: flex; align-items: center; justify-content: center;
  font-size: .72rem; font-weight: 800; color: #fff; flex-shrink: 0; letter-spacing: .03em;
}
.q-info { flex: 1; min-width: 0; }
.q-name { font-weight: 600; font-size: .88rem; color: var(--text-heading); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.q-email { font-size: .75rem; color: var(--text-muted); font-family: var(--font-mono); margin-top: 2px; }
.q-actions { display: flex; align-items: center; gap: .5rem; flex-shrink: 0; }
.confirm-label { font-size: .78rem; color: var(--danger); font-weight: 600; margin-right: .25rem; }

/* ── Buttons ── */
.btn { display: inline-flex; align-items: center; gap: .35rem; padding: .45rem .875rem; border-radius: 8px; font-size: .78rem; font-weight: 700; cursor: pointer; border: none; font-family: inherit; transition: all .15s; white-space: nowrap; }
.btn:disabled { opacity: .45; cursor: not-allowed; }
.btn--approve { background: var(--success); color: #fff; min-width: 88px; justify-content: center; }
.btn--approve:hover:not(:disabled) { opacity: .85; transform: translateY(-1px); }
.btn--reject { background: transparent; border: 1px solid var(--border); color: var(--danger); }
.btn--reject:hover:not(:disabled) { background: color-mix(in srgb, var(--danger) 10%, transparent); border-color: var(--danger); }
.btn--danger { background: var(--danger); color: #fff; }
.btn--danger:hover:not(:disabled) { opacity: .85; }
.btn--ghost { background: transparent; border: 1px solid var(--border); color: var(--text-muted); }
.btn--ghost:hover { background: var(--bg-input); }

/* ── Responsive ── */
@media (max-width: 1024px) {
  .bento { grid-template-columns: 1fr 1fr; }
  .span2 { grid-column: span 2; }
  .span1 { grid-column: span 1; }
}
@media (max-width: 768px) {
  .bento { grid-template-columns: 1fr; }
  .span2, .span1 { grid-column: span 1; }
  .kpi-strip { flex-wrap: wrap; }
  .kpi-divider { display: none; }
  .kpi-cell { flex: 1 1 45%; border-bottom: 1px solid var(--border); }
  .cp-title { font-size: 1.5rem; }
}
@media (max-width: 600px) {
  .queue-row { flex-wrap: wrap; }
  .q-actions { width: 100%; justify-content: flex-end; }
  .sb-date { display: none; }
}
</style>
