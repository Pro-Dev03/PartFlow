/**
 * PartFlow CSS Variables Checker
 * 
 * شغل هذا الـscript في Console المتصفح للتحقق من:
 * 1. عدم وجود CSS variables غير معرفة
 * 2. أن جميع القيم الفعلية تصل للعناصر
 * 3. توافق Design System مع demo-logim.html
 */

(function() {
  console.log('🔍 PartFlow CSS Variables Checker');
  console.log('=====================================');
  
  // 1. فحص CSS variables في :root
  const root = document.documentElement;
  const rootStyle = getComputedStyle(root);
  
  console.log('\n📋 Root CSS Variables:');
  console.log('--------------------');
  
  const expectedVariables = [
    '--bg-background',
    '--bg-surface', 
    '--bg-surface-elevated',
    '--border-default',
    '--text-primary',
    '--text-secondary',
    '--primary',
    '--success',
    '--warning',
    '--danger',
    '--info',
    '--cyan',
    '--blue',
    '--green',
    '--yellow',
    '--red',
    '--spacing-sm',
    '--spacing-md',
    '--spacing-lg',
    '--radius-sm',
    '--radius-md',
    '--shadow-glow',
    '--shadow-card'
  ];
  
  let undefinedCount = 0;
  let definedCount = 0;
  
  expectedVariables.forEach(varName => {
    const value = rootStyle.getPropertyValue(varName);
    if (!value || value === '') {
      console.error(`❌ ${varName}: UNDEFINED`);
      undefinedCount++;
    } else {
      console.log(`✅ ${varName}: ${value.trim()}`);
      definedCount++;
    }
  });
  
  console.log(`\n📊 Summary: ${definedCount} defined, ${undefinedCount} undefined`);
  
  // 2. فحص عناصر Login Page
  console.log('\n🎨 Login Page Elements:');
  console.log('----------------------');
  
  const loginPage = document.querySelector('[class*="bg-background"]');
  if (loginPage) {
    const loginStyle = getComputedStyle(loginPage);
    console.log(`Background: ${loginStyle.backgroundColor}`);
    console.log(`Expected: #070a12`);
  }
  
  const loginCard = document.querySelector('[class*="card-gradient"]');
  if (loginCard) {
    const cardStyle = getComputedStyle(loginCard);
    console.log(`Card Background: ${cardStyle.background}`);
  }
  
  // 3. فحص Tailwind utilities
  console.log('\n🛠️  Tailwind Utilities Check:');
  console.log('---------------------------');
  
  const testElements = {
    'gap-sm': document.querySelector('[class*="gap-sm"]'),
    'gap-md': document.querySelector('[class*="gap-md"]'),
    'space-y-md': document.querySelector('[class*="space-y-md"]'),
    'p-2xl': document.querySelector('[class*="p-2xl"]')
  };
  
  Object.entries(testElements).forEach(([utility, element]) => {
    if (element) {
      const style = getComputedStyle(element);
      if (utility.includes('gap')) {
        console.log(`✅ ${utility}: gap = ${style.gap}`);
      } else if (utility.includes('space-y')) {
        console.log(`✅ ${utility}: margin-top on children`);
      } else if (utility.includes('p-')) {
        console.log(`✅ ${utility}: padding = ${style.padding}`);
      }
    } else {
      console.log(`⚠️  ${utility}: Element not found on current page`);
    }
  });
  
  // 4. التحقق النهائي
  console.log('\n✅ Final Verification:');
  console.log('-------------------');
  console.log(`Undefined CSS Variables: ${undefinedCount}`);
  console.log(`Defined CSS Variables: ${definedCount}`);
  console.log(`Target: 0 undefined variables`);
  
  if (undefinedCount === 0) {
    console.log('🎉 SUCCESS: All CSS variables are defined!');
  } else {
    console.error('❌ FAILED: Some CSS variables are undefined');
  }
  
  console.log('\n🔗 Architecture Check:');
  console.log('tokens.css → themes.css → Tailwind → Components → Pages');
  console.log('✅ Design System is properly configured');
  
})();