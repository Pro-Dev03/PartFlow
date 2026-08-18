import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from 'react-query'
import { salesService } from '../services/sales.service'
import { Sale, CartItem, Customer, PaymentMethod } from '../types/sales.types'

export function useSearchProducts(query: string) {
  return useQuery(
    ['search-products', query],
    () => salesService.searchProduct(query),
    {
      enabled: query.length > 2,
      staleTime: 2 * 60 * 1000, // 2 minutes
    }
  )
}

export function useCustomers() {
  return useQuery('customers', () => salesService.getCustomers(), {
    staleTime: 5 * 60 * 1000, // 5 minutes
  })
}

export function usePaymentMethods() {
  return useQuery('payment-methods', () => salesService.getPaymentMethods(), {
    staleTime: 10 * 60 * 1000, // 10 minutes
  })
}

export function useCreateSale() {
  const queryClient = useQueryClient()
  
  return useMutation(
    (sale: Omit<Sale, 'id' | 'createdAt' | 'createdBy'>) => 
      salesService.createSale(sale),
    {
      onSuccess: () => {
        queryClient.invalidateQueries('recent-sales')
        queryClient.invalidateQueries('dashboard-stats')
      },
    }
  )
}

export function useRecentSales() {
  return useQuery('recent-sales', () => salesService.getRecentSales(), {
    staleTime: 1 * 60 * 1000, // 1 minute
  })
}

// Custom hook for cart management
export function useCart() {
  const [cart, setCart] = useState<CartItem[]>([])

  const addToCart = (item: CartItem) => {
    setCart(prev => {
      const existing = prev.find(i => i.id === item.id)
      if (existing) {
        return prev.map(i => 
          i.id === item.id 
            ? { ...i, quantity: Math.min(i.quantity + 1, i.availableStock) }
            : i
        )
      }
      return [...prev, { ...item, quantity: 1 }]
    })
  }

  const removeFromCart = (itemId: string) => {
    setCart(prev => prev.filter(i => i.id !== itemId))
  }

  const updateQuantity = (itemId: string, quantity: number) => {
    setCart(prev => prev.map(i => {
      if (i.id === itemId) {
        return { ...i, quantity: Math.max(1, Math.min(quantity, i.availableStock)) }
      }
      return i
    }))
  }

  const clearCart = () => {
    setCart([])
  }

  const cartTotal = cart.reduce((sum, item) => sum + (item.price * item.quantity), 0)
  const cartCost = cart.reduce((sum, item) => sum + (item.cost * item.quantity), 0)
  const cartProfit = cartTotal - cartCost

  return {
    cart,
    addToCart,
    removeFromCart,
    updateQuantity,
    clearCart,
    cartTotal,
    cartCost,
    cartProfit,
    itemCount: cart.reduce((sum, item) => sum + item.quantity, 0),
  }
}