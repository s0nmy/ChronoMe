import React from 'react';
import { Card, CardContent, CardHeader, CardTitle } from './ui/card';
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, PieChart, Pie, Cell } from 'recharts';
import { BarChart3, PieChart as PieChartIcon } from 'lucide-react';
import type { Entry, Project } from '../types';
import { getThisWeekRange, isSameDay } from '../utils/time';

interface WorkChartsProps {
  entries: Entry[];
  projects: Project[];
}

export function WorkCharts({ entries, projects }: WorkChartsProps) {
  // Generate weekly data
  const getWeeklyData = () => {
    const weekDays = ['月', '火', '水', '木', '金', '土', '日'];
    const { start: weekStart } = getThisWeekRange();
    
    return weekDays.map((day, index) => {
      const dayDate = new Date(weekStart);
      dayDate.setDate(weekStart.getDate() + index);
      
      // Find sessions for this day
      const dayEntries = entries.filter(entry => 
        isSameDay(entry.startedAt, dayDate) && entry.endedAt
      );
      
      const totalHours = dayEntries.reduce((sum, entry) => sum + entry.durationSec, 0) / 3600;
      
      return {
        day,
        hours: Math.round(totalHours * 10) / 10
      };
    });
  };

  // Generate project distribution data
  const getProjectDistribution = () => {
    const projectMap = new Map<string, { duration: number; color: string; name: string }>();
    
    entries.filter(e => e.endedAt).forEach(entry => {
      const project = projects.find(p => p.id === entry.projectId);
      const projectName = project?.name || 'Unknown Project';
      const current = projectMap.get(entry.projectId) || { duration: 0, color: project?.color || '#666', name: projectName };
      projectMap.set(entry.projectId, {
        ...current,
        duration: current.duration + entry.durationSec
      });
    });

    const totalDuration = Array.from(projectMap.values()).reduce((sum, p) => sum + p.duration, 0);
    
    return Array.from(projectMap.entries()).map(([projectId, data]) => ({
      projectId,
      projectName: data.name,
      hours: Math.round((data.duration / 3600) * 10) / 10,
      percentage: totalDuration > 0 ? Math.round((data.duration / totalDuration) * 100) : 0,
      color: data.color
    }));
  };

  const weeklyData = getWeeklyData();
  const projectData = getProjectDistribution();

  return (
    <div className="grid gap-6 md:grid-cols-2">
      {/* Weekly Hours Chart */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <BarChart3 className="w-5 h-5" />
            週間作業時間
          </CardTitle>
        </CardHeader>
        <CardContent>
          <ResponsiveContainer width="100%" height={250}>
            <BarChart data={weeklyData}>
              <CartesianGrid strokeDasharray="3 3" />
              <XAxis dataKey="day" />
              <YAxis />
              <Tooltip 
                formatter={(value) => [`${value}時間`, '作業時間']}
                labelStyle={{ color: 'hsl(var(--foreground))' }}
                contentStyle={{ 
                  backgroundColor: 'hsl(var(--card))',
                  border: '1px solid hsl(var(--border))',
                  borderRadius: 'var(--radius)'
                }}
              />
              <Bar dataKey="hours" fill="hsl(var(--chart-1))" radius={[4, 4, 0, 0]} />
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
            <div className="space-y-4">
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
                  <Tooltip 
                    formatter={(value) => [`${value}時間`, 'プロジェクト時間']}
                    labelStyle={{ color: 'hsl(var(--foreground))' }}
                    contentStyle={{ 
                      backgroundColor: 'hsl(var(--card))',
                      border: '1px solid hsl(var(--border))',
                      borderRadius: 'var(--radius)'
                    }}
                  />
                </PieChart>
              </ResponsiveContainer>
              <div className="space-y-2">
                {projectData.map((item) => (
                  <div key={item.projectId} className="flex items-center gap-2 text-sm">
                    <div 
                      className="w-3 h-3 rounded-full" 
                      style={{ backgroundColor: item.color }}
                    />
                    <span className="flex-1">{item.projectName}</span>
                    <span className="text-muted-foreground">{item.hours}時間</span>
                  </div>
                ))}
              </div>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
