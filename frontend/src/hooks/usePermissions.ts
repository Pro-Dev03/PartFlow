import { useAuthStore } from '../stores/authStore';
import {
  hasPermission,
  hasAnyPermission,
  hasAllPermissions,
  getUserPermissions,
  type Permission,
} from '../lib/permissions';

export function usePermissions() {
  const { role } = useAuthStore();
  const userRole = role || 'employee';

  return {
    hasPermission: (permission: Permission) => hasPermission(userRole, permission),
    hasAnyPermission: (permissions: Permission[]) => hasAnyPermission(userRole, permissions),
    hasAllPermissions: (permissions: Permission[]) => hasAllPermissions(userRole, permissions),
    getPermissions: () => getUserPermissions(userRole),
    role: userRole,
  };
}