-- Update inspections table to support template-based inspections (USED-PARTS-ACQUISITION.md)

-- Add new columns to inspections table
ALTER TABLE inspections 
ADD COLUMN IF NOT EXISTS template_type VARCHAR(50),
ADD COLUMN IF NOT EXISTS checkpoint_results JSONB DEFAULT '{}',
ADD COLUMN IF NOT EXISTS acquisition_item_id UUID REFERENCES acquisition_items(id);

-- Create index for template type
CREATE INDEX IF NOT EXISTS idx_inspections_template_type ON inspections(template_type);
CREATE INDEX IF NOT EXISTS idx_inspections_acquisition_item ON inspections(acquisition_item_id);

-- Update existing inspections to use default template type if needed
UPDATE inspections 
SET template_type = 'general' 
WHERE template_type IS NULL;