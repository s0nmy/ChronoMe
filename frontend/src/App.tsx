import { useEffect, useState } from 'react'
import './App.css'
import {
  fetchMonthlyReport,
  fetchTags,
  fetchWeeklyReport,
  type MonthlyReport,
  type Tag,
  type WeeklyReport,
} from './api'

const today = new Date()
const isoDate = today.toISOString().slice(0, 10)
const isoMonth = today.toISOString().slice(0, 7)

function secondsToHours(seconds: number) {
  return (seconds / 3600).toFixed(1)
}

function App() {
  const [tags, setTags] = useState<Tag[]>([])
  const [weeklyReport, setWeeklyReport] = useState<WeeklyReport | null>(null)
  const [monthlyReport, setMonthlyReport] = useState<MonthlyReport | null>(null)

  const [weekStart, setWeekStart] = useState(isoDate)
  const [month, setMonth] = useState(isoMonth)
  const [status, setStatus] = useState<string | null>(null)

  useEffect(() => {
    loadTags()
    loadWeekly(isoDate)
    loadMonthly(isoMonth)
  }, [])

  const loadTags = async () => {
    try {
      const data = await fetchTags()
      setTags(data)
    } catch (err) {
      setStatus((err as Error).message)
    }
  }

  const loadWeekly = async (start?: string) => {
    try {
      const report = await fetchWeeklyReport(start)
      setWeeklyReport(report)
    } catch (err) {
      setStatus((err as Error).message)
    }
  }

  const loadMonthly = async (targetMonth: string) => {
    try {
      const report = await fetchMonthlyReport(targetMonth)
      setMonthlyReport(report)
    } catch (err) {
      setStatus((err as Error).message)
    }
  }

  const handleWeeklySubmit = (event: React.FormEvent) => {
    event.preventDefault()
    loadWeekly(weekStart)
  }

  const handleMonthlySubmit = (event: React.FormEvent) => {
    event.preventDefault()
    loadMonthly(month)
  }

  return (
    <main className="app">
      <header>
        <h1>ChronoMe API Explorer</h1>
        <p>Quickly verify tag & report APIs via the Vite dev server proxy.</p>
      </header>

      {status && (
        <p className="status" role="alert">
          {status}
        </p>
      )}

      <section>
        <div className="section-header">
          <h2>Tags</h2>
          <button onClick={loadTags}>Reload</button>
        </div>
        {tags.length === 0 ? (
          <p>No tags yet.</p>
        ) : (
          <ul className="tag-list">
            {tags.map((tag) => (
              <li key={tag.id}>
                <span className="tag-swatch" style={{ background: tag.color }} />
                {tag.name} <small>{tag.color}</small>
              </li>
            ))}
          </ul>
        )}
      </section>

      <section>
        <div className="section-header">
          <h2>Weekly Report</h2>
          <form onSubmit={handleWeeklySubmit}>
            <label>
              Week start (YYYY-MM-DD)
              <input
                type="date"
                value={weekStart}
                onChange={(e) => setWeekStart(e.target.value)}
              />
            </label>
            <button type="submit">Fetch</button>
          </form>
        </div>

        {weeklyReport ? (
          <div className="report-card">
            <p>
              Week of <strong>{weeklyReport.week_start}</strong> ·{' '}
              {secondsToHours(weeklyReport.total_seconds)} h
            </p>
            <table>
              <thead>
                <tr>
                  <th>Day</th>
                  <th>Hours</th>
                </tr>
              </thead>
              <tbody>
                {weeklyReport.days.map((day) => (
                  <tr key={day.date}>
                    <td>{day.date}</td>
                    <td>{secondsToHours(day.total_seconds)}</td>
                  </tr>
                ))}
              </tbody>
            </table>

            <div className="report-grid weekly-breakdown">
              <div>
                <h3>Projects</h3>
                {weeklyReport.projects.length ? (
                  <ul>
                    {weeklyReport.projects.map((project) => (
                      <li key={project.project_id ?? 'unassigned'}>
                        <span
                          className="tag-swatch"
                          style={{ background: project.color || '#d1d5db' }}
                        />
                        {project.name || 'Unassigned'} · {secondsToHours(project.total_seconds)} h
                      </li>
                    ))}
                  </ul>
                ) : (
                  <p className="muted">No project data yet.</p>
                )}
              </div>
              <div>
                <h3>Tags</h3>
                {weeklyReport.tags.length ? (
                  <ul>
                    {weeklyReport.tags.map((tag) => (
                      <li key={tag.tag_id}>
                        <span className="tag-swatch" style={{ background: tag.color }} />
                        {tag.name} · {secondsToHours(tag.total_seconds)} h
                      </li>
                    ))}
                  </ul>
                ) : (
                  <p className="muted">No tag usage for this week.</p>
                )}
              </div>
            </div>
          </div>
        ) : (
          <p>Loading weekly report...</p>
        )}
      </section>

      <section>
        <div className="section-header">
          <h2>Monthly Report</h2>
          <form onSubmit={handleMonthlySubmit}>
            <label>
              Month (YYYY-MM)
              <input
                type="month"
                value={month}
                onChange={(e) => setMonth(e.target.value)}
              />
            </label>
            <button type="submit">Fetch</button>
          </form>
        </div>

        {monthlyReport ? (
          <div className="report-card">
            <p>
              Month <strong>{monthlyReport.month}</strong> ·{' '}
              {secondsToHours(monthlyReport.total_seconds)} h
            </p>
            <div className="report-grid">
              <div>
                <h3>Weekly Totals</h3>
                <ul>
                  {monthlyReport.weeks.map((week) => (
                    <li key={week.week_start}>
                      {week.week_start}: {secondsToHours(week.total_seconds)} h
                    </li>
                  ))}
                </ul>
              </div>
              <div>
                <h3>Projects</h3>
                {monthlyReport.projects.length ? (
                  <ul>
                    {monthlyReport.projects.map((project, index) => (
                      <li key={`${project.project_id ?? 'none'}-${index}`}>
                        <span
                          className="tag-swatch"
                          style={{ background: project.color || '#d1d5db' }}
                        />
                        {project.name || 'Unassigned'} · {secondsToHours(project.total_seconds)} h
                      </li>
                    ))}
                  </ul>
                ) : (
                  <p className="muted">No project data.</p>
                )}
              </div>
              <div>
                <h3>Tags</h3>
                {monthlyReport.tags.length ? (
                  <ul>
                    {monthlyReport.tags.map((tag) => (
                      <li key={tag.tag_id}>
                        <span className="tag-swatch" style={{ background: tag.color }} />
                        {tag.name} · {secondsToHours(tag.total_seconds)} h
                      </li>
                    ))}
                  </ul>
                ) : (
                  <p className="muted">No tag usage this month.</p>
                )}
              </div>
            </div>
          </div>
        ) : (
          <p>Loading monthly report...</p>
        )}
      </section>
    </main>
  )
}

export default App
