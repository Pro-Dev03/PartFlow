import type { CSSProperties } from 'react';
import type { AssistantInteraction } from '../../lib/assistant-interaction';

type AssistantCueProps = {
  interaction: AssistantInteraction;
  placement: 'above-avatar' | 'above-header-avatar';
};

const toneStyles: Record<AssistantInteraction['tone'], { accent: string; background: string; text: string }> = {
  neutral: { accent: '#94a3b8', background: 'rgba(248, 250, 252, 0.96)', text: '#334155' },
  data: { accent: '#06b6d4', background: 'rgba(236, 254, 255, 0.96)', text: '#155e75' },
  finance: { accent: '#d97706', background: 'rgba(255, 251, 235, 0.96)', text: '#92400e' },
  success: { accent: '#10b981', background: 'rgba(236, 253, 245, 0.96)', text: '#065f46' },
  attention: { accent: '#f97316', background: 'rgba(255, 247, 237, 0.96)', text: '#9a3412' },
  error: { accent: '#ef4444', background: 'rgba(254, 242, 242, 0.96)', text: '#991b1b' },
  offline: { accent: '#64748b', background: 'rgba(248, 250, 252, 0.96)', text: '#475569' },
};

export default function AssistantCue({ interaction, placement }: AssistantCueProps) {
  const tone = toneStyles[interaction.tone];
  const isFloating = placement === 'above-avatar';
  const style: CSSProperties = {
    position: 'absolute',
    ...(isFloating
      ? { left: '50%', bottom: 'calc(100% + 12px)', transform: 'translateX(-50%)' }
      : { right: '-156px', bottom: 'calc(100% + 6px)' }),
    display: 'flex',
    alignItems: 'center',
    gap: '7px',
    width: 'max-content',
    maxWidth: isFloating ? '240px' : '220px',
    padding: '8px 11px',
    border: `1px solid ${tone.accent}55`,
    borderInlineStart: `3px solid ${tone.accent}`,
    borderRadius: '10px',
    background: tone.background,
    color: tone.text,
    boxShadow: '0 12px 28px rgba(15, 23, 42, 0.16), 0 2px 6px rgba(15, 23, 42, 0.08)',
    backdropFilter: 'blur(10px)',
    WebkitBackdropFilter: 'blur(10px)',
    whiteSpace: 'nowrap',
    fontSize: '11px',
    fontWeight: 650,
    lineHeight: 1.35,
    direction: 'rtl',
    zIndex: 2,
    animation: 'assistantCueIn 180ms cubic-bezier(0.2, 0.8, 0.2, 1)',
  };

  return (
    <div role="status" aria-live="polite" style={style}>
      <span style={{ display: 'inline-flex', flexShrink: 0, color: tone.accent }} aria-hidden="true">
        {interaction.icon}
      </span>
      <span>{interaction.label}</span>
      <span
        aria-hidden="true"
        style={{
          position: 'absolute',
          ...(isFloating ? { left: '50%', bottom: '-5px' } : { right: '18px', bottom: '-5px' }),
          width: '9px',
          height: '9px',
          borderRight: `1px solid ${tone.accent}55`,
          borderBottom: `1px solid ${tone.accent}55`,
          background: tone.background,
          transform: 'rotate(45deg)',
        }}
      />
    </div>
  );
}
