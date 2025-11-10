import { randomUUID } from 'node:crypto'
import { Router, Request, Response } from 'express'
import db from '../db/connection.js'
import {
  AllocationRequestSchema,
  distributeAllocations,
  type AllocationRequest,
} from '../lib/distribute.js'

const router = Router()

router.post('/', (req: Request, res: Response) => {
  const parsed = AllocationRequestSchema.safeParse(req.body)
  if (!parsed.success) {
    return res.status(422).json({ errors: parsed.error.flatten() })
  }

  let allocations
  try {
    allocations = distributeAllocations(parsed.data)
  } catch (error) {
    return res.status(422).json({ error: (error as Error).message })
  }

  const requestId = randomUUID()
  const now = new Date().toISOString()

  const insertRequest = db.prepare(
    'INSERT INTO allocation_requests (id, total_minutes, created_at) VALUES (?, ?, ?)',
  )
  const insertAllocation = db.prepare(
    `INSERT INTO task_allocations
      (request_id, task_id, ratio, allocated_minutes, min_minutes, max_minutes, created_at, updated_at)
      VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
  )

  const persist = db.transaction((input: AllocationRequest) => {
    insertRequest.run(requestId, input.total_minutes, now)

    allocations.forEach((allocation) => {
      insertAllocation.run(
        requestId,
        allocation.task_id,
        allocation.ratio,
        allocation.allocated_minutes,
        allocation.min_minutes ?? null,
        allocation.max_minutes ?? null,
        now,
        now,
      )
    })
  })

  persist(parsed.data)

  return res.status(201).json({
    request_id: requestId,
    total_minutes: parsed.data.total_minutes,
    allocations: allocations.map((allocation) => ({
      task_id: allocation.task_id,
      ratio: allocation.ratio,
      allocated_minutes: allocation.allocated_minutes,
    })),
  })
})

export default router
