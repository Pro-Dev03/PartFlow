package inspections

import (
	"encoding/json"
)

// InspectionTemplate represents a template for inspecting a specific type of part
// Based on USED-PARTS-ACQUISITION.md - Section 9: قوالب الفحص حسب نوع القطعة
type InspectionTemplate struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`        // GPU, CPU, Laptop, Motherboard, etc.
	DisplayName string                 `json:"display_name"` // Display name in Arabic
	Category    string                 `json:"category"`    // GPU, CPU, Storage, etc.
	Checkpoints []InspectionCheckpoint `json:"checkpoints"`
}

// InspectionCheckpoint represents a single inspection checkpoint
type InspectionCheckpoint struct {
	ID          string `json:"id"`
	Name        string `json:"name"`        // Arabic name
	NameEn      string `json:"name_en"`     // English name
	Description string `json:"description"` // Description of what to check
	Required    bool   `json:"required"`    // Is this checkpoint required?
	Category    string `json:"category"`    // physical, functional, performance, etc.
}

// Predefined inspection templates for different part types
var InspectionTemplates = map[string]InspectionTemplate{
	"gpu": {
		ID:          "gpu",
		Name:        "GPU",
		DisplayName: "كارت شاشة",
		Category:    "GPU",
		Checkpoints: []InspectionCheckpoint{
			{
				ID:          "gpu_power_on",
				Name:        "التشغيل",
				NameEn:      "Power On",
				Description: "هل يبدأ الكارت بشكل طبيعي؟",
				Required:    true,
				Category:    "functional",
			},
			{
				ID:          "gpu_vram",
				Name:        "ذاكرة الفيديو",
				NameEn:      "VRAM",
				Description: "فحص ذاكرة الفيديو باستخدام أدوات مثل MemTestGPU",
				Required:    true,
				Category:    "functional",
			},
			{
				ID:          "gpu_temperature",
				Name:        "الحرارة",
				NameEn:      "Temperature",
				Description: "فحص حرارة الكارت تحت الحمل (should be < 85°C)",
				Required:    true,
				Category:    "performance",
			},
			{
				ID:          "gpu_fans",
				Name:        "المراوح",
				NameEn:      "Fans",
				Description: "هل المراوح تعمل بشكل صحيح؟",
				Required:    true,
				Category:    "physical",
			},
			{
				ID:          "gpu_display_outputs",
				Name:        "مخارج العرض",
				NameEn:      "Display Outputs",
				Description: "فحص جميع مخارج العرض (HDMI, DisplayPort, DVI)",
				Required:    true,
				Category:    "functional",
			},
			{
				ID:          "gpu_physical_condition",
				Name:        "الحالة الخارجية",
				NameEn:      "Physical Condition",
				Description: "فحص الحالة الخارجية (خدوش، تلف، أوساخ)",
				Required:    false,
				Category:    "physical",
			},
		},
	},
	"cpu": {
		ID:          "cpu",
		Name:        "CPU",
		DisplayName: "معالج",
		Category:    "CPU",
		Checkpoints: []InspectionCheckpoint{
			{
				ID:          "cpu_post",
				Name:        "POST",
				NameEn:      "POST",
				Description: "هل يمر المعالج من POST بنجاح؟",
				Required:    true,
				Category:    "functional",
			},
			{
				ID:          "cpu_stability",
				Name:        "الاستقرار",
				NameEn:      "Stability",
				Description: "فحص الاستقرار باستخدام Prime95 أو أداة مشابهة",
				Required:    true,
				Category:    "performance",
			},
			{
				ID:          "cpu_temperature",
				Name:        "الحرارة",
				NameEn:      "Temperature",
				Description: "فحص حرارة المعالج تحت الحمل",
				Required:    true,
				Category:    "performance",
			},
			{
				ID:          "cpu_pins",
				Name:        "الإبر/الموصلات",
				NameEn:      "Pins/Contacts",
				Description: "فحص حالة الإبر (لـ Desktop) أو الموصلات (لـ Laptop)",
				Required:    true,
				Category:    "physical",
			},
		},
	},
	"laptop": {
		ID:          "laptop",
		Name:        "Laptop",
		DisplayName: "لابتوب",
		Category:    "Laptop",
		Checkpoints: []InspectionCheckpoint{
			{
				ID:          "laptop_screen",
				Name:        "الشاشة",
				NameEn:      "Screen",
				Description: "فحص الشاشة (dead pixels, backlight, cracks)",
				Required:    true,
				Category:    "physical",
			},
			{
				ID:          "laptop_battery",
				Name:        "البطارية",
				NameEn:      "Battery",
				Description: "فحص صحة البطارية وسعة الشحن",
				Required:    true,
				Category:    "functional",
			},
			{
				ID:          "laptop_keyboard",
				Name:        "لوحة المفاتيح",
				NameEn:      "Keyboard",
				Description: "فحص جميع المفاتيح",
				Required:    true,
				Category:    "physical",
			},
			{
				ID:          "laptop_touchpad",
				Name:        "التشباد",
				NameEn:      "Touchpad",
				Description: "فحص التشباد",
				Required:    true,
				Category:    "functional",
			},
			{
				ID:          "laptop_usb",
				Name:        "منافذ USB",
				NameEn:      "USB Ports",
				Description: "فحص جميع منافذ USB",
				Required:    true,
				Category:    "functional",
			},
			{
				ID:          "laptop_wifi",
				Name:        "الواي فاي",
				NameEn:      "Wi-Fi",
				Description: "فحص كارت الواي فاي",
				Required:    true,
				Category:    "functional",
			},
			{
				ID:          "laptop_temperature",
				Name:        "الحرارة",
				NameEn:      "Temperature",
				Description: "فحص حرارة المعالج وكارت الشاشة",
				Required:    true,
				Category:    "performance",
			},
			{
				ID:          "laptop_charger",
				Name:        "الشاحن",
				NameEn:      "Charger",
				Description: "فحص الشاحن ومنفذ الشحن",
				Required:    true,
				Category:    "functional",
			},
		},
	},
	"motherboard": {
		ID:          "motherboard",
		Name:        "Motherboard",
		DisplayName: "بوردة",
		Category:    "Motherboard",
		Checkpoints: []InspectionCheckpoint{
			{
				ID:          "mb_post",
				Name:        "POST",
				NameEn:      "POST",
				Description: "هل تمر البوردة من POST بنجاح؟",
				Required:    true,
				Category:    "functional",
			},
			{
				ID:          "mb_ram_slots",
				Name:        "فتحات الرام",
				NameEn:      "RAM Slots",
				Description: "فحص جميع فتحات الرام",
				Required:    true,
				Category:    "functional",
			},
			{
				ID:          "mb_usb",
				Name:        "منافذ USB",
				NameEn:      "USB Ports",
				Description: "فحص جميع منافذ USB",
				Required:    true,
				Category:    "functional",
			},
			{
				ID:          "mb_pcie",
				Name:        "فتحات PCIe",
				NameEn:      "PCIe Slots",
				Description: "فحص فتحات PCIe",
				Required:    true,
				Category:    "functional",
			},
			{
				ID:          "mb_storage_interfaces",
				Name:        "واجهات التخزين",
				NameEn:      "Storage Interfaces",
				Description: "فحص واجهات SATA و NVMe",
				Required:    true,
				Category:    "functional",
			},
		},
	},
	"ram": {
		ID:          "ram",
		Name:        "RAM",
		DisplayName: "ذاكرة RAM",
		Category:    "Memory",
		Checkpoints: []InspectionCheckpoint{
			{
				ID:          "ram_detection",
				Name:        "الاكتشاف",
				NameEn:      "Detection",
				Description: "هل يتم اكتشاف الرام في BIOS؟",
				Required:    true,
				Category:    "functional",
			},
			{
				ID:          "ram_capacity",
				Name:        "السعة",
				NameEn:      "Capacity",
				Description: "هل السعة صحيحة؟",
				Required:    true,
				Category:    "functional",
			},
			{
				ID:          "ram_memtest",
				Name:        "فحص MemTest",
				NameEn:      "MemTest",
				Description: "فحص الرام باستخدام MemTest86",
				Required:    true,
				Category:    "performance",
			},
			{
				ID:          "ram_physical",
				Name:        "الحالة الخارجية",
				NameEn:      "Physical Condition",
				Description: "فحص حالة القطع المعدنية",
				Required:    false,
				Category:    "physical",
			},
		},
	},
	"storage": {
		ID:          "storage",
		Name:        "Storage",
		DisplayName: "تخزين (SSD/HDD)",
		Category:    "Storage",
		Checkpoints: []InspectionCheckpoint{
			{
				ID:          "storage_detection",
				Name:        "الاكتشاف",
				NameEn:      "Detection",
				Description: "هل يتم اكتشاف القرص؟",
				Required:    true,
				Category:    "functional",
			},
			{
				ID:          "storage_capacity",
				Name:        "السعة",
				NameEn:      "Capacity",
				Description: "هل السعة صحيحة؟",
				Required:    true,
				Category:    "functional",
			},
			{
				ID:          "storage_smart",
				Name:        "فحص SMART",
				NameEn:      "SMART Test",
				Description: "فحص حالة SMART",
				Required:    true,
				Category:    "performance",
			},
			{
				ID:          "storage_speed",
				Name:        "السرعة",
				NameEn:      "Speed",
				Description: "فحص سرعة القراءة والكتابة",
				Required:    false,
				Category:    "performance",
			},
			{
				ID:          "storage_bad_sectors",
				Name:        "القطع التالفة",
				NameEn:      "Bad Sectors",
				Description: "فحص وجود قطع تالفة",
				Required:    true,
				Category:    "performance",
			},
		},
	},
	"psu": {
		ID:          "psu",
		Name:        "PSU",
		DisplayName: "مزود طاقة",
		Category:    "Power",
		Checkpoints: []InspectionCheckpoint{
			{
				ID:          "psu_power_on",
				Name:        "التشغيل",
				NameEn:      "Power On",
				Description: "هل يعمل مزود الطاقة؟",
				Required:    true,
				Category:    "functional",
			},
			{
				ID:          "psu_voltage",
				Name:        "الجهد",
				NameEn:      "Voltage",
				Description: "فحص جهد التيار (+12V, +5V, +3.3V)",
				Required:    true,
				Category:    "performance",
			},
			{
				ID:          "psu_load",
				Name:        "الحمل",
				NameEn:      "Load Test",
				Description: "فحص الاستقرار تحت الحمل",
				Required:    true,
				Category:    "performance",
			},
			{
				ID:          "psu_cables",
				Name:        "الكابلات",
				NameEn:      "Cables",
				Description: "فحص حالة الكابلات والموصلات",
				Required:    false,
				Category:    "physical",
			},
		},
	},
}

// GetTemplate retrieves an inspection template by type
func GetTemplate(templateType string) (*InspectionTemplate, bool) {
	template, exists := InspectionTemplates[templateType]
	return &template, exists
}

// GetAllTemplates retrieves all inspection templates
func GetAllTemplates() []InspectionTemplate {
	templates := make([]InspectionTemplate, 0, len(InspectionTemplates))
	for _, template := range InspectionTemplates {
		templates = append(templates, template)
	}
	return templates
}

// GetTemplateForProduct determines the appropriate template based on product category
func GetTemplateForProduct(category string) (*InspectionTemplate, bool) {
	// Map product categories to inspection templates
	categoryMap := map[string]string{
		"GPU":        "gpu",
		"Graphics":   "gpu",
		"CPU":        "cpu",
		"Processor":  "cpu",
		"Laptop":     "laptop",
		"Notebook":   "laptop",
		"Motherboard": "motherboard",
		"RAM":        "ram",
		"Memory":     "ram",
		"SSD":        "storage",
		"HDD":        "storage",
		"Storage":    "storage",
		"PSU":        "psu",
		"Power":      "psu",
	}

	templateType, exists := categoryMap[category]
	if !exists {
		return nil, false
	}

	return GetTemplate(templateType)
}

// CheckpointResult represents the result of a single checkpoint inspection
type CheckpointResult struct {
	CheckpointID string `json:"checkpoint_id"`
	Status      string `json:"status"`      // pass, fail, pending
	Notes       string `json:"notes"`
	Images      []string `json:"images"`
}

// InspectionFromTemplate creates an inspection from a template with results
type InspectionFromTemplate struct {
	TemplateID     string              `json:"template_id"`
	TemplateType   string              `json:"template_type"`
	CheckpointResults []CheckpointResult `json:"checkpoint_results"`
	OverallStatus  string              `json:"overall_status"` // passed, failed, needs_repair
	Notes          string              `json:"notes"`
}

// CalculateOverallStatus calculates the overall inspection status based on checkpoint results
func CalculateOverallStatus(template InspectionTemplate, results []CheckpointResult) string {
	requiredMap := make(map[string]bool)
	for _, checkpoint := range template.Checkpoints {
		if checkpoint.Required {
			requiredMap[checkpoint.ID] = true
		}
	}

	// Check if any required checkpoints failed
	for _, result := range results {
		if requiredMap[result.CheckpointID] && result.Status == "fail" {
			return "failed"
		}
	}

	// Check if any required checkpoints are still pending
	for _, result := range results {
		if requiredMap[result.CheckpointID] && result.Status == "pending" {
			return "needs_repair"
		}
	}

	// All required checkpoints passed
	return "passed"
}

// MarshalCheckpointResults converts checkpoint results to JSON for storage
func MarshalCheckpointResults(results []CheckpointResult) (string, error) {
	data, err := json.Marshal(results)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// UnmarshalCheckpointResults converts JSON string to checkpoint results
func UnmarshalCheckpointResults(data string) ([]CheckpointResult, error) {
	var results []CheckpointResult
	err := json.Unmarshal([]byte(data), &results)
	if err != nil {
		return nil, err
	}
	return results, nil
}