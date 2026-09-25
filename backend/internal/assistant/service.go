package assistant

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/partflow/smart-store/internal/accounting"
	"github.com/partflow/smart-store/internal/dashboard"
)

type Intent string

const (
	IntentGreeting              Intent = "GREETING"
	IntentDailySummary          Intent = "DAILY_SUMMARY"
	IntentAdvice                Intent = "ADVICE"
	IntentAlerts                Intent = "ALERTS"
	IntentNavigation            Intent = "NAVIGATION"
	IntentMonthlySummary        Intent = "MONTHLY_SUMMARY"
	IntentSales                 Intent = "SALES"
	IntentInventory             Intent = "INVENTORY"
	IntentLowStock              Intent = "LOW_STOCK"
	IntentDebts                 Intent = "DEBTS"
	IntentProfit                Intent = "PROFIT"
	IntentPurchases             Intent = "PURCHASES"
	IntentSuppliers             Intent = "SUPPLIERS"
	IntentExpenses              Intent = "EXPENSES"
	IntentSettings              Intent = "SETTINGS"
	IntentRecent                Intent = "RECENT"
	IntentCustomers             Intent = "CUSTOMERS"
	IntentPayments              Intent = "PAYMENTS"
	IntentReturns               Intent = "RETURNS"
	IntentReports               Intent = "REPORTS"
	IntentSystemStatus          Intent = "SYSTEM_STATUS"
	IntentAssistantIdentity     Intent = "ASSISTANT_IDENTITY"
	IntentAssistantCapabilities Intent = "ASSISTANT_CAPABILITIES"
	IntentAssistantBehavior     Intent = "ASSISTANT_BEHAVIOR"
	IntentClarification         Intent = "CLARIFICATION"
	IntentThanks                Intent = "THANKS"
	IntentGoodbye               Intent = "GOODBYE"
	IntentSmallTalk             Intent = "SMALL_TALK"
	IntentUnknown               Intent = "UNKNOWN"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ReplyRequest struct {
	Message      string    `json:"message" binding:"required"`
	Conversation []Message `json:"conversation,omitempty"`
}

type ReplyResponse struct {
	Reply   string            `json:"reply"`
	Intent  Intent            `json:"intent"`
	State   string            `json:"state"`
	Actions []AssistantAction `json:"actions,omitempty"`
}

type AssistantAction struct {
	Label string `json:"label"`
	Path  string `json:"path"`
}

type Service struct {
	db *sqlx.DB
}

func NewService(db *sqlx.DB) *Service { return &Service{db: db} }

func (s *Service) Reply(ctx context.Context, request ReplyRequest) (ReplyResponse, error) {
	intent := detectIntent(request.Message, request.Conversation)
	if intent == IntentNavigation {
		action := navigationAction(request.Message)
		return ReplyResponse{Reply: "تمام، فتحت لك القسم المطلوب.", Intent: intent, State: "speaking", Actions: []AssistantAction{action}}, nil
	}
	if isConversationalIntent(intent) {
		return ReplyResponse{
			Reply:  buildConversationalReply(intent, request.Message, request.Conversation),
			Intent: intent,
			State:  "speaking",
		}, nil
	}
	if s.db == nil {
		return ReplyResponse{}, fmt.Errorf("assistant database is unavailable")
	}

	stats, err := dashboard.NewService(s.db).GetDashboardStats(ctx)
	if err != nil {
		return ReplyResponse{}, fmt.Errorf("load current store context: %w", err)
	}

	summary, err := s.loadSummary(ctx)
	if err != nil {
		return ReplyResponse{}, fmt.Errorf("load current assistant context: %w", err)
	}

	return ReplyResponse{
		Reply:   buildReply(intent, request.Message, request.Conversation, stats, summary),
		Intent:  intent,
		State:   "speaking",
		Actions: actionsForIntent(intent, stats),
	}, nil
}

type storeSummary struct {
	SalesCount        int
	OutstandingCount  int
	OutstandingTotal  float64
	SupplierTotal     float64
	PurchasesCount    int
	ExpensesToday     float64
	MonthlySales      float64
	MonthlyExpenses   float64
	MonthlyPurchases  float64
	TaxRate           string
	LowStockNames     []string
	LastSale          string
	LastPurchase      string
	ProductQuantities map[string]int
	InventoryValue    float64
	InventoryRetail   float64
	PotentialProfit   float64
	BestMarginProduct string
	BestMarginRate    float64
	ExpenseAverage7d  float64
}

func (s *Service) loadSummary(ctx context.Context) (storeSummary, error) {
	var result storeSummary
	now := accounting.StoreNow()
	todayStart, todayEnd, err := accounting.StoreDayBounds(now)
	if err != nil {
		return result, err
	}
	monthStart, monthEnd, err := accounting.StoreMonthBounds(now)
	if err != nil {
		return result, err
	}
	rollingStartDate, rollingEndDate, err := accounting.StoreDateRange(now, 6)
	if err != nil {
		return result, err
	}
	rollingStart, _, err := accounting.StoreDateBounds(rollingStartDate)
	if err != nil {
		return result, err
	}
	rollingEnd, _, err := accounting.StoreDateBounds(rollingEndDate)
	if err != nil {
		return result, err
	}
	if isSQLite(s.db) {
		if err := s.db.GetContext(ctx, &result.SalesCount, `SELECT COUNT(*) FROM sales WHERE datetime(COALESCE(sale_date, created_at)) >= datetime(?) AND datetime(COALESCE(sale_date, created_at)) < datetime(?) AND LOWER(COALESCE(status, 'completed')) = 'completed'`, todayStart, todayEnd); err != nil && err != sql.ErrNoRows {
			return result, err
		}
		if err := s.db.GetContext(ctx, &result.OutstandingCount, `SELECT COUNT(DISTINCT customer_id) FROM debts WHERE COALESCE(remaining_amount, 0) > 0`); err != nil && err != sql.ErrNoRows {
			return result, err
		}
		if err := s.db.GetContext(ctx, &result.OutstandingTotal, `SELECT COALESCE(SUM(remaining_amount), 0) FROM debts WHERE COALESCE(remaining_amount, 0) > 0`); err != nil && err != sql.ErrNoRows {
			return result, err
		}
		_ = s.db.GetContext(ctx, &result.SupplierTotal, `SELECT COALESCE(SUM(current_balance), 0) FROM suppliers WHERE COALESCE(is_active, 1) = 1`)
		_ = s.db.GetContext(ctx, &result.PurchasesCount, `SELECT COUNT(*) FROM purchases WHERE datetime(COALESCE(purchase_date, created_at)) >= datetime(?) AND datetime(COALESCE(purchase_date, created_at)) < datetime(?) AND LOWER(COALESCE(status, '')) NOT IN ('cancelled', 'reversed')`, todayStart, todayEnd)
		_ = s.db.GetContext(ctx, &result.ExpensesToday, `SELECT COALESCE(SUM(amount), 0) FROM expenses WHERE datetime(expense_date) >= datetime(?) AND datetime(expense_date) < datetime(?) AND LOWER(COALESCE(status, 'approved')) IN ('approved', 'paid', 'completed', 'archived')`, todayStart, todayEnd)
		_ = s.db.GetContext(ctx, &result.MonthlySales, `SELECT COALESCE(SUM(total_amount), 0) FROM sales WHERE datetime(COALESCE(sale_date, created_at)) >= datetime(?) AND datetime(COALESCE(sale_date, created_at)) < datetime(?) AND LOWER(COALESCE(status, 'completed')) NOT IN ('cancelled', 'canceled', 'reversed')`, monthStart, monthEnd)
		_ = s.db.GetContext(ctx, &result.MonthlyExpenses, `SELECT COALESCE(SUM(amount), 0) FROM expenses WHERE datetime(expense_date) >= datetime(?) AND datetime(expense_date) < datetime(?) AND LOWER(COALESCE(status, 'approved')) IN ('approved', 'paid', 'completed', 'archived')`, monthStart, monthEnd)
		_ = s.db.GetContext(ctx, &result.MonthlyPurchases, `SELECT COALESCE(SUM(total_amount), 0) FROM purchases WHERE datetime(COALESCE(purchase_date, created_at)) >= datetime(?) AND datetime(COALESCE(purchase_date, created_at)) < datetime(?) AND LOWER(COALESCE(status, '')) NOT IN ('cancelled', 'reversed')`, monthStart, monthEnd)
		_ = s.db.GetContext(ctx, &result.InventoryValue, `SELECT COALESCE(SUM(ii.purchase_cost), 0) FROM inventory_items ii JOIN products p ON p.id = ii.product_id WHERE UPPER(COALESCE(ii.status, '')) = 'AVAILABLE' AND p.deleted_at IS NULL`)
		_ = s.db.GetContext(ctx, &result.InventoryRetail, `SELECT COALESCE(SUM(ii.selling_price), 0) FROM inventory_items ii JOIN products p ON p.id = ii.product_id WHERE UPPER(COALESCE(ii.status, '')) = 'AVAILABLE' AND p.deleted_at IS NULL`)
		result.PotentialProfit = result.InventoryRetail - result.InventoryValue
		_ = s.db.GetContext(ctx, &result.ExpenseAverage7d, `SELECT COALESCE(SUM(amount), 0) / 7.0 FROM expenses WHERE datetime(expense_date) >= datetime(?) AND datetime(expense_date) < datetime(?) AND LOWER(COALESCE(status, 'approved')) IN ('approved', 'paid', 'completed', 'archived')`, rollingStart, rollingEnd)
		_ = s.db.GetContext(ctx, &result.BestMarginProduct, `SELECT name FROM products WHERE deleted_at IS NULL AND COALESCE(selling_price, 0) > 0 ORDER BY (COALESCE(selling_price, 0) - COALESCE(cost_price, 0)) DESC LIMIT 1`)
		if result.BestMarginProduct != "" {
			_ = s.db.GetContext(ctx, &result.BestMarginRate, `SELECT COALESCE((selling_price - cost_price) * 100.0 / NULLIF(selling_price, 0), 0) FROM products WHERE deleted_at IS NULL AND name = ? LIMIT 1`, result.BestMarginProduct)
		}
		_ = s.db.GetContext(ctx, &result.TaxRate, `SELECT COALESCE(value, '0') FROM settings WHERE key = 'tax_rate' LIMIT 1`)
		rows, err := s.db.QueryxContext(ctx, `SELECT p.name FROM products p JOIN inventory_items i ON i.product_id = p.id WHERE p.deleted_at IS NULL AND UPPER(COALESCE(i.status, '')) = 'AVAILABLE' GROUP BY p.id, p.name, p.min_stock_level HAVING COUNT(i.id) <= MAX(1, COALESCE(p.min_stock_level, 3)) ORDER BY p.name LIMIT 5`)
		if err != nil {
			return result, err
		}
		defer rows.Close()
		for rows.Next() {
			var name string
			if err := rows.Scan(&name); err != nil {
				return result, err
			}
			result.LowStockNames = append(result.LowStockNames, name)
		}
		_ = s.db.GetContext(ctx, &result.LastSale, `SELECT COALESCE(invoice_number, sale_number, id) FROM sales WHERE LOWER(COALESCE(status, 'completed')) = 'completed' ORDER BY COALESCE(sale_date, created_at) DESC LIMIT 1`)
		_ = s.db.GetContext(ctx, &result.LastPurchase, `SELECT COALESCE(invoice_number, purchase_number, id) FROM purchases WHERE LOWER(COALESCE(status, '')) NOT IN ('cancelled', 'reversed') ORDER BY COALESCE(purchase_date, created_at) DESC LIMIT 1`)
		result.ProductQuantities, err = loadProductQuantities(ctx, s.db)
		if err != nil {
			return result, err
		}
		return result, rows.Err()
	}

	if err := s.db.GetContext(ctx, &result.SalesCount, `SELECT COUNT(*) FROM sales WHERE COALESCE(sale_date, created_at) >= $1 AND COALESCE(sale_date, created_at) < $2 AND LOWER(COALESCE(status, 'completed')) = 'completed'`, todayStart, todayEnd); err != nil && err != sql.ErrNoRows {
		return result, err
	}
	if err := s.db.GetContext(ctx, &result.OutstandingCount, `SELECT COUNT(DISTINCT customer_id) FROM debts WHERE COALESCE(remaining_amount, 0) > 0`); err != nil && err != sql.ErrNoRows {
		return result, err
	}
	if err := s.db.GetContext(ctx, &result.OutstandingTotal, `SELECT COALESCE(SUM(remaining_amount), 0) FROM debts WHERE COALESCE(remaining_amount, 0) > 0`); err != nil && err != sql.ErrNoRows {
		return result, err
	}
	_ = s.db.GetContext(ctx, &result.SupplierTotal, `SELECT COALESCE(SUM(current_balance), 0) FROM suppliers WHERE COALESCE(is_active, true) = true`)
	_ = s.db.GetContext(ctx, &result.PurchasesCount, `SELECT COUNT(*) FROM purchases WHERE COALESCE(purchase_date, created_at) >= $1 AND COALESCE(purchase_date, created_at) < $2 AND LOWER(COALESCE(status, '')) NOT IN ('cancelled', 'reversed')`, todayStart, todayEnd)
	_ = s.db.GetContext(ctx, &result.ExpensesToday, `SELECT COALESCE(SUM(amount), 0) FROM expenses WHERE expense_date >= $1 AND expense_date < $2 AND LOWER(COALESCE(status, 'approved')) IN ('approved', 'paid', 'completed', 'archived')`, todayStart, todayEnd)
	_ = s.db.GetContext(ctx, &result.MonthlySales, `SELECT COALESCE(SUM(total_amount), 0) FROM sales WHERE COALESCE(sale_date, created_at) >= $1 AND COALESCE(sale_date, created_at) < $2 AND LOWER(COALESCE(status, 'completed')) NOT IN ('cancelled', 'canceled', 'reversed')`, monthStart, monthEnd)
	_ = s.db.GetContext(ctx, &result.MonthlyExpenses, `SELECT COALESCE(SUM(amount), 0) FROM expenses WHERE expense_date >= $1 AND expense_date < $2 AND LOWER(COALESCE(status, 'approved')) IN ('approved', 'paid', 'completed', 'archived')`, monthStart, monthEnd)
	_ = s.db.GetContext(ctx, &result.MonthlyPurchases, `SELECT COALESCE(SUM(total_amount), 0) FROM purchases WHERE COALESCE(purchase_date, created_at) >= $1 AND COALESCE(purchase_date, created_at) < $2 AND LOWER(COALESCE(status, '')) NOT IN ('cancelled', 'reversed')`, monthStart, monthEnd)
	_ = s.db.GetContext(ctx, &result.InventoryValue, `SELECT COALESCE(SUM(ii.purchase_cost), 0) FROM inventory_items ii JOIN products p ON p.id = ii.product_id WHERE UPPER(COALESCE(ii.status, '')) = 'AVAILABLE' AND p.deleted_at IS NULL`)
	_ = s.db.GetContext(ctx, &result.InventoryRetail, `SELECT COALESCE(SUM(ii.selling_price), 0) FROM inventory_items ii JOIN products p ON p.id = ii.product_id WHERE UPPER(COALESCE(ii.status, '')) = 'AVAILABLE' AND p.deleted_at IS NULL`)
	result.PotentialProfit = result.InventoryRetail - result.InventoryValue
	_ = s.db.GetContext(ctx, &result.ExpenseAverage7d, `SELECT COALESCE(SUM(amount), 0) / 7.0 FROM expenses WHERE expense_date >= $1 AND expense_date < $2 AND LOWER(COALESCE(status, 'approved')) IN ('approved', 'paid', 'completed', 'archived')`, rollingStart, rollingEnd)
	_ = s.db.GetContext(ctx, &result.BestMarginProduct, `SELECT name FROM products WHERE deleted_at IS NULL AND COALESCE(selling_price, 0) > 0 ORDER BY (COALESCE(selling_price, 0) - COALESCE(cost_price, 0)) DESC LIMIT 1`)
	if result.BestMarginProduct != "" {
		_ = s.db.GetContext(ctx, &result.BestMarginRate, `SELECT COALESCE((selling_price - cost_price) * 100.0 / NULLIF(selling_price, 0), 0) FROM products WHERE deleted_at IS NULL AND name = $1 LIMIT 1`, result.BestMarginProduct)
	}
	_ = s.db.GetContext(ctx, &result.TaxRate, `SELECT COALESCE(value, '0') FROM settings WHERE key = 'tax_rate' LIMIT 1`)
	rows, err := s.db.QueryxContext(ctx, `SELECT p.name FROM products p JOIN inventory_items i ON i.product_id = p.id WHERE p.deleted_at IS NULL AND UPPER(COALESCE(i.status, '')) = 'AVAILABLE' GROUP BY p.id, p.name, p.min_stock_level HAVING COUNT(i.id) <= GREATEST(1, COALESCE(p.min_stock_level, 3)) ORDER BY p.name LIMIT 5`)
	if err != nil {
		return result, err
	}
	defer rows.Close()
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return result, err
		}
		result.LowStockNames = append(result.LowStockNames, name)
	}
	_ = s.db.GetContext(ctx, &result.LastSale, `SELECT COALESCE(invoice_number, sale_number, id) FROM sales WHERE LOWER(COALESCE(status, 'completed')) = 'completed' ORDER BY COALESCE(sale_date, created_at) DESC LIMIT 1`)
	_ = s.db.GetContext(ctx, &result.LastPurchase, `SELECT COALESCE(invoice_number, purchase_number, id) FROM purchases WHERE LOWER(COALESCE(status, '')) NOT IN ('cancelled', 'reversed') ORDER BY COALESCE(purchase_date, created_at) DESC LIMIT 1`)
	result.ProductQuantities, err = loadProductQuantities(ctx, s.db)
	if err != nil {
		return result, err
	}
	return result, rows.Err()
}

func loadProductQuantities(ctx context.Context, db *sqlx.DB) (map[string]int, error) {
	rows, err := db.QueryxContext(ctx, `
		SELECT p.name, COUNT(i.id)
		FROM products p
		LEFT JOIN inventory_items i
			ON i.product_id = p.id
			AND UPPER(COALESCE(i.status, '')) = 'AVAILABLE'
		WHERE p.deleted_at IS NULL
		GROUP BY p.id, p.name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	quantities := make(map[string]int)
	for rows.Next() {
		var name string
		var quantity int
		if err := rows.Scan(&name, &quantity); err != nil {
			return nil, err
		}
		quantities[normalize(name)] = quantity
	}
	return quantities, rows.Err()
}

func detectIntent(message string, conversation []Message) Intent {
	text := normalize(message)
	text = strings.Trim(text, "؟?!.,،؛:ـ")
	if text == "" {
		return IntentClarification
	}
	switch {
	case regexp.MustCompile(`((افتح|فتح|اذهب|روح|خذني|انتقل|اعرض|اعمل|ابدأ|ابدا|سجل|نفذ).*(عملية\s*بيع|بيع\s*جديد|نقطة\s*البيع|المبيعات|POS|البيع|المخزون|الديون|التقارير|المشتريات|المصروفات|العملاء|الموردين|الأرشيف)|بيع\s*جديد|عملية\s*بيع)`).MatchString(text):
		return IntentNavigation
	case regexp.MustCompile(`ليش.*(تكرر|بتكرر|تعيد)|تكرر.*(كلام|رد)|نفس.*(الكلام|الرد)|كرر`).MatchString(text):
		return IntentAssistantBehavior
	case regexp.MustCompile(`مين\s*(انت|إنت|أنت)|من\s*(انت|إنت|أنت)|(?:انت|إنت|أنت)\s*من|who\s*are\s*you`).MatchString(text):
		return IntentAssistantIdentity
	case regexp.MustCompile(`شو\s*(بتقدر|تقدر|فيك)\s*(تعمل|تساعد)|قدراتك|ماذا\s*تستطيع|شو\s*بتساعد`).MatchString(text):
		return IntentAssistantCapabilities
	case regexp.MustCompile(`شكرا|شكرًا|يعطيك\s*العافية|thanks|thank\s*you`).MatchString(text):
		return IntentThanks
	case regexp.MustCompile(`باي|مع\s*السلامة|إلى\s*اللقاء|goodbye|bye`).MatchString(text):
		return IntentGoodbye
	case regexp.MustCompile(`تمام|ممتاز|ماشي|حسنا|حسنًا|اوكي|أوكي|كيفك|شو\s*الأخبار|شو\s*الاخبار|كيف\s*حالك`).MatchString(text):
		return IntentSmallTalk
	case regexp.MustCompile(`تنبيه|تنبيهات|يحتاج\s*انتباه|شو\s*في\s*مشاكل|في\s*مشكلة|تحذير`).MatchString(text):
		return IntentAlerts
	case regexp.MustCompile(`نصيحة|نصائح|اقتراح|اقتراحات|شو\s*(أعمل|اعمل|لازم|بتنصح)|ماذا\s*أفعل|كيف\s*أحسن|أولوياتي`).MatchString(text):
		return IntentAdvice
	case regexp.MustCompile(`مرحبا|اهلا|صباح|مساء|السلام|هلا`).MatchString(text):
		return IntentGreeting
	}

	contextText := text
	if regexp.MustCompile(`^(و|وال|طيب|ثم|كمان|ايش|شو|مين|كم)`).MatchString(text) {
		if previous := lastUserMessage(conversation); previous != "" {
			contextText = normalize(previous) + " " + text
		}
	}
	switch {
	case regexp.MustCompile(`ربح|أرباح|profit`).MatchString(contextText):
		return IntentProfit
	case regexp.MustCompile(`شراء|مشتريات|purchase`).MatchString(contextText):
		return IntentPurchases
	case regexp.MustCompile(`مورد|موردين|supplier`).MatchString(contextText):
		return IntentSuppliers
	case regexp.MustCompile(`مصروف|صرفنا|expenses`).MatchString(contextText):
		return IntentExpenses
	case regexp.MustCompile(`ضريبة|خصم|إعدادات|settings`).MatchString(contextText):
		return IntentSettings
	case regexp.MustCompile(`عميل|عملاء|customers`).MatchString(contextText):
		return IntentCustomers
	case regexp.MustCompile(`دفعة|دفعات|تحصيل اليوم|payments`).MatchString(contextText):
		return IntentPayments
	case regexp.MustCompile(`مرتجع|مرتجعات|returns`).MatchString(contextText):
		return IntentReturns
	case regexp.MustCompile(`تقرير|تقارير|reports`).MatchString(contextText):
		return IntentReports
	case regexp.MustCompile(`حالة النظام|النظام شغال|system status`).MatchString(contextText):
		return IntentSystemStatus
	case regexp.MustCompile(`آخر.*بيع|اخر.*بيع|آخر.*شراء|اخر.*شراء|recent`).MatchString(contextText):
		return IntentRecent
	case regexp.MustCompile(`دين|ديون|مستحق|تحصيل|مديون`).MatchString(contextText):
		return IntentDebts
	case regexp.MustCompile(`مخزون|منتجات|منتج|بضاعة`).MatchString(contextText):
		if regexp.MustCompile(`منخفض|ناقصة|تخلص|قربت`).MatchString(contextText) {
			return IntentLowStock
		}
		return IntentInventory
	case regexp.MustCompile(`مبيعات|بعنا|بيع|دخل|sales`).MatchString(contextText):
		return IntentSales
	case regexp.MustCompile(`هذا\s*الشهر|هالشهر|الشهر|هذا\s*الشه?ر`).MatchString(contextText):
		return IntentMonthlySummary
	case regexp.MustCompile(`ملخص|وضع المتجر|كيف.*اليوم|شو.*اليوم|أولويات`).MatchString(contextText):
		return IntentDailySummary
	default:
		return IntentUnknown
	}
}

func lastUserMessage(conversation []Message) string {
	for i := len(conversation) - 1; i >= 0; i-- {
		if strings.EqualFold(conversation[i].Role, "user") {
			return conversation[i].Content
		}
	}
	return ""
}

func isConversationalIntent(intent Intent) bool {
	switch intent {
	case IntentGreeting, IntentAssistantIdentity, IntentAssistantCapabilities, IntentAssistantBehavior,
		IntentClarification, IntentThanks, IntentGoodbye, IntentSmallTalk, IntentUnknown:
		return true
	default:
		return false
	}
}

func navigationAction(message string) AssistantAction {
	text := normalize(message)
	switch {
	case regexp.MustCompile(`عملية\s*بيع|بيع\s*جديد|نقطة\s*البيع|POS|المبيعات|البيع`).MatchString(text):
		return AssistantAction{Label: "فتح نقطة البيع", Path: "/app/sales"}
	case strings.Contains(text, "مخزون"):
		return AssistantAction{Label: "فتح المخزون", Path: "/app/inventory"}
	case strings.Contains(text, "ديون"):
		return AssistantAction{Label: "فتح الديون", Path: "/app/debts"}
	case strings.Contains(text, "تقارير"):
		return AssistantAction{Label: "فتح التقارير", Path: "/app/reports"}
	case strings.Contains(text, "مشتريات"):
		return AssistantAction{Label: "فتح المشتريات", Path: "/app/purchases"}
	case strings.Contains(text, "مصروفات"):
		return AssistantAction{Label: "فتح المصروفات", Path: "/app/expenses"}
	case strings.Contains(text, "عملاء"):
		return AssistantAction{Label: "فتح العملاء", Path: "/app/customers"}
	case strings.Contains(text, "موردين") || strings.Contains(text, "موردون"):
		return AssistantAction{Label: "فتح الموردين", Path: "/app/suppliers"}
	default:
		return AssistantAction{Label: "فتح لوحة التحكم", Path: "/app/dashboard"}
	}
}

func buildConversationalReply(intent Intent, message string, conversation []Message) string {
	switch intent {
	case IntentGreeting:
		return greetingReply(message)
	case IntentAssistantIdentity:
		if strings.Count(strings.ToLower(strings.Join(conversationContents(conversation), " ")), "مين انت") > 0 {
			return "لسه معك 😄 أنا مساعد PartFlow الذكي، وبساعدك تتابع شغل المحل وبياناته."
		}
		return "أنا مساعد PartFlow الذكي 👋 بساعدك تتابع المبيعات، المخزون، الديون، المشتريات والأرباح من البيانات الحالية."
	case IntentAssistantCapabilities:
		return "بقدر أراجع معك المبيعات، الأرباح، المخزون، الديون، المشتريات والموردين، وأفهم أسئلتك المتتابعة."
	case IntentAssistantBehavior:
		return "معك حق، كررت نفس الرد. كان المفروض أفهم سؤالك بدل ما أرجع لنفس الإجابة. خلينا نكمل."
	case IntentClarification:
		return "معك 👋 شو حابب تعرف؟"
	case IntentThanks:
		return "العفو 👋 أنا معك."
	case IntentGoodbye:
		return "مع السلامة، أنا موجود لما تحتاجني."
	case IntentSmallTalk:
		return "تمام، أنا معك. شو حابب نراجع؟"
	default:
		return "ما قدرت أفهم قصدك تمامًا. ممكن تسألني عن المبيعات، المخزون، الديون أو المشتريات."
	}
}

func conversationContents(conversation []Message) []string {
	contents := make([]string, 0, len(conversation))
	for _, item := range conversation {
		contents = append(contents, item.Content)
	}
	return contents
}

func buildReply(intent Intent, message string, conversation []Message, stats *dashboard.DashboardStats, summary storeSummary) string {
	money := func(value float64) string { return fmt.Sprintf("₪%s", formatNumber(value)) }
	switch intent {
	case IntentGreeting:
		return greetingReply(message)
	case IntentDailySummary:
		return fmt.Sprintf("ملخص اليوم: سجلت %d مبيعات بإجمالي %s، وصافي الربح %s. الديون المستحقة %s، والمخزون المنخفض %d منتجات.", summary.SalesCount, money(stats.TodaySales), money(stats.TodayProfit), money(stats.OutstandingDebts), stats.LowStockCount)
	case IntentAdvice:
		return buildAdviceReply(message, stats, summary, money)
	case IntentAlerts:
		return buildAlertsReply(stats, summary, money)
	case IntentMonthlySummary:
		return fmt.Sprintf("هذا الشهر: المبيعات %s، المشتريات %s، والمصروفات %s.", money(summary.MonthlySales), money(summary.MonthlyPurchases), money(summary.MonthlyExpenses))
	case IntentSales:
		if stats.TodaySales == 0 {
			return "ما عندي مبيعات مسجلة اليوم حتى الآن."
		}
		return fmt.Sprintf("مبيعات اليوم %s من %d عمليات بيع.", money(stats.TodaySales), summary.SalesCount)
	case IntentProfit:
		return fmt.Sprintf("صافي ربح اليوم %s.", money(stats.TodayProfit))
	case IntentPurchases:
		return fmt.Sprintf("اليوم سجلت %d عمليات شراء.", summary.PurchasesCount)
	case IntentSuppliers:
		return fmt.Sprintf("إجمالي الرصيد المستحق للموردين حاليًا %s.", money(summary.SupplierTotal))
	case IntentExpenses:
		return fmt.Sprintf("المصروفات المسجلة اليوم %s.", money(summary.ExpensesToday))
	case IntentSettings:
		return fmt.Sprintf("نسبة الضريبة الحالية %s%%.", valueOrNone(summary.TaxRate))
	case IntentRecent:
		return fmt.Sprintf("آخر عملية بيع: %s. آخر عملية شراء: %s.", valueOrNone(summary.LastSale), valueOrNone(summary.LastPurchase))
	case IntentCustomers:
		return fmt.Sprintf("عدد العملاء المسجلين في النظام %d عميل.", stats.TotalCustomers)
	case IntentPayments:
		return fmt.Sprintf("المبالغ المحصلة اليوم %s.", money(stats.TodayCollected))
	case IntentReturns:
		return fmt.Sprintf("إجمالي قيمة المرتجعات المسجلة %s.", money(stats.TotalRefunded))
	case IntentReports:
		return fmt.Sprintf("ملخص التقارير الحالي: مبيعات %s، ربح %s، ومصروفات اليوم %s.", money(stats.NetSales), money(stats.TodayProfit), money(summary.ExpensesToday))
	case IntentSystemStatus:
		return "النظام متصل وقادر على قراءة بيانات PartFlow الحالية."
	case IntentDebts:
		if summary.OutstandingCount == 0 {
			return "ما في ديون غير مسددة حاليًا."
		}
		return fmt.Sprintf("عندك %d عملاء عليهم ديون بإجمالي %s.", summary.OutstandingCount, money(summary.OutstandingTotal))
	case IntentLowStock:
		if len(summary.LowStockNames) == 0 {
			return "ما في منتجات منخفضة المخزون حاليًا."
		}
		return fmt.Sprintf("المنتجات منخفضة المخزون: %s.", strings.Join(summary.LowStockNames, "، "))
	case IntentInventory:
		if name, quantity, ok := findProductQuantity(message, conversation, summary.ProductQuantities); ok {
			return fmt.Sprintf("الكمية الحالية من %s هي %d.", name, quantity)
		}
		return fmt.Sprintf("حاليًا يوجد %d منتجات منخفضة المخزون. أقدر أطلع لك التفاصيل أو أراجع معك المبيعات.", stats.LowStockCount)
	default:
		return fmt.Sprintf("أقدر أساعدك ببيانات PartFlow الحالية. مبيعات اليوم %s، والربح %s. جرّب تسألني عن المبيعات أو الديون أو المخزون.", money(stats.TodaySales), money(stats.TodayProfit))
	}
}

func buildAdviceReply(message string, stats *dashboard.DashboardStats, summary storeSummary, money func(float64) string) string {
	text := normalize(message)
	if regexp.MustCompile(`هامش|ربحية|ربح.*منتج|منتج.*ربح`).MatchString(text) {
		if summary.BestMarginProduct != "" {
			return fmt.Sprintf("تحليل الربحية: المنتج صاحب أعلى هامش مسجل حاليًا هو %s بنسبة تقريبية %.1f%%. لا تعتمد الخصم عليه قبل مقارنة سرعة بيعه، وتكلفة المخزون المتجمد.", summary.BestMarginProduct, summary.BestMarginRate)
		}
		return "لا أملك تفاصيل تكلفة كافية لحساب هامش كل منتج بدقة. سجّل تكلفة الشراء وسعر البيع، ثم أقدر أقارن المنتجات ربحياً بدون تخمين."
	}
	if regexp.MustCompile(`مصروف.*(غريب|مرتفع|زائد)|ارتفاع.*مصروف|مصاريف.*اليوم`).MatchString(text) {
		if summary.ExpenseAverage7d > 0 && stats.TodayExpenses > summary.ExpenseAverage7d*1.5 {
			return fmt.Sprintf("تنبيه مصروفات: مصروفات اليوم %s أعلى من متوسط آخر 7 أيام البالغ %s. راجع الإيصالات والتصنيف قبل اعتماد أي مصروف إضافي.", money(stats.TodayExpenses), money(summary.ExpenseAverage7d))
		}
		return fmt.Sprintf("مصروفات اليوم %s، ومتوسط آخر 7 أيام %s. راقب الاتجاه أسبوعيًا، ولا تعتبر الارتفاع مشكلة قبل مقارنة نفس أيام الأسبوع.", money(stats.TodayExpenses), money(summary.ExpenseAverage7d))
	}
	if regexp.MustCompile(`سيولة.*متوق|توقع.*سيولة|تدفق.*نقد|نقد.*متوقع`).MatchString(text) {
		return fmt.Sprintf("توقع السيولة المبدئي: ابدأ من رصيد التحصيل الحالي %s، اطرح المصروفات اليومية %s، وأجّل المشتريات غير الضرورية حتى تتأكد من الديون المتوقع تحصيلها %s.", money(stats.TodayCollected), money(stats.TodayExpenses), money(stats.OutstandingDebts))
	}
	if regexp.MustCompile(`إعادة.*طلب|اعادة.*طلب|كم.*أطلب|كم.*اطلب|طلب.*مخزون`).MatchString(text) {
		if stats.LowStockCount > 0 {
			return fmt.Sprintf("إعادة الطلب: عندك %d منتجات منخفضة. ابدأ بالأسرع حركة، واطلب كمية تغطي دورة البيع القادمة فقط بدل تجميد السيولة في مخزون زائد.", stats.LowStockCount)
		}
		return "لا توجد منتجات منخفضة حاليًا. لا أنصح بإعادة طلب عشوائية؛ استخدم سرعة البيع وحد إعادة الطلب لكل منتج قبل الشراء."
	}
	if regexp.MustCompile(`عملاء.*(تحليل|أفضل|متوقف)|أفضل.*عملاء|عملاء.*ربح|تقسيم.*عملاء`).MatchString(text) {
		return fmt.Sprintf("تحليل العملاء: لديك %d عميلًا مسجلًا. قسّمهم إلى عملاء نشطين، متأخرين، وعملاء لم يشتروا مؤخرًا، ثم خصص المتابعة والعروض لكل مجموعة بدل إرسال عرض واحد للجميع.", stats.TotalCustomers)
	}
	if regexp.MustCompile(`محاسب|محاسبة|حسابات|قيد|فاتورة|ضريبة|تكلفة|دفتر`).MatchString(text) {
		return fmt.Sprintf("محاسبيًا: افصل مبيعات الشهر %s عن المشتريات %s والمصروفات %s، وسجّل كل مصروف وفاتورة في يومها. قبل إغلاق الشهر طابق المبيعات والتحصيل والمبالغ المستحقة.", money(summary.MonthlySales), money(summary.MonthlyPurchases), money(summary.MonthlyExpenses))
	}
	if regexp.MustCompile(`مالي|سيولة|نقد|كاش|تحصيل|تمويل|ديون|دين\b`).MatchString(text) {
		if stats.OutstandingDebts > 0 {
			return fmt.Sprintf("ماليًا: الأولوية لتحسين السيولة عبر تحصيل الديون المستحقة %s، ثم تجميد المشتريات غير الضرورية ومراجعة المصروفات قبل أي توسع.", money(stats.OutstandingDebts))
		}
		return fmt.Sprintf("ماليًا: السيولة تبدو مستقرة من ناحية الديون. راقب المصروفات %s يوميًا، ولا تلتزم بشراء جديد إلا بعد مقارنة التكلفة بسرعة دوران المنتج.", money(summary.ExpensesToday))
	}
	if regexp.MustCompile(`إدار|ادار|تشغيل|موظف|تنظيم|مورد|مشتريات|إجراء|اجراء`).MatchString(text) {
		if stats.LowStockCount > 0 {
			return fmt.Sprintf("إداريًا: رتّب العمل حول %d منتجات منخفضة المخزون، حدّد مسؤولًا عن الطلب والاستلام، واجعل مراجعة المخزون والتحصيل نقطة ثابتة في نهاية كل يوم.", stats.LowStockCount)
		}
		return fmt.Sprintf("إداريًا: ثبّت روتينًا يوميًا لمراجعة المبيعات والمصروفات والتحصيل، وروتينًا أسبوعيًا لمراجعة المشتريات والموردين. لديك %d عملية شراء اليوم، فراجعها قبل إعادة الطلب.", summary.PurchasesCount)
	}
	if regexp.MustCompile(`مستقبل|خطة|تخطيط|أخطط|نمو|توسع|هدف|الشهر الجاي|الشهر القادم|بكرا|غد`).MatchString(text) {
		return fmt.Sprintf("للتخطيط: استخدم مبيعات الشهر %s كمؤشر أساس، وحدد هدفًا قابلًا للقياس للشهر القادم. ابدأ بتحسين المنتجات الأسرع حركة، ثم راقب الربح %s قبل زيادة المصروفات أو التوسع.", money(summary.MonthlySales), money(stats.TodayProfit))
	}
	if regexp.MustCompile(`عرض|عروض|خصم|تسويق|حملة|زبائن|عملاء`).MatchString(text) {
		return "للعروض: لا تبدأ بخصم عام. اختر منتجًا سريع الحركة أو اربط منتجًا بطيء الحركة بمنتج مطلوب، وحدد مدة وكمية واضحة. احسب تكلفة القطعة وهامش الربح أولًا حتى لا يتحول العرض إلى خسارة."
	}
	if stats.LowStockCount > 0 && stats.OutstandingDebts > 0 {
		return fmt.Sprintf("ابدأ اليوم بخطوتين: اطلب %d منتجات منخفضة المخزون، ثم تابع تحصيل الديون المستحقة %s. بعد ذلك راقب المبيعات قبل نهاية اليوم.", stats.LowStockCount, money(stats.OutstandingDebts))
	}
	if stats.LowStockCount > 0 {
		return fmt.Sprintf("اقتراحي الأول اليوم: راجع %d منتجات منخفضة المخزون وابدأ طلب القطع الأسرع حركة، حتى لا تخسر عملية بيع بسبب نفادها.", stats.LowStockCount)
	}
	if stats.OutstandingDebts > 0 {
		return fmt.Sprintf("الوضع يحتاج متابعة مالية: تواصل مع العملاء أصحاب الديون المستحقة %s، ثم راجع المبيعات والتحصيل في نهاية اليوم.", money(stats.OutstandingDebts))
	}
	if stats.TodaySales == 0 {
		return "المخزون والديون مستقران حاليًا. اقتراحي العملي: راجع عروض المنتجات وابدأ بمتابعة أول عملية بيع اليوم."
	}
	if stats.TodayProfit > 0 {
		return fmt.Sprintf("الأداء جيد حتى الآن. حافظ على المنتجات التي تحرك المبيعات، وراقب الربح الحالي %s قبل اعتماد أي خصم جديد.", money(stats.TodayProfit))
	}
	return "لا توجد أولوية حرجة الآن. اقتراحي أن تراجع آخر المبيعات والمشتريات وتحدّث حد إعادة الطلب للمنتجات المهمة."
}

func actionsForIntent(intent Intent, stats *dashboard.DashboardStats) []AssistantAction {
	actions := make([]AssistantAction, 0, 3)
	if stats.LowStockCount > 0 && (intent == IntentAdvice || intent == IntentAlerts || intent == IntentLowStock) {
		actions = append(actions, AssistantAction{Label: "فتح المخزون المنخفض", Path: "/app/inventory?low_stock_only=true"})
	}
	if stats.OutstandingDebts > 0 && (intent == IntentAdvice || intent == IntentAlerts || intent == IntentDebts) {
		actions = append(actions, AssistantAction{Label: "متابعة الديون", Path: "/app/debts"})
	}
	switch intent {
	case IntentAdvice, IntentAlerts, IntentDailySummary, IntentMonthlySummary, IntentProfit:
		actions = append(actions, AssistantAction{Label: "فتح التقارير", Path: "/app/reports"})
	case IntentExpenses:
		actions = append(actions, AssistantAction{Label: "مراجعة المصروفات", Path: "/app/expenses"})
	case IntentSales:
		actions = append(actions, AssistantAction{Label: "فتح المبيعات", Path: "/app/sales"})
	}
	return actions
}

func buildAlertsReply(stats *dashboard.DashboardStats, summary storeSummary, money func(float64) string) string {
	alerts := make([]string, 0, 3)
	if stats.LowStockCount > 0 {
		alerts = append(alerts, fmt.Sprintf("%d منتجات عند الحد الأدنى أو أقل", stats.LowStockCount))
	}
	if stats.OutstandingDebts > 0 {
		alerts = append(alerts, fmt.Sprintf("ديون مستحقة بقيمة %s", money(stats.OutstandingDebts)))
	}
	if stats.TodaySales == 0 {
		alerts = append(alerts, "لا توجد مبيعات مسجلة اليوم حتى الآن")
	}
	if len(alerts) == 0 {
		return "ما عندي تنبيهات حرجة حاليًا. المخزون والديون مستقران حسب البيانات الحالية."
	}
	return fmt.Sprintf("التنبيهات الحالية: %s. ابدأ بالأولوية الأولى ثم راقب التحديثات بعد أي بيع أو تحصيل.", strings.Join(alerts, "، "))
}

func greetingReply(message string) string {
	text := normalize(message)
	switch {
	case strings.Contains(text, "صباح"):
		return "صباح الخير 🌤️ جاهز نبدأ؟"
	case strings.Contains(text, "مساء"):
		return "مساء الخير 👋 شو حابب نراجع اليوم؟"
	case strings.Contains(text, "السلام"):
		return "وعليكم السلام! أنا جاهز أساعدك."
	default:
		return "أهلا فيك 👋 شو حابب نراجع؟"
	}
}

func findProductQuantity(message string, conversation []Message, quantities map[string]int) (string, int, bool) {
	contextText := normalize(message)
	for _, item := range conversation {
		contextText += " " + normalize(item.Content)
	}
	for name, quantity := range quantities {
		if name != "" && strings.Contains(contextText, name) {
			return name, quantity, true
		}
	}
	if len(quantities) == 1 {
		for name, quantity := range quantities {
			return name, quantity, true
		}
	}
	return "", 0, false
}

func formatNumber(value float64) string {
	return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.2f", value), "0"), ".")
}
func valueOrNone(value string) string {
	if strings.TrimSpace(value) == "" {
		return "لا توجد بيانات"
	}
	return value
}
func normalize(value string) string { return strings.ToLower(strings.TrimSpace(value)) }
func isSQLite(db *sqlx.DB) bool {
	name := strings.ToLower(db.DriverName())
	return name == "sqlite" || name == "sqlite3"
}
