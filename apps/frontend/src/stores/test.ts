import { defineStore } from 'pinia'
import { ref } from 'vue'
import client from '@/api/client'
import type { components } from '@tkaprep/shared-types'

type Test = components['schemas']['TestDetailResponse']

export const useTestStore = defineStore('test', () => {
  const tests = ref<Test[]>([])
  const total = ref(0)
  const isLoading = ref(false)

  async function fetchPublished(params?: { category_id?: string; page?: number; limit?: number }) {
    isLoading.value = true
    try {
      const { data } = await client.GET('/tests', {
        params: {
          query: {
            status: 'published',
            category_id: params?.category_id as string | undefined,
            page: params?.page,
            limit: params?.limit ?? 20,
          },
        },
      })
      if (data) {
        tests.value = data.data
        total.value = data.total
      }
    } finally {
      isLoading.value = false
    }
  }

  async function fetchAll(params?: { page?: number; limit?: number }) {
    isLoading.value = true
    try {
      const { data } = await client.GET('/tests', {
        params: { query: { page: params?.page, limit: params?.limit ?? 20 } },
      })
      if (data) {
        tests.value = data.data
        total.value = data.total
      }
    } finally {
      isLoading.value = false
    }
  }

  return { tests, total, isLoading, fetchPublished, fetchAll }
})
