import { Card, CardContent, CardHeader, CardTitle } from '../../../components/ui/card';
import { Button } from '../../../components/ui/button';
import { Calendar } from 'lucide-react';
import { DateRange } from '../types/reports.types';

interface DateRangeSelectorProps {
  dateRanges: DateRange[];
  selectedRange: string;
  onSelectRange: (value: string) => void;
}

export function DateRangeSelector({ dateRanges, selectedRange, onSelectRange }: DateRangeSelectorProps) {
  return (
    <Card className="report-range-card">
      <CardHeader className="report-range-header">
        <CardTitle className="report-range-title">
          <Calendar className="w-5 h-5 text-cyan-400" />
          نطاق التاريخ
        </CardTitle>
      </CardHeader>
      <CardContent className="report-range-content">
        <div className="report-range-group">
          {dateRanges.map((range) => (
            <Button
              key={range.value}
              variant={selectedRange === range.value ? 'primary' : 'secondary'}
              className={`report-range-button ${selectedRange === range.value ? 'active' : ''}`}
              onClick={() => onSelectRange(range.value)}
            >
              {range.label}
            </Button>
          ))}
        </div>
      </CardContent>
    </Card>
  );
}
