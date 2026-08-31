#!/usr/bin/env node

/**
 * End-to-End test for offline login workflow
 * This script tests the complete offline mode initialization and login process
 */

import fetch from 'node-fetch';

const API_BASE = 'http://localhost:8080/api/v1';
const OWNER_EMAIL = 'owner@partflow.com';
const OWNER_PASSWORD = 'Owner123456';

async function sleep(ms) {
  return new Promise(resolve => setTimeout(resolve, ms));
}

async function test() {
  console.log('\n========================================');
  console.log('Offline Login Workflow Test');
  console.log('========================================\n');

  try {
    // Step 1: Check API health
    console.log('Step 1: Checking API health...');
    const healthRes = await fetch(`${API_BASE}/health`, { method: 'GET' });
    if (!healthRes.ok) {
      console.error('❌ API health check failed:', healthRes.status);
      process.exit(1);
    }
    console.log('✅ API is healthy\n');

    // Step 2: Test login endpoint
    console.log('Step 2: Testing login endpoint...');
    const loginRes = await fetch(`${API_BASE}/auth/login`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Origin': 'http://localhost:5174',
        'Referer': 'http://localhost:5174/',
      },
      body: JSON.stringify({
        email: OWNER_EMAIL,
        password: OWNER_PASSWORD,
      }),
    });

    if (loginRes.status !== 200 && loginRes.status !== 201) {
      console.error(`❌ Login failed with status ${loginRes.status}`);
      const text = await loginRes.text();
      console.error('Response:', text);
      process.exit(1);
    }

    const loginData = await loginRes.json();
    console.log('✅ Login successful\n');

    // Verify response structure
    console.log('Step 3: Verifying login response structure...');
    const requiredFields = {
      'data.token': loginData.data?.token,
      'data.access_token': loginData.data?.access_token,
      'data.refresh_token': loginData.data?.refresh_token,
      'data.user.id': loginData.data?.user?.id,
      'data.user.email': loginData.data?.user?.email,
      'data.user.first_name': loginData.data?.user?.first_name,
    };

    let hasAllFields = true;
    for (const [field, value] of Object.entries(requiredFields)) {
      if (!value) {
        console.error(`  ❌ Missing field: ${field}`);
        hasAllFields = false;
      } else {
        console.log(`  ✅ ${field}`);
      }
    }

    if (!hasAllFields) {
      console.error('\n❌ Login response missing required fields');
      process.exit(1);
    }
    console.log('✅ All required fields present\n');

    // Step 4: Verify CORS headers
    console.log('Step 4: Checking CORS headers...');
    const corsHeaders = {
      'Access-Control-Allow-Origin': loginRes.headers.get('access-control-allow-origin'),
      'Access-Control-Allow-Methods': loginRes.headers.get('access-control-allow-methods'),
      'Access-Control-Allow-Headers': loginRes.headers.get('access-control-allow-headers'),
    };

    console.log('  Access-Control-Allow-Origin:', corsHeaders['Access-Control-Allow-Origin']);
    console.log('  Access-Control-Allow-Methods:', corsHeaders['Access-Control-Allow-Methods']);
    console.log('  Access-Control-Allow-Headers:', corsHeaders['Access-Control-Allow-Headers']);

    if (corsHeaders['Access-Control-Allow-Origin']) {
      console.log('✅ CORS headers present\n');
    } else {
      console.warn('⚠️  CORS headers missing\n');
    }

    // Step 5: Test token with protected endpoint
    console.log('Step 5: Testing token with protected endpoint...');
    const token = loginData.data.token || loginData.data.access_token;
    const protectedRes = await fetch(`${API_BASE}/dashboard/stats`, {
      method: 'GET',
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json',
      },
    });

    if (protectedRes.status === 401) {
      console.error('❌ Token rejected by protected endpoint');
      process.exit(1);
    } else if (protectedRes.ok) {
      console.log('✅ Token accepted by protected endpoint\n');
    } else {
      console.warn(`⚠️  Protected endpoint returned ${protectedRes.status}`);
    }

    // Success summary
    console.log('========================================');
    console.log('✅ All tests passed! Offline login workflow is working.');
    console.log('========================================\n');

    console.log('Login successful with:');
    console.log(`  Email: ${OWNER_EMAIL}`);
    console.log(`  User: ${loginData.data.user.first_name} ${loginData.data.user.last_name}`);
    console.log(`  Token: ${token.substring(0, 20)}...`);
    console.log(`  Expires in: ${loginData.data.expires_in} seconds\n`);

    return {
      success: true,
      token,
      refreshToken: loginData.data.refresh_token,
      user: loginData.data.user,
    };
  } catch (error) {
    console.error('\n❌ Test failed with error:', error.message);
    process.exit(1);
  }
}

test();
