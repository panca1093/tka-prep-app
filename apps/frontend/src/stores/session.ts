import { defineStore } from 'pinia'
import { ref } from 'vue'
import client from '@/api/client'
import type { components } from '@tkaprep/shared-types'

type Session = components['schemas']['SessionResponse']
type Test = components['schemas']['TestDetailResponse']
type QuestionDetail = components['schemas']['QuestionDetailResponse']
type QuestionType = 'mcq' | 'true_false' | 'multi_correct'

export interface LoadedStatement {
  id: string
  text: string
  is_correct?: boolean  // not exposed during session, only in review
}

export interface LoadedQuestion {
  id: string
  question_type: QuestionType
  text: string
  options: { id: string; label: string; text: string }[]
  statements: LoadedStatement[]
  order_index: number
}

export const useSessionStore = defineStore('session', () => {
  const session = ref<Session | null>(null)
  const test = ref<Test | null>(null)
  const questions = ref<LoadedQuestion[]>([])
  const currentIndex = ref(0)
  const isLoading = ref(false)
  const error = ref<string | null>(null)

  // MCQ: question_id → selected option_id (or null)
  const answers = ref<Record<string, string | null>>({})
  // PGK: question_id → set of selected option_ids
  const pgkAnswers = ref<Record<string, Set<string>>>({})
  // B/S: question_id → {statement_id → boolean}
  const bsAnswers = ref<Record<string, Record<string, boolean>>>({})

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
          question_type: q.question_type as QuestionType,
          text: q.text,
          options: q.options.map((o) => ({ id: o.id, label: o.label, text: o.text })),
          statements: q.statements.map((s) => ({ id: s.id, text: s.text })),
          order_index: tq.order_index,
        }))

      // Hydrate answers from session
      const newAnswers: Record<string, string | null> = {}
      const newPGK: Record<string, Set<string>> = {}
      const newBS: Record<string, Record<string, boolean>> = {}

      for (const a of sess.answers) {
        const qid = a.question_id
        if (a.selected_option_id) {
          // MCQ or PGK
          const q = questions.value.find((q) => q.id === qid)
          if (q?.question_type === 'multi_correct') {
            if (!newPGK[qid]) newPGK[qid] = new Set()
            newPGK[qid].add(a.selected_option_id)
          } else {
            newAnswers[qid] = a.selected_option_id
          }
        } else if (a.statement_id && a.boolean_answer != null) {
          if (!newBS[qid]) newBS[qid] = {}
          newBS[qid][a.statement_id] = a.boolean_answer
        }
      }

      answers.value = newAnswers
      pgkAnswers.value = newPGK
      bsAnswers.value = newBS

      // Hydrate flags
      const newFlagged = new Set<string>()
      for (const f of sess.flags ?? []) {
        if (f.is_flagged) newFlagged.add(f.question_id)
      }
      flagged.value = newFlagged
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Unknown error'
    } finally {
      isLoading.value = false
    }
  }

  // MCQ: toggle or clear selection
  async function saveMCQAnswer(questionId: string, optionId: string | null) {
    if (!session.value) return
    answers.value[questionId] = optionId
    await client.POST('/sessions/{sessionId}/answers', {
      params: { path: { sessionId: session.value.id } },
      body: { question_id: questionId, selected_option_id: optionId },
    })
  }

  // PGK: toggle an option in the selected set
  async function togglePGKOption(questionId: string, optionId: string) {
    if (!session.value) return
    if (!pgkAnswers.value[questionId]) pgkAnswers.value[questionId] = new Set()
    const set = pgkAnswers.value[questionId]
    if (set.has(optionId)) set.delete(optionId)
    else set.add(optionId)
    pgkAnswers.value[questionId] = new Set(set)
    await client.POST('/sessions/{sessionId}/answers', {
      params: { path: { sessionId: session.value.id } },
      body: { question_id: questionId, selected_option_ids: [...set] },
    })
  }

  // B/S: set Benar (true) or Salah (false) for one statement
  async function saveBSStatement(questionId: string, statementId: string, value: boolean) {
    if (!session.value) return
    if (!bsAnswers.value[questionId]) bsAnswers.value[questionId] = {}
    bsAnswers.value[questionId][statementId] = value
    const statementAnswers = Object.entries(bsAnswers.value[questionId]).map(([sid, v]) => ({
      statement_id: sid,
      is_correct: v,
    }))
    await client.POST('/sessions/{sessionId}/answers', {
      params: { path: { sessionId: session.value.id } },
      body: { question_id: questionId, statement_answers: statementAnswers },
    })
  }

  async function toggleFlag(questionId: string) {
    if (!session.value) return
    const nowFlagged = !flagged.value.has(questionId)
    if (nowFlagged) flagged.value.add(questionId)
    else flagged.value.delete(questionId)
    flagged.value = new Set(flagged.value)
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
    pgkAnswers.value = {}
    bsAnswers.value = {}
    flagged.value = new Set()
    currentIndex.value = 0
    error.value = null
  }

  function isAnswered(q: LoadedQuestion): boolean {
    if (q.question_type === 'mcq') return answers.value[q.id] != null
    if (q.question_type === 'multi_correct') return (pgkAnswers.value[q.id]?.size ?? 0) > 0
    if (q.question_type === 'true_false') return Object.keys(bsAnswers.value[q.id] ?? {}).length > 0
    return false
  }

  function answeredCount() {
    return questions.value.filter(isAnswered).length
  }

  function flaggedCount() {
    return flagged.value.size
  }

  return {
    session, test, questions, currentIndex, isLoading, error,
    answers, pgkAnswers, bsAnswers, flagged,
    startOrResume, load,
    saveMCQAnswer, togglePGKOption, saveBSStatement,
    toggleFlag, submit, reset,
    isAnswered, answeredCount, flaggedCount,
  }
})
