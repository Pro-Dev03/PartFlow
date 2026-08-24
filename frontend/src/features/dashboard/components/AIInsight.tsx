import { Card, CardContent, CardHeader, CardTitle } from '../../../components/ui/card';
import { Button } from '../../../components/ui/button';
import { Sparkles, TrendingUp } from 'lucide-react';

interface AIInsightProps {
  title: string;
  description: string;
  actionLabel: string;
  onAction: () => void;
}

export function AIInsight({ title, description, actionLabel, onAction }: AIInsightProps) {
  return (
    <Card variant="ai">
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Sparkles className="w-5 h-5" style={{ color: '#22d3ee' }} />
          AI Insight
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="flex gap-3.5">
          <div className="w-5 h-5 rounded-lg flex items-center justify-center flex-shrink-0" style={{ background: 'rgba(34, 211, 238, 0.1)' }}>
            <TrendingUp className="w-3 h-3" style={{ color: '#22d3ee' }} />
          </div>
          <div>
            <p style={{ fontSize: '13px', fontWeight: '600', color: 'var(--text-primary)', marginBottom: '4px' }}>
              {title}
            </p>
            <p style={{ fontSize: '11px', color: 'var(--text-secondary)', marginBottom: '8px' }}>
              {description}
            </p>
            <Button
              variant="secondary"
              onClick={onAction}
              className="text-xs"
            >
              {actionLabel} ←
            </Button>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}