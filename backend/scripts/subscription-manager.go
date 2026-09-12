package main

import (
	"bufio"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/ssh/terminal"
)

const banner = `
============================================================
 مدير اشتراكات PartFlow
============================================================
 إدارة الحسابات والتجديدات والإيقاف والحذف.
============================================================
`

type Account struct {
	ID                    string     `db:"id"`
	Email                 string     `db:"email"`
	FirstName             string     `db:"first_name"`
	LastName              string     `db:"last_name"`
	Phone                 string     `db:"phone"`
	IsActive              bool       `db:"is_active"`
	SubscriptionStatus    string     `db:"subscription_status"`
	SubscriptionExpiresAt *time.Time `db:"subscription_expires_at"`
	CreatedAt             time.Time  `db:"created_at"`
	UpdatedAt             time.Time  `db:"updated_at"`
}

type Config struct {
	DBURL string
}

// normalizeEmail keeps the administrator allow-list and database lookups
// consistent with the authentication middleware.
func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// configuredAdminEmails mirrors the backend's PARTFLOW_ADMIN_EMAILS policy.
// The owner address remains the backwards-compatible bootstrap administrator.
func configuredAdminEmails() []string {
	raw := strings.TrimSpace(os.Getenv("PARTFLOW_ADMIN_EMAILS"))

	seen := make(map[string]struct{})
	admins := make([]string, 0, 2)
	for _, value := range strings.Split(raw, ",") {
		email := normalizeEmail(value)
		if email == "" {
			continue
		}
		if _, exists := seen[email]; exists {
			continue
		}
		seen[email] = struct{}{}
		admins = append(admins, email)
	}
	if len(admins) == 0 {
		admins = append(admins, "owner@partflow.com")
	}
	return admins
}

func defaultAdminEmail() string {
	return configuredAdminEmails()[0]
}

func isAdminEmail(email string) bool {
	normalized := normalizeEmail(email)
	for _, configured := range configuredAdminEmails() {
		if normalized == configured {
			return true
		}
	}
	return false
}

func accountKind(email string) string {
	if isAdminEmail(email) {
		return "أدمن"
	}
	return "مشترك"
}

func main() {
	createCmd := flag.NewFlagSet("create", flag.ContinueOnError)
	createEmail := createCmd.String("email", "", "البريد الإلكتروني للحساب")
	createPassword := createCmd.String("password", "", "كلمة المرور للحساب")
	createFirstName := createCmd.String("first-name", "", "الاسم الأول")
	createLastName := createCmd.String("last-name", "", "اسم العائلة")
	createPhone := createCmd.String("phone", "", "رقم الهاتف")
	createDays := createCmd.Int("days", 30, "عدد أيام الاشتراك")
	createAdminFlag := createCmd.Bool("admin", false, "إنشاء حساب كمدير/أدمن")

	createAdminCmd := flag.NewFlagSet("create-admin", flag.ContinueOnError)
	createAdminEmail := createAdminCmd.String("email", "", "البريد الإلكتروني للحساب الإداري")
	createAdminPassword := createAdminCmd.String("password", "", "كلمة المرور للحساب الإداري")
	createAdminFirstName := createAdminCmd.String("first-name", "", "الاسم الأول")
	createAdminLastName := createAdminCmd.String("last-name", "", "اسم العائلة")
	createAdminPhone := createAdminCmd.String("phone", "", "رقم الهاتف")
	createAdminDays := createAdminCmd.Int("days", 30, "عدد أيام الاشتراك")

	renewCmd := flag.NewFlagSet("renew", flag.ContinueOnError)
	renewEmail := renewCmd.String("email", "", "البريد الإلكتروني للحساب")
	renewDays := renewCmd.Int("days", 30, "عدد الأيام المراد إضافتها")

	disableCmd := flag.NewFlagSet("disable", flag.ContinueOnError)
	disableEmail := disableCmd.String("email", "", "البريد الإلكتروني للحساب")
	disableReason := disableCmd.String("reason", "انتهت مدة الاشتراك", "سبب الإيقاف")

	deleteCmd := flag.NewFlagSet("delete", flag.ContinueOnError)
	deleteEmail := deleteCmd.String("email", "", "البريد الإلكتروني للحساب")

	listCmd := flag.NewFlagSet("list", flag.ContinueOnError)
	subscribersCmd := flag.NewFlagSet("subscribers", flag.ContinueOnError)
	statusCmd := flag.NewFlagSet("status", flag.ContinueOnError)
	statusEmail := statusCmd.String("email", "", "البريد الإلكتروني للحساب")

	summaryCmd := flag.NewFlagSet("summary", flag.ContinueOnError)
	adminPasswordCmd := flag.NewFlagSet("change-admin-password", flag.ContinueOnError)
	adminPasswordEmail := adminPasswordCmd.String("email", "", "admin account email (defaults to PARTFLOW_ADMIN_EMAILS)")

	if len(os.Args) < 2 {
		printBanner()
		cfg, err := loadConfig()
		if err != nil {
			fatalError(err)
		}
		db, err := connectDB(cfg)
		if err != nil {
			fatalError(err)
		}
		defer db.Close()
		runInteractiveApp(db)
		return
	}

	if os.Args[1] == "app" || os.Args[1] == "ui" || os.Args[1] == "menu" {
		printBanner()
		cfg, err := loadConfig()
		if err != nil {
			fatalError(err)
		}
		db, err := connectDB(cfg)
		if err != nil {
			fatalError(err)
		}
		defer db.Close()
		runInteractiveApp(db)
		return
	}

	if os.Args[1] == "help" || os.Args[1] == "-h" || os.Args[1] == "--help" {
		printBanner()
		printUsageAndExit(createCmd, createAdminCmd, renewCmd, disableCmd, deleteCmd, listCmd, subscribersCmd, statusCmd, summaryCmd, adminPasswordCmd)
	}

	cfg, err := loadConfig()
	if err != nil {
		fatalError(err)
	}

	db, err := connectDB(cfg)
	if err != nil {
		fatalError(err)
	}
	defer db.Close()

	switch os.Args[1] {
	case "create", "create-subscriber", "subscriber", "مشترك":
		if err := createCmd.Parse(os.Args[2:]); err != nil {
			fatalError(err)
		}
		if err := validateRequired("email", *createEmail); err != nil {
			fatalError(err)
		}
		if err := validateRequired("password", *createPassword); err != nil {
			fatalError(err)
		}
		if err := validateRequired("first-name", *createFirstName); err != nil {
			fatalError(err)
		}
		if err := validateRequired("last-name", *createLastName); err != nil {
			fatalError(err)
		}
		if *createDays <= 0 {
			fatalError(errors.New("يجب أن تكون الأيام أكبر من صفر"))
		}
		if err := validatePasswordStrength(*createPassword); err != nil {
			fatalError(err)
		}
		if err := createAccount(db, *createEmail, *createPassword, *createFirstName, *createLastName, *createPhone, *createDays); err != nil {
			fatalError(err)
		}
		if *createAdminFlag {
			fmt.Printf("⚠️  لاحظ: الحساب تم إنشاؤه كحساب فعلي، ويجب إضافة البريد إلى PARTFLOW_ADMIN_EMAILS لتمكين صلاحيات الإدارة.")
			fmt.Printf("\nالبريد الحالي المسموح: %s\n", strings.Join(configuredAdminEmails(), ", "))
		}
		fmt.Println("\n✅ تم إنشاء الحساب بنجاح.")
	case "create-admin", "admin-create":
		if err := createAdminCmd.Parse(os.Args[2:]); err != nil {
			fatalError(err)
		}
		if err := validateRequired("email", *createAdminEmail); err != nil {
			fatalError(err)
		}
		if err := validateRequired("password", *createAdminPassword); err != nil {
			fatalError(err)
		}
		if err := validateRequired("first-name", *createAdminFirstName); err != nil {
			fatalError(err)
		}
		if err := validateRequired("last-name", *createAdminLastName); err != nil {
			fatalError(err)
		}
		if *createAdminDays <= 0 {
			fatalError(errors.New("يجب أن تكون الأيام أكبر من صفر"))
		}
		if err := validatePasswordStrength(*createAdminPassword); err != nil {
			fatalError(err)
		}
		if err := createAccount(db, *createAdminEmail, *createAdminPassword, *createAdminFirstName, *createAdminLastName, *createAdminPhone, *createAdminDays); err != nil {
			fatalError(err)
		}
		fmt.Printf("⚠️  لإعطاء هذا الحساب صلاحية الإدارة فعلياً، أضف البريد إلى PARTFLOW_ADMIN_EMAILS أو استخدم owner@partflow.com كمدير أساسي.\n")
		fmt.Println("✅ تم إنشاء حساب الإدارة بنجاح.")
	case "renew":
		if err := renewCmd.Parse(os.Args[2:]); err != nil {
			fatalError(err)
		}
		if err := validateRequired("email", *renewEmail); err != nil {
			fatalError(err)
		}
		if *renewDays <= 0 {
			fatalError(errors.New("يجب أن تكون الأيام أكبر من صفر"))
		}
		if err := renewAccount(db, *renewEmail, *renewDays); err != nil {
			fatalError(err)
		}
		fmt.Println("✅ تم تجديد الاشتراك بنجاح.")
	case "disable":
		if err := disableCmd.Parse(os.Args[2:]); err != nil {
			fatalError(err)
		}
		if err := validateRequired("email", *disableEmail); err != nil {
			fatalError(err)
		}
		if err := disableAccount(db, *disableEmail, *disableReason); err != nil {
			fatalError(err)
		}
		fmt.Println("✅ تم إيقاف الحساب بنجاح.")
	case "delete":
		if err := deleteCmd.Parse(os.Args[2:]); err != nil {
			fatalError(err)
		}
		if err := validateRequired("email", *deleteEmail); err != nil {
			fatalError(err)
		}
		if err := deleteAccount(db, *deleteEmail); err != nil {
			fatalError(err)
		}
		fmt.Println("✅ تم حذف الحساب بنجاح.")
	case "list":
		if err := listCmd.Parse(os.Args[2:]); err != nil {
			fatalError(err)
		}
		if err := printAccounts(db); err != nil {
			fatalError(err)
		}
	case "subscribers", "مشتركين":
		if err := subscribersCmd.Parse(os.Args[2:]); err != nil {
			fatalError(err)
		}
		if err := printSubscribers(db); err != nil {
			fatalError(err)
		}
	case "status":
		if err := statusCmd.Parse(os.Args[2:]); err != nil {
			fatalError(err)
		}
		if err := validateRequired("email", *statusEmail); err != nil {
			fatalError(err)
		}
		if err := printAccountStatus(db, *statusEmail); err != nil {
			fatalError(err)
		}
	case "summary":
		if err := summaryCmd.Parse(os.Args[2:]); err != nil {
			fatalError(err)
		}
		if err := printSummaryOverview(db); err != nil {
			fatalError(err)
		}
	case "change-admin-password", "admin-password", "change-admin", "تغيير-كلمة-مرور-الأدمن":
		if err := adminPasswordCmd.Parse(os.Args[2:]); err != nil {
			fatalError(err)
		}
		email := normalizeEmail(*adminPasswordEmail)
		if email == "" {
			email = defaultAdminEmail()
		}
		fmt.Printf("حساب الإدارة المستهدف: %s\n", email)
		password, err := promptPassword("كلمة المرور الجديدة (لن تظهر أثناء الكتابة): ")
		if err != nil {
			fatalError(fmt.Errorf("فشل في قراءة كلمة المرور: %w", err))
		}
		if err := changeAdminPassword(db, email, password); err != nil {
			fatalError(err)
		}
		fmt.Println("✅ تم تغيير كلمة مرور حساب الإدارة وإبطال جلساته القديمة.")
	case "help", "-h", "--help":
		printUsageAndExit(createCmd, createAdminCmd, renewCmd, disableCmd, deleteCmd, listCmd, subscribersCmd, statusCmd, summaryCmd, adminPasswordCmd)
	default:
		printUsageAndExit(createCmd, createAdminCmd, renewCmd, disableCmd, deleteCmd, listCmd, subscribersCmd, statusCmd, summaryCmd, adminPasswordCmd)
	}
}

func loadConfig() (*Config, error) {
	loadDotEnv()

	// Prefer the same database connection the app already uses. The backend and
	// the scripts both should work against the same primary database, which may be
	// Supabase, PostgreSQL, or a local installation depending on the environment.
	candidates := []string{
		os.Getenv("DATABASE_URL"),
		os.Getenv("SUPABASE_DATABASE_URL"),
		os.Getenv("DATABASE_URL_DIRECT"),
	}

	var dbURL string
	for _, candidate := range candidates {
		trimmed := strings.TrimSpace(candidate)
		if trimmed != "" {
			dbURL = trimmed
			break
		}
	}

	if dbURL == "" {
		return nil, errors.New("لا يوجد رابط قاعدة بيانات متوفر. قم بإعداد DATABASE_URL أو SUPABASE_DATABASE_URL أو DATABASE_URL_DIRECT")
	}
	if strings.HasPrefix(dbURL, "=") {
		return nil, errors.New("تنسيق رابط قاعدة البيانات غير صحيح: استخدم DATABASE_URL=postgresql://... أو SUPABASE_DATABASE_URL=postgresql://...")
	}
	return &Config{DBURL: dbURL}, nil
}

// loadDotEnv loads the first .env file found while walking from the current
// directory up to the project root. This lets the script work when launched
// from backend/scripts, backend, or the repository root without putting
// credentials in source code. Existing environment variables always win.
func loadDotEnv() {
	directory, err := os.Getwd()
	if err != nil {
		return
	}

	for level := 0; level < 3; level++ {
		envPath := filepath.Join(directory, ".env")
		if _, err := os.Stat(envPath); err == nil {
			_ = godotenv.Load(envPath)
			return
		}

		parent := filepath.Dir(directory)
		if parent == directory {
			return
		}
		directory = parent
	}
}

func createAccount(db *sqlx.DB, email, password, firstName, lastName, phone string, days int) error {
	email = normalizeEmail(email)
	if strings.TrimSpace(email) == "" {
		return errors.New("البريد الإلكتروني مطلوب")
	}
	if strings.TrimSpace(password) == "" {
		return errors.New("كلمة المرور مطلوبة")
	}

	if len([]rune(password)) < 6 {
		return errors.New("password must be at least 6 characters")
	}
	if days <= 0 {
		return errors.New("subscription days must be greater than zero")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("فشل في تشفير كلمة المرور: %w", err)
	}

	var userID string
	err = db.Get(&userID, "SELECT id FROM users WHERE email = $1 LIMIT 1", email)
	if err == nil {
		return fmt.Errorf("الحساب موجود بالفعل: %s", email)
	}

	if !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("failed to check existing account: %w", err)
	}

	expiresAt := time.Now().AddDate(0, 0, days)
	accountID := uuid.New().String()

	_, err = db.Exec(`
		INSERT INTO users (
			id, email, password_hash, first_name, last_name, phone,
			is_active, subscription_status, subscription_expires_at,
			created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, true, 'active', $7, NOW(), NOW())
	`, accountID, email, string(hashed), strings.TrimSpace(firstName), strings.TrimSpace(lastName), strings.TrimSpace(phone), expiresAt)
	if err != nil {
		return fmt.Errorf("فشل في إدخال الحساب: %w", err)
	}

	printAccountSummary(email, "تم الإنشاء", expiresAt, "نشط")
	return nil
}

func renewAccount(db *sqlx.DB, email string, days int) error {
	email = normalizeEmail(email)
	var account Account
	err := db.Get(&account, `
		SELECT id, email, first_name, last_name, phone, is_active,
		       subscription_status, subscription_expires_at, created_at, updated_at
		FROM users WHERE email = $1 LIMIT 1
	`, strings.ToLower(email))
	if err != nil {
		return fmt.Errorf("account not found: %s", email)
	}

	now := time.Now().UTC()
	newExpiry := now.AddDate(0, 0, days)
	if account.SubscriptionExpiresAt != nil && account.SubscriptionExpiresAt.After(now) {
		newExpiry = account.SubscriptionExpiresAt.AddDate(0, 0, days)
	}

	_, err = db.Exec(`
		UPDATE users
		SET subscription_status = 'active',
		    subscription_expires_at = $1,
		    is_active = true,
		    updated_at = NOW()
		WHERE email = $2
	`, newExpiry, strings.ToLower(email))
	if err != nil {
		return fmt.Errorf("فشل في تجديد الاشتراك: %w", err)
	}

	printAccountSummary(email, "تم التجديد", newExpiry, "نشط")
	return nil
}

func disableAccount(db *sqlx.DB, email, reason string) error {
	email = normalizeEmail(email)
	var exists int
	err := db.Get(&exists, `SELECT 1 FROM users WHERE email = $1 LIMIT 1`, strings.ToLower(email))
	if err != nil {
		return fmt.Errorf("الحساب غير موجود: %s", email)
	}

	_, err = db.Exec(`
		UPDATE users
		SET is_active = false,
		    subscription_status = 'canceled',
		    updated_at = NOW()
		WHERE email = $1
	`, strings.ToLower(email))
	if err != nil {
		return fmt.Errorf("disable account: %w", err)
	}

	fmt.Printf("\n🛑 تم إيقاف الحساب: %s\nالسبب: %s\n", email, reason)
	return nil
}

func deleteAccount(db *sqlx.DB, email string) error {
	email = normalizeEmail(email)
	result, err := db.Exec(`DELETE FROM users WHERE email = $1`, strings.ToLower(email))
	if err != nil {
		return fmt.Errorf("فشل في حذف الحساب: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("فشل في قراءة نتيجة الحذف: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("الحساب غير موجود: %s", email)
	}

	fmt.Printf("\n🗑️ تم حذف الحساب: %s\n", email)
	return nil
}

func printAccounts(db *sqlx.DB) error {
	var accounts []Account
	if err := db.Select(&accounts, `
		SELECT id, email, first_name, last_name, phone, is_active,
		       subscription_status, subscription_expires_at, created_at, updated_at
		FROM users
		ORDER BY created_at DESC
	`); err != nil {
		return fmt.Errorf("فشل في عرض الحسابات: %w", err)
	}

	if len(accounts) == 0 {
		fmt.Println("لا توجد حسابات حاليًا.")
		return nil
	}

	fmt.Println("\n📋 الحسابات")
	fmt.Printf("Configured admin accounts: %s\n", strings.Join(configuredAdminEmails(), ", "))
	fmt.Println("--------------------------------------------------------------------------------------------------------------------------------")
	fmt.Printf("%-28s %-12s %-12s %-18s %-16s %-12s\n", "البريد", "الحالة", "الاشتراك", "تاريخ الانتهاء", "الأيام المتبقية", "الاسم")
	fmt.Println("--------------------------------------------------------------------------------------------------------------------------------")
	for _, account := range accounts {
		displayEmail := account.Email
		if isAdminEmail(account.Email) {
			displayEmail += " [ADMIN]"
		}
		expires := "غير محدد"
		daysLeft := "-"
		if account.SubscriptionExpiresAt != nil {
			expires = account.SubscriptionExpiresAt.Format("2006-01-02")
			daysLeft = fmt.Sprintf("%d", int(time.Until(*account.SubscriptionExpiresAt).Hours()/24))
			if time.Until(*account.SubscriptionExpiresAt) < 0 {
				daysLeft = "منتهي"
			}
		}
		status := "غير نشط"
		if account.IsActive {
			status = account.SubscriptionStatus
		}
		fullName := strings.TrimSpace(account.FirstName + " " + account.LastName)
		if fullName == "" {
			fullName = "-"
		}
		fmt.Printf("%-28s %-12s %-12s %-18s %-16s %-12s\n", displayEmail, status, account.SubscriptionStatus, expires, daysLeft, fullName)
	}
	fmt.Println("--------------------------------------------------------------------------------------------------------------------------------")
	return nil
}

func printSubscribers(db *sqlx.DB) error {
	var accounts []Account
	if err := db.Select(&accounts, `
		SELECT id, email, first_name, last_name, phone, is_active,
		       subscription_status, subscription_expires_at, created_at, updated_at
		FROM users
		WHERE is_active = true AND subscription_status IN ('active', 'trial')
		ORDER BY created_at DESC
	`); err != nil {
		return fmt.Errorf("فشل في عرض المشتركين: %w", err)
	}

	if len(accounts) == 0 {
		fmt.Println("لا يوجد مشتركون نشطون حاليًا.")
		return nil
	}

	fmt.Println("\n👥 المشتركون النشطون")
	fmt.Println("--------------------------------------------------------------------")
	fmt.Printf("%-28s %-14s %-18s %-20s\n", "البريد", "الحالة", "الاشتراك", "ينتهي في")
	fmt.Println("--------------------------------------------------------------------")
	subscriberCount := 0
	for _, account := range accounts {
		if isAdminEmail(account.Email) {
			continue
		}
		subscriberCount++
		expires := "غير محدد"
		if account.SubscriptionExpiresAt != nil {
			expires = account.SubscriptionExpiresAt.Format(time.RFC3339)
		}
		fmt.Printf("%-28s %-14s %-18s %-20s\n", account.Email, account.SubscriptionStatus, account.SubscriptionStatus, expires)
	}
	fmt.Println("--------------------------------------------------------------------")
	fmt.Printf("Subscriber count (excluding admins): %d\n", subscriberCount)
	return nil
}

func printSummaryOverview(db *sqlx.DB) error {
	var total, active, expired, disabled int
	if err := db.Get(&total, `SELECT COUNT(*) FROM users`); err != nil {
		return fmt.Errorf("فشل في حساب العدد الإجمالي: %w", err)
	}
	if err := db.Get(&active, `SELECT COUNT(*) FROM users WHERE is_active = true AND subscription_status = 'active'`); err != nil {
		return fmt.Errorf("فشل في حساب المشتركين النشطين: %w", err)
	}
	if err := db.Get(&expired, `SELECT COUNT(*) FROM users WHERE subscription_status = 'expired' OR subscription_expires_at < NOW()`); err != nil {
		return fmt.Errorf("فشل في حساب المتأخرين: %w", err)
	}
	if err := db.Get(&disabled, `SELECT COUNT(*) FROM users WHERE is_active = false OR subscription_status = 'canceled' OR subscription_status = 'cancelled'`); err != nil {
		return fmt.Errorf("فشل في حساب المعطلين: %w", err)
	}

	fmt.Println("\n============================================================")
	fmt.Println("لوحة مراقبة اشتراكات PartFlow")
	fmt.Println("============================================================")
	fmt.Printf("إجمالي الحسابات: %d\n", total)
	fmt.Printf("نشط: %d\n", active)
	fmt.Printf("منتهي: %d\n", expired)
	fmt.Printf("معطل/ملغي: %d\n", disabled)
	fmt.Println("============================================================")
	return nil
}

func printAccountStatus(db *sqlx.DB, email string) error {
	email = normalizeEmail(email)
	var account Account
	err := db.Get(&account, `
		SELECT id, email, first_name, last_name, phone, is_active,
		       subscription_status, subscription_expires_at, created_at, updated_at
		FROM users WHERE email = $1 LIMIT 1
	`, strings.ToLower(email))
	if err != nil {
		return fmt.Errorf("الحساب غير موجود: %s", email)
	}

	daysLeft := "غير محدد"
	if account.SubscriptionExpiresAt != nil {
		daysLeft = formatDaysRemaining(*account.SubscriptionExpiresAt)
	}

	fmt.Println("\n👤 حالة الحساب")
	fmt.Println("------------------------------------------------------------")
	fmt.Printf("البريد: %s\n", account.Email)
	fmt.Printf("الاسم: %s %s\n", account.FirstName, account.LastName)
	fmt.Printf("الهاتف: %s\n", account.Phone)
	fmt.Printf("نشط: %t\n", account.IsActive)
	fmt.Printf("حالة الاشتراك: %s\n", account.SubscriptionStatus)
	fmt.Printf("نوع الحساب: %s\n", accountKind(account.Email))
	fmt.Printf("الأيام المتبقية: %s\n", daysLeft)
	if account.SubscriptionExpiresAt != nil {
		fmt.Printf("تاريخ الانتهاء: %s\n", account.SubscriptionExpiresAt.Format(time.RFC3339))
	} else {
		fmt.Printf("تاريخ الانتهاء: لا يوجد\n")
	}
	fmt.Println("------------------------------------------------------------")
	return nil
}

func printBanner() {
	fmt.Print(banner)
}

func printUsageAndExit(createCmd, createAdminCmd, renewCmd, disableCmd, deleteCmd, listCmd, subscribersCmd, statusCmd, summaryCmd, adminPasswordCmd *flag.FlagSet) {
	fmt.Println("الاستخدام:")
	fmt.Println("  go run ./scripts/subscription-manager.go app")
	fmt.Println("  go run ./scripts/subscription-manager.go summary")
	fmt.Println("  go run ./scripts/subscription-manager.go create --email user@example.com --password Pass123 --first-name علي --last-name السعدي --days 30")
	fmt.Println("  go run ./scripts/subscription-manager.go create-admin --email admin@example.com --password Pass123 --first-name Ali --last-name Admin --days 30")
	fmt.Println("  go run ./scripts/subscription-manager.go renew --email user@example.com --days 60")
	fmt.Println("  go run ./scripts/subscription-manager.go disable --email user@example.com")
	fmt.Println("  go run ./scripts/subscription-manager.go delete --email user@example.com")
	fmt.Println("  go run ./scripts/subscription-manager.go list")
	fmt.Println("  go run ./scripts/subscription-manager.go subscribers")
	fmt.Println("  go run ./scripts/subscription-manager.go status --email user@example.com")
	fmt.Println("  go run ./scripts/subscription-manager.go create-subscriber --email user@example.com --password Pass123 --first-name Ali --last-name User --days 30")
	fmt.Println("  go run ./scripts/subscription-manager.go change-admin-password --email owner@partflow.com")
	fmt.Println("  تُدخل كلمة مرور الأدمن تفاعلياً ولا تُرسل داخل سطر الأوامر.")
	fmt.Println("  حسابات الأدمن تُحدد عبر PARTFLOW_ADMIN_EMAILS؛ والافتراضي owner@partflow.com")
	fmt.Println("")
	fmt.Println("الخيارات:")
	createCmd.PrintDefaults()
	createAdminCmd.PrintDefaults()
	renewCmd.PrintDefaults()
	disableCmd.PrintDefaults()
	deleteCmd.PrintDefaults()
	listCmd.PrintDefaults()
	subscribersCmd.PrintDefaults()
	statusCmd.PrintDefaults()
	summaryCmd.PrintDefaults()
	adminPasswordCmd.PrintDefaults()
	os.Exit(1)
}

func connectDB(cfg *Config) (*sqlx.DB, error) {
	db, err := sqlx.Connect("postgres", cfg.DBURL)
	if err != nil {
		return nil, fmt.Errorf("فشل في الاتصال بقاعدة البيانات: %w", err)
	}
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("فشل في اختبار الاتصال بقاعدة البيانات: %w", err)
	}
	return db, nil
}

func runInteractiveApp(db *sqlx.DB) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("Administrator account: %s (admin)\n", defaultAdminEmail())
	for {
		fmt.Println("\n========================================")
		fmt.Println("   لوحة إدارة اشتراكات PartFlow")
		fmt.Println("========================================")
		fmt.Println("1) عرض جميع الحسابات")
		fmt.Println("2) عرض المشتركين النشطين")
		fmt.Println("3) إنشاء مشترك جديد")
		fmt.Println("4) تجديد اشتراك")
		fmt.Println("5) إيقاف حساب")
		fmt.Println("6) حذف حساب")
		fmt.Println("7) تغيير كلمة مرور الأدمن")
		fmt.Println("8) حالة حساب")
		fmt.Println("9) الخروج")
		fmt.Println("========================================")
		fmt.Print("اختر رقم الخيار: ")

		choice, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("حدث خطأ في قراءة الإدخال.")
			return
		}

		switch strings.TrimSpace(choice) {
		case "0":
			if err := printSummaryOverview(db); err != nil {
				fmt.Println(err)
			}
		case "1":
			if err := printAccounts(db); err != nil {
				fmt.Println(err)
			}
		case "2":
			if err := printSubscribers(db); err != nil {
				fmt.Println(err)
			}
		case "3":
			createInteractiveAccount(db, reader)
		case "4":
			renewInteractiveAccount(db, reader)
		case "5":
			disableInteractiveAccount(db, reader)
		case "6":
			deleteInteractiveAccount(db, reader)
		case "7":
			changeInteractiveAdminPassword(db, reader)
		case "8":
			showInteractiveStatus(db, reader)
		case "9", "exit", "خروج":
			fmt.Println("تم الخروج من لوحة الإدارة.")
			return
		default:
			fmt.Println("الخيار غير صحيح. الرجاء اختيار رقم من 0 إلى 9.")
		}
	}
}

func createInteractiveAccount(db *sqlx.DB, reader *bufio.Reader) {
	fmt.Print("البريد الإلكتروني: ")
	email, _ := reader.ReadString('\n')
	password, err := promptPassword("Subscriber password (hidden): ", reader)
	if err != nil {
		fmt.Printf("Failed to read password: %v\n", err)
		return
	}
	fmt.Print("الاسم الأول: ")
	firstName, _ := reader.ReadString('\n')
	fmt.Print("اسم العائلة: ")
	lastName, _ := reader.ReadString('\n')
	fmt.Print("رقم الهاتف: ")
	phone, _ := reader.ReadString('\n')
	fmt.Print("عدد أيام الاشتراك: ")
	daysInput, _ := reader.ReadString('\n')
	days := 30
	if trimmed := strings.TrimSpace(daysInput); trimmed != "" {
		if parsed, err := strconv.Atoi(trimmed); err == nil && parsed > 0 {
			days = parsed
		}
	}
	if err := createAccount(db, strings.TrimSpace(email), strings.TrimSpace(password), strings.TrimSpace(firstName), strings.TrimSpace(lastName), strings.TrimSpace(phone), days); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("تم إنشاء الحساب بنجاح.")
}

func renewInteractiveAccount(db *sqlx.DB, reader *bufio.Reader) {
	fmt.Print("البريد الإلكتروني للحساب: ")
	email, _ := reader.ReadString('\n')
	fmt.Print("عدد الأيام المراد إضافتها: ")
	daysInput, _ := reader.ReadString('\n')
	days := 30
	if trimmed := strings.TrimSpace(daysInput); trimmed != "" {
		if parsed, err := strconv.Atoi(trimmed); err == nil && parsed > 0 {
			days = parsed
		}
	}
	if err := renewAccount(db, strings.TrimSpace(email), days); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("تم تجديد الاشتراك بنجاح.")
}

func disableInteractiveAccount(db *sqlx.DB, reader *bufio.Reader) {
	fmt.Print("البريد الإلكتروني للحساب: ")
	email, _ := reader.ReadString('\n')
	fmt.Print("سبب الإيقاف: ")
	reason, _ := reader.ReadString('\n')
	if err := disableAccount(db, strings.TrimSpace(email), strings.TrimSpace(reason)); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("تم إيقاف الحساب بنجاح.")
}

func deleteInteractiveAccount(db *sqlx.DB, reader *bufio.Reader) {
	fmt.Print("البريد الإلكتروني للحساب: ")
	email, _ := reader.ReadString('\n')
	if err := deleteAccount(db, strings.TrimSpace(email)); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("تم حذف الحساب بنجاح.")
}

func changeInteractiveAdminPassword(db *sqlx.DB, reader *bufio.Reader) {
	email := defaultAdminEmail()
	fmt.Printf("Administrator account: %s\n", email)
	password, err := promptPassword("New admin password (hidden): ", reader)
	if err != nil {
		fmt.Printf("Failed to read password: %v\n", err)
		return
	}
	confirmation, err := promptPassword("Repeat new admin password (hidden): ", reader)
	if err != nil {
		fmt.Printf("Failed to read password confirmation: %v\n", err)
		return
	}
	if password != confirmation {
		fmt.Println("Passwords do not match.")
		return
	}
	if err := changeAdminPassword(db, email, password); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Admin password changed successfully; old sessions were revoked.")
}

// promptPassword hides input when the script is attached to a terminal. When
// stdin is piped (for automation), it safely falls back to reading one line.
func promptPassword(prompt string, readers ...*bufio.Reader) (string, error) {
	fmt.Print(prompt)
	if terminal.IsTerminal(int(os.Stdin.Fd())) {
		value, err := terminal.ReadPassword(int(os.Stdin.Fd()))
		fmt.Println()
		return string(value), err
	}

	reader := (*bufio.Reader)(nil)
	if len(readers) > 0 {
		reader = readers[0]
	}
	if reader == nil {
		reader = bufio.NewReader(os.Stdin)
	}
	line, err := reader.ReadString('\n')
	if err != nil && len(line) == 0 {
		return "", err
	}
	return strings.TrimRight(line, "\r\n"), nil
}

func changeAdminPassword(db *sqlx.DB, email, password string) error {
	email = normalizeEmail(email)
	if !isAdminEmail(email) {
		return fmt.Errorf("%s is not configured as an administrator; set PARTFLOW_ADMIN_EMAILS first (configured: %s)", email, strings.Join(configuredAdminEmails(), ", "))
	}
	if err := changePassword(db, email, password); err != nil {
		return err
	}
	return nil
}

func showInteractiveStatus(db *sqlx.DB, reader *bufio.Reader) {
	fmt.Print("البريد الإلكتروني للحساب: ")
	email, _ := reader.ReadString('\n')
	if err := printAccountStatus(db, strings.TrimSpace(email)); err != nil {
		fmt.Println(err)
		return
	}
}

func changePassword(db *sqlx.DB, email, password string) error {
	email = normalizeEmail(email)
	if strings.TrimSpace(email) == "" {
		return errors.New("البريد الإلكتروني مطلوب")
	}
	if strings.TrimSpace(password) == "" {
		return errors.New("كلمة المرور الجديدة مطلوبة")
	}

	if len([]rune(password)) < 6 {
		return errors.New("password must be at least 6 characters")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("فشل في تشفير كلمة المرور: %w", err)
	}

	result, err := db.Exec(`
		UPDATE users
		SET password_hash = $1,
		    updated_at = NOW()
		WHERE email = $2
	`, string(hashed), email)
	if err != nil {
		return fmt.Errorf("فشل في تحديث كلمة المرور: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("فشل في قراءة نتيجة التحديث: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("الحساب غير موجود: %s", email)
	}

	if err := revokeRefreshTokens(db, email); err != nil {
		return err
	}

	return nil
}

func revokeRefreshTokens(db *sqlx.DB, email string) error {
	_, err := db.Exec(`
		DELETE FROM refresh_tokens
		WHERE user_id = (SELECT id FROM users WHERE email = $1 LIMIT 1)
	`, normalizeEmail(email))
	if err != nil && !isMissingRefreshTokenTable(err) {
		return fmt.Errorf("failed to revoke old sessions: %w", err)
	}
	return nil
}

func isMissingRefreshTokenTable(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "refresh_tokens") &&
		(strings.Contains(message, "does not exist") || strings.Contains(message, "no such table"))
}

func validateRequired(name, value string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("%s مطلوب", name)
	}
	return nil
}

func validatePasswordStrength(password string) error {
	if len([]rune(strings.TrimSpace(password))) < 6 {
		return errors.New("password must be at least 6 characters")
	}
	return nil
}

func formatDaysRemaining(expiresAt time.Time) string {
	duration := time.Until(expiresAt)
	if duration < 0 {
		return "منتهي"
	}
	return fmt.Sprintf("%d يوم", int(duration.Hours()/24))
}

func printAccountSummary(email, action string, expiresAt time.Time, status string) {
	fmt.Printf("\n📌 الإجراء: %s\n", action)
	fmt.Printf("البريد: %s\n", email)
	fmt.Printf("الحالة: %s\n", status)
	fmt.Printf("تاريخ الانتهاء: %s\n", expiresAt.Format(time.RFC3339))
}

func fatalError(err error) {
	fmt.Printf("\n❌ خطأ: %v\n\n", err)
	os.Exit(1)
}
