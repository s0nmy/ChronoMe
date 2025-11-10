import express from 'express'
import distributionRouter from './routes/distribution.js'
import { runMigrations } from './db/migrate.js'

runMigrations()

const app = express()
app.use(express.json())

app.get('/healthz', (_req, res) => {
  res.json({ ok: true })
})

app.use('/api/allocations', distributionRouter)

app.use(
  (
    err: Error,
    _req: express.Request,
    res: express.Response,
    _next: express.NextFunction,
  ) => {
    console.error(err)
    res.status(500).json({ error: 'internal server error' })
  },
)

const port = process.env.PORT ? Number(process.env.PORT) : 4000
app.listen(port, () => {
  console.log(`Allocation API listening on http://localhost:${port}`)
})
