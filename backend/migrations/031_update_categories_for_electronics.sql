-- Update categories for electronics store (instead of car parts)
-- Migration: 031_update_categories_for_electronics.sql
-- Date: 2026-08-26

-- First, delete existing categories (or you can update them if you prefer)
DELETE FROM categories WHERE name IN (
    'أجزاء محرك',
    'إطارات',
    'تبريد',
    'تخزين',
    'زيوت',
    'كروت شاشة',
    'كهرباء',
    'لوحات أم',
    'مصادر طاقة',
    'معالجات'
);

-- Insert new electronics-specific categories
INSERT INTO categories (id, name, description, icon, color, is_active, created_at, updated_at) VALUES
    (gen_random_uuid(), 'هواتف', 'هواتف ذكية وأجهزة لوحية', 'smartphone', '#3b82f6', true, NOW(), NOW()),
    (gen_random_uuid(), 'لابتوب', 'أجهزة الكمبيوتر المحمولة', 'laptop', '#8b5cf6', true, NOW(), NOW()),
    (gen_random_uuid(), 'كمبيوتر مكتبي', 'أجهزة الكمبيوتر المكتبي', 'monitor', '#10b981', true, NOW(), NOW()),
    (gen_random_uuid(), 'قطع الكمبيوتر', 'معالجات، رام، كروت شاشة، لوحات أم', 'cpu', '#f59e0b', true, NOW(), NOW()),
    (gen_random_uuid(), 'تخزين', 'هارد ديسك، SSD، فلاشات', 'hard-drive', '#ef4444', true, NOW(), NOW()),
    (gen_random_uuid(), 'شاشات', 'شاشات الكمبيوتر والتلفزيون', 'monitor', '#06b6d4', true, NOW(), NOW()),
    (gen_random_uuid(), 'كاميرات', 'كاميرات رقمية وكاميرات أمنية', 'camera', '#ec4899', true, NOW(), NOW()),
    (gen_random_uuid(), 'طابعات', 'طابعات وماسحات ضوئية', 'printer', '#6366f1', true, NOW(), NOW()),
    (gen_random_uuid(), 'شبكات', 'راوترات، مودمات، كابلات', 'wifi', '#14b8a6', true, NOW(), NOW()),
    (gen_random_uuid(), 'إكسسوارات', 'سماعات، كيبورد، ماوس، شواحن', 'headphones', '#f97316', true, NOW(), NOW()),
    (gen_random_uuid(), 'صوتيات', 'مكبرات صوت وأنظمة صوتية', 'speaker', '#a855f7', true, NOW(), NOW()),
    (gen_random_uuid(), 'كيبلات', 'كابلات ووصلات متنوعة', 'cable', '#64748b', true, NOW(), NOW())
ON CONFLICT (name) DO UPDATE SET
    description = EXCLUDED.description,
    icon = EXCLUDED.icon,
    color = EXCLUDED.color,
    updated_at = NOW();
