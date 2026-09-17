package assistant

import (
	"fmt"
	"strings"
	"testing"

	"github.com/partflow/smart-store/internal/dashboard"
)

func TestDetectIntentUnderstandsNaturalArabic(t *testing.T) {
	cases := []struct {
		message string
		want    Intent
	}{
		{"شو بعنا اليوم؟", IntentSales},
		{"كيف وضع المتجر اليوم؟", IntentDailySummary},
		{"شو المنتجات اللي قربت تخلص؟", IntentLowStock},
		{"مين عليه ديون؟", IntentDebts},
		{"كم الربح؟", IntentProfit},
		{"ما الذي حدث في هذا الشهر؟", IntentMonthlySummary},
		{"مين انت؟", IntentAssistantIdentity},
		{"من انت", IntentAssistantIdentity},
		{"شو بتقدر تعمل؟", IntentAssistantCapabilities},
		{"؟", IntentClarification},
		{"ليش بتكرر نفس الكلام؟", IntentAssistantBehavior},
		{"شكراً", IntentThanks},
		{"شو بتنصحني أعمل؟", IntentAdvice},
		{"شو التنبيهات؟", IntentAlerts},
		{"افتح نقطة البيع", IntentNavigation},
	}
	for _, tc := range cases {
		if got := detectIntent(tc.message, nil); got != tc.want {
			t.Fatalf("detectIntent(%q) = %s, want %s", tc.message, got, tc.want)
		}
	}
}

func TestBuildReplyUsesCurrentMonthValues(t *testing.T) {
	reply := buildReply(IntentMonthlySummary, "ما الذي حدث في هذا الشهر؟", nil, &dashboard.DashboardStats{}, storeSummary{
		MonthlySales: 8500, MonthlyPurchases: 2400, MonthlyExpenses: 700,
	})
	for _, expected := range []string{"8500", "2400", "700"} {
		if !strings.Contains(reply, expected) {
			t.Fatalf("monthly reply %q does not contain %q", reply, expected)
		}
	}
}

func TestBuildReplyUsesCurrentValuesAndDoesNotInventZeroData(t *testing.T) {
	stats := &dashboard.DashboardStats{TodaySales: 1450, TodayProfit: 500, OutstandingDebts: 50, LowStockCount: 3}
	reply := buildReply(IntentDailySummary, "كيف وضع المتجر؟", nil, stats, storeSummary{SalesCount: 8})
	for _, expected := range []string{"1450", "500", "50", "3", "8"} {
		if !strings.Contains(reply, expected) {
			t.Fatalf("summary reply %q does not contain current value %q", reply, expected)
		}
	}

	zeroReply := buildReply(IntentSales, "كم بعنا؟", nil, &dashboard.DashboardStats{}, storeSummary{})
	if zeroReply != "ما عندي مبيعات مسجلة اليوم حتى الآن." {
		t.Fatalf("zero sales reply = %q", zeroReply)
	}
}

func TestBuildReplyUsesLiveProductQuantityForFollowUp(t *testing.T) {
	reply := buildReply(IntentInventory, "وكم صارت؟", []Message{{Role: "user", Content: "كم كمية كرت شاشة؟"}}, &dashboard.DashboardStats{}, storeSummary{
		ProductQuantities: map[string]int{"كرت شاشة": 4},
	})
	if !strings.Contains(reply, "4") {
		t.Fatalf("quantity reply = %q, want current quantity", reply)
	}
}

func TestGreetingReplyUsesMessageContext(t *testing.T) {
	if got := greetingReply("صباح الخير"); got != "صباح الخير 🌤️ جاهز نبدأ؟" {
		t.Fatalf("morning greeting = %q", got)
	}
}

func TestFollowUpIntentKeepsConversationContext(t *testing.T) {
	got := detectIntent("والربح؟", []Message{
		{Role: "user", Content: "كم مبيعات اليوم؟"},
		{Role: "assistant", Content: "مبيعات اليوم وصلت لـ ₪11680."},
	})
	if got != IntentProfit {
		t.Fatalf("follow-up intent = %s, want %s", got, IntentProfit)
	}
}

func TestConversationalRepliesDoNotLeakBusinessFallback(t *testing.T) {
	identity := buildConversationalReply(IntentAssistantIdentity, "مين انت؟", nil)
	if strings.Contains(identity, "11680") || strings.Contains(identity, "5580") {
		t.Fatalf("identity reply leaked business data: %q", identity)
	}
	if got := buildConversationalReply(IntentClarification, "؟", nil); got != "معك 👋 شو حابب تعرف؟" {
		t.Fatalf("clarification reply = %q", got)
	}
	if got := buildConversationalReply(IntentAssistantBehavior, "ليش بتكرر نفس الكلام؟", nil); !strings.Contains(got, "معك حق") {
		t.Fatalf("behavior reply = %q", got)
	}
}

func TestRepeatedIdentityQuestionChangesReply(t *testing.T) {
	first := buildConversationalReply(IntentAssistantIdentity, "مين انت؟", nil)
	second := buildConversationalReply(IntentAssistantIdentity, "مين انت؟", []Message{{Role: "user", Content: "مين انت؟"}})
	if first == second {
		t.Fatalf("repeated identity reply did not change: %q", first)
	}
}

func TestAdvicePrioritizesCurrentStoreRisks(t *testing.T) {
	stats := &dashboard.DashboardStats{LowStockCount: 3, OutstandingDebts: 250, TodaySales: 1200, TodayProfit: 300}
	reply := buildAdviceReply("شو بتنصحني أعمل؟", stats, storeSummary{}, func(value float64) string { return fmt.Sprintf("₪%.0f", value) })
	if !strings.Contains(reply, "3 منتجات") || !strings.Contains(reply, "₪250") {
		t.Fatalf("advice reply did not prioritize current risks: %q", reply)
	}
}

func TestAdviceCoversBusinessPlanningTopics(t *testing.T) {
	stats := &dashboard.DashboardStats{TodayProfit: 300, OutstandingDebts: 250}
	summary := storeSummary{MonthlySales: 8500, MonthlyPurchases: 2400, MonthlyExpenses: 700, PurchasesCount: 2}
	money := func(value float64) string { return fmt.Sprintf("₪%.0f", value) }
	cases := []struct {
		message string
		terms   []string
	}{
		{"كيف أرتب حساباتي محاسبيًا؟", []string{"محاسبيًا", "₪8500", "₪2400"}},
		{"كيف أحسن وضعي المالي؟", []string{"ماليًا", "₪250"}},
		{"أعطني نصيحة إدارية", []string{"إداريًا", "عملية شراء"}},
		{"كيف أخطط للشهر القادم؟", []string{"للتخطيط", "هدفًا", "الشهر"}},
		{"شو رأيك أعمل عروض؟", []string{"للعروض", "هامش الربح"}},
	}
	for _, tc := range cases {
		reply := buildAdviceReply(tc.message, stats, summary, money)
		for _, term := range tc.terms {
			if !strings.Contains(reply, term) {
				t.Fatalf("advice %q = %q, missing %q", tc.message, reply, term)
			}
		}
	}
}

func TestAlertsDoNotInventProblems(t *testing.T) {
	stats := &dashboard.DashboardStats{TodaySales: 1200, TodayProfit: 300}
	reply := buildAlertsReply(stats, storeSummary{}, func(value float64) string { return fmt.Sprintf("₪%.0f", value) })
	if reply != "ما عندي تنبيهات حرجة حاليًا. المخزون والديون مستقران حسب البيانات الحالية." {
		t.Fatalf("stable alerts reply = %q", reply)
	}
}

func TestAdviceReturnsDirectActionsForCurrentRisks(t *testing.T) {
	actions := actionsForIntent(IntentAdvice, &dashboard.DashboardStats{LowStockCount: 2, OutstandingDebts: 180})
	if len(actions) != 3 {
		t.Fatalf("action count = %d, want 3", len(actions))
	}
	for _, action := range actions {
		if action.Label == "" || action.Path == "" {
			t.Fatalf("invalid assistant action: %+v", action)
		}
	}
}

func TestNavigationActionOpensPointOfSale(t *testing.T) {
	action := navigationAction("افتح نقطة البيع")
	if action.Path != "/app/sales" || action.Label != "فتح نقطة البيع" {
		t.Fatalf("navigation action = %+v", action)
	}
}

func TestNavigationIntentSupportsStartingSale(t *testing.T) {
	if got := detectIntent("ابدأ عملية بيع", nil); got != IntentNavigation {
		t.Fatalf("intent = %s, want %s", got, IntentNavigation)
	}
	if action := navigationAction("ابدأ عملية بيع"); action.Path != "/app/sales" {
		t.Fatalf("sale action = %+v", action)
	}
}

func TestNavigationIntentSupportsUnvocalizedSaleCommands(t *testing.T) {
	for _, message := range []string{"بيع جديد", "ابدا بعملية بيع"} {
		if got := detectIntent(message, nil); got != IntentNavigation {
			t.Fatalf("%q intent = %s, want %s", message, got, IntentNavigation)
		}
	}
}
