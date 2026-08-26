import { Card, CardContent, CardHeader, CardTitle } from '../../../components/ui/card';
import { FileText } from 'lucide-react';
import { ReportType } from '../types/reports.types';

interface ReportTypeSelectorProps {
  reportTypes: ReportType[];
  selectedReport: string;
  onSelectReport: (id: string) => void;
}

export function ReportTypeSelector({ reportTypes, selectedReport, onSelectReport }: ReportTypeSelectorProps) {
  return (
    <Card>
      <CardHeader>
        <CardTitle style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
          <FileText className="w-5 h-5 text-cyan-400" />
          نوع التقرير
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(200px, 1fr))', gap: '14px' }}
             className="grid-cols-1 sm:grid-cols-2 lg:grid-cols-3">
          {reportTypes.map((report) => {
            const Icon = report.icon;
            return (
              <button
                key={report.id}
                onClick={() => onSelectReport(report.id)}
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  gap: '10px',
                  padding: '14px',
                  borderRadius: '10px',
                  border: selectedReport === report.id 
                    ? '1px solid rgba(34, 211, 238, 0.35)' 
                    : '1px solid var(--border-default)',
                  background: selectedReport === report.id
                    ? 'linear-gradient(135deg, rgba(34, 211, 238, 0.17), rgba(59, 130, 246, 0.12))'
                    : 'var(--bg-surface-elevated)',
                  cursor: 'pointer',
                  transition: '180ms ease',
                  color: selectedReport === report.id ? 'var(--primary)' : 'var(--text-primary)'
                }}
                className="report-type-button"
              >
                <div style={{
                  width: '32px',
                  height: '32px',
                  borderRadius: '8px',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  background: selectedReport === report.id
                    ? 'rgba(34, 211, 238, 0.2)'
                    : 'rgba(34, 211, 238, 0.1)'
                }}>
                  <Icon className="w-4.5 h-4.5 text-cyan-400" />
                </div>
                <span style={{ fontSize: '13px', fontWeight: '500' }}>{report.label}</span>
              </button>
            );
          })}
        </div>
      </CardContent>
    </Card>
  );
}
