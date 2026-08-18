// Offline Queue for PartFlow PWA
// This handles operations when the user is offline and syncs them when back online

interface OfflineOperation {
  id: string;
  type: 'CREATE' | 'UPDATE' | 'DELETE';
  endpoint: string;
  data: any;
  timestamp: number;
  status: 'pending' | 'completed' | 'failed';
  retryCount: number;
  error?: string;
}

class OfflineQueue {
  private queue: OfflineOperation[] = [];
  private storageKey = 'offline_queue';
  private syncInProgress = false;

  constructor() {
    this.loadFromStorage();
    this.setupOnlineListener();
  }

  private loadFromStorage() {
    try {
      const stored = localStorage.getItem(this.storageKey);
      if (stored) {
        this.queue = JSON.parse(stored);
      }
    } catch (error) {
      console.error('Failed to load offline queue:', error);
      this.queue = [];
    }
  }

  private saveToStorage() {
    try {
      localStorage.setItem(this.storageKey, JSON.stringify(this.queue));
    } catch (error) {
      console.error('Failed to save offline queue:', error);
    }
  }

  private setupOnlineListener() {
    window.addEventListener('online', () => {
      console.log('[Offline Queue] Back online, processing queue...');
      this.processQueue();
    });

    window.addEventListener('offline', () => {
      console.log('[Offline Queue] Gone offline, operations will be queued');
    });
  }

  add(operation: Omit<OfflineOperation, 'id' | 'timestamp' | 'status' | 'retryCount'>) {
    const newOperation: OfflineOperation = {
      id: this.generateId(),
      timestamp: Date.now(),
      status: 'pending',
      retryCount: 0,
      ...operation
    };

    this.queue.push(newOperation);
    this.saveToStorage();
    
    console.log('[Offline Queue] Added operation:', newOperation.type, newOperation.endpoint);
    
    // Try to process immediately if online
    if (navigator.onLine) {
      this.processQueue();
    }
  }

  private generateId(): string {
    return `${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
  }

  async processQueue() {
    if (this.syncInProgress || !navigator.onLine || this.queue.length === 0) {
      return;
    }

    this.syncInProgress = true;
    console.log(`[Offline Queue] Processing ${this.queue.length} operations...`);

    for (const operation of this.queue) {
      if (operation.status === 'pending') {
        try {
          await this.executeOperation(operation);
          this.markAsCompleted(operation);
        } catch (error: any) {
          this.markAsFailed(operation, error);
        }
      }
    }

    this.clearCompleted();
    this.syncInProgress = false;
    
    console.log('[Offline Queue] Queue processing completed');
  }

  private async executeOperation(operation: OfflineOperation): Promise<void> {
    const { type, endpoint, data } = operation;
    
    console.log(`[Offline Queue] Executing ${type} ${endpoint}`);
    
    // Import API client dynamically to avoid circular dependencies
    const { apiClient } = await import('../services/api/client');
    
    switch (type) {
      case 'CREATE':
        await apiClient.post(endpoint, data);
        break;
      case 'UPDATE':
        await apiClient.put(endpoint, data);
        break;
      case 'DELETE':
        await apiClient.delete(endpoint);
        break;
      default:
        throw new Error(`Unknown operation type: ${type}`);
    }
  }

  private markAsCompleted(operation: OfflineOperation) {
    operation.status = 'completed';
    this.saveToStorage();
    console.log(`[Offline Queue] Operation completed:`, operation.id);
  }

  private markAsFailed(operation: OfflineOperation, error: any) {
    operation.status = 'failed';
    operation.error = error.message;
    operation.retryCount++;
    
    // Retry up to 3 times
    if (operation.retryCount < 3) {
      operation.status = 'pending';
    }
    
    this.saveToStorage();
    console.error(`[Offline Queue] Operation failed:`, operation.id, error);
  }

  private clearCompleted() {
    this.queue = this.queue.filter(op => op.status !== 'completed');
    this.saveToStorage();
  }

  getPendingCount(): number {
    return this.queue.filter(op => op.status === 'pending').length;
  }

  getFailedCount(): number {
    return this.queue.filter(op => op.status === 'failed').length;
  }

  getAll(): OfflineOperation[] {
    return [...this.queue];
  }

  clearAll() {
    this.queue = [];
    this.saveToStorage();
  }
}

export const offlineQueue = new OfflineQueue();