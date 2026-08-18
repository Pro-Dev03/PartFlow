import { ReactNode } from 'react'
import { usePermissions } from './usePermissions'

interface PermissionWrapperProps {
  role: string
  resource: string
  action: string
  fallback?: ReactNode
  children: ReactNode
}

export function PermissionWrapper({
  role,
  resource,
  action,
  fallback = null,
  children,
}: PermissionWrapperProps) {
  const { hasPermission } = usePermissions(role as any)

  if (!hasPermission(resource, action)) {
    return <>{fallback}</>
  }

  return <>{children}</>
}

interface PermissionWrapperPropsAny {
  role: string
  resource: string
  actions: string[]
  requireAll?: boolean
  fallback?: ReactNode
  children: ReactNode
}

export function PermissionWrapperAny({
  role,
  resource,
  actions,
  requireAll = false,
  fallback = null,
  children,
}: PermissionWrapperPropsAny) {
  const { hasAnyPermission, hasAllPermissions } = usePermissions(role as any)

  const hasAccess = requireAll
    ? hasAllPermissions(resource, actions)
    : hasAnyPermission(resource, actions)

  if (!hasAccess) {
    return <>{fallback}</>
  }

  return <>{children}</>
}
