export function Auth() {
  return (
    <div className="min-h-screen flex items-center justify-center bg-background">
      <div className="bg-surface p-8 rounded-lg border border-border w-full max-w-md">
        <h1 className="text-2xl font-bold text-center mb-6">تسجيل الدخول</h1>
        <form className="space-y-4">
          <div>
            <label className="block text-sm font-medium mb-2">البريد الإلكتروني</label>
            <input
              type="email"
              className="w-full p-2 border border-border rounded-lg"
              placeholder="example@email.com"
            />
          </div>
          <div>
            <label className="block text-sm font-medium mb-2">كلمة المرور</label>
            <input
              type="password"
              className="w-full p-2 border border-border rounded-lg"
              placeholder="••••••••"
            />
          </div>
          <button
            type="submit"
            className="w-full p-2 bg-primary text-white rounded-lg hover:bg-primary-600 transition-colors"
          >
            تسجيل الدخول
          </button>
        </form>
      </div>
    </div>
  )
}
