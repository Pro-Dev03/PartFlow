# PartFlow Engineering Rules

**Status:** Active engineering contract  
**Effective date:** 2026-09-16  
**Scope:** Backend, SQLite, PostgreSQL, API, Frontend, Electron, Worker, migrations, tests, and reports.

## Evidence Classification Standard (mandatory baseline)

هذه الوثيقة تضع قاعدة ثابتة لتقييم أي ادعاء فني أو تجميعي قبل أي رفع للحالة:

- **PASS** = مثبت باختبار قابل لإعادة التشغيل أو دليل حي يثبت كامل الادعاء عبر الطبقات المطلوبة (`DB -> Query -> Backend -> API` أو `DB -> Repository -> Service -> API -> Frontend` حسب النطاق). لا يُقبل ادعاء PASS من خلال قراءة الكود فقط.
- **PARTIAL** = بعض الطبقات مثبتة، لكن طبقة أو أكثر ما زالت مفتوحة أو غير مكتملة. هذا يعني أن جزء من السلسلة مثبت، ولكن السلسلة الكاملة أو المستهلك النهائي لم يُثبت بعد.
- **NOT PROVEN** = لا يوجد دليل كافٍ لإثبات الادعاء. لا يُسمح بإنشاء Aggregation أو fallback لإخفاء الاختلافات أو لتقريب القيم إلى "مطابقة" دون توثيق واضح.
- **Browser E2E** = يحتاج اختبار متصفح فعلي مع بيانات حقيقية أو متحكم فيها؛ إذا لم يوجد، يبقى في حالة `NOT PROVEN` أو `PARTIAL` بحسب ما تم تثبيته في Backend/API فقط.

**Baseline الحالي متوقف (2026-09-16):**

- `Net Sales`, `Tax`, `Gross Profit`, `Net Profit` = **PASS** فقط داخل backend monthly reconciliation validated.
- `Monthly Returns` = **PARTIAL**: بعض الحسابات واردة ضمن الربح الشهري، لكن لا يوجد إثبات مباشر مستقل لقيمة `Monthly Returns` كقيمة رسمية.
- `Monthly Purchases` = **NOT PROVEN**.
- `Monthly Supplier Balance` = **NOT PROVEN**.
- `Monthly Customer Debt` = **NOT PROVEN**.
- `Dashboard/Reports UI parity` = **PARTIAL**.
- `Browser E2E` = **NOT PROVEN**.

أية خطوة تطوير لاحقة يجب أن تبني على هذه الحالة الموثقة، ولا يجوز رفع أي قيمة من `PARTIAL` أو `NOT PROVEN` إلى `PASS` بدون evidence مباشر ومكتوب ومرئي.

هذه الوثيقة هي المرجع العملي عند إضافة ميزة أو إصلاح عطل. لا تهدف إلى إعادة تصميم النظام دفعة واحدة؛ بل تمنع التناقضات الجديدة وتحدد خطوات نقل الوحدات تدريجيًا إلى بنية قابلة للصيانة.

قرارات الملكية ومصادر الحقيقة الحالية لكل Entity وقيمة مالية موثقة في [ARCHITECTURE-BASELINE.md](ARCHITECTURE-BASELINE.md). عند التعارض، يجب تحديث الـ Baseline وتسجيل حالة الدليل بدل إنشاء مصدر جديد بصمت.

## 1. Single Source of Truth

كل قيمة مالية أو تشغيلية يجب أن تملك مصدرًا معتمدًا واحدًا.

| القيمة | المصدر المعتمد | المستهلكون |
|---|---|---|
| إجمالي البيع والضريبة والخصم والدفع | `sales` + `sale_items` عبر Sales Service | POS، الفواتير، dashboard، reports |
| تكلفة البيع وCOGS والربح | Sales/Accounting Service وقواعد الربح المركزية | sales response، dashboard، reports |
| تكلفة الشراء | `purchase_items.unit_price` ثم `inventory_items.purchase_cost` عند الاستلام | purchases، inventory، supplier ledger |
| المخزون الحالي | Inventory Service وحالة `inventory`/`inventory_items` | inventory، POS، dashboard |
| رصيد العميل والدين | Customer/Debt Service وledger المعتمد | customers، debts، dashboard، reports |
| رصيد المورد | Supplier Ledger Service | suppliers، purchases، supplier returns |
| المرتجع وقيمته | Returns Service و`returns`/`return_items` | returns، sales، profit، inventory |
| التاريخ المالي | Business Date الخاص بالكيان | التقارير، التجميعات، aging |

القواعد:

- لا يعيد Frontend حساب قيمة مالية موجودة في Backend إلا لتنسيق العرض، مثل rounding/display formatting.
- لا يقرأ تقرير أو بطاقة من مصدرين ثم يستخدم `Math.max` أو fallback لإخفاء اختلاف المصدرين دون تسجيله.
- أي قيمة مشتقة يجب أن تحدد: مصدرها، معادلتها، مالكها، واختبار regression لها.

## 2. Entity Ownership

كل Entity يملك Service واحدًا لقواعد الأعمال وRepository واحدًا للوصول إلى التخزين. يمكن لبقية الوحدات القراءة عبر contract، لكن لا تعدل الجداول مباشرة.

| Entity | Business owner | Persistence owner | التعديل المسموح |
|---|---|---|---|
| Product/Category/Brand | Products Service | Products Repository | Products وعمليات الإدارة فقط |
| Inventory/Inventory Item | Inventory Service | Inventory Repository | Purchase Receive، Sale، Return عبر service contract فقط |
| Purchase/Purchase Item | Purchases Service | Purchases Repository | إنشاء وتعديل وعكس الشراء والاستلام |
| Sale/Sale Item | Sales Service | Sales Repository | إنشاء البيع، الدفع، الإلغاء، العكس |
| Customer/Debt | Customers/Debts Service | Customer/Debt Repository | إنشاء الدين وتسويته عبر service |
| Supplier/Supplier Balance | Suppliers/Supplier Ledger Service | Supplier repositories | شراء، دفع، مرتجع عبر contracts |
| Payment | Payments Service | Payments Repository | إنشاء/إكمال/إلغاء الدفع |
| Customer Return | Returns Service | Returns Repository | إنشاء، موافقة، refund، إكمال، عكس |
| Supplier Return | Supplier Returns Service | Supplier Return Repository | إنشاء البنود والإكمال والعكس |
| Expense | Expenses Service | Expenses Repository | إنشاء واعتماد ورفض وتكرار |
| Dashboard/Aggregation | Dashboard/Aggregation Service | Summary repositories | قراءة وتجميع، لا يملك transaction data |
| Notification | Notifications Service/Worker | Notifications Repository | إنشاء الإشعار وتغيير حالته |
| User/Auth/Subscription | Auth Service + Cloud authority | Auth repository/cloud API | الجلسات والصلاحية فقط |

قواعد الحدود:

- Handler يحول HTTP إلى DTO ويستدعي Service؛ لا يضع transaction business logic.
- Service يملك قواعد الأعمال والمعاملات.
- Repository يملك SQL والتحويل بين DB models وdomain models.
- Worker يستدعي Services؛ لا يكرر SQL أو قواعد مالية داخل `cmd/worker`.
- Frontend لا يكتب إلى DB ولا يقرر حالة مالية.
- أي استثناء يحتاج توثيقًا في نفس التغيير واختبار boundary.

## 3. Official Business Date

`created_at` هو timestamp لإنشاء السجل أو الحدث، وليس Business Date تلقائيًا.

| Entity | Official Business Date | `created_at` |
|---|---|---|
| Sale | `sales.sale_date` | وقت إنشاء السجل والترتيب |
| Purchase | `purchases.purchase_date` | وقت إنشاء السجل |
| Customer Return | `returns.return_date` | وقت إنشاء طلب المرتجع |
| Supplier Return | تاريخ المرتجع في عقد Supplier Return | وقت الإنشاء |
| Customer/Supplier Payment | `payments.payment_date` | وقت إدخال الدفعة |
| Expense | `expenses.expense_date` | وقت إنشاء المصروف |
| Debt | تاريخ إنشاء الدين إن توفر، و`due_date` للاستحقاق فقط | لا يستخدم كبديل صامت لتاريخ مالي |
| Opening Stock | `inventory_movements.business_date` مع `source_type=OPENING_STOCK` | وقت تسجيل الحركة |
| Inventory event | تاريخ الحدث المتخصص مثل `purchase_date` أو `sold_at` أو `business_date` | وقت كتابة event row |

القواعد:

- كل filter أو group مالي يستخدم Business Date الرسمي.
- legacy fallback إلى `created_at` مسموح فقط إذا كان الحقل الرسمي غير موجود أو NULL، ويجب تحويله عبر Store Timezone وتوثيقه.
- كل Entity جديد يجب أن يضيف Business Date إلى contract قبل إضافة التقارير.
- اختبار boundary مطلوب حول منتصف الليل في Store Timezone.

## 4. Database Contract

PostgreSQL وSQLite يمثلان نفس الـ domain contract، حتى لو اختلفت صياغة SQL.

يجب أن يحدد كل تغيير:

- الجداول والحقول والأنواع والـ nullability والـ defaults.
- المفاتيح الفريدة والعلاقات وقواعد الحالة.
- PostgreSQL migration المقابلة.
- SQLite bootstrap/upgrade المقابل.
- backward compatibility مع البيانات القديمة.
- خطة rollback أو migration recovery إن كان التغيير destructive.

بوابات إلزامية لكل تغيير schema:

1. Bootstrap من قاعدة PostgreSQL فارغة.
2. Bootstrap من SQLite فارغة.
3. Legacy upgrade من نسخة سابقة.
4. قراءة وكتابة entity في المحركين.
5. اختبار nullable/legacy rows عند الحاجة.
6. عدم حذف البيانات لإخفاء عدم التوافق.

لا تعتبر طبقة compatibility schema عقدًا نهائيًا. هي جسر انتقال يجب أن يكون له سبب إزالة أو تثبيت موثق.

## 5. API Contract

كل endpoint يملك response contract ثابتًا:

```json
{
  "success": true,
  "data": {},
  "meta": {}
}
```

وقالب الخطأ:

```json
{
  "success": false,
  "error": {
    "code": "STABLE_ERROR_CODE",
    "message": "Human-readable message"
  }
}
```

القواعد:

- DTOs العامة لا تتغير بصمت.
- إضافة field يجب ألا تكسر المستهلكين، وتغيير المعنى يحتاج version أو migration contract.
- كل mutation يحدد status transitions وidempotency behavior وtransaction boundary.
- كل endpoint جديد يملك test response shape وtest error shape.
- عند تغيير DTO يجب مراجعة كل Frontend consumers وWorker وsync snapshot والتقارير.

## 6. Frontend Data Flow

المسار المعتمد:

```text
DB -> Repository -> Service -> Handler/API DTO -> api client -> Query/Store -> UI
```

القواعد:

- Frontend يعرض القيم القادمة من Backend.
- يسمح فقط بتحويلات العرض: locale، currency formatting، labels، sorting البصري.
- لا يعيد حساب tax/COGS/profit/debt/balance/return totals.
- React Query cache لا يصبح مصدر حقيقة دائمًا؛ invalidation بعد mutation إلزامي.
- response normalization يكون في API boundary واحد، لا داخل كل صفحة.
- إذا احتاجت الشاشة قيمة جديدة، يضاف field إلى API contract بدل حسابه محليًا.

## 7. Regression Rule

أي تعديل في DB أو Repository أو Service أو API يجب أن يحدد:

- البطاقات المتأثرة.
- الصفحات المتأثرة.
- التقارير والتجميعات المتأثرة.
- عمليات المزامنة والـ Worker المتأثرة.
- اختبارات الوحدة والتكامل وE2E المطلوبة.

لا يغلق التغيير قبل تشغيل regression scope المرتبط به. لا يكفي تشغيل اختبار الملف المعدل إذا كان التغيير يغير قيمة مالية مشتركة.

### 7.1 Purchases maintenance contract

هذا هو النموذج المرجعي لصيانة دورة Purchases. أي تغيير مستقبلي في Purchase أو Purchase Item يجب أن يبدأ بمراجعة هذه السلسلة، ثم يحدد ما الذي تغير منها وما الذي سيعاد اختباره.

| Layer | Owner / consumers | Contract |
|---|---|---|
| Source of Truth | `purchases`, `purchase_items`, `inventory`, `inventory_items`, `supplier_ledger`, `payments`, `supplier_returns` | `purchases.purchase_date` هو Business Date؛ تكلفة الشراء من `purchase_items.unit_price` ثم `inventory_items.purchase_cost` عند الاستلام؛ الرصيد والتأثيرات المالية من ledger/services الرسمية |
| Owner Service | `backend/internal/purchases/service.go` + `backend/internal/purchases/repository.go` | يملك إنشاء وتعديل واستلام ودفع وعكس الشراء، وإنشاء بنود الشراء، وتسجيل debit الشراء؛ لا تكرر القواعد المالية خارجه |
| API Consumers | `backend/internal/api/router.go` وPurchase handlers | `POST/GET/PUT /api/purchases`، `POST /:id/receive`، `POST /:id/payment`، `POST /:id/reverse`، بنود الشراء، وSmart Delete؛ الـ Handler يحول DTO فقط ويستدعي Service |
| Frontend Consumers | `frontend/src/features/purchases/pages/`، `frontend/src/features/purchases/hooks/usePurchases`، `frontend/src/services/api/endpoints` | قائمة المشتريات، الإنشاء، التعديل، التفاصيل، الاستلام، الدفع، العكس؛ الواجهة تعرض response ولا تعيد حساب totals أو balance، وتعمل invalidation بعد mutation |
| Reports / Dashboard Consumers | Dashboard stats/activity، purchase summaries، inventory/report aggregations، supplier summary/ledger، Supplier Returns | أي تغيير في total/paid/remaining/status/cost/received quantity يجب أن يراجع بطاقات Dashboard وActivity وتقارير الشراء والمخزون ورصيد المورد والمرتجعات |
| Regression Tests | `TestPurchaseLifecycleSupplierBalanceAndReturnLedgerSQLite`، Purchases package tests، Supplier Return lifecycle، Supplier Ledger payment lifecycle | الحد الأدنى يثبت Product -> Purchase -> Receive -> Inventory -> Supplier Balance -> Supplier Payment -> Supplier Return، مع partial/full/overpayment/duplicate/return-credit cases عند تأثرها |

**حالة Purchases المعتمدة حاليًا**

- **Purchases = PASS للـ Backend:** الدليل القابل لإعادة التشغيل يغطي DB/Repository/Service لدورة الشراء الكاملة.
- **API/UI/E2E = PARTIAL:** لا يوجد بعد دليل واحد مكتمل من API إلى واجهة المتصفح إلى E2E لكل الدورة؛ لا يجوز رفع الحالة إلى PASS بالاعتماد على اختبارات Backend فقط.
- لا تعني حالة Backend PASS أن المستهلكين الأماميين أو التقارير اجتازوا تلقائيًا؛ يجب تسجيل كل طبقة بشكل مستقل.

**Change Impact المطلوب لأي تغيير Purchase**

```text
Purchase Change Impact
What changes:
Source-of-truth fields/tables:
Owner service/repository:
API routes/DTOs:
Frontend consumers:
Inventory and supplier consumers:
Dashboard/reports/aggregations:
Sync/worker paths:
SQLite/PostgreSQL contracts:
Data preservation and state-transition impact:
Required focused tests:
Required regression packages:
Evidence status: PASS / PARTIAL / FAIL / NOT PROVEN
```

**بوابة الاعتماد الخاصة بـ Purchases**

1. إذا تغيرت قواعد Purchase أو Receive أو Payment أو Return، شغّل اختبار lifecycle الأساسي:
  `go test ./internal/purchases -run TestPurchaseLifecycleSupplierBalanceAndReturnLedgerSQLite -count=1`
2. شغّل الحزمة المتأثرة والمستهلك المالي المرتبط:
  `go test ./internal/purchases ./internal/supplierreturns ./internal/suppliers -count=1`
3. عند تغيير API أو DTO أو Frontend، أضف اختبار response/error أو اختبار UI/E2E المناسب، ثم شغّل فحوص Frontend المطلوبة.
4. عند تغيير schema أو query، تحقق من SQLite وPostgreSQL وlegacy/null behavior حسب الحقول المتأثرة.
5. راجع `docs/MASTER-EVIDENCE-MATRIX.md` و`docs/PHASE-2-IMPLEMENTATION-REPORT-2026-09-16.md`، ولا تغير حالة Purchases إلى PASS للواجهة قبل وجود دليل قابل لإعادة التشغيل لها.

### 7.2 Supplier Ledger maintenance contract

يستخدم Supplier Ledger نفس نموذج الصيانة. لا يُعتبر `suppliers.current_balance` مصدرًا مستقلاً عن الحركات؛ هو projection سريع يجب أن يتطابق مع الرصيد المحسوب من حركات `supplier_ledger` بعد تطبيق sign/type contract.

| Layer | Owner / consumers | Contract |
|---|---|---|
| Source of Truth | `supplier_ledger` كسجل الحركات، و`suppliers.current_balance` كـ current projection | المالك النهائي لقواعد الرصيد هو Supplier Ledger Service؛ `supplier_ledger.balance` يمثل الرصيد بعد الحركة، و`current_balance` يجب أن يساوي الرصيد الحالي القابل للقراءة. أي اختلاف بينهما evidence failure وليس سببًا لإضافة fallback في Frontend |
| Owner Service | `backend/internal/suppliers/service.go` + `backend/internal/suppliers/repository.go` | يملك قراءة الرصيد والـ ledger، قبول الدفع، منع overpayment وduplicate reference، إضافة debit/credit والتحديث المتسق للـ projection |
| Purchase operation | `backend/internal/purchases/service.go` | Purchase ينشئ `debit/PURCHASE` بقيمة الفاتورة ويرفع رصيد المورد؛ لا يجوز تسجيل Purchase debit من شاشة أو Repository آخر خارج contract |
| Supplier Payment operation | `suppliers.Service.AddPayment` وPurchase payment flow | Supplier Payment ينشئ `credit/PAYMENT` ويخفض الرصيد؛ يجب رفض المبلغ غير الصحيح، overpayment، وduplicate reference، مع تحديث `payments` و`supplier_ledger` و`suppliers.current_balance` معًا |
| Supplier Return operation | `backend/internal/supplierreturns/service.go` | إكمال Supplier Return يغير inventory إلى `RETURNED`، ينشئ `credit/SUPPLIER_RETURN` بقيمة التكلفة، ويخفض رصيد المورد؛ لا يعتمد على تحديث واجهة فقط |
| Credits / Adjustments | `AddLedgerEntry`, `AddDebt`, `CreateDebtEntry`, ومسارات supplier debt عند استخدامها | الـ debit يزيد الرصيد، والـ credit/adjustment الموثق يخفضه حسب نوع الحركة؛ كل adjustment يجب أن يملك reference وdescription وسببًا قابلًا للتدقيق واختبارًا مستقلًا. لا تستخدم `current_balance` كـ manual override بلا ledger entry |
| API Consumers | Supplier routes، Purchase payment routes، Supplier Return routes | `/api/suppliers/:id/ledger`، `/api/suppliers/:id/payments`، debt-summary/overdue، `/api/purchases/:id/payment`، و`/api/supplier-returns/:id/complete`؛ الـ handlers تحوّل DTO وتستدعي الخدمات |
| Frontend Consumers | `frontend/src/features/suppliers/`، `frontend/src/features/purchases/`، `frontend/src/features/supplier-returns/`، `frontend/src/services/api/endpoints.ts` | Supplier list/cards تعرض outstanding، Supplier modal تعرض ledger/payments/return credits، Purchase details تعرض supplier ledger، Supplier Returns تنفذ credit flow؛ الواجهة لا تحسب الرصيد ولا تخفي اختلاف current_balance وledger |
| Reports / Dashboard Consumers | Supplier summary، Purchases stats/details، Reports supplier breakdown، Dashboard/activity، overdue supplier views | تعتمد على total purchases، payments، return credits، outstanding/current balance، aging/overdue؛ أي تغيير في ledger movement أو sign يجب أن يراجع هذه المستهلكات كلها |
| Regression Tests | `TestSupplierLedgerPaymentLifecycleSQLite`، `TestSupplierLedgerAdjustmentReconcilesProjectionSQLite`، `TestSupplierReturnCreditsLedgerAndRemovesInventorySQLite`، `TestPurchaseLifecycleSupplierBalanceAndReturnLedgerSQLite` | تثبت payment partial/full، overpayment rejection، duplicate-reference rejection، Purchase debit، Supplier Return credit، adjustment debit، inventory effect، تطابق projection، وعدم تكرار الحركة. Generic credit غير المرتبط بـ Supplier Return وAPI/UI/E2E تحتاج evidence إضافي قبل PASS الشامل |

**عمليات تغيير رصيد المورد**

| Operation | Ledger movement | Projection effect | Evidence status |
|---|---|---|---|
| Purchase | `debit/PURCHASE` | `current_balance += amount` | PASS على Backend ضمن Purchase lifecycle |
| Supplier Payment | `credit/PAYMENT` | `current_balance -= amount` | PASS في اختبار Supplier Ledger payment lifecycle على Backend |
| Supplier Return | `credit/SUPPLIER_RETURN` | `current_balance -= refund_amount` | PASS في اختبار Supplier Return وPurchase lifecycle على Backend |
| Supplier Credit / Adjustment | `credit` أو `ADJUSTMENT` مع reference | يخفض projection وفق نوع الحركة | PARTIAL: المسار موجود، لكن لا يوجد lifecycle مستقل يغطي كل أنواع credit/adjustment |
| Supplier Debt / Additional Debit | `debit` مع reference/due date | `current_balance += amount` | PARTIAL: API/service موجودان، والدليل الشامل عبر المستهلكين غير مكتمل |

**حالة Supplier Ledger المعتمدة حاليًا**

- **Backend operations: PASS:** الدليل الحالي يثبت Purchase debit، Supplier Payment، Supplier Return credit، ورصيد المورد النهائي ضمن اختبارات SQLite المركزة.
- **Supplier Ledger overall: PARTIAL:** adjustment debit أصبح مثبتًا باختبار مستقل، لكن generic credit غير المرتبط بـ Supplier Return ومسارات Supplier Debt الكاملة ليست كلها مثبتة في lifecycle مستقل شامل.
- **API/UI/E2E: PARTIAL:** routes والمستهلكون موجودون، لكن لا يوجد دليل قابل لإعادة التشغيل يغطي السلسلة كاملة من API إلى Frontend إلى E2E.
- لا توجد حالة `FAIL` مثبتة حاليًا لهذه الحدود؛ أي فرع غير مغطى يبقى `PARTIAL` أو `NOT PROVEN`، ولا يرفع إلى PASS بقراءة الكود فقط.

**بوابة الاعتماد الخاصة بـ Supplier Ledger**

1. عند تغيير Purchase debit أو supplier payment أو return credit شغّل:
  `go test ./internal/purchases -run TestPurchaseLifecycleSupplierBalanceAndReturnLedgerSQLite -count=1`
2. شغّل اختبار الدفع والمرتجع معًا:
  `go test ./internal/suppliers -run TestSupplierLedgerPaymentLifecycleSQLite -count=1`
  `go test ./internal/supplierreturns -run TestSupplierReturnCreditsLedgerAndRemovesInventorySQLite -count=1`
3. شغّل regression للحزم المرتبطة:
  `go test ./internal/purchases ./internal/supplierreturns ./internal/suppliers -count=1`
4. عند تغيير route/DTO/frontend، أضف response/error أو UI/E2E evidence، وراجع Supplier list، ledger modal، purchase details، returns، reports، dashboard، وoverdue views.
5. عند تغيير schema أو sign/type mapping، تحقق من SQLite وPostgreSQL، legacy columns (`type`/`transaction_type`)، nullable/reference rows، وتطابق `supplier_ledger.balance` مع `suppliers.current_balance`.
6. راجع `docs/MASTER-EVIDENCE-MATRIX.md` وسجل الحالة كـ `PASS` أو `PARTIAL` أو `FAIL` أو `NOT PROVEN` لكل boundary. لا تنتقل إلى Customer Debt قبل إغلاق فجوات Supplier Ledger المحددة بهذا العقد.

### 7.3 Customer Debt maintenance contract

هذا هو نموذج الصيانة القياسي لدورة Customer Debt. الدين والدفعة وCustomer Credit حركات مالية مختلفة؛ لا يجوز تسجيل Customer Credit كـ Payment أو جعل `customers.current_balance` يتغير دون حركة مقابلة قابلة للتدقيق.

| Layer | Owner / consumers | Contract |
|---|---|---|
| Source of Truth | `debts` للحالة والـ remaining، `customer_ledger` لسجل الحركات، `customers.current_balance` كـ projection، و`payments` للدفعات المقبولة | Customer Debt Service يملك قواعد الرصيد؛ `debit` يزيد الدين، `credit/PAYMENT` يخفضه، و`credit` المرتبط بمرتجع يمثل Customer Credit أو Debt Adjustment وليس Payment. يجب أن يتطابق signed ledger balance مع `customers.current_balance` وألا يصبح `debts.remaining_amount` سالبًا |
| Owner Service | `backend/internal/customers/service.go` + `backend/internal/customers/repository.go` | يملك إنشاء debt من Credit Sale، حدود الائتمان، partial/full payment، منع overpayment وduplicate reference، وتوزيع الدفع على الديون المفتوحة |
| Credit Sale / Debt | Sales Service ينشئ Credit Sale، وCustomer Service ينشئ `debts` وdebit ledger entry | Credit Sale يحتاج customer؛ إنشاء الدين يجب أن يملك reference إلى sale، ويظهر مرة واحدة في `debts` و`customer_ledger` |
| Customer Payment | `ProcessDebtPaymentWithReference` و`AddPayment` | يسجل Payment في `payments` و`customer_ledger`، يخصم من الدين مرة واحدة، ويرفض overpayment وduplicate reference قبل أي mutation |
| Customer Return / Debt Adjustment | `backend/internal/returns/service.go` + `backend/internal/returns/repository.go` | `DEBT_ADJUSTMENT` يطبق الجزء الممكن على debt، يمنع الدين السالب، ويحفظ الزيادة كـ `customer_credit` وحركة credit؛ لا ينشئ payment row للـ credit، ويحمي completion من التكرار |
| API Consumers | Customer routes، Debt routes، Sales routes، Return routes | `/api/customers/:id/ledger`، `/api/customers/:id/debt-summary`، `/api/customers/:id/payments`، `/api/customers/:id/debt-payments`، `/api/customers/:id/debts`، `/api/debts/*`، و`/api/v1/returns/:id/complete` |
| Frontend Consumers | `frontend/src/features/debts/`، `frontend/src/features/customers/`، `frontend/src/features/returns/`، `frontend/src/services/api/endpoints.ts` | Debt list/payment modal، customer ledger/profile، Create Return وReturn Details؛ تعرض remaining/debt_adjustment/customer_credit من API ولا تحول Customer Credit إلى paid amount |
| Reports / Dashboard Consumers | Dashboard debt cards، overdue debt views، Customers stats، Debts report، Reports page، customer financial timeline | تعتمد على `debts.remaining_amount` للحالي والمتأخر، وledger للحركة والتسلسل؛ يجب مراجعة total debt، paid، outstanding، overdue، customer count، وreturn credit عند أي تغيير |
| Regression Tests | `TestCustomerDebtLifecyclePaymentAndReturnCreditSQLite`، `TestSQLiteDebtAdjustmentCapsDebtAndCreatesCustomerCredit`، debt summary tests، returns SQLite tests | تثبت Credit Sale -> Debt -> Partial -> Full -> zero، overpayment، duplicate payment، return adjustment، customer credit، duplicate completion، non-negative debt، ledger/projection parity، وعدم تسجيل credit كـ payment |

**حالة Customer Debt الحالية**

- **Backend DB/Repository/Service: PASS:** lifecycle قابل لإعادة التشغيل يثبت الدين والدفعات والمرتجع والـ credit والتطابق وعدم التكرار.
- **Duplicate Payment: PASS:** أضيف reference validation في Customer Service/API قبل تعديل الدين أو الرصيد.
- **Customer Credit: PASS في مسار `DEBT_ADJUSTMENT`:** الزيادة بعد إغلاق الدين تحفظ كـ `customer_credit` وحركة ledger، ولا تسجل كـ Payment.
- **API/UI/E2E: PARTIAL:** package/API contract وFrontend build/consumer checks تحتاج live route/browser evidence قبل PASS الشامل.
- أي generic credit adjustment خارج Customer Return يبقى `NOT PROVEN` ما لم يوجد contract واختبار مستقل.

**بوابة الاعتماد الخاصة بـ Customer Debt**

1. شغّل lifecycle الأساسي:
  `go test ./internal/customers -run TestCustomerDebtLifecyclePaymentAndReturnCreditSQLite -count=1`
2. شغّل اختبارات المرتجعات والدين المرتبطة:
  `go test ./internal/returns -run 'TestSQLiteDebtAdjustmentCapsDebtAndCreatesCustomerCredit|TestServiceCreateReturnLinksActiveDebtForDebtAdjustment' -count=1`
3. شغّل regression للحزم المتأثرة:
  `go test ./internal/customers ./internal/returns ./internal/sales ./internal/api -count=1`
4. عند تغيير route/DTO/frontend، تحقق من customer ledger، debt list/payment modal، Create/Return Details، Dashboard، وDebts Report، وسجل live API/UI/E2E كـ PASS أو PARTIAL أو NOT PROVEN.
5. عند تغيير payment أو return SQL، تحقق من SQLite/PostgreSQL، duplicate/idempotency، signed ledger balance، `remaining_amount >= 0`، وأن Customer Credit لا يدخل في `payments` أو paid totals.

## 8. Financial Rules Centralization

القواعد المركزية المطلوبة:

- Sales totals, tax, discount, payment status, change.
- Purchase totals, tax, received/returned quantity, supplier debit.
- Payment allocation, partial payment, duplicate reference, overpayment.
- Debt creation, remaining amount, overdue status, settlement.
- Customer and supplier returns, refund/credit, inventory effect.
- COGS, gross profit, net profit, tax-adjusted refund.
- Cash flow and ledger balances.

كل قاعدة مالية يجب أن تكون قابلة للاستدعاء من أكثر من surface، وأن تملك اختبارات zero/partial/full/invalid/legacy cases. لا توضع معادلة جديدة في Dashboard أو Reports فقط.

## 9. Data Preservation

- لا تستخدم `DELETE`, reset، أو إعادة تهيئة قاعدة البيانات لإخفاء compatibility failure.
- السجل التجاري المؤكد يعكس أو يلغى وفق state machine، ولا يحذف بلا أثر.
- migration تحفظ الصفوف القديمة أو تحولها صراحة مع audit.
- أي repair script يملك dry-run، count قبل/بعد، وbackup/rollback plan.
- عمليات حذف البيانات الإدارية تبقى صريحة، محمية، ومسجلة في audit.

## 10. Evidence-Based Status

الحالات الوحيدة المسموحة:

- **PASS:** دليل قابل لإعادة التشغيل يغطي الادعاء المطلوب.
- **PARTIAL:** بعض الطبقات أو السيناريوهات ناجحة، لكن السلسلة كاملة غير مثبتة.
- **FAIL:** اختبار قابل لإعادة التشغيل يثبت مخالفة أو regression.
- **NOT PROVEN:** لا يوجد دليل كافٍ، دون استنتاج النجاح أو الفشل من قراءة الكود فقط.

كل تقرير evidence يجب أن يذكر الأمر، البيئة، التاريخ، artifact/trace، وما الذي لم يختبره.

## 11. Change Impact قبل التعديل

يجب أن يسبق كل تغيير هذا القالب في PR أو تقرير العمل:

```text
Change Impact
What changes:
Which entities:
Which services/repositories:
Which APIs/DTOs:
Which frontend screens/components:
Which dashboards/reports/aggregations:
Which sync/worker paths:
Which database migrations/contracts:
Which tests:
Data preservation plan:
Evidence status expected:
```

إذا لم يمكن ملء بند، يكتب `NOT APPLICABLE` مع السبب، لا يترك فارغًا.

## 12. Final Verification Gate

قبل اعتبار المرحلة مكتملة، يجب تشغيل ما ينطبق من التالي وتسجيل النتيجة:

```text
go test ./...
go vet ./...                 # عند تغييرات Backend
npm run test:run             # عند تغييرات Frontend
npm run build:check          # Typecheck + build
npm run lint                 # عند تغييرات Frontend
git diff --check
Focused integration tests
Focused E2E                    # عندما تتطلب المرحلة UI/API/DB evidence
```

لا تستخدم “نجح البناء” كبديل عن اختبار السلوك. ولا تستخدم “نجح اختبار الوحدة” كبديل عن E2E عندما يكون الادعاء عبر DB -> Backend -> API -> UI.

## 13. التطبيق التدريجي

لا يعاد تصميم PartFlow بالكامل دفعة واحدة. الترتيب المعتمد:

### المرحلة 1: تثبيت العقود

- اعتماد هذه الوثيقة.
- إنشاء Entity Ownership Matrix محدثة.
- تثبيت Business Date لكل Entity.
- تسجيل API DTOs والحقول المالية الأساسية.

### المرحلة 2: المال والمخزون

- Sales ثم Purchases ثم Payments/Debts ثم Returns.
- توحيد COGS/Profit/Ledger.
- إضافة regression لكل بطاقة وتقرير متأثر.

### المرحلة 3: Database parity

- توحيد bootstrap PostgreSQL/SQLite.
- إضافة legacy upgrade fixtures.
- تقليل compatibility SQL تدريجيًا دون حذف بيانات.

### المرحلة 4: Frontend/report consumers

- إزالة الحسابات المالية المكررة من الصفحات.
- توحيد response normalization في API boundary.
- ربط كل بطاقة وتقرير بمصدره المعتمد.

### المرحلة 5: Evidence closure

- إضافة lifecycle E2E مركزة لكل workflow.
- حفظ traces/network/DB evidence عند الحاجة.
- عدم إعلان PASS قبل اكتمال الطبقات المطلوبة.

## 14. قاعدة القرار

عند التعارض بين إصلاح سريع وقاعدة طويلة المدى، نختار الحل الذي:

1. يحافظ على البيانات.
2. يضع القاعدة في مالك الـ Entity الصحيح.
3. يحافظ على API contract أو يغيره بشكل معلن.
4. يثبت أثره باختبار قابل لإعادة التشغيل.
5. يحد التغيير في وحدة واحدة قبل توسيعه.

الهدف هو أن يصبح إصلاح PartFlow بعد سنة أو سنتين عملية متوقعة: نعرف المالك، المصدر، التاريخ الرسمي، العقود، المستهلكين، الاختبارات، ودليل الحالة قبل لمس الكود.
