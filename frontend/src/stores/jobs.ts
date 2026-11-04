import { defineStore } from 'pinia'
import type { Job } from '../types/task'

interface JobsState {
  jobs: Record<string, Job>
  order: string[]
}

export const useJobsStore = defineStore('jobs', {
  state: (): JobsState => ({
    jobs: {},
    order: [],
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
  },
})
