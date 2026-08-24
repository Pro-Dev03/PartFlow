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
    <Card>
      <CardHeader>
        <CardTitle style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
          <Calendar className="w-5 h-5 text-cyan-400" />
          نطاق التاريخ
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div style={{ display: 'flex', gap: '10px' }}>
          {dateRanges.map((range) => (
            <Button
              key={range.value}
              variant={selectedRange === range.value ? 'primary' : 'secondary'}
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
