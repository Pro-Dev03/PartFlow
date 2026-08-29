import { useEffect, useRef, useState } from 'react';
import { X, ArrowUpRight, Check, CheckCheck, Send } from 'lucide-react';
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

interface UnifiedAIChatProps {
  onClose?: () => void;
  position?: { x: number; y: number };
  embedded?: boolean;
  isOffline?: boolean;
  assistantContext?: AssistantContext;
  onSendMessage?: (text: string) => void;
}

const defaultContext: AssistantContext = {
  lowStockCount: 0,
  overdueDebtsCount: 0,
  salesToday: 0,
  salesYesterday: 0,
  lowStockItems: [],
  overdueDebts: [],
};

export default function UnifiedAIChat({
  onClose,
  position,
  embedded = false,
  isOffline,
  assistantContext,
  onSendMessage
}: UnifiedAIChatProps) {
  const [messages, setMessages] = useState<ChatMessage[]>([
    {
      id: 'welcome',
      role: 'assistant',
      text: 'مرحبًا 👋 أنا أمان، مساعدك الذكي\n\nيمكنني مساعدتك في إدارة متجرك. ماذا تريد أن تفعل؟',
      timestamp: new Date(),
      status: 'delivered',
    },
  ]);
  const [inputText, setInputText] = useState('');
  const [isTyping, setIsTyping] = useState(false);
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const offlineMode = isOffline ?? (typeof navigator !== 'undefined' && !navigator.onLine);
  const context = assistantContext || defaultContext;

  const quickActions = [
    { id: 'inventory', icon: '📦', label: 'إدارة المخزون', action: 'أظهر لي المخزون' },
    { id: 'debts', icon: '💰', label: 'متابعة الديون والتحصيل', action: 'هل توجد ديون متأخرة؟' },
    { id: 'sales', icon: '📊', label: 'تحليل المبيعات والأرباح', action: 'أعطني ملخص المبيعات' },
    { id: 'suggestions', icon: '⚡', label: 'الحصول على اقتراحات عملية', action: 'ما هي أولويات اليوم؟' },
  ];

  const suggestedQuestions = [
    'ما الذي يحدث الآن؟',
    'ما هي أولويات اليوم؟',
    'هل توجد ديون متأخرة؟',
    'أظهر لي المخزون',
    'أعطني ملخص المبيعات',
  ];

  const scrollToBottom = () => {
    if (typeof messagesEndRef.current?.scrollIntoView === 'function') {
      messagesEndRef.current.scrollIntoView({ behavior: 'smooth' });
    }
  };

  useEffect(() => {
    scrollToBottom();
  }, [messages]);

  const handleSendMessage = async () => {
    if (inputText.trim()) {
      const userMessage: ChatMessage = {
        id: crypto.randomUUID(),
        role: 'user',
        text: inputText.trim(),
        timestamp: new Date(),
        status: 'sent',
      };

      setMessages(prev => [...prev, userMessage]);
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
        setMessages(prev => [...prev, assistantMessage]);
        setIsTyping(false);
      }, 500 + Math.random() * 500);

      onSendMessage?.(userMessage.text);
    }
  };

  const handleSuggestedQuestion = (question: string) => {
    const userMessage: ChatMessage = {
      id: crypto.randomUUID(),
      role: 'user',
      text: question,
      timestamp: new Date(),
      status: 'sent',
    };
    setMessages(prev => [...prev, userMessage]);
    setIsTyping(true);

    setTimeout(() => {
      const response = generateAssistantReply(question, context);
      const assistantMessage: ChatMessage = {
        id: crypto.randomUUID(),
        role: 'assistant',
        text: response,
        timestamp: new Date(),
        status: 'delivered',
      };
      setMessages(prev => [...prev, assistantMessage]);
      setIsTyping(false);
    }, 500 + Math.random() * 500);

    onSendMessage?.(question);
  };

  const showSuggestions = messages.length === 0 || (messages.length === 1 && messages[0].role === 'assistant');

  const MessageStatus = ({ status }: { status?: string }) => {
    if (!status) return null;
    return (
      <span className="inline-flex items-center mr-1">
        {status === 'sent' && <Check className="w-3 h-3 opacity-70" />}
        {status === 'delivered' && <CheckCheck className="w-3 h-3 text-cyan-300" />}
      </span>
    );
  };

  const MessageBubble = ({ message }: { message: ChatMessage }) => {
    const isUser = message.role === 'user';
    return (
      <div
        className={`flex items-end gap-3 ${isUser ? 'justify-end' : 'justify-start'} animate-in fade-in slide-in-from-bottom-2 duration-300`}
      >
        {!isUser && (
          <div className="w-8 h-8 rounded-full bg-gradient-to-br from-blue-400 to-cyan-400 flex items-center justify-center shrink-0 shadow-md">
            <span className="text-sm">✨</span>
          </div>
        )}
        <div
          className={`max-w-[75%] px-4 py-3 rounded-2xl transition-all duration-200 ${
            isUser
              ? 'bg-gradient-to-br from-slate-900 to-slate-800 text-white rounded-br-sm shadow-md'
              : 'bg-white border border-slate-200 text-slate-800 shadow-sm dark:border-slate-700 dark:bg-slate-800 dark:text-slate-100 rounded-bl-sm'
          }`}
        >
          <p className="text-sm whitespace-pre-line leading-relaxed font-normal">{message.text}</p>
          <div className={`mt-2 flex items-center justify-end gap-2 text-[11px] ${isUser ? 'text-slate-300' : 'text-slate-400 dark:text-slate-500'}`}>
            {message.timestamp.toLocaleTimeString('ar-IL', {
              hour: '2-digit',
              minute: '2-digit',
            })}
            {isUser && <MessageStatus status={message.status} />}
          </div>
        </div>
        {isUser && (
          <div className="w-8 h-8 rounded-full bg-gradient-to-br from-slate-300 to-slate-400 dark:from-slate-600 dark:to-slate-500 flex items-center justify-center shrink-0 shadow-md">
            <span className="text-sm">👤</span>
          </div>
        )}
      </div>
    );
  };

  const popupStyle = !embedded ? {
    position: 'fixed' as const,
    right: typeof window !== 'undefined' ? '20px' : 'auto',
    left: typeof window !== 'undefined' ? 'auto' : Math.min(position?.x || 0, typeof window !== 'undefined' ? window.innerWidth - 400 : 0),
    top: Math.min((position?.y || 0) + 65, typeof window !== 'undefined' ? window.innerHeight - 600 : 0),
    width: typeof window !== 'undefined' && window.innerWidth < 640 ? 'calc(100vw - 40px)' : '380px',
    maxWidth: 'calc(100vw - 40px)',
    zIndex: 9999,
  } : {};

  const chatContainer = !embedded ? (
    <div style={popupStyle} className="animate-in fade-in slide-in-from-bottom-4 duration-300">
      <div className="bg-white rounded-2xl shadow-[0_20px_60px_rgba(15,23,42,0.15)] border border-slate-200 overflow-hidden">
        <div className="flex items-center justify-between px-4 py-3 border-b border-slate-200 bg-white">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 rounded-full bg-gradient-to-br from-blue-400 to-cyan-400 flex items-center justify-center shadow-md">
              <span className="text-lg">🤖</span>
            </div>
              <div className="flex flex-col">
                <h3 className="text-sm font-semibold text-slate-900">أمان</h3>
                <div className="flex items-center gap-1.5">
                  <span className={`h-2 w-2 rounded-full ${offlineMode ? 'bg-amber-500' : 'bg-emerald-500'} ${!offlineMode ? 'animate-pulse' : ''}`} />
                  <span className="text-xs text-slate-500">{offlineMode ? 'أوفلاين' : 'متصل الآن'}</span>
                </div>
              </div>
          </div>
          {onClose && (
            <button
              type="button"
              onClick={onClose}
              aria-label="إغلاق"
              className="flex h-8 w-8 items-center justify-center rounded-full border border-slate-200 bg-slate-100 text-slate-600 transition-all duration-150 hover:border-slate-300 hover:bg-slate-200 hover:text-slate-800"
            >
              <X className="h-4 w-4" strokeWidth={2} />
            </button>
          )}
        </div>

        <div className="flex flex-col bg-slate-50" style={{ height: 'min(380px, 60vh)' }}>
          <div className="flex-1 overflow-y-auto p-4 space-y-4 custom-scrollbar">
            {messages.map((message) => (
              <MessageBubble key={message.id} message={message} />
            ))}

            {isTyping && (
              <div className="flex items-end gap-3 justify-start animate-in fade-in duration-300">
                <div className="w-8 h-8 rounded-full bg-gradient-to-br from-blue-400 to-cyan-400 flex items-center justify-center shrink-0 shadow-md">
                  <span className="text-sm">✨</span>
                </div>
                <div className="bg-white border border-slate-200 px-4 py-3 rounded-2xl rounded-bl-sm shadow-sm">
                  <div className="flex gap-1.5 items-center">
                    <div className="w-2 h-2 bg-gradient-to-r from-blue-400 to-cyan-400 rounded-full animate-bounce" style={{ animationDelay: '0ms' }} />
                    <div className="w-2 h-2 bg-gradient-to-r from-blue-400 to-cyan-400 rounded-full animate-bounce" style={{ animationDelay: '150ms' }} />
                    <div className="w-2 h-2 bg-gradient-to-r from-blue-400 to-cyan-400 rounded-full animate-bounce" style={{ animationDelay: '300ms' }} />
                  </div>
                </div>
              </div>
            )}

            <div ref={messagesEndRef} />
          </div>

          {showSuggestions && (
            <div className="border-t border-slate-200 bg-white px-4 py-3">
              <div className="grid grid-cols-2 gap-2">
                {quickActions.map((action) => (
                  <button
                    key={action.id}
                    type="button"
                    onClick={() => handleSuggestedQuestion(action.action)}
                    className="flex items-center gap-2 px-3 py-2.5 rounded-xl border border-slate-200 bg-slate-50 text-slate-700 text-xs font-medium transition-all duration-150 hover:border-slate-300 hover:bg-slate-100 hover:shadow-sm active:scale-95"
                  >
                    <span className="text-base">{action.icon}</span>
                    <span className="text-right flex-1">{action.label}</span>
                  </button>
                ))}
              </div>
            </div>
          )}

          <div className="border-t border-slate-200 bg-white p-3">
            <div className="flex items-center gap-2 rounded-xl border border-slate-200 bg-slate-50 px-3 py-2 transition-all duration-150 focus-within:border-sky-300 focus-within:bg-white focus-within:shadow-[0_0_0_3px_rgba(14,165,233,0.1)]">
              <input
                type="text"
                value={inputText}
                onChange={(e) => setInputText(e.target.value)}
                onKeyDown={(e) => e.key === 'Enter' && handleSendMessage()}
                placeholder="اكتب رسالتك..."
                aria-label="رسالة جديدة"
                className="flex-1 bg-transparent text-sm text-slate-700 outline-none placeholder:text-slate-400 text-right"
                dir="rtl"
              />
              <button
                type="button"
                onClick={handleSendMessage}
                disabled={!inputText.trim() || isTyping}
                aria-label="إرسال"
                className="flex h-8 w-8 shrink-0 items-center justify-center rounded-lg bg-gradient-to-br from-slate-900 to-slate-800 text-white shadow-sm transition-all duration-150 hover:from-slate-800 hover:to-slate-700 disabled:cursor-not-allowed disabled:from-slate-200 disabled:to-slate-300 disabled:text-slate-400"
              >
                <Send className="h-4 w-4" strokeWidth={2} />
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  ) : (
    <div className="space-y-6 animate-in fade-in duration-500" dir="rtl">
      <div className="flex items-center gap-4">
        <div style={{ width: '56px', height: '56px' }} className="animate-pulse">
          <NeonAIBot size={56} />
        </div>
        <div>
          <h1 className="text-2xl font-bold text-slate-900 dark:text-slate-100 bg-gradient-to-r from-blue-600 to-cyan-600 bg-clip-text text-transparent">مساعد PartFlow الذكي</h1>
          <p className="text-sm text-slate-600 dark:text-slate-400">مساعدك الخاص لإدارة متجرك</p>
        </div>
      </div>

      <div className="bg-white rounded-2xl shadow-xl border border-slate-200 overflow-hidden">
        <div className="h-[600px] flex flex-col">
          <div className="flex-1 overflow-y-auto p-6 space-y-4 bg-slate-50 custom-scrollbar">
            {messages.map((message) => (
              <MessageBubble key={message.id} message={message} />
            ))}

            {isTyping && (
              <div className="flex items-end gap-3 justify-start animate-in fade-in duration-300">
                <div className="w-8 h-8 rounded-full bg-gradient-to-br from-blue-400 to-cyan-400 flex items-center justify-center shrink-0 shadow-md">
                  <span className="text-sm">✨</span>
                </div>
                <div className="bg-white border border-slate-200 px-4 py-3 rounded-2xl rounded-bl-sm shadow-sm">
                  <div className="flex gap-1.5 items-center">
                    <div className="w-2 h-2 bg-gradient-to-r from-blue-400 to-cyan-400 rounded-full animate-bounce" style={{ animationDelay: '0ms' }} />
                    <div className="w-2 h-2 bg-gradient-to-r from-blue-400 to-cyan-400 rounded-full animate-bounce" style={{ animationDelay: '150ms' }} />
                    <div className="w-2 h-2 bg-gradient-to-r from-blue-400 to-cyan-400 rounded-full animate-bounce" style={{ animationDelay: '300ms' }} />
                  </div>
                </div>
              </div>
            )}

            <div ref={messagesEndRef} />
          </div>

          {showSuggestions && (
            <div className="border-t border-slate-200 bg-white px-6 py-4">
              <div className="grid grid-cols-2 gap-3">
                {quickActions.map((action) => (
                  <button
                    key={action.id}
                    type="button"
                    onClick={() => handleSuggestedQuestion(action.action)}
                    className="flex items-center gap-3 px-4 py-3 rounded-xl border border-slate-200 bg-slate-50 text-slate-700 text-sm font-medium transition-all duration-150 hover:border-slate-300 hover:bg-slate-100 hover:shadow-sm active:scale-95"
                  >
                    <span className="text-lg">{action.icon}</span>
                    <span className="text-right flex-1">{action.label}</span>
                  </button>
                ))}
              </div>
            </div>
          )}

          <div className="border-t border-slate-200 bg-white p-4">
            <div className="flex items-center gap-3 rounded-xl border border-slate-200 bg-slate-50 px-4 py-3 transition-all duration-150 focus-within:border-sky-300 focus-within:bg-white focus-within:shadow-[0_0_0_3px_rgba(14,165,233,0.1)]">
              <input
                type="text"
                value={inputText}
                onChange={(e) => setInputText(e.target.value)}
                onKeyDown={(e) => e.key === 'Enter' && handleSendMessage()}
                placeholder="اكتب رسالتك..."
                aria-label="رسالة جديدة"
                className="flex-1 bg-transparent text-sm text-slate-700 outline-none placeholder:text-slate-400 text-right"
                dir="rtl"
              />
              <button
                type="button"
                onClick={handleSendMessage}
                disabled={!inputText.trim() || isTyping}
                aria-label="إرسال"
                className="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-gradient-to-br from-slate-900 to-slate-800 text-white shadow-sm transition-all duration-150 hover:from-slate-800 hover:to-slate-700 disabled:cursor-not-allowed disabled:from-slate-200 disabled:to-slate-300 disabled:text-slate-400"
              >
                <Send className="h-5 w-5" strokeWidth={2} />
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  );

  return chatContainer;
}
