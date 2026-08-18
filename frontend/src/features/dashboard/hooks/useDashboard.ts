import { useQuery } from 'react-query'
import { dashboardService } from '../services/dashboard.service'

export function useDashboardStats() {
  return useQuery('dashboard-stats', () => dashboardService.getStats(), {
    staleTime: 5 * 60 * 1000, // 5 minutes
    refetchOnWindowFocus: false,
  })
}

export function useDashboardAlerts() {
  return useQuery('dashboard-alerts', () => dashboardService.getAlerts(), {
    staleTime: 2 * 60 * 1000, // 2 minutes
    refetchOnWindowFocus: false,
  })
}

export function useQuickActions() {
  return useQuery('quick-actions', () => dashboardService.getQuickActions(), {
    staleTime: 10 * 60 * 1000, // 10 minutes
  })
}

export function useRecentActivity() {
  return useQuery('recent-activity', () => dashboardService.getRecentActivity(), {
    staleTime: 1 * 60 * 1000, // 1 minute
    refetchOnWindowFocus: false,
  })
}