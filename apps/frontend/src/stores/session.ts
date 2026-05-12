import { defineStore } from 'pinia'
import { ref } from 'vue'
import client from '@/api/client'
import type { components } from '@tkaprep/shared-types'

type Session = components['schemas']['SessionResponse']
type Test = components['schemas']['TestDetailResponse']
type QuestionDetail = components['schemas']['QuestionDetailResponse']

export interface LoadedQuestion {
  id: string
  text: string
  options: { id: string; label: string; text: string }[]
  order_index: number
}

export const useSessionStore = defineStore('session', () => {
  const session = ref<Session | null>(null)
  const test = ref<Test | null>(null)
  const questions = ref<LoadedQuestion[]>([])
  const currentIndex = ref(0)
  const isLoading = ref(false)
  const error = ref<string | null>(null)

  // selectedOptionId keyed by questionId
  const answers = ref<Record<string, string | null>>({})
  const flagged = ref<Set<string>>(new Set())

  async function startOrResume(testId: string): Promise<string> {
    const { data, error: err } = await client.POST('/tests/{testId}/sessions', {
      params: { path: { testId } },
    })
    if (err) throw new Error('Failed to start session')
    return data.id
  }

  async function load(sessionId: string) {
    isLoading.value = true
    error.value = null
    try {
      const { data: sess, error: sessErr } = await client.GET('/sessions/{sessionId}', {
        params: { path: { sessionId } },
      })
      if (sessErr || !sess) throw new Error('Failed to load session')
      session.value = sess

      const { data: t, error: testErr } = await client.GET('/tests/{testId}', {
        params: { path: { testId: sess.test_id } },
      })
      if (testErr || !t) throw new Error('Failed to load test')
      test.value = t

      // Fetch all questions in parallel
      const qResults = await Promise.all(
        t.questions.map(async (tq) => {
          const { data: q } = await client.GET('/questions/{questionId}', {
            params: { path: { questionId: tq.question_id } },
          })
          return { tq, q: q as QuestionDetail | undefined }
        })
      )

      questions.value = qResults
        .filter((r): r is { tq: typeof r.tq; q: QuestionDetail } => r.q != null)
        .sort((a, b) => a.tq.order_index - b.tq.order_index)
        .map(({ tq, q }) => ({
          id: q.id,
          text: q.text,
          options: q.options.map((o) => ({ id: o.id, label: o.label, text: o.text })),
          order_index: tq.order_index,
        }))

      // Hydrate answers + flags from session
      answers.value = {}
      const newFlagged = new Set<string>()
      for (const a of sess.answers) {
        answers.value[a.question_id] = a.selected_option_id ?? null
        if (a.is_flagged) newFlagged.add(a.question_id)
      }
      flagged.value = newFlagged
    } finally {
      isLoading.value = false
    }
  }

  async function saveAnswer(questionId: string, selectedOptionId: string | null) {
    if (!session.value) return
    answers.value[questionId] = selectedOptionId
    await client.POST('/sessions/{sessionId}/answers', {
      params: { path: { sessionId: session.value.id } },
      body: { question_id: questionId, selected_option_id: selectedOptionId },
    })
  }

  async function toggleFlag(questionId: string) {
    if (!session.value) return
    const nowFlagged = !flagged.value.has(questionId)
    if (nowFlagged) flagged.value.add(questionId)
    else flagged.value.delete(questionId)
    await client.POST('/sessions/{sessionId}/flag', {
      params: { path: { sessionId: session.value.id } },
      body: { question_id: questionId, is_flagged: nowFlagged },
    })
  }

  async function submit(): Promise<string> {
    if (!session.value) throw new Error('No session')
    const { data, error: err } = await client.POST('/sessions/{sessionId}/submit', {
      params: { path: { sessionId: session.value.id } },
    })
    if (err || !data) throw new Error('Submit failed')
    reset()
    return data.id
  }

  function reset() {
    session.value = null
    test.value = null
    questions.value = []
    answers.value = {}
    flagged.value = new Set()
    currentIndex.value = 0
    error.value = null
  }

  function answeredCount() {
    return questions.value.filter((q) => answers.value[q.id] != null).length
  }

  function flaggedCount() {
    return flagged.value.size
  }

  return {
    session, test, questions, currentIndex, isLoading, error,
    answers, flagged,
    startOrResume, load, saveAnswer, toggleFlag, submit, reset,
    answeredCount, flaggedCount,
  }
})
