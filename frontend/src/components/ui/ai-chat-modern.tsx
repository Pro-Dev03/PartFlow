import { useEffect, useRef, useState } from 'react';
import { X, Send, Loader2 } from 'lucide-react';
import { generateAssistantReply } from '../../lib/assistant-response';
import type { AssistantContext } from '../../lib/assistant-response';
import NeonAIBot from './neon-ai-bot';

type ChatMessage = {
  id: string;
  role: 'assistant' | 'user';
  text: string;
  timestamp: Date;
  status?: 'sending' | 'sent' | 'delivered';
};

interface AIChatModernProps {
  onClose?: () => void;
  position?: { x: number; y: number };
  embedded?: boolean;
  isOffline?: boolean;
  assistantContext?: AssistantContext;
  onSendMessage?: (text: string) => void;
}

interface QuickAction {
  icon: string;
  label: string;
  query: string;
}

const defaultContext: AssistantContext = {
  lowStockCount: 0,
  overdueDebtsCount: 0,
  salesToday: 0,
  salesYesterday: 0,
  lowStockItems: [],
  overdueDebts: [],
};

export default function AIChatModern({
  onClose,
  position,
  embedded = false,
  isOffline,
  assistantContext,
  onSendMessage,
}: AIChatModernProps) {
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [inputText, setInputText] = useState('');
  const [isTyping, setIsTyping] = useState(false);
  const [isInitialized, setIsInitialized] = useState(false);
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const inputRef = useRef<HTMLTextAreaElement>(null);

  const offlineMode = isOffline ?? (typeof navigator !== 'undefined' && !navigator.onLine);
  const context = assistantContext || defaultContext;

  const quickActions: QuickAction[] = [
    { icon: '📦', label: 'إدارة المخزون', query: 'ما حال المخزون؟' },
    { icon: '💰', label: 'متابعة الديون والتحصيل', query: 'أخبرني عن الديون المتأخرة' },
    { icon: '📊', label: 'تحليل المبيعات والأرباح', query: 'ما الذي حدث في هذا الشهر؟' },
    { icon: '⚡', label: 'الحصول على اقتراحات عملية', query: 'هل لديك اقتراحات؟' },
  ];

  useEffect(() => {
    if (!isInitialized) {
      const welcomeMessage: ChatMessage = {
        id: crypto.randomUUID(),
        role: 'assistant',
        text: '👋 مرحباً! أنا مساعدك الذكي.\n\nكيف يمكنني مساعدتك؟',
        timestamp: new Date(),
        status: 'delivered',
      };
      setMessages([welcomeMessage]);
      setIsInitialized(true);
    }
  }, [isInitialized]);

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages, isTyping]);

  useEffect(() => {
    if (inputRef.current) {
      inputRef.current.style.height = 'auto';
      const newHeight = Math.min(inputRef.current.scrollHeight, 80);
      inputRef.current.style.height = `${newHeight}px`;
    }
  }, [inputText]);

  const showQuickActions = messages.length === 1 && messages[0].role === 'assistant';

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
    `;
    document.head.appendChild(style);
    return () => {
      document.head.removeChild(style);
    };
  }, []);

  const handleSendMessage = () => {
    if (!inputText.trim()) return;

    const userMessage: ChatMessage = {
      id: crypto.randomUUID(),
      role: 'user',
      text: inputText.trim(),
      timestamp: new Date(),
      status: 'sent',
    };

    setMessages((prev) => [...prev, userMessage]);
    setInputText('');
    setIsTyping(true);

    setTimeout(() => {
      const response = generateAssistantReply(userMessage.text, context);

      const assistantMessage: ChatMessage = {
        id: crypto.randomUUID(),
        role: 'assistant',
        text: response,
        timestamp: new Date(),
        status: 'delivered',
      };

      setMessages((prev) => [...prev, assistantMessage]);
      setIsTyping(false);
    }, 600);

    onSendMessage?.(inputText);
  };

  const handleQuickAction = (query: string) => {
    const userMessage: ChatMessage = {
      id: crypto.randomUUID(),
      role: 'user',
      text: query,
      timestamp: new Date(),
      status: 'sent',
    };

    setMessages((prev) => [...prev, userMessage]);
    setIsTyping(true);

    setTimeout(() => {
      const response = generateAssistantReply(query, context);

      const assistantMessage: ChatMessage = {
        id: crypto.randomUUID(),
        role: 'assistant',
        text: response,
        timestamp: new Date(),
        status: 'delivered',
      };

      setMessages((prev) => [...prev, assistantMessage]);
      setIsTyping(false);
    }, 600);

    onSendMessage?.(query);
  };

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
            <NeonAIBot size={32} />
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
          <p style={{
            fontSize: '11px',
            marginTop: '4px',
            opacity: 0.7,
            margin: 0
          }}>
            {message.timestamp.toLocaleTimeString('ar-IL', {
              hour: '2-digit',
              minute: '2-digit',
            })}
          </p>
        </div>
      </div>
    );
  };

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
          flexShrink: 0
        }}
      >
        <div style={{ display: 'flex', alignItems: 'center', gap: '12px', flex: 1 }}>
          <div style={{ width: '40px', height: '40px', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
            <NeonAIBot size={40} />
          </div>
          <div>
            <h2 style={{ color: '#ffffff', fontWeight: 'bold', fontSize: '14px', margin: 0 }}>أمان</h2>
            <p style={{ color: 'rgba(255, 255, 255, 0.8)', fontSize: '12px', margin: 0 }}>المساعد الذكي • {offlineMode ? 'أوفلاين' : 'متصل الآن'}</p>
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
                <NeonAIBot size={32} />
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
            onChange={(e) => setInputText(e.target.value)}
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
