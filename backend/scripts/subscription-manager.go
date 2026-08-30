package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
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

func main() {
	createCmd := flag.NewFlagSet("create", flag.ContinueOnError)
	createEmail := createCmd.String("email", "", "البريد الإلكتروني للحساب")
	createPassword := createCmd.String("password", "", "كلمة المرور للحساب")
	createFirstName := createCmd.String("first-name", "", "الاسم الأول")
	createLastName := createCmd.String("last-name", "", "اسم العائلة")
	createPhone := createCmd.String("phone", "", "رقم الهاتف")
	createDays := createCmd.Int("days", 30, "عدد أيام الاشتراك")

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
		printUsageAndExit(createCmd, renewCmd, disableCmd, deleteCmd, listCmd, subscribersCmd, statusCmd, summaryCmd)
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
	case "create":
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
		if err := createAccount(db, *createEmail, *createPassword, *createFirstName, *createLastName, *createPhone, *createDays); err != nil {
			fatalError(err)
		}
		fmt.Println("✅ تم إنشاء الحساب بنجاح.")
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
	case "help", "-h", "--help":
		printUsageAndExit(createCmd, renewCmd, disableCmd, deleteCmd, listCmd, subscribersCmd, statusCmd, summaryCmd)
	default:
		printUsageAndExit(createCmd, renewCmd, disableCmd, deleteCmd, listCmd, subscribersCmd, statusCmd, summaryCmd)
	}
}

func loadConfig() (*Config, error) {
	dbURL := os.Getenv("DATABASE_URL")
	if strings.TrimSpace(dbURL) == "" {
		return nil, errors.New("المتغير DATABASE_URL مطلوب. مثال: export DATABASE_URL='postgres://postgres:password@localhost:5432/partflow?sslmode=disable'")
	}
	return &Config{DBURL: dbURL}, nil
}

func createAccount(db *sqlx.DB, email, password, firstName, lastName, phone string, days int) error {
	if strings.TrimSpace(email) == "" {
		return errors.New("البريد الإلكتروني مطلوب")
	}
	if strings.TrimSpace(password) == "" {
		return errors.New("كلمة المرور مطلوبة")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("فشل في تشفير كلمة المرور: %w", err)
	}

	var userID string
	err = db.Get(&userID, "SELECT id FROM users WHERE email = $1 LIMIT 1", strings.ToLower(email))
	if err == nil {
		return fmt.Errorf("الحساب موجود بالفعل: %s", email)
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
	`, accountID, strings.ToLower(email), string(hashed), firstName, lastName, phone, expiresAt)
	if err != nil {
		return fmt.Errorf("فشل في إدخال الحساب: %w", err)
	}

	printAccountSummary(email, "تم الإنشاء", expiresAt, "نشط")
	return nil
}

func renewAccount(db *sqlx.DB, email string, days int) error {
	var account Account
	err := db.Get(&account, `
		SELECT id, email, first_name, last_name, phone, is_active,
		       subscription_status, subscription_expires_at, created_at, updated_at
		FROM users WHERE email = $1 LIMIT 1
	`, strings.ToLower(email))
	if err != nil {
		return fmt.Errorf("account not found: %s", email)
	}

	newExpiry := time.Now().AddDate(0, 0, days)
	if account.SubscriptionExpiresAt != nil && account.SubscriptionExpiresAt.After(time.Now()) {
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
	fmt.Println("--------------------------------------------------------------------------------------------------------------------------------")
	fmt.Printf("%-28s %-12s %-12s %-18s %-16s %-12s\n", "البريد", "الحالة", "الاشتراك", "تاريخ الانتهاء", "الأيام المتبقية", "الاسم")
	fmt.Println("--------------------------------------------------------------------------------------------------------------------------------")
	for _, account := range accounts {
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
		fmt.Printf("%-28s %-12s %-12s %-18s %-16s %-12s\n", account.Email, status, account.SubscriptionStatus, expires, daysLeft, fullName)
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
	for _, account := range accounts {
		expires := "غير محدد"
		if account.SubscriptionExpiresAt != nil {
			expires = account.SubscriptionExpiresAt.Format(time.RFC3339)
		}
		fmt.Printf("%-28s %-14s %-18s %-20s\n", account.Email, account.SubscriptionStatus, account.SubscriptionStatus, expires)
	}
	fmt.Println("--------------------------------------------------------------------")
	fmt.Printf("إجمالي المشتركين النشطين: %d\n", len(accounts))
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
	var account Account
	err := db.Get(&account, `
		SELECT id, email, first_name, last_name, phone, is_active,
		       subscription_status, subscription_expires_at, created_at, updated_at
		FROM users WHERE email = $1 LIMIT 1
	`, strings.ToLower(email))
	if err != nil {
		return fmt.Errorf("الحساب غير موجود: %s", email)
	}

	fmt.Println("\n👤 حالة الحساب")
	fmt.Println("------------------------------------------------------------")
	fmt.Printf("البريد: %s\n", account.Email)
	fmt.Printf("الاسم: %s %s\n", account.FirstName, account.LastName)
	fmt.Printf("الهاتف: %s\n", account.Phone)
	fmt.Printf("نشط: %t\n", account.IsActive)
	fmt.Printf("حالة الاشتراك: %s\n", account.SubscriptionStatus)
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

func printUsageAndExit(createCmd, renewCmd, disableCmd, deleteCmd, listCmd, subscribersCmd, statusCmd, summaryCmd *flag.FlagSet) {
	fmt.Println("الاستخدام:")
	fmt.Println("  go run ./scripts/subscription-manager.go app")
	fmt.Println("  go run ./scripts/subscription-manager.go summary")
	fmt.Println("  go run ./scripts/subscription-manager.go create --email user@example.com --password Pass123 --first-name علي --last-name السعدي --days 30")
	fmt.Println("  go run ./scripts/subscription-manager.go renew --email user@example.com --days 60")
	fmt.Println("  go run ./scripts/subscription-manager.go disable --email user@example.com")
	fmt.Println("  go run ./scripts/subscription-manager.go delete --email user@example.com")
	fmt.Println("  go run ./scripts/subscription-manager.go list")
	fmt.Println("  go run ./scripts/subscription-manager.go subscribers")
	fmt.Println("  go run ./scripts/subscription-manager.go status --email user@example.com")
	fmt.Println("")
	fmt.Println("الخيارات:")
	createCmd.PrintDefaults()
	renewCmd.PrintDefaults()
	disableCmd.PrintDefaults()
	deleteCmd.PrintDefaults()
	listCmd.PrintDefaults()
	subscribersCmd.PrintDefaults()
	statusCmd.PrintDefaults()
	summaryCmd.PrintDefaults()
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
	for {
		fmt.Println("\n========================================")
		fmt.Println("   لوحة إدارة اشتراكات PartFlow")
		fmt.Println("========================================")
		fmt.Println("1) عرض جميع الحسابات")
		fmt.Println("2) عرض المشتركين النشطين")
		fmt.Println("3) إنشاء حساب جديد")
		fmt.Println("4) تجديد اشتراك")
		fmt.Println("5) إيقاف حساب")
		fmt.Println("6) حذف حساب")
		fmt.Println("7) تغيير كلمة المرور")
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
			changeInteractivePassword(db, reader)
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
	fmt.Print("كلمة المرور: ")
	password, _ := reader.ReadString('\n')
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

func changeInteractivePassword(db *sqlx.DB, reader *bufio.Reader) {
	fmt.Print("البريد الإلكتروني للحساب: ")
	email, _ := reader.ReadString('\n')
	fmt.Print("كلمة المرور الجديدة: ")
	password, _ := reader.ReadString('\n')
	if err := changePassword(db, strings.TrimSpace(email), strings.TrimSpace(password)); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("تم تغيير كلمة المرور بنجاح.")
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
	if strings.TrimSpace(email) == "" {
		return errors.New("البريد الإلكتروني مطلوب")
	}
	if strings.TrimSpace(password) == "" {
		return errors.New("كلمة المرور الجديدة مطلوبة")
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
	`, string(hashed), strings.ToLower(email))
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

	return nil
}

func validateRequired(name, value string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("%s مطلوب", name)
	}
	return nil
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
