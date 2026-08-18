import { useState } from 'react'
import { clsx } from 'clsx'
import { PermissionWrapper } from '@/components/auth/PermissionWrapper'
import type { User } from '../types/settings'

export function UsersManagement() {
  const [users, setUsers] = useState<User[]>([])
  const [showAddModal, setShowAddModal] = useState(false)
  const [selectedUser, setSelectedUser] = useState<User | null>(null)
  const [loading, setLoading] = useState(false)

  // TODO: Fetch users from API

  const handleAddUser = () => {
    setShowAddModal(true)
  }

  const handleEditUser = (user: User) => {
    setSelectedUser(user)
    setShowAddModal(true)
  }

  const handleDeleteUser = (userId: string) => {
    // TODO: Show confirmation and delete user
    console.log('Delete user:', userId)
  }

  const handleToggleStatus = (userId: string, isActive: boolean) => {
    // TODO: Toggle user status
    console.log('Toggle user status:', userId, isActive)
  }

  const roleLabels = {
    owner: 'المالك',
    manager: 'المدير',
    employee: 'الموظف',
    accountant: 'المحاسب',
  }

  const roleColors = {
    owner: 'bg-primary-10 text-primary',
    manager: 'bg-success-10 text-success',
    employee: 'bg-info-10 text-info',
    accountant: 'bg-warning-10 text-warning',
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h3 className="text-lg font-semibold text-text">إدارة المستخدمين</h3>
          <p className="text-sm text-muted">إدارة صلاحيات وأدوار المستخدمين</p>
        </div>
        <PermissionWrapper role="owner" resource="users" action="create">
          <button
            onClick={handleAddUser}
            className="px-4 py-2 bg-primary text-white rounded-lg hover:bg-primary-600 transition-colors"
          >
            إضافة مستخدم
          </button>
        </PermissionWrapper>
      </div>

      {/* Users List */}
      {users.length > 0 ? (
        <div className="bg-surface rounded-lg overflow-hidden">
          <table className="w-full">
            <thead className="bg-muted-10 border-b border-border">
              <tr>
                <th className="px-4 py-3 text-right text-sm font-medium text-muted">المستخدم</th>
                <th className="px-4 py-3 text-right text-sm font-medium text-muted">البريد الإلكتروني</th>
                <th className="px-4 py-3 text-right text-sm font-medium text-muted">الدور</th>
                <th className="px-4 py-3 text-right text-sm font-medium text-muted">الحالة</th>
                <th className="px-4 py-3 text-right text-sm font-medium text-muted">آخر دخول</th>
                <th className="px-4 py-3 text-right text-sm font-medium text-muted">إجراءات</th>
              </tr>
            </thead>
            <tbody>
              {users.map((user) => (
                <tr key={user.id} className="border-b border-border hover:bg-muted-5">
                  <td className="px-4 py-3">
                    <div className="flex items-center gap-3">
                      <div className="w-8 h-8 rounded-full bg-primary-10 flex items-center justify-center text-primary font-medium">
                        {user.name.charAt(0)}
                      </div>
                      <div>
                        <p className="font-medium text-text">{user.name}</p>
                        {user.phone && (
                          <p className="text-sm text-muted">{user.phone}</p>
                        )}
                      </div>
                    </div>
                  </td>
                  <td className="px-4 py-3 text-muted">{user.email}</td>
                  <td className="px-4 py-3">
                    <span className={clsx('px-2 py-1 rounded-full text-xs font-medium', roleColors[user.role])}>
                      {roleLabels[user.role]}
                    </span>
                  </td>
                  <td className="px-4 py-3">
                    <span className={clsx(
                      'px-2 py-1 rounded-full text-xs font-medium',
                      user.isActive ? 'bg-success-10 text-success' : 'bg-muted-10 text-muted'
                    )}>
                      {user.isActive ? 'نشط' : 'معطل'}
                    </span>
                  </td>
                  <td className="px-4 py-3 text-muted text-sm">
                    {user.lastLogin || 'لم يسجل دخول'}
                  </td>
                  <td className="px-4 py-3">
                    <div className="flex gap-2">
                      <PermissionWrapper role="owner" resource="users" action="update">
                        <button
                          onClick={() => handleEditUser(user)}
                          className="text-primary hover:text-primary-600 text-sm"
                        >
                          تعديل
                        </button>
                      </PermissionWrapper>
                      <PermissionWrapper role="owner" resource="users" action="delete">
                        <button
                          onClick={() => handleDeleteUser(user.id)}
                          className="text-danger hover:text-danger-600 text-sm"
                        >
                          حذف
                        </button>
                      </PermissionWrapper>
                      <button
                        onClick={() => handleToggleStatus(user.id, !user.isActive)}
                        className="text-muted hover:text-text text-sm"
                      >
                        {user.isActive ? 'تعطيل' : 'تفعيل'}
                      </button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      ) : (
        <div className="text-center py-12 bg-muted-10 rounded-lg">
          <p className="text-muted mb-4">لا يوجد مستخدمين</p>
          <PermissionWrapper role="owner" resource="users" action="create">
            <button
              onClick={handleAddUser}
              className="px-4 py-2 bg-primary text-white rounded-lg hover:bg-primary-600 transition-colors"
            >
              إضافة مستخدم أول
            </button>
          </PermissionWrapper>
        </div>
      )}

      {/* Add/Edit User Modal */}
      {showAddModal && (
        <div className="fixed inset-0 bg-black-50 flex items-center justify-center z-50">
          <div className="bg-surface rounded-lg p-6 max-w-md w-full mx-4">
            <h3 className="text-lg font-semibold text-text mb-4">
              {selectedUser ? 'تعديل مستخدم' : 'إضافة مستخدم جديد'}
            </h3>
            {/* TODO: Add user form */}
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-text mb-2">الاسم</label>
                <input
                  type="text"
                  className="w-full px-4 py-2 border border-border rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent"
                  placeholder="الاسم الكامل"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-text mb-2">البريد الإلكتروني</label>
                <input
                  type="email"
                  className="w-full px-4 py-2 border border-border rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent"
                  placeholder="email@example.com"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-text mb-2">الدور</label>
                <select className="w-full px-4 py-2 border border-border rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent">
                  <option value="employee">موظف</option>
                  <option value="manager">مدير</option>
                  <option value="accountant">محاسب</option>
                  <option value="owner">مالك</option>
                </select>
              </div>
            </div>
            <div className="flex justify-end gap-3 mt-6">
              <button
                onClick={() => {
                  setShowAddModal(false)
                  setSelectedUser(null)
                }}
                className="px-4 py-2 border border-border rounded-lg hover:bg-muted-10 transition-colors"
              >
                إلغاء
              </button>
              <button
                onClick={() => {/* TODO: Save user */}}
                className="px-4 py-2 bg-primary text-white rounded-lg hover:bg-primary-600 transition-colors"
              >
                حفظ
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
