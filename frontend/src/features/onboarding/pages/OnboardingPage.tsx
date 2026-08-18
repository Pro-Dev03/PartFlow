import { useState } from 'react'
import { clsx } from 'clsx'
import { OrganizationStep } from '../components/OrganizationStep'
import { UserStep } from '../components/UserStep'
import { PreferencesStep } from '../components/PreferencesStep'
import { CompletionStep } from '../components/CompletionStep'
import type { OnboardingData } from '../types/onboarding'

export function OnboardingPage() {
  const [currentStep, setCurrentStep] = useState(1)
  const [data, setData] = useState<Partial<OnboardingData>>({
    step: 1,
    organization: {
      name: '',
      type: 'computer_store',
      currency: 'ILS',
      language: 'ar',
    },
    user: {
      name: '',
      email: '',
      password: '',
    },
    preferences: {
      businessHours: {
        sunday: { open: '09:00', close: '18:00', closed: false },
        monday: { open: '09:00', close: '18:00', closed: false },
        tuesday: { open: '09:00', close: '18:00', closed: false },
        wednesday: { open: '09:00', close: '18:00', closed: false },
        thursday: { open: '09:00', close: '18:00', closed: false },
        friday: { open: '09:00', close: '14:00', closed: false },
        saturday: { open: '09:00', close: '14:00', closed: true },
      },
      lowStockThreshold: 5,
      defaultWarrantyDays: 30,
    },
    completed: false,
  })

  const totalSteps = 4

  const steps = [
    { id: 1, title: 'المؤسسة', description: 'معلومات متجرك' },
    { id: 2, title: 'الحساب', description: 'إنشاء حساب المدير' },
    { id: 3, title: 'التفضيلات', description: 'ضبط الإعدادات الأساسية' },
    { id: 4, title: 'الإنتهاء', description: 'جاهز للبدء' },
  ]

  const handleNext = () => {
    if (currentStep < totalSteps) {
      setCurrentStep(currentStep + 1)
      setData({ ...data, step: currentStep + 1 })
    }
  }

  const handleBack = () => {
    if (currentStep > 1) {
      setCurrentStep(currentStep - 1)
      setData({ ...data, step: currentStep - 1 })
    }
  }

  const handleComplete = async () => {
    try {
      // TODO: Submit onboarding data to API
      console.log('Completing onboarding:', data)
      setData({ ...data, completed: true })
      // TODO: Redirect to dashboard
    } catch (error) {
      console.error('Failed to complete onboarding:', error)
    }
  }

  const updateData = (section: keyof OnboardingData, updates: any) => {
    setData({
      ...data,
      [section]: {
        ...data[section],
        ...updates,
      },
    })
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-primary-5 to-secondary-5 flex items-center justify-center p-4">
      <div className="max-w-4xl w-full">
        {/* Header */}
        <div className="text-center mb-8">
          <h1 className="text-3xl font-bold text-text mb-2">مرحبًا بك في PartFlow</h1>
          <p className="text-muted">لنجهز متجرك خلال دقائق</p>
        </div>

        {/* Progress Steps */}
        <div className="flex items-center justify-center mb-8">
          {steps.map((step, index) => (
            <div key={step.id} className="flex items-center">
              <div className="flex flex-col items-center">
                <div
                  className={clsx(
                    'w-10 h-10 rounded-full flex items-center justify-center font-medium transition-colors',
                    currentStep > step.id
                      ? 'bg-success text-white'
                      : currentStep === step.id
                      ? 'bg-primary text-white'
                      : 'bg-muted text-muted'
                  )}
                >
                  {currentStep > step.id ? '✓' : step.id}
                </div>
                <div className="mt-2 text-center">
                  <p className={clsx(
                    'text-sm font-medium',
                    currentStep === step.id ? 'text-primary' : 'text-muted'
                  )}>
                    {step.title}
                  </p>
                  <p className="text-xs text-muted">{step.description}</p>
                </div>
              </div>
              {index < steps.length - 1 && (
                <div
                  className={clsx(
                    'w-16 h-1 mx-2',
                    currentStep > step.id ? 'bg-success' : 'bg-muted'
                  )}
                />
              )}
            </div>
          ))}
        </div>

        {/* Step Content */}
        <div className="bg-surface rounded-lg shadow-lg p-6">
          {currentStep === 1 && (
            <OrganizationStep
              data={data.organization}
              onChange={(updates) => updateData('organization', updates)}
              onNext={handleNext}
            />
          )}

          {currentStep === 2 && (
            <UserStep
              data={data.user}
              onChange={(updates) => updateData('user', updates)}
              onNext={handleNext}
              onBack={handleBack}
            />
          )}

          {currentStep === 3 && (
            <PreferencesStep
              data={data.preferences}
              onChange={(updates) => updateData('preferences', updates)}
              onNext={handleNext}
              onBack={handleBack}
            />
          )}

          {currentStep === 4 && (
            <CompletionStep
              data={data}
              onComplete={handleComplete}
              onBack={handleBack}
            />
          )}
        </div>

        {/* Skip Option */}
        {currentStep < totalSteps && (
          <div className="text-center mt-4">
            <button
              onClick={() => {/* TODO: Skip onboarding */}}
              className="text-sm text-muted hover:text-text underline"
            >
              تخطي الإعداد لاحقًا
            </button>
          </div>
        )}
      </div>
    </div>
  )
}
