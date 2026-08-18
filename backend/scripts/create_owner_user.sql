-- Script to create a new owner user (highest privilege role)
-- This script creates a user with full administrative access

-- Step 1: Check existing organizations
DO $$
DECLARE
    org_id UUID;
    owner_role_id UUID;
    user_exists INTEGER;
BEGIN
    -- Get the first organization (or create one if none exists)
    SELECT id INTO org_id FROM organizations LIMIT 1;
    
    IF org_id IS NULL THEN
        -- Create a default organization if none exists
        INSERT INTO organizations (name, slug, email, phone, address, city, country, subscription_plan, subscription_status)
        VALUES ('Default Organization', 'default-org', 'contact@partflow.com', '+1234567890', 'Default Address', 'Default City', 'Default Country', 'free', 'active')
        RETURNING id INTO org_id;
        
        RAISE NOTICE 'Created default organization with ID: %', org_id;
    ELSE
        RAISE NOTICE 'Using existing organization with ID: %', org_id;
    END IF;
    
    -- Step 2: Check if owner role exists, if not create it
    SELECT id INTO owner_role_id FROM roles WHERE name = 'owner' AND organization_id = org_id LIMIT 1;
    
    IF owner_role_id IS NULL THEN
        -- Create owner role with all permissions (as JSONB array)
        INSERT INTO roles (organization_id, name, description, permissions, is_system)
        VALUES (
            org_id,
            'owner',
            'Full access to all features',
            '["products.read", "products.create", "products.update", "products.delete", "products.archive",
                "inventory.read", "inventory.adjust", "inventory.transfer", "inventory.inspect",
                "sales.read", "sales.create", "sales.cancel", "sales.refund",
                "customers.read", "customers.create", "customers.update", "customers.delete",
                "debts.read", "debts.manage",
                "suppliers.read", "suppliers.create", "suppliers.update", "suppliers.delete",
                "purchases.read", "purchases.create", "purchases.receive",
                "expenses.read", "expenses.create", "expenses.update", "expenses.delete",
                "returns.read", "returns.create", "returns.approve",
                "warranties.read", "warranties.claim",
                "reports.read", "reports.export",
                "users.read", "users.create", "users.update", "users.delete",
                "settings.manage",
                "audit.read"]'::JSONB,
            true
        )
        RETURNING id INTO owner_role_id;
        
        RAISE NOTICE 'Created owner role with ID: %', owner_role_id;
    ELSE
        RAISE NOTICE 'Using existing owner role with ID: %', owner_role_id;
    END IF;
    
    -- Step 3: Check if user already exists
    SELECT COUNT(*) INTO user_exists FROM users WHERE email = 'owner@partflow.com';
    
    IF user_exists > 0 THEN
        RAISE NOTICE 'User owner@partflow.com already exists. Skipping creation.';
    ELSE
        -- Step 4: Create the owner user
        -- Password: OwnerPass123! (change this in production)
        INSERT INTO users (
            id,
            organization_id,
            email,
            password_hash,
            first_name,
            last_name,
            phone,
            role_id,
            is_active,
            is_verified,
            created_at,
            updated_at
        )
        SELECT
            gen_random_uuid(),
            org_id,
            'owner@partflow.com',
            crypt('OwnerPass123!', gen_salt('bf', 12)),  -- Password: OwnerPass123!
            'System',
            'Owner',
            '+1234567890',
            owner_role_id,
            true,
            true,
            NOW(),
            NOW();
            
        RAISE NOTICE 'Successfully created owner user: owner@partflow.com';
        RAISE NOTICE 'Password: OwnerPass123!';
        RAISE NOTICE 'IMPORTANT: Change this password in production!';
    END IF;
    
END $$;

-- Step 5: Verify the created user
SELECT 
    u.id,
    u.email,
    u.first_name,
    u.last_name,
    u.is_active,
    u.is_verified,
    r.name as role_name,
    r.description as role_description,
    o.name as organization_name
FROM users u
JOIN roles r ON u.role_id = r.id
JOIN organizations o ON u.organization_id = o.id
WHERE u.email = 'owner@partflow.com';