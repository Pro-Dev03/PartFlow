// Conflict Resolution for PartFlow PWA
// This handles conflicts when syncing offline changes with server data

interface SyncableData {
  id: string;
  version: number;
  timestamp: number;
  [key: string]: any;
}

/**
 * Strategy: Last-Write-Wins with timestamp
 * The data with the most recent timestamp wins
 */
export function resolveConflict(localData: SyncableData, serverData: SyncableData): SyncableData {
  if (!localData || !serverData) {
    return localData || serverData;
  }

  // Compare timestamps
  if (localData.timestamp > serverData.timestamp) {
    console.log('[Conflict Resolution] Local data wins (newer timestamp)');
    return {
      ...localData,
      version: Math.max(localData.version, serverData.version) + 1
    };
  } else if (serverData.timestamp > localData.timestamp) {
    console.log('[Conflict Resolution] Server data wins (newer timestamp)');
    return serverData;
  } else {
    // Timestamps are equal, prefer server data
    console.log('[Conflict Resolution] Timestamps equal, server data wins');
    return serverData;
  }
}

/**
 * Strategy: Version-based conflict detection
 * Returns true if there's a conflict (versions don't match)
 */
export function detectConflict(localData: SyncableData, serverData: SyncableData): boolean {
  if (!localData || !serverData) {
    return false;
  }

  const hasConflict = localData.version !== serverData.version;
  
  if (hasConflict) {
    console.log('[Conflict Resolution] Conflict detected:', {
      localVersion: localData.version,
      serverVersion: serverData.version
    });
  }
  
  return hasConflict;
}

/**
 * Strategy: Manual merge (for complex data structures)
 * Allows custom merge logic for specific fields
 */
export function manualMerge(
  localData: SyncableData,
  serverData: SyncableData,
  mergeStrategy: (local: any, server: any) => any
): SyncableData {
  const mergedData = mergeStrategy(localData, serverData);
  
  return {
    ...mergedData,
    id: localData.id,
    version: Math.max(localData.version, serverData.version) + 1,
    timestamp: Date.now()
  };
}

/**
 * Strategy: Client wins (for user-initiated changes)
 * Always prefer local data when the user made the change
 */
export function clientWins(localData: SyncableData, serverData: SyncableData): SyncableData {
  console.log('[Conflict Resolution] Client wins strategy');
  return {
    ...localData,
    version: Math.max(localData.version, serverData.version) + 1,
    timestamp: Date.now()
  };
}

/**
 * Strategy: Server wins (for automated/synced data)
 * Always prefer server data (e.g., inventory counts)
 */
export function serverWins(localData: SyncableData, serverData: SyncableData): SyncableData {
  console.log('[Conflict Resolution] Server wins strategy');
  return serverData;
}

/**
 * Field-level merge strategy
 * Merges specific fields from both sources
 */
export function fieldLevelMerge(
  localData: SyncableData,
  serverData: SyncableData,
  clientFields: string[]
): SyncableData {
  const merged: any = { ...serverData };
  
  // Keep client fields from local data
  clientFields.forEach(field => {
    if (localData[field] !== undefined) {
      merged[field] = localData[field];
    }
  });
  
  merged.version = Math.max(localData.version, serverData.version) + 1;
  merged.timestamp = Date.now();
  
  console.log('[Conflict Resolution] Field-level merge applied');
  return merged;
}

/**
 * Smart conflict resolution based on operation type
 */
export function smartResolve(
  operation: 'CREATE' | 'UPDATE' | 'DELETE',
  localData: SyncableData,
  serverData: SyncableData
): SyncableData {
  switch (operation) {
    case 'CREATE':
      // For new items, prefer client (it should be new on server too)
      return clientWins(localData, serverData);
    
    case 'UPDATE':
      // For updates, use field-level merge for important user fields
      return fieldLevelMerge(localData, serverData, ['name', 'description', 'custom_fields']);
    
    case 'DELETE':
      // For deletes, prefer server (it might have been deleted already)
      return serverWins(localData, serverData);
    
    default:
      return resolveConflict(localData, serverData);
  }
}

/**
 * Conflict history tracking
 */
class ConflictHistory {
  private history: Map<string, any[]> = new Map();
  private maxHistorySize = 10;

  recordConflict(id: string, conflict: any) {
    if (!this.history.has(id)) {
      this.history.set(id, []);
    }
    
    const history = this.history.get(id)!;
    history.push({
      ...conflict,
      timestamp: Date.now()
    });
    
    // Keep only last N conflicts
    if (history.length > this.maxHistorySize) {
      history.shift();
    }
  }

  getConflictHistory(id: string): any[] {
    return this.history.get(id) || [];
  }

  clearHistory(id: string) {
    this.history.delete(id);
  }
}

export const conflictHistory = new ConflictHistory();