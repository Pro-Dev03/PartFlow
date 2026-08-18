import { Sale, CartItem, Customer, PaymentMethod } from '../types/sales.types'

// Mock service - replace with actual API calls
export const salesService = {
  async searchProduct(query: string): Promise<CartItem[]> {
    // Replace with actual API call
    return [
      {
        id: '1',
        productId: 'prod-1',
        productName: 'RTX 4070',
        barcode: 'RTX4070-001',
        serialNumber: 'SN123456',
        condition: 'used',
        price: 2350,
        cost: 1850,
        quantity: 1,
        availableStock: 1,
      },
    ]
  },

  async getCustomers(): Promise<Customer[]> {
    // Replace with actual API call
    return [
      {
        id: 'cust-1',
        name: 'أحمد محمد',
        phone: '0501234567',
        email: 'ahmed@example.com',
        address: 'الرياض',
        outstandingBalance: 1250,
        totalPurchases: 8450,
      },
      {
        id: 'cust-2',
        name: 'سارة أحمد',
        phone: '0507654321',
        email: 'sara@example.com',
        address: 'جدة',
        outstandingBalance: 0,
        totalPurchases: 3200,
      },
    ]
  },

  async getPaymentMethods(): Promise<PaymentMethod[]> {
    return [
      { id: '1', name: 'نقد', icon: '💵', value: 'cash' },
      { id: '2', name: 'بطاقة', icon: '💳', value: 'card' },
      { id: '3', name: 'تحويل', icon: '🏦', value: 'transfer' },
      { id: '4', name: 'دين', icon: '📝', value: 'debt' },
    ]
  },

  async createSale(sale: Omit<Sale, 'id' | 'createdAt' | 'createdBy'>): Promise<Sale> {
    // Replace with actual API call
    return {
      ...sale,
      id: `sale-${Date.now()}`,
      createdAt: new Date(),
      createdBy: 'current-user',
    }
  },

  async getRecentSales(): Promise<Sale[]> {
    // Replace with actual API call
    return [
      {
        id: 'sale-1',
        customerId: 'cust-1',
        customerName: 'أحمد محمد',
        items: [
          {
            id: '1',
            productId: 'prod-1',
            productName: 'RTX 4070',
            barcode: 'RTX4070-001',
            condition: 'used',
            price: 2350,
            cost: 1850,
            quantity: 1,
            availableStock: 1,
          },
        ],
        subtotal: 2350,
        tax: 0,
        discount: 0,
        total: 2350,
        paymentMethod: 'cash',
        paymentStatus: 'paid',
        paidAmount: 2350,
        remainingAmount: 0,
        createdAt: new Date(Date.now() - 1000 * 60 * 30),
        createdBy: 'محمد',
      },
    ]
  },
}