-- Create part_types table and related tables for the part types system
BEGIN;

-- Create part_types table
CREATE TABLE IF NOT EXISTS part_types (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name_ar VARCHAR(255) NOT NULL,
    name_en VARCHAR(255) NOT NULL,
    icon VARCHAR(100),
    color VARCHAR(7),
    is_active BOOLEAN DEFAULT true,
    sort_order INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_part_types_active ON part_types(is_active);
CREATE INDEX IF NOT EXISTS idx_part_types_sort ON part_types(sort_order);

-- Create part_specifications table
CREATE TABLE IF NOT EXISTS part_specifications (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name_ar VARCHAR(255) NOT NULL,
    name_en VARCHAR(255) NOT NULL,
    data_type VARCHAR(50) NOT NULL, -- text, number, select, boolean
    options TEXT[] DEFAULT '{}',
    is_required BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_part_specifications_type ON part_specifications(data_type);

-- Create type_specifications junction table
CREATE TABLE IF NOT EXISTS type_specifications (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    part_type_id UUID NOT NULL REFERENCES part_types(id) ON DELETE CASCADE,
    specification_id UUID NOT NULL REFERENCES part_specifications(id) ON DELETE CASCADE,
    sort_order INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(part_type_id, specification_id)
);

CREATE INDEX IF NOT EXISTS idx_type_specifications_part_type ON type_specifications(part_type_id);
CREATE INDEX IF NOT EXISTS idx_type_specifications_spec ON type_specifications(specification_id);

-- Create item_specification_values table
CREATE TABLE IF NOT EXISTS item_specification_values (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    inventory_item_id UUID NOT NULL REFERENCES inventory(id) ON DELETE CASCADE,
    specification_id UUID NOT NULL REFERENCES part_specifications(id) ON DELETE CASCADE,
    value_text TEXT,
    value_number NUMERIC,
    value_boolean BOOLEAN,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(inventory_item_id, specification_id)
);

CREATE INDEX IF NOT EXISTS idx_item_spec_values_item ON item_specification_values(inventory_item_id);
CREATE INDEX IF NOT EXISTS idx_item_spec_values_spec ON item_specification_values(specification_id);

-- Insert default part types
INSERT INTO part_types (name_ar, name_en, icon, color, sort_order) VALUES
('معالج', 'Processor', 'cpu', '#3B82F6', 1),
('كرت شاشة', 'Graphics Card', 'gpu', '#10B981', 2),
('ذاكرة عشوائية', 'RAM', 'memory', '#F59E0B', 3),
('تخزين', 'Storage', 'hdd', '#EF4444', 4),
('لوحة أم', 'Motherboard', 'motherboard', '#8B5CF6', 5),
('مصدر طاقة', 'Power Supply', 'power', '#EC4899', 6),
('تبريد', 'Cooling', 'fan', '#06B6D4', 7),
('حالة', 'Case', 'case', '#6B7280', 8)
ON CONFLICT DO NOTHING;

COMMIT;