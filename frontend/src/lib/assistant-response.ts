export type AssistantContext = {
  lowStockCount: number;
  overdueDebtsCount: number;
  salesToday: number;
  salesYesterday: number;
  todayProfit?: number;
  totalCustomers?: number;
  totalProducts?: number;
  lowStockItems?: Array<{
    name?: string;
    quantity?: number;
    minQuantity?: number;
    category?: string;
  }>;
  overdueDebts?: Array<{
    customerName?: string;
    amount?: number;
    days?: number;
    phone?: string;
  }>;
  topSellingProducts?: Array<{
    name?: string;
    quantity?: number;
    revenue?: number;
  }>;
};

function normalizeMessage(message: string): string {
  return message.trim().toLowerCase();
}

function formatCurrency(value?: number): string {
  const safe = Number(value ?? 0);
  return new Intl.NumberFormat('ar-EG', {
    maximumFractionDigits: 0,
  }).format(safe);
}

const assistantName = 'أمان';

function pick<T>(arr: T[], index: number): T {
  return arr[index % arr.length];
}

function topItemName(items: Array<{ name?: string }>[], fallback = 'منتج') {
  return items?.[0]?.name || items?.[0]?.product_name || fallback;
}

function randomBetween(min: number, max: number) {
  return Math.floor(Math.random() * (max - min + 1)) + min;
}

function mentionStock(context: AssistantContext, prefix = ''): string {
  if (context.lowStockCount <= 0) {
    return prefix ? `${prefix}المخزون تمام وما في نقص حالياً، فبتركز على البيع وخلاص.` : 'المخزون تمام وما في نقص حالياً، فبتركز على البيع وخلاص.';
  }
  const names = (context.lowStockItems || [])
    .slice(0, 3)
    .map((item) => item.name || 'منتج')
    .join(' و ');
  const count = context.lowStockCount;
  const phrases = [
    `في ${count} عنصر تحتاج طلب، أبرزهم ${names}. لا تنسى تجهيز الطلب قبل ما ينفد ويوقفك عن البيع.`,
    `عندك ${count} عنصر ناقص، وهم ${names}. الأفضل تطلبهم اليوم عشان ما تتأخر على الزبائن.`,
    `${count} قطع وصلوا للحد الأدنى، منهم ${names}. خذ بالك منهم عشان ما يصير لك نقص في وسط البيع.`,
  ];
  return pick(phrases, randomBetween(0, 2));
}

function mentionDebts(context: AssistantContext, prefix = ''): string {
  if (context.overdueDebtsCount <= 0) {
    return prefix ? `${prefix}الديون تمام وكل العملاء منضبطين، ماشي الحال.` : 'الديون تمام وكل العملاء منضبطين، ماشي الحال.';
  }
  const top = context.overdueDebts?.[0];
  const count = context.overdueDebtsCount;
  const base = `في ${count} دين متأخر يحتاج متابعة.`;
  const details = top
    ? ` وأبعدهم ${top.customerName}، مبلغه ${formatCurrency(top.amount)} وتأخر ${top.days ?? 0} يوم.`
    : '';
  const suffixes = [
    'هالشي يضغط على السيولة، فالاتصال بهم اليوم يخليك مرتاح.',
    'لا ت procrastinate عليهن، خذ لك دقيقة وتصل بهم وتخلص منه.',
    'التحصيل اليوم بيسهل عليك الأيام الجاية، فلا تتوانى.',
  ];
  return `${base}${details} ${pick(suffixes, randomBetween(0, 2))}`;
}

function mentionSales(context: AssistantContext, prefix = ''): string {
  const today = Number(context.salesToday ?? 0);
  const yesterday = Number(context.salesYesterday ?? 0);
  const profit = Number(context.todayProfit ?? 0);

  if (today <= 0 && yesterday <= 0) {
    return prefix ? `${prefix}ما في مبيعات مسجلة حالياً. ممكن تبدأ بتسجيل عملية بيع عشان نعرف شو وضع المتجر.` : 'ما في مبيعات مسجلة حالياً. ممكن تبدأ بتسجيل عملية بيع عشان نعرف شو وضع المتجر.';
  }

  const profitLine = profit > 0 ? ` والربح اليوم تقريباً ${formatCurrency(profit)}.` : '';
  let deltaLine = '';
  if (yesterday > 0) {
    const delta = ((today - yesterday) / yesterday) * 100;
    if (delta > 0) {
      deltaLine = ` وهاد أفضل من أمس بنسبة ${Math.abs(delta).toFixed(1)}%. يعني الحركة ماشية وماشية.`;
    } else if (delta < 0) {
      deltaLine = ` وهاد أقل من أمس بنسبة ${Math.abs(delta).toFixed(1)}%. ممكن تحتاج تشوف شو السبب وتعدل الخطة.`;
    } else {
      deltaLine = ' وهاد زي أمس بالضبط، يعني مستقر.';
    }
  } else {
    deltaLine = ' وهاد هو رقم اليوم.';
  }

  const phrases = [
    `المبيعات اليوم وصلت لـ ${formatCurrency(today)}.${profitLine}${deltaLine}`,
    `رقم المبيعات اليوم ${formatCurrency(today)}.${profitLine}${deltaLine}`,
    `اليوم بيعت ${formatCurrency(today)}.${profitLine}${deltaLine}`,
  ];
  return prefix ? `${prefix}${pick(phrases, randomBetween(0, 2))}` : pick(phrases, randomBetween(0, 2));
}

function mentionTopProducts(context: AssistantContext): string | undefined {
  const items = context.topSellingProducts || [];
  if (!items.length) return undefined;
  const top = items[0];
  const name = top.name || 'منتج';
  const qty = Number(top.quantity ?? 0);
  const rev = Number(top.revenue ?? 0);
  const phrases = [
    `والمنتج اللي بيشتغل أحسن هو ${name}، باعنا ${qty} وحدة ودرّ علينا ${formatCurrency(rev)}.`,
    `أحسن مبيع اليوم هو ${name}، مبيعاته ${qty} وحدة وإيراده ${formatCurrency(rev)}.`,
    `النجم اليوم هو ${name}، باع ${qty} وحدة وجاب إيراد ${formatCurrency(rev)}.`,
  ];
  return pick(phrases, randomBetween(0, 2));
}

function smallTalk(message: string): string | undefined {
  const text = normalizeMessage(message);

  if (/شخبارك|شو\s*حالك|كيف\s*حالك|كيف\s*أموري/.test(text)) {
    const responses = [
      'تمام تمام، الحمد لله. كل يوم بيعرفني على وضع المتجر وخلاص. شو في جديد عندك؟',
      'الحمد لله بخير، وكل ما أشوف وضعك أحسن بفرح. شو بدك نراجع اليوم؟',
      'ماشي الحال، وكل يوم أتعلم منك شي جديد. شو اللي يخوفك اليوم؟',
      'عامل زي الفل، والحمد لله. شو رأيك نطلع على المخزون أو الديون؟',
    ];
    return pick(responses, randomBetween(0, 3));
  }

  if (/من\s*أنت|who\s*are\s*you|أنت\s*من|انت\s*من|من\s*هو\s*أنت/.test(text)) {
    const responses = [
      `أنا ${assistantName}، مساعدك الشخصي في PartFlow. بقلب عليك، بتابع المخزون، الديون، والمبيعات، وبقولك شو اللي يحتاج انتباهك أول.`,
      `اسمي ${assistantName}، وأنا كالسكرتير بتاعك في المتجر. كل شي تبي تسأله عن المخزون أو الديون أو المبيعات، أنا معك.`,
      `أنا ${assistantName}، صاحبك في PartFlow. هون عشان أسهل عليك حياتك وأقولك شو لازم تعمله اليوم.`,
    ];
    return pick(responses, randomBetween(0, 2));
  }

  if (/مرحبا|اهلا|hello|hi|السلام|السلام\s*عليكم|هاي|مرحبة|اهلاً|هلا|صباح\s*الخير|مساء\s*الخير/.test(text)) {
    const responses = [
      `أهلاً بك! شو بدك اليوم؟ أقدر أطلعلك على شو عم يصير في المتجر بثواني.`,
      `هاي! شو الأخبار؟ إذا بدك نراجع المخزون أو الديون أو المبيعات، قل لي.`,
      `أهلاً وسهلاً! شو تبي نساعدك فيه اليوم؟`,
      `مرحباً! أنا ${assistantName}، هون عشان أسهل عليك اليوم. شو بدك نبدأ بيه؟`,
    ];
    return pick(responses, randomBetween(0, 3));
  }

  if (/شكرا|شكرا\s*لك|شكرا\s*جزيلا|thanks|thank\s*you|merci/.test(text)) {
    const responses = [
      'عفواً، دايماً معك. شو حابب نساعدك بيه تاني؟',
      'لا شكر على واجب! إذا عندك سؤال تاني، أنا موجود.',
      'على الرحب والسعة! شو رأيك نراجع شو صار في المتجر؟',
      'كل شيء عشانك. شو تبي نعمل تاني؟',
    ];
    return pick(responses, randomBetween(0, 3));
  }

  if (/باي|مع\s*السلامة|bye|goodbye|إلى\s*اللقاء/.test(text)) {
    const responses = [
      'مع السلامة! إذا صار أي شي، ارجع لي وأنا موجود.',
      'باي! خد بالك من المتجر وأنا موجود إذا احتجتني.',
      'إلى اللقاء! لا تنسى تراجع المخزون والديون من وقت للتاني.',
      'سلامات! نراك قريب إن شاء الله.',
    ];
    return pick(responses, randomBetween(0, 3));
  }

  if (/تمام|حسنا|oke|ok|موافق|ماشي|ممتاز/.test(text)) {
    const responses = [
      'تمام، إذا بدك حاجة تانية قل لي.',
      'ماشي الحال! شو رأيك نراجع باقي الأقسام؟',
      'حسناً، دايماً موجود عشان أساعدك.',
      'ممتاز، هيا نكمل.',
    ];
    return pick(responses, randomBetween(0, 3));
  }

  return undefined;
}

export function generateAssistantReply(message: string, context: AssistantContext): string {
  const text = normalizeMessage(message);
  const isUrgent = /مستعجل|urgent|طارئ|طوارئ|حالة\s*طارئة|عاجل|حالاً|سريع|بسرعة/.test(text);
  const isFriendly = /كيف\s*الحال|كيف\s*أنت|شخبارك|شو\s*حالك|مرحبا|اهلا|السلام|هاي|هلا/.test(text);
  const isFirm = /مهم|ضروري|لازم|يجب|أكيد|مطلوب|بشكل\s*صارم|بالتأكيد/.test(text);

  if (!text) {
    const responses = [
      `أهلاً، أنا ${assistantName}. شو بدك اليوم؟ أقدر أطلعلك على شو عم يصير في المتجر وأقول لك شو اللي يحتاج انتباهك أول.`,
      `مرحباً! أنا ${assistantName}، شو في جديد؟ إذا بدك نراجع المخزون أو الديون أو المبيعات، قل لي.`,
      `أهلاً بك! أنا ${assistantName}، هون عشان أسهل عليك. شو تبي نساعدك فيه اليوم؟`,
    ];
    return pick(responses, randomBetween(0, 2));
  }

  const smallTalkReply = smallTalk(message);
  if (smallTalkReply) {
    return smallTalkReply;
  }

  const asksWhoAmI = /من\s*أنت|who\s*are\s*you|من\s*كنت|أنت\s*من|انت\s*من|أنت\s*من\s*هو|من\s*هو\s*أنت/.test(text);
  if (asksWhoAmI) {
    const responses = [
      `أنا مساعدك الشخصي في PartFlow، وأنا ${assistantName}. بقلب عليك، بتابع المخزون، أراقب الديون، وأقولك شو عم يصير في المتجر.`,
      `أنا ${assistantName}، مساعدك الشخصي هنا. هون عشان أسهل عليك حياتك وأقولك شو لازم تعمله أول.`,
      `اسمي ${assistantName}، وأنا صاحبك في المتجر. بتابع المخزون، الديون، والمبيعات، وأعطيك الخلاصة بسرعة.`,
    ];
    return pick(responses, randomBetween(0, 2));
  }

  const asksHowAreYou = /كيف\s*الحال|كيف\s*أنت|how\s*are\s*you|كيف\s*حالك|شخبارك|شو\s*حالك|كيف\s*الشي/.test(text);
  if (asksHowAreYou) {
    const responses = [
      'أنا بخير، الحمد لله، ورايح على الخدمة. شو اللي بدك نركز عليه اليوم؟',
      'تمام، الحمد لله، وكل يوم أتعلم منك شي جديد. شو رأيك نطلع على شو عم يصير في المتجر؟',
      'أنا بخير وجاهز لمساعدتك في متجرك. كيف حالك أنت؟ شو في جديد؟',
      'ماشي الحال، وكل ما أشوف وضعك أحسن بفرح. شو في جديد عندك؟',
      'عامل زي الفل، والحمد لله. شو اللي يخوفك اليوم؟',
      'أنا بخير، الحمد لله. وأستطيع متابعة المخزون والديون والمبيعات لك. شو تبي نراجع أول؟',
    ];
    return pick(responses, randomBetween(0, 5));
  }

  const hasGreeting = /^(مرحبا|اهلا|hello|hi|السلام|السلام عليكم|هاي|مرحبة|اهلاً|هلا|هاي|هلا|صباح\s*الخير|مساء\s*الخير)/.test(text);
  if (hasGreeting) {
    const responses = [
      `أهلاً، أنا ${assistantName}. شو بدك اليوم؟ أقدر أطلعلك على شو عم يصير في المتجر وأقول لك شو اللي يحتاج انتباهك أول.`,
      `هاي! شو الأخبار؟ إذا بدك نراجع المخزون أو الديون أو المبيعات، قل لي.`,
      `أهلاً وسهلاً! شو تبي نساعدك فيه اليوم؟`,
    ];
    return pick(responses, randomBetween(0, 2));
  }

  const asksAboutCurrentStatus = /ما.*يحدث.*الآن|ماذا.*يحدث.*الآن|ما.*يحدث|حالة.*المتجر|حالة.*الآن|status|what.*happening|what.*is.*happening|شو\s*عم\s*يصير|شو\s*صار|شو\s*الوضيعة|شو\s*الخبر|الآن|وضع.*المتجر|شنو.*الوضع|شو.*في/.test(text);
  if (asksAboutCurrentStatus) {
    const issues: string[] = [];
    if (context.lowStockCount > 0) issues.push(`${context.lowStockCount} عناصر تحتاج إعادة طلب`);
    if (context.overdueDebtsCount > 0) issues.push(`${context.overdueDebtsCount} ديون متأخرة تحتاج متابعة`);

    if (issues.length === 0) {
      const topProductsLine = mentionTopProducts(context);
      const profitLine = typeof context.todayProfit === 'number' && context.todayProfit > 0 ? ` وربح اليوم ${formatCurrency(context.todayProfit)}.` : '';
      const salesLine = context.salesToday > 0 ? ` مبيعات اليوم ${formatCurrency(context.salesToday)}.` : '';
      const stableResponses = [
        `أولويات اليوم واضحة: ما في تنبيهات حرجة في المخزون أو الديون، والمتجر مستقر.${salesLine}${profitLine} الأن الأفضل إنك تركز على المبيعات وتبقي الوضع مستمر على هذا النمط.`,
        `تمام ماشي الحال اليوم. ما في مشاكل في المخزون أو الديون.${salesLine}${profitLine} خد بالك من المبيعات وكمل بهذا الزخم.`,
        `اليوم ماشي زي ماshould be: ما في نقص ولا ديون متأخرة.${salesLine}${profitLine} فبتركز على البيع وتكبّش زخم.`,
      ];
      if (topProductsLine) {
        return `${pick(stableResponses, randomBetween(0, 2))} ${topProductsLine}`;
      }
      return pick(stableResponses, randomBetween(0, 2));
    }

    const stockLine = mentionStock(context);
    const debtLine = mentionDebts(context);
    const topProductsLine = mentionTopProducts(context);

    if (isUrgent) {
      const urgentResponses = [
        `مستعجل؟ نعم، أولويات اليوم واضحة: ${issues.join(' و ')}. خذ هالخطوات الآن: طلب المخزون، واتباع التحصيل، ثم راجع المبيعات. هذا هو الطريق الأسرع لتأمين المتجر.`,
        `عندك ${issues.join(' و ')}. لا تنتظر حتى يتأخر الزبائن أو يضغط الموردين. خد بالك من هالمواضيع اليوم.`,
        `مستعجل صحيح. ${stockLine} ${debtLine} ابدأ بهم من دلوقتي عشان ما تكبر المشكلة.`,
      ];
      return pick(urgentResponses, randomBetween(0, 2));
    }

    const normalResponses = [
      `أولويات اليوم واضحة: ${issues.join(' و ')}. يعني أهم شي الآن إنك تراجع المخزون والديون قبل ما تشتغل على أي حاجة تانية، لأن هالشي مباشر على تشغيل المتجر.`,
      `شفت شو في؟ عندك ${issues.join(' و ')}. فلا تستهين بهم، خدهم بجدية اليوم.`,
      `الوضع واضح: ${issues.join(' و ')}. يعني لازم تبدأ بهم من الصباح عشان تنهي يومك ومرتاح.`,
    ];
    return pick(normalResponses, randomBetween(0, 2));
  }

  const asksAboutLowStock = /مخزون|stock|تجديد|طلب.*مخزون|محتاج.*مخزون|نقص.*مخزون|القطع.*ناقصة|كم.*الكمية|المنتجات.*منخفض|شو.*مخزون|كم.*المنتجات|منتجات|اطلب/.test(text);
  if (asksAboutLowStock) {
    if (context.lowStockCount > 0) {
      const stockLine = mentionStock(context);
      if (isUrgent) {
        const urgentResponses = [
          `مستعجل؟ نعم، ${stockLine} لازم تجهز الطلب اليوم قبل ما يطلع النقص على البيع.`,
          `عاجل؟ ${stockLine} لا تنتظر حتى الغد، خد بالك منهم حالاً.`,
          `مستعجل صحيح. ${stockLine} وخليك على بالك إن الزبائن ما بيستنوا.`,
        ];
        return pick(urgentResponses, randomBetween(0, 2));
      }
      const normalResponses = [
        stockLine,
        `${stockLine} وخلينا نتابع هالشي كل يوم عشان ما يصير مفاجأة.`,
        `${stockLine} وممكن نضيف على هالطلبات من وقت للتاني عشان ما ينفد.`,
      ];
      return pick(normalResponses, randomBetween(0, 2));
    }
    const total = context.totalProducts ?? 0;
    const extra = total > 0 ? ` وعدد المنتجات الكلي في المتجر ${formatCurrency(total)} منتج.` : '';
    const goodResponses = [
      `المخزون مستقر الآن، ما في عناصر تحت الحد الأدنى في الوقت الحالي.${extra} يعني الأمور جيدة، ويمكنك تركز على المبيعات والتحصيل.`,
      `تمام، المخزون كويس وما في نقص حالياً.${extra} فبتركز على البيع وخلاص.`,
      `ماشي الحال، المخزون منضبط.${extra} خد بالك من المبيعات واستمر على هذا النمط.`,
    ];
    return pick(goodResponses, randomBetween(0, 2));
  }

  const asksAboutDebt = /دين|ديون|debt|تحصيل|متأخر|مستحق|مديون|مستحقات|التحصيل|شو\s*في\s*الديون|عملاء.*متأخرين/.test(text);
  if (asksAboutDebt) {
    if (context.overdueDebtsCount > 0) {
      const debtLine = mentionDebts(context);
      if (isUrgent) {
        const urgentResponses = [
          `مستعجل؟ نعم، هذا مهم. ${debtLine} لازم تتصل وتطلب التحصيل اليوم، لأن التأخير يضغط على السيولة بشكل مباشر.`,
          `عاجل؟ ${debtLine} لا تنتظر حتى يتأخر المبلغ أكتر، خد بالك منهم اليوم.`,
          `مستعجل صحيح. ${debtLine} والتحصيل اليوم بيوفر عليك ضغط الأيام الجاية.`,
        ];
        return pick(urgentResponses, randomBetween(0, 2));
      }
      const normalResponses = [
        debtLine,
        `${debtLine} وممكن تعطيهم خيار دفع مرن عشان يخلصو بسرعة.`,
        `${debtLine} وترا إنك ما بتاخد فلوس منهم، فالخطوة الأولى هي الاتصال بهم.`,
      ];
      return pick(normalResponses, randomBetween(0, 2));
    }
    const totalCustomers = context.totalCustomers ?? 0;
    const extra = totalCustomers > 0 ? ` وعدد العملاء الكلي ${formatCurrency(totalCustomers)} عميل.` : '';
    const goodResponses = [
      `ما في ديون متأخرة الآن، وهذا إشارة جيدة.${extra} الوضع المالي مستقر نسبياً، ويمكنك تركز على البيع بدون ضغوط إضافية.`,
      `تمام، الديون منضبطة وما في متأخرات.${extra} فبتركز على المبيعات وخلاص.`,
      `ماشي الحال، العملاء كلهم يسددوا في وقتهم.${extra} خد بالك من الزبائن الجداد عشان ما يصير مشاكل.`,
    ];
    return pick(goodResponses, randomBetween(0, 2));
  }

  const asksAboutSales = /مبيعات|sales|ربح|performance|أداء|بيع|سوق|إيرادات|كم.*باع|كم.*بيعت|شو.*المبيعات|مبيعات.*اليوم|أرباح|دخل/.test(text);
  if (asksAboutSales) {
    if (typeof context.salesToday === 'number') {
      const salesLine = mentionSales(context);
      const topProductsLine = mentionTopProducts(context);
      if (isUrgent) {
        const urgentResponses = [
          `المبيعات اليوم تحتاج متابعة. ${salesLine} ${topProductsLine ? topProductsLine : ''} خذ بالك من المخزون والديون لأن الخسارة في أي من هاتين الجهتين تسرّع المشكلة.`,
          `عاجل؟ ${salesLine} ${topProductsLine ? topProductsLine : ''} فلا تتهاون وابقَ على المبيعات وراجع كل شيء بشكل دوري.`,
        ];
        return pick(urgentResponses, randomBetween(0, 1));
      }
      const normalResponses = [
        `${salesLine} ${topProductsLine ? topProductsLine : ''} يعني الحركة التجارية تسير بشكل جيد، خاصة إذا استمر هذا الأداء في بقية اليوم.`,
        `${salesLine} ${topProductsLine ? topProductsLine : ''} فالحمد لله الوضع ماشي وماشي.`,
        `${salesLine} ${topProductsLine ? topProductsLine : ''} ودا معناه إن الزبائن راضين وبيشتروا منك.`,
      ];
      return pick(normalResponses, randomBetween(0, 2));
    }
    const noDataResponses = [
      'ما في بيانات مبيعات كافية الآن، لكن أقدر أساعدك في مراجعة المخزون أو الديون إذا بدك. شو اللي تريد تركز عليه أول؟',
      'لأ مابي مبيعات مسجلة حالياً. ممكن تبدأ بتسجيل عملية بيع عشان نعرف شو وضع المتجر.',
      'البيانات مش موجودة دلوقتي، بس نقدر نراجع المخزون أو الديون إذا بدك.',
    ];
    return pick(noDataResponses, randomBetween(0, 2));
  }

  const asksAboutSummary = /ما.*أهم|أهم.*اليوم|مشكلة|ترتيب|أولويات|summary|ملخص|ماذا.*تحدث|كيف.*الحالة|الحالة|ما.*الحال|شو.*أهم.*الشي|شو.*محتاج|اقتراح|نصيحة|خطة/.test(text);
  if (asksAboutSummary) {
    const issues: string[] = [];
    if (context.lowStockCount > 0) issues.push(`${context.lowStockCount} عناصر تحتاج إعادة طلب`);
    if (context.overdueDebtsCount > 0) issues.push(`${context.overdueDebtsCount} ديون متأخرة تحتاج متابعة`);

    if (issues.length === 0) {
      const profitText =
        typeof context.todayProfit === 'number' && context.todayProfit > 0
          ? ` وربح اليوم ${formatCurrency(context.todayProfit)}.`
          : '';
      const salesText = context.salesToday > 0 ? ` ومبيعات اليوم ${formatCurrency(context.salesToday)}.` : '';
      const topProductsLine = mentionTopProducts(context);
      const goodResponses = [
        `الحالة جيدة اليوم، ما في تنبيهات حرجة في المخزون أو الديون، والمتجر مستقر.${salesText}${profitText} الأفضل الآن إنك تركز على المبيعات وتبقي الوضع مستمر بنفس الوتيرة.`,
        `ماشي الحال اليوم. ما في مشاكل.${salesText}${profitText} فبتركز على البيع وتكبّش زخم.`,
        `تمام، اليوم سهل. ما في مشاكل ولا تنبيهات.${salesText}${profitText} فخد بالك من الزبائن وكمل.`,
      ];
      if (topProductsLine) {
        return `${pick(goodResponses, randomBetween(0, 2))} ${topProductsLine}`;
      }
      return pick(goodResponses, randomBetween(0, 2));
    }

    const stockLine = mentionStock(context);
    const debtLine = mentionDebts(context);

    if (isUrgent) {
      const urgentResponses = [
        `مستعجل؟ تمام، أولويات اليوم واضحة: ${issues.join(' و ')}. خذ هالخطوات الآن: 1) اطلب المخزون الناقص، 2) تواصل مع المتأخرين، 3) راقب المبيعات خلال الساعات القادمة. هذا هو أقل شيء تضمن به أن المتجر يظل مستقر.`,
        `عندك ${issues.join(' و ')}. فلا تنتظر، خد بالك منهم اليوم عشان ما تكبر المشكلة.`,
        `مستعجل صحيح. ${stockLine} ${debtLine} ابدأ بهم من دلوقتي.`,
      ];
      return pick(urgentResponses, randomBetween(0, 2));
    }

    if (isFirm) {
      const firmResponses = [
        `أقترح تبدأ اليوم بثلاثة أشياء فقط: المخزون اللي يحتاج طلب، الديون المتأخرة، ثم المبيعات اللي تحتاج دعم. هذا هو ترتيب العمل الصحيح، ومش لازم تعقد الأمور.`,
        `خلينا نكون واضحين: المخزون أولاً، الديون ثانياً، والمبيعات ثالثاً. بهالترتيب بتضمن إن المتجر يظل مستقر.`,
        `أفضل ترتيب لليوم: طلب المخزون، تحصيل الديون، ثم دعم المبيعات. مش لازم تعقد الأمور أكتر من هيك.`,
      ];
      return pick(firmResponses, randomBetween(0, 2));
    }

    const normalResponses = [
      `أولويات اليوم واضحة: ${issues.join(' و ')}. إذا بدك، أرتب لك اللي لازم تشتغل عليه أول، بدءاً من اللي يضغط على السيولة ومن ثم المخزون، حتى تتعامل مع المتجر بشكل أسهل.`,
      `شفت شو في؟ عندك ${issues.join(' و ')}. فلا تستهين بهم، خدهم بجدية اليوم.`,
      `الوضع واضح: ${issues.join(' و ')}. يعني لازم تبدأ بهم من الصباح عشان تنهي يومك ومرتاح.`,
    ];
    return pick(normalResponses, randomBetween(0, 2));
  }

  const asksForHelp = /ساعد|مساعدة|أريد.*توصية|نصيحة|ماذا.*افعل|what.*do|هل.*أحتاج|أحتاج.*مساعدة|شو.*الخطوة|شو.*أعمل/.test(text);
  if (asksForHelp) {
    if (isUrgent) {
      const urgentResponses = [
        'مستعجل؟ تمام، خذ هالخطوات الآن: 1) اطلب المخزون الناقص، 2) تواصل مع المتأخرين، 3) راقب المبيعات خلال الساعات القادمة. هذا هو أقل شيء تضمن به أن المتجر يظل مستقر.',
        'عاجل؟ خد بالك من هالمواضيع: المخزون الناقص، الديون المتأخرة، والمبيعات. ولا تتهن من هالثلاثة.',
        'مستعجل صحيح. ابدأ بالمخزون والديون، وبعدين شوف المبيعات. بهالترتيب بتضمن إن المتجر يظل مستقر.',
      ];
      return pick(urgentResponses, randomBetween(0, 2));
    }
    if (isFirm) {
      const firmResponses = [
        'أقترح تبدأ اليوم بثلاثة أشياء فقط: المخزون اللي يحتاج طلب، الديون المتأخرة، ثم المبيعات اللي تحتاج دعم. هذا هو ترتيب العمل الصحيح، ومش لازم تعقد الأمور.',
        'خلينا نكون واضحين: المخزون أولاً، الديون ثانياً، والمبيعات ثالثاً. بهالترتيب بتضمن إن المتجر يظل مستقر.',
        'أفضل ترتيب لليوم: طلب المخزون، تحصيل الديون، ثم دعم المبيعات. مش لازم تعقد الأمور أكتر من هيك.',
      ];
      return pick(firmResponses, randomBetween(0, 2));
    }
    const normalResponses = [
      'أقترح تبدأ اليوم بثلاثة أشياء فقط: المخزون اللي يحتاج طلب، الديون المتأخرة، ثم المبيعات اللي تحتاج دعم. بهذه الطريقة بتحافظ على السيولة وتبقى الأمور تحت السيطرة دون ما تضيّع وقتك.',
      'خلينا نبدأ بالمخزون، بعدها الديون، وبعدين المبيعات. بهالترتيب بتضمن إن المتجر يظل مستقر وما في مفاجآت.',
      'أفضل خطوة لليوم: طلب المخزون الناقص، تحصيل الديون المتأخرة، ثم دعم المبيعات. مش لازم تعقد الأمور أكتر من هيك.',
    ];
    return pick(normalResponses, randomBetween(0, 2));
  }

  if (isUrgent) {
    const urgentResponses = [
      'أهلاً، خلّيها واضحة: أول شيء لازم تحلّه اليوم هو المخزون الناقص، وبعده الديون المتأخرة، ثم المبيعات. هالترتيب يضمن أن المتجر ما يتهت على نفسه.',
      'عاجل؟ خد بالك من هالمواضيع: المخزون الناقص، الديون المتأخرة، والمبيعات. ولا تتهن من هالثلاثة.',
      'مستعجل صحيح. ابدأ بالمخزون والديون، وبعدين شوف المبيعات. بهالترتيب بتضمن إن المتجر يظل مستقر.',
    ];
    return pick(urgentResponses, randomBetween(0, 2));
  }

  if (isFirm) {
    const firmResponses = [
      'أول ما يهمك اليوم هو ترتيب الأولويات: المخزون، الديون، ثم المبيعات. هذا هو الترتيب الصحيح، ومش لازم تعقد الأمور أكثر من هيك.',
      'خلينا نكون واضحين: المخزون أولاً، الديون ثانياً، والمبيعات ثالثاً. بهالترتيب بتضمن إن المتجر يظل مستقر.',
      'أفضل ترتيب لليوم: طلب المخزون، تحصيل الديون، ثم دعم المبيعات. مش لازم تعقد الأمور أكتر من هيك.',
    ];
    return pick(firmResponses, randomBetween(0, 2));
  }

  const fallbackResponses = [
    `أرى أن أول ما يهمك اليوم هو متابعة المخزون والديون، ثم الحفاظ على زخم المبيعات. إذا بدك، أجهّز لك خطة عملية لساعتك القادمة خطوة بخطوة، بشكل بسيط ومباشر.`,
    'خلينا نراجع شو عم يصير في المتجر. المخزون، الديون، والمبيعات - هالثلاثة هم الأساس.',
    'إذا بدك نراجع أي قسم من المتجر، قل لي. أنا موجود للمخزون، الديون، المبيعات، وأي حاجة تانية.',
    'شو رأيك نراجع المخزون أول؟ بعدها نروح للديون وبعدين المبيعات. بهالترتيب بنكون منظمين.',
    'كل يوم في فرصة نتحسن فيه. شو رأيك نراجع شو صار في المتجر اليوم ونخطط للغد؟',
    'إذا بدك نصيحة عملية: خد بالك من المخزون والديون، وبعدين كبّش زخم في المبيعات. بهالطريقة بتكون رايح في الطريق الصح.',
    'شو في جديد؟ إذا عندك سؤال عن المخزون أو الديون أو المبيعات، أنا معك.',
    'كل متجر زي الإنسان، يحتاج فحص دوري. شو رأيك نراجع الوضع من وقت للتاني؟',
  ];
  return pick(fallbackResponses, randomBetween(0, 7));
}
