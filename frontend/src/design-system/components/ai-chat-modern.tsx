import { useEffect, useRef, useState } from 'react';
import { X, Send, Loader2 } from 'lucide-react';
import { assistantApi } from '../../services/api/endpoints';
import { formatStoreTime } from '../../utils/store-time';
import type { AssistantContext } from '../../lib/assistant-response';
import { getAssistantInteraction, getAssistantNavigationPath, getSystemInteraction, inferAssistantIntent, type AssistantInteraction } from '../../lib/assistant-interaction';
import NeonAIBot from './neon-ai-bot';
import AssistantCue from './assistant-cue';

type AssistantAction = { label: string; path: string };

type AssistantState = 'idle' | 'greeting' | 'listening' | 'thinking' | 'speaking' | 'attention' | 'error' | 'offline';

type ChatMessage = {
  id: string;
  role: 'assistant' | 'user';
  text: string;
  timestamp: Date;
  status?: 'sending' | 'sent' | 'delivered';
  actions?: AssistantAction[];
};

interface AIChatModernProps {
  onClose?: () => void;
  position?: { x: number; y: number };
  embedded?: boolean;
  isOffline?: boolean;
  assistantContext?: AssistantContext;
  onSendMessage?: (text: string) => void;
  onStateChange?: (state: AssistantState) => void;
  onInteractionChange?: (interaction: AssistantInteraction | null) => void;
}

interface QuickAction {
  icon: string;
  label: string;
  query: string;
}

export default function AIChatModern({
  onClose,
  position,
  embedded = false,
  isOffline,
  assistantContext,
  onSendMessage,
  onStateChange,
  onInteractionChange,
}: AIChatModernProps) {
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [inputText, setInputText] = useState('');
  const [isTyping, setIsTyping] = useState(false);
  const [isInitialized, setIsInitialized] = useState(false);
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const inputRef = useRef<HTMLTextAreaElement>(null);
  const mountedRef = useRef(true);
  const stateTimerRef = useRef<number | null>(null);
  const messagesRef = useRef<ChatMessage[]>([]);

  const offlineMode = isOffline ?? (typeof navigator !== 'undefined' && !navigator.onLine);
  const [assistantState, setAssistantState] = useState<AssistantState>(offlineMode ? 'offline' : 'idle');
  const [interaction, setInteraction] = useState<AssistantInteraction>(() => (
    offlineMode ? getSystemInteraction('offline') : getAssistantInteraction('GREETING')
  ));

  const clearStateTimer = () => {
    if (stateTimerRef.current !== null) {
      window.clearTimeout(stateTimerRef.current);
      stateTimerRef.current = null;
    }
  };

  const transitionToIdle = (delay: number) => {
    clearStateTimer();
    stateTimerRef.current = window.setTimeout(() => {
      if (mountedRef.current) setAssistantState('idle');
    }, delay);
  };

  useEffect(() => {
    mountedRef.current = true;
    return () => {
      mountedRef.current = false;
      clearStateTimer();
      onStateChange?.('idle');
    };
  }, [onStateChange]);

  useEffect(() => {
    messagesRef.current = messages;
  }, [messages]);

  useEffect(() => {
    if (offlineMode) {
      clearStateTimer();
      setAssistantState('offline');
      setInteraction(getSystemInteraction('offline'));
    } else if (isTyping) {
      setAssistantState('thinking');
    } else {
      setAssistantState((current) => current === 'offline' ? 'idle' : current);
    }
  }, [offlineMode, isTyping]);

  useEffect(() => {
    onStateChange?.(assistantState);
  }, [assistantState, onStateChange]);

  useEffect(() => {
    onInteractionChange?.(interaction);
    return () => onInteractionChange?.(null);
  }, [interaction, onInteractionChange]);

  const quickActions: QuickAction[] = [
    { icon: '📦', label: 'إدارة المخزون', query: 'ما حال المخزون؟' },
    { icon: '💰', label: 'متابعة الديون والتحصيل', query: 'أخبرني عن الديون المتأخرة' },
    { icon: '📊', label: 'تحليل المبيعات والأرباح', query: 'ما الذي حدث في هذا الشهر؟' },
    { icon: '⚡', label: 'الحصول على اقتراحات عملية', query: 'هل لديك اقتراحات؟' },
  ];

  useEffect(() => {
    if (!isInitialized) {
      setAssistantState(offlineMode ? 'offline' : 'greeting');
      const welcomeMessage: ChatMessage = {
        id: crypto.randomUUID(),
        role: 'assistant',
        text: '👋 مرحباً! أنا مساعدك الذكي.\n\nكيف يمكنني مساعدتك؟',
        timestamp: new Date(),
        status: 'delivered',
      };
      setMessages([welcomeMessage]);
      setIsInitialized(true);
      if (!offlineMode) {
        transitionToIdle(900);
      }
    }
  }, [isInitialized, offlineMode]);

  useEffect(() => {
    if (typeof messagesEndRef.current?.scrollIntoView === 'function') {
      messagesEndRef.current.scrollIntoView({ behavior: 'smooth' });
    }
  }, [messages, isTyping]);

  useEffect(() => {
    if (inputRef.current) {
      inputRef.current.style.height = 'auto';
      const newHeight = Math.min(inputRef.current.scrollHeight, 80);
      inputRef.current.style.height = `${newHeight}px`;
    }
  }, [inputText]);

  const showQuickActions = messages.length === 1 && messages[0].role === 'assistant';
  const stateLabels: Record<AssistantState, string> = {
    idle: 'جاهز',
    greeting: 'يرحب بك',
    listening: 'يستمع',
    thinking: 'يفكر',
    speaking: 'يرد',
    attention: 'يحتاج انتباهك',
    offline: 'غير متصل',
    error: 'تعذر الرد',
  };

  useEffect(() => {
    const style = document.createElement('style');
    style.textContent = `
      @keyframes chatBounce {
        0%, 100% { transform: translateY(0); }
        50% { transform: translateY(-4px); }
      }
      @keyframes chatSpin {
        from { transform: rotate(0deg); }
        to { transform: rotate(360deg); }
      }
      @keyframes assistantCueIn {
        from { opacity: 0; transform: translateY(4px) scale(0.98); }
        to { opacity: 1; transform: translateY(0) scale(1); }
      }
    `;
    document.head.appendChild(style);
    return () => {
      document.head.removeChild(style);
    };
  }, []);

  const requestReply = async (text: string) => {
    if (!text || isTyping) return;
    if (offlineMode) {
      clearStateTimer();
      setAssistantState('offline');
      setMessages((prev) => [...prev, { id: crypto.randomUUID(), role: 'assistant', text: 'المساعد غير متاح بدون اتصال حاليًا.', timestamp: new Date(), status: 'delivered' }]);
      return;
    }

    const userMessage: ChatMessage = {
      id: crypto.randomUUID(),
      role: 'user',
      text,
      timestamp: new Date(),
      status: 'sent',
    };

    setMessages((prev) => [...prev, userMessage]);
    messagesRef.current = [...messagesRef.current, userMessage];
    setInputText('');
    const localNavigationPath = getAssistantNavigationPath(text);
    if (localNavigationPath) {
      setInteraction(getAssistantInteraction('NAVIGATION', text));
      setAssistantState('speaking');
      window.setTimeout(() => {
        if (!mountedRef.current) return;
        window.location.hash = localNavigationPath;
        onClose?.();
      }, 220);
      return;
    }
    setIsTyping(true);
    clearStateTimer();
    setInteraction(getAssistantInteraction(inferAssistantIntent(text), text));
    setAssistantState('thinking');
    try {
      const response = await assistantApi.reply(userMessage.text, messagesRef.current.slice(-10).map((item) => ({ role: item.role, content: item.text })));
      const reply = response?.data?.reply ?? response?.reply;
      if (!reply) throw new Error('Assistant returned no reply');
      const assistantMessage: ChatMessage = {
        id: crypto.randomUUID(),
        role: 'assistant',
        text: reply,
        timestamp: new Date(),
        status: 'delivered',
        actions: Array.isArray(response?.data?.actions) ? response.data.actions : [],
      };
      setMessages((prev) => [...prev, assistantMessage]);
      messagesRef.current = [...messagesRef.current, assistantMessage];
      const intent = String(response?.data?.intent ?? response?.intent ?? '').toUpperCase();
      setInteraction(getAssistantInteraction(intent, text));
      if (intent === 'NAVIGATION' && assistantMessage.actions?.[0]?.path) {
        window.location.hash = assistantMessage.actions[0].path;
        onClose?.();
        return;
      }
      const hasAttention = (intent === 'LOW_STOCK' && Number(assistantContext?.lowStockCount ?? 0) > 0)
        || (intent === 'DEBTS' && Number(assistantContext?.overdueDebtsCount ?? 0) > 0);
      if (hasAttention) {
        setAssistantState('attention');
        stateTimerRef.current = window.setTimeout(() => {
          if (!mountedRef.current) return;
          setAssistantState('speaking');
          transitionToIdle(Math.min(1800, Math.max(700, reply.length * 18)));
        }, 420);
      } else {
        setAssistantState('speaking');
        transitionToIdle(Math.min(1800, Math.max(700, reply.length * 18)));
      }
      onSendMessage?.(userMessage.text);
    } catch {
      clearStateTimer();
      setInteraction(getSystemInteraction('error'));
      setAssistantState('error');
      setMessages((prev) => [...prev, { id: crypto.randomUUID(), role: 'assistant', text: 'تعذر جلب بيانات المتجر حاليًا. جرّب مرة ثانية.', timestamp: new Date(), status: 'delivered' }]);
      transitionToIdle(1200);
    } finally {
      setIsTyping(false);
    }
  };

  const handleSendMessage = () => { void requestReply(inputText.trim()); };
  const handleQuickAction = (query: string) => { void requestReply(query); };

  const MessageBubble = ({ message }: { message: ChatMessage }) => {
    const isUser = message.role === 'user';

    return (
      <div
        style={{
          display: 'flex',
          justifyContent: isUser ? 'flex-end' : 'flex-start',
          direction: 'rtl',
          marginBottom: '12px',
          gap: '8px'
        }}
      >
        {!isUser && (
          <div style={{ width: '32px', height: '32px', display: 'flex', alignItems: 'center', justifyContent: 'center', flexShrink: 0 }}>
            <NeonAIBot size={32} state={assistantState} />
          </div>
        )}
        <div
          style={{
            maxWidth: '85%',
            padding: '12px',
            borderRadius: '8px',
            backgroundColor: isUser ? '#06b6d4' : '#ffffff',
            color: isUser ? '#ffffff' : '#1e293b',
            border: isUser ? 'none' : '2px solid #06b6d4',
            textAlign: 'right'
          }}
        >
          <p style={{
            fontSize: '14px',
            lineHeight: '1.5',
            whiteSpace: 'pre-line',
            margin: 0
          }}>{message.text}</p>
          {message.actions && message.actions.length > 0 && (
            <div style={{ display: 'flex', flexWrap: 'wrap', gap: '6px', marginTop: '10px', direction: 'rtl' }}>
              {message.actions.map((action) => (
                <button
                  key={`${message.id}-${action.path}`}
                  type="button"
                  onClick={() => { window.location.hash = action.path; }}
                  style={{
                    border: '1px solid rgba(6, 182, 212, 0.35)',
                    borderRadius: '7px',
                    padding: '5px 8px',
                    background: 'rgba(236, 254, 255, 0.8)',
                    color: '#155e75',
                    fontSize: '11px',
                    fontWeight: 600,
                    cursor: 'pointer',
                  }}
                >
                  {action.label}
                </button>
              ))}
            </div>
          )}
          <p style={{
            fontSize: '11px',
            marginTop: '4px',
            opacity: 0.7,
            margin: 0
          }}>
            {formatStoreTime(message.timestamp, 'ar-IL')}
          </p>
        </div>
      </div>
    );
  };

  const cue = assistantState === 'offline'
    ? getSystemInteraction('offline')
    : assistantState === 'error'
      ? getSystemInteraction('error')
      : assistantState === 'listening'
        ? { ...interaction, label: 'أستمع لك...' }
        : interaction;
  const chatContent = (
    <div
      style={{
        display: 'flex',
        flexDirection: 'column',
        height: '100%',
        width: '100%',
        backgroundColor: '#ffffff',
        borderRadius: '12px',
        boxShadow: '0 25px 50px -12px rgba(0, 0, 0, 0.25)',
        border: '2px solid #06b6d4',
        overflow: 'hidden',
        direction: 'rtl'
      }}
    >
      <div
        style={{
          background: 'linear-gradient(to right, #3b82f6, #06b6d4)',
          padding: '16px',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          flexShrink: 0,
          position: 'relative'
        }}
      >
        <div style={{ display: 'flex', alignItems: 'center', gap: '12px', flex: 1 }}>
          <div style={{ width: '40px', height: '40px', display: 'flex', alignItems: 'center', justifyContent: 'center', position: 'relative' }}>
            {assistantState !== 'idle' && (
              <AssistantCue interaction={cue} placement="above-header-avatar" />
            )}
            <NeonAIBot size={40} state={assistantState} />
          </div>
          <div>
            <h2 style={{ color: '#ffffff', fontWeight: 'bold', fontSize: '14px', margin: 0 }}>أمان</h2>
            <p style={{ color: 'rgba(255, 255, 255, 0.8)', fontSize: '12px', margin: 0 }}>المساعد الذكي • {offlineMode ? 'أوفلاين' : stateLabels[assistantState]}</p>
          </div>
        </div>

        {onClose && (
          <button
            onClick={onClose}
            style={{
              width: '32px',
              height: '32px',
              borderRadius: '8px',
              backgroundColor: 'rgba(255, 255, 255, 0.2)',
              color: '#ffffff',
              border: 'none',
              cursor: 'pointer',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              transition: 'background-color 0.15s'
            }}
            onMouseOver={(e) => e.currentTarget.style.backgroundColor = 'rgba(255, 255, 255, 0.3)'}
            onMouseOut={(e) => e.currentTarget.style.backgroundColor = 'rgba(255, 255, 255, 0.2)'}
          >
            <X style={{ width: '16px', height: '16px' }} />
          </button>
        )}
      </div>

      <div style={{ height: '320px', display: 'flex', flexDirection: 'column' }}>
        <div
          style={{
            flex: 1,
            overflowY: 'auto',
            padding: '16px',
            backgroundColor: '#ffffff',
            display: 'flex',
            flexDirection: 'column'
          }}
        >
          {messages.map((msg) => (
            <MessageBubble key={msg.id} message={msg} />
          ))}

          {isTyping && (
            <div style={{ display: 'flex', justifyContent: 'flex-start', gap: '8px' }}>
              <div style={{ width: '32px', height: '32px', display: 'flex', alignItems: 'center', justifyContent: 'center', flexShrink: 0 }}>
                <NeonAIBot size={32} state="thinking" />
              </div>
              <div
                style={{
                  backgroundColor: '#ffffff',
                  border: '2px solid #06b6d4',
                  padding: '12px',
                  borderRadius: '8px'
                }}
              >
                <div style={{ display: 'flex', gap: '4px' }}>
                  <div style={{ width: '8px', height: '8px', backgroundColor: '#06b6d4', borderRadius: '50%', animation: 'chatBounce 0.6s infinite' }} />
                  <div style={{ width: '8px', height: '8px', backgroundColor: '#06b6d4', borderRadius: '50%', animation: 'chatBounce 0.6s infinite', animationDelay: '0.15s' }} />
                  <div style={{ width: '8px', height: '8px', backgroundColor: '#06b6d4', borderRadius: '50%', animation: 'chatBounce 0.6s infinite', animationDelay: '0.3s' }} />
                </div>
              </div>
            </div>
          )}

          {showQuickActions && (
            <div style={{ padding: '0 16px 8px 16px' }}>
              <div style={{ display: 'flex', flexWrap: 'wrap', gap: '8px' }}>
                {quickActions.map((action) => (
                  <button
                    key={action.query}
                    onClick={() => handleQuickAction(action.query)}
                    style={{
                      padding: '6px 12px',
                      borderRadius: '8px',
                      border: '1px solid #e2e8f0',
                      backgroundColor: '#f8fafc',
                      color: '#334155',
                      fontSize: '12px',
                      cursor: 'pointer',
                      transition: 'background-color 0.15s'
                    }}
                    onMouseOver={(e) => e.currentTarget.style.backgroundColor = '#f1f5f9'}
                    onMouseOut={(e) => e.currentTarget.style.backgroundColor = '#f8fafc'}
                  >
                    {action.label}
                  </button>
                ))}
              </div>
            </div>
          )}

          <div ref={messagesEndRef} />
        </div>
      </div>

      <div
        style={{
          borderTop: '1px solid #e2e8f0',
          padding: '12px',
          flexShrink: 0,
          backgroundColor: '#ffffff'
        }}
      >
        <div style={{ display: 'flex', alignItems: 'flex-end', gap: '8px' }}>
          <textarea
            ref={inputRef}
            value={inputText}
            onChange={(e) => {
              const nextValue = e.target.value;
              setInputText(nextValue);
              if (!isTyping && !offlineMode) {
                clearStateTimer();
                setAssistantState(nextValue.trim() ? 'listening' : 'idle');
              }
            }}
            onKeyDown={(e) => e.key === 'Enter' && !e.shiftKey && (e.preventDefault(), handleSendMessage())}
            placeholder="اكتب رسالتك..."
            rows={1}
            style={{
              flex: 1,
              padding: '8px 12px',
              backgroundColor: '#f8fafc',
              color: '#334155',
              border: '1px solid #e2e8f0',
              borderRadius: '8px',
              resize: 'none',
              outline: 'none',
              fontSize: '14px',
              textAlign: 'right',
              direction: 'rtl',
              fontFamily: 'inherit'
            }}
          />
          <button
            onClick={handleSendMessage}
            disabled={!inputText.trim() || isTyping}
            style={{
              padding: '8px 16px',
              backgroundColor: '#06b6d4',
              color: '#ffffff',
              borderRadius: '8px',
              border: 'none',
              cursor: !inputText.trim() || isTyping ? 'not-allowed' : 'pointer',
              opacity: !inputText.trim() || isTyping ? 0.5 : 1,
              fontSize: '14px',
              flexShrink: 0,
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center'
            }}
          >
            {isTyping ? <Loader2 style={{ width: '16px', height: '16px', animation: 'chatSpin 1s linear infinite' }} /> : <Send style={{ width: '16px', height: '16px' }} />}
          </button>
        </div>
      </div>
    </div>
  );

  if (embedded) {
    return <div className="h-full">{chatContent}</div>;
  }

  const positionStyle: React.CSSProperties = {
    position: 'fixed',
    left: Math.min(position?.x || 0, typeof window !== 'undefined' ? window.innerWidth - 400 : 0),
    top: Math.min((position?.y || 0) + 80, typeof window !== 'undefined' ? window.innerHeight - 500 : 0),
    width: '384px',
    zIndex: 1000,
  };

  return (
    <div style={positionStyle} className="animate-in fade-in slide-in-from-bottom-4 duration-300 w-96">
      {chatContent}
    </div>
  );
}
