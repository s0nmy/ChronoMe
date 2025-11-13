import React from 'react';
import { Card, CardContent, CardHeader, CardTitle } from './ui/card';
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, PieChart, Pie, Cell, type TooltipProps } from 'recharts';
import { PieChart as PieChartIcon } from 'lucide-react';
import type { Entry, Project } from '../types';
import { getThisWeekRange, isSameDay } from '../utils/time';
import { ProjectBarChartIcon } from './icons/ProjectBarChartIcon';

type WeeklyChartDatum = {
  day: string;
  hours: number;
} & Record<string, number | string>;

const TotalHoursTooltip = ({ active, payload, label }: TooltipProps<number, string>) => {
  if (!active || !payload?.length) {
    return null;
  }

  const totalHours = payload[0]?.payload?.hours;

  return (
    <div
      style={{
        backgroundColor: 'rgba(255, 255, 255, 0.95)',
        border: '1px solid rgba(0, 0, 0, 0.1)',
        borderRadius: '8px',
        padding: '8px 12px',
        color: '#000',
        fontSize: '12px',
        boxShadow: '0 4px 12px rgba(0, 0, 0, 0.08)'
      }}
    >
      <div style={{ fontWeight: 600, marginBottom: 4 }}>{label}</div>
      <div>総時間: {typeof totalHours === 'number' ? `${totalHours}時間` : '-'}</div>
    </div>
  );
};

const ProjectDistributionTooltip = ({ active, payload }: TooltipProps<number, string>) => {
  if (!active || !payload?.length) {
    return null;
  }

  const dataPoint = payload[0]?.payload as { projectName?: string; hours?: number };

  return (
    <div
      style={{
        backgroundColor: 'rgba(255, 255, 255, 0.95)',
        border: '1px solid rgba(0, 0, 0, 0.1)',
        borderRadius: '8px',
        padding: '8px 12px',
        color: '#000',
        fontSize: '12px',
        boxShadow: '0 4px 12px rgba(0, 0, 0, 0.08)'
      }}
    >
      <div style={{ fontWeight: 600, marginBottom: 4 }}>{dataPoint?.projectName || 'プロジェクト'}</div>
      <div>総時間: {typeof dataPoint?.hours === 'number' ? `${dataPoint.hours}時間` : '-'}</div>
    </div>
  );
};

export function WorkCharts({ entries, projects }: { entries: Entry[]; projects: Project[] }) {
  const projectLookup = React.useMemo(() => {
    const map = new Map<string, Project>();
    projects.forEach(project => {
      map.set(project.id, project);
    });
    return map;
  }, [projects]);

  const projectData = React.useMemo(() => {
    const projectMap = new Map<
      string,
      {
        duration: number;
        color: string;
        name: string;
      }
    >();

    entries.forEach(entry => {
      if (!entry.endedAt || !entry.projectId) {
        return;
      }

      const project = projectLookup.get(entry.projectId);
      const current = projectMap.get(entry.projectId) || {
        duration: 0,
        color: project?.color || '#666',
        name: project?.name || 'Unknown Project'
      };

      projectMap.set(entry.projectId, {
        ...current,
        duration: current.duration + entry.durationSec
      });
    });

    const totalDuration = Array.from(projectMap.values()).reduce((sum, project) => sum + project.duration, 0);

    return Array.from(projectMap.entries())
      .map(([projectId, data]) => ({
        projectId,
        projectName: data.name,
        hours: Math.round((data.duration / 3600) * 10) / 10,
        percentage: totalDuration > 0 ? Math.round((data.duration / totalDuration) * 100) : 0,
        color: data.color,
        durationSec: data.duration
      }))
      .sort((a, b) => b.durationSec - a.durationSec);
  }, [entries, projectLookup]);

  const chartProjects = React.useMemo(() => {
    if (projectData.length > 0) {
      return projectData;
    }

    return projects.map(project => ({
      projectId: project.id,
      projectName: project.name,
      hours: 0,
      percentage: 0,
      color: project.color,
      durationSec: 0
    }));
  }, [projectData, projects]);

  const weeklyData = React.useMemo(() => {
    const weekDays = ['月', '火', '水', '木', '金', '土', '日'];
    const { start: weekStart } = getThisWeekRange();

    return weekDays.map((day, index) => {
      const dayDate = new Date(weekStart);
      dayDate.setDate(weekStart.getDate() + index);

      const datum: WeeklyChartDatum = { day, hours: 0 };
      chartProjects.forEach(project => {
        datum[project.projectId] = 0;
      });

      entries.forEach(entry => {
        if (!entry.endedAt || !entry.projectId) {
          return;
        }

        if (!isSameDay(entry.startedAt, dayDate)) {
          return;
        }

        const hours = entry.durationSec / 3600;
        const key = entry.projectId;

        datum[key] = ((datum[key] as number) || 0) + hours;
        datum.hours += hours;
      });

      datum.hours = Math.round(datum.hours * 10) / 10;

      chartProjects.forEach(project => {
        const key = project.projectId;
        datum[key] = Math.round(((datum[key] as number) || 0) * 10) / 10;
      });

      return datum;
    });
  }, [chartProjects, entries]);



  return (
    <div className="grid gap-6 md:grid-cols-2">
      {/* Weekly Hours Chart */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <ProjectBarChartIcon />
            週間作業時間
          </CardTitle>
        </CardHeader>
        <CardContent>
          <ResponsiveContainer width="100%" height={250}>
            <BarChart data={weeklyData}>
              <CartesianGrid strokeDasharray="3 3" />
              <XAxis dataKey="day" />
              <YAxis />
              <Tooltip content={<TotalHoursTooltip />} cursor={{ fill: 'rgba(0,0,0,0.04)' }} />
              {chartProjects.map((project, index) => (
                <Bar
                  key={project.projectId}
                  dataKey={project.projectId}
                  stackId="hours"
                  fill={project.color}
                  radius={index === chartProjects.length - 1 ? [4, 4, 0, 0] : [0, 0, 0, 0]}
                />
              ))}
            </BarChart>
          </ResponsiveContainer>
        </CardContent>
      </Card>

      {/* Project Distribution Chart */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <PieChartIcon className="w-5 h-5" />
            プロジェクト別時間配分
          </CardTitle>
        </CardHeader>
        <CardContent>
          {projectData.length === 0 ? (
            <div className="flex items-center justify-center h-[250px] text-muted-foreground">
              データがありません
            </div>
          ) : (
            <div className="flex justify-center">
              <div className="w-full max-w-sm">
                <ResponsiveContainer width="100%" height={200}>
                  <PieChart>
                    <Pie
                      data={projectData}
                      cx="50%"
                      cy="50%"
                      outerRadius={80}
                      dataKey="hours"
                      label={({ percentage }) => `${percentage}%`}
                    >
                      {projectData.map((entry, index) => (
                        <Cell key={`cell-${index}`} fill={entry.color} />
                      ))}
                    </Pie>
                    <Tooltip content={<ProjectDistributionTooltip />} cursor={{ fill: 'rgba(0,0,0,0.04)' }} />
                  </PieChart>
                </ResponsiveContainer>
              </div>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
