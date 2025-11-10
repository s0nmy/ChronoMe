import { z } from 'zod'

export const TaskRatioSchema = z
  .object({
    task_id: z.string().min(1, 'task_id is required'),
    ratio: z.number().positive('ratio must be positive'),
    min_minutes: z.number().int().nonnegative().optional(),
    max_minutes: z.number().int().positive().optional(),
  })
  .refine(
    (value) => {
      if (value.max_minutes === undefined) {
        return true
      }
      const min = value.min_minutes ?? 0
      return min <= value.max_minutes
    },
    {
      message: 'min_minutes cannot exceed max_minutes',
      path: ['max_minutes'],
    },
  )

export const AllocationRequestSchema = z
  .object({
    total_minutes: z.number().int().positive(),
    tasks: z.array(TaskRatioSchema).min(1, 'at least one task is required'),
  })
  .refine(
    (value) => {
      const ids = new Set(value.tasks.map((task) => task.task_id))
      return ids.size === value.tasks.length
    },
    { message: 'task_id must be unique', path: ['tasks'] },
  )

export type AllocationRequest = z.infer<typeof AllocationRequestSchema>

export interface AllocationResult {
  task_id: string
  ratio: number
  allocated_minutes: number
  min_minutes?: number
  max_minutes?: number
}

interface TaskState {
  task_id: string
  ratio: number
  min_minutes: number
  max_minutes: number
  allocation: number
  remainder: number
  normalized: number
  index: number
}

const EPSILON = 1e-9

export function distributeAllocations(input: AllocationRequest): AllocationResult[] {
  const totalMinutes = input.total_minutes
  const ratioSum = input.tasks.reduce((sum, task) => sum + task.ratio, 0)
  if (ratioSum <= 0) {
    throw new Error('sum of ratios must be greater than zero')
  }

  const states: TaskState[] = input.tasks.map((task, index) => {
    const min = task.min_minutes ?? 0
    const max = task.max_minutes ?? Number.POSITIVE_INFINITY
    if (min > max) {
      throw new Error(`min_minutes cannot exceed max_minutes for task ${task.task_id}`)
    }
    return {
      task_id: task.task_id,
      ratio: task.ratio,
      min_minutes: min,
      max_minutes: max,
      allocation: min,
      remainder: 0,
      normalized: task.ratio / ratioSum,
      index,
    }
  })

  let allocated = states.reduce((sum, task) => sum + task.allocation, 0)
  if (allocated > totalMinutes) {
    throw new Error('total_minutes is smaller than the sum of min_minutes')
  }

  const boundedTasks = states.filter((task) => Number.isFinite(task.max_minutes))
  if (boundedTasks.length === states.length) {
    const maxSum = boundedTasks.reduce((sum, task) => sum + task.max_minutes, 0)
    if (totalMinutes > maxSum) {
      throw new Error('total_minutes exceeds the sum of max_minutes')
    }
  }

  const remainingPool = totalMinutes - allocated
  if (remainingPool === 0) {
    return states.map((task) => ({
      task_id: task.task_id,
      ratio: task.ratio,
      allocated_minutes: task.allocation,
      min_minutes: task.min_minutes || undefined,
      max_minutes: Number.isFinite(task.max_minutes) ? task.max_minutes : undefined,
    }))
  }

  let carried = 0
  states.forEach((task) => {
    const desired = remainingPool * task.normalized
    const capacity = Math.max(0, Math.floor(task.max_minutes - task.allocation))
    if (capacity <= 0) {
      task.remainder = -1
      return
    }
    const baseAdd = Math.min(Math.floor(desired), capacity)
    task.allocation += baseAdd
    carried += baseAdd
    task.remainder = desired - baseAdd
  })

  let remaining = totalMinutes - (allocated + carried)
  while (remaining > 0) {
    const eligible = states
      .filter((task) => task.allocation + EPSILON < task.max_minutes)
      .sort((a, b) => {
        if (Math.abs(b.remainder - a.remainder) > EPSILON) {
          return b.remainder - a.remainder
        }
        if (Math.abs(b.normalized - a.normalized) > EPSILON) {
          return b.normalized - a.normalized
        }
        return a.index - b.index
      })

    if (!eligible.length) {
      throw new Error('unable to satisfy max constraints with provided total_minutes')
    }

    if (eligible.length === 1) {
      const task = eligible[0]
      const available = task.max_minutes - task.allocation
      const chunk = Math.min(available, remaining)
      if (chunk <= 0) {
        throw new Error('unable to satisfy max constraints with provided total_minutes')
      }
      task.allocation += chunk
      remaining -= chunk
      continue
    }

    let distributedThisRound = 0
    for (const task of eligible) {
      if (remaining === 0) {
        break
      }
      if (task.allocation + EPSILON >= task.max_minutes) {
        continue
      }
      task.allocation += 1
      remaining -= 1
      distributedThisRound++
    }

    if (distributedThisRound === 0) {
      throw new Error('unable to distribute remaining minutes due to max constraints')
    }
  }

  return states.map((task) => ({
    task_id: task.task_id,
    ratio: task.ratio,
    allocated_minutes: Math.trunc(task.allocation),
    min_minutes: task.min_minutes || undefined,
    max_minutes: Number.isFinite(task.max_minutes) ? task.max_minutes : undefined,
  }))
}
