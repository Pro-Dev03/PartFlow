import { Card, CardContent } from '../../../design-system/components/card';
import { ReportType } from '../types/reports.types';

interface ReportTypeSelectorProps {
  reportTypes: ReportType[];
  selectedReport: string;
  onSelectReport: (id: string) => void;
}

export function ReportTypeSelector({ reportTypes, selectedReport, onSelectReport }: ReportTypeSelectorProps) {
  return (
    <Card className="report-selector-card">
      <CardContent className="report-selector-content">
        {(['period', 'current'] as const).map((group) => {
          const groupReports = reportTypes.filter(report => (report.group || 'period') === group);
          return (
            <div key={group} className="report-selector-group">
              <div className="report-selector-heading">
                <div className="report-selector-heading-main">
                  {group === 'period' ? 'الأداء المالي' : 'حالة المتجر'}
                </div>
              </div>
              <div className="report-selector-grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3">
                {groupReports.map((report) => {
                  const Icon = report.icon;
                  const isActive = selectedReport === report.id;

                  return (
                    <button
                      key={report.id}
                      onClick={() => onSelectReport(report.id)}
                      className={`report-type-button ${isActive ? 'active' : ''}`}
                    >
                      <div className={`report-type-icon ${isActive ? 'active' : ''}`}>
                        <Icon className="w-4.5 h-4.5 text-cyan-400" />
                      </div>
                      <span>{report.label}</span>
                    </button>
                  );
                })}
              </div>
            </div>
          );
        })}
      </CardContent>
    </Card>
  );
}
