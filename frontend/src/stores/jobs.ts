import { defineStore } from 'pinia'
import type { Job } from '../types/task'

interface JobsState {
  jobs: Record<string, Job>
  order: string[]
  // 任务完成回调（按任务类型注册）：全局 jobUpdate 分发后通知对应页面清空表单
  doneHandlers: Record<string, (() => void) | undefined>
}

export const useJobsStore = defineStore('jobs', {
  state: (): JobsState => ({
    jobs: {},
    order: [],
    doneHandlers: {},
  }),
  getters: {
    jobList(state): Job[] {
      return state.order.map((id) => state.jobs[id]).filter(Boolean)
    },
    hasActive(): boolean {
      return this.jobList.some((j) => j.status === 'pending' || j.status === 'running')
    },
  },
  actions: {
    upsertJob(job: Job) {
      if (!this.jobs[job.jobId]) {
        // 新任务插到队首：任务队列 latest on top
        this.order.unshift(job.jobId)
      }
      this.jobs[job.jobId] = job
    },
    updateJob(job: Partial<Job> & { jobId: string }) {
      const existing = this.jobs[job.jobId]
      if (existing) {
        Object.assign(existing, job)
      } else {
        this.upsertJob(job as Job)
      }
    },
    clearJobs() {
      this.jobs = {}
      this.order = []
    },
    // 各页面在 onMounted 注册自己任务类型的完成回调（清空表单等）
    registerDoneHandler(kind: string, fn: () => void) {
      this.doneHandlers[kind] = fn
    },
  },
})
