import { SmartDeleteResult } from '../types/api';

/**
 * SmartDelete Handler Utility (PRODUCT-PHILOSOPHY.md)
 *
 * This utility handles the SmartDelete pattern where the backend decides
 * whether to DELETE, REVERSE, or BLOCK based on the entity state.
 *
 * From user perspective: it's just "delete"
 * Internally: the system decides the appropriate action
 * 
 * Philosophy: "The system works for the store owner, not the store owner works for the system"
 */

export interface SmartDeleteHandlerOptions {
  onSuccess?: (result: SmartDeleteResult) => void;
  onBlocked?: (result: SmartDeleteResult) => void;
  onError?: (error: Error) => void;
  showConfirmation?: boolean;
  confirmationMessage?: string;
}

/**
 * Handle SmartDelete response with appropriate UI feedback
 */
export async function handleSmartDelete(
  deleteFunction: () => Promise<SmartDeleteResult>,
  options: SmartDeleteHandlerOptions = {}
): Promise<boolean> {
  const {
    onSuccess,
    onBlocked,
    onError,
    showConfirmation = true,
    confirmationMessage = 'هل تريد حذف هذه العملية؟'
  } = options;

  try {
    // Show confirmation if required
    if (showConfirmation) {
      const confirmed = window.confirm(confirmationMessage);
      if (!confirmed) {
        return false;
      }
    }

    // Execute the delete operation
    const result = await deleteFunction();

    // Handle different action types
    switch (result.action) {
      case 'deleted':
        // Simple delete succeeded
        if (onSuccess) {
          onSuccess(result);
        }
        showSuccessMessage(result.message);
        return true;

      case 'reversed':
        // Reverse operation succeeded (delete with inventory adjustment)
        if (onSuccess) {
          onSuccess(result);
        }
        showSuccessMessage(result.message);
        return true;

      case 'blocked':
        // Delete was blocked - show details to user
        if (onBlocked) {
          onBlocked(result);
        }
        showBlockedMessage(result);
        return false;

      default:
        throw new Error(`Unknown action type: ${result.action}`);
    }
  } catch (error) {
    if (onError) {
      onError(error as Error);
    }
    showErrorMessage(error);
    return false;
  }
}

/**
 * Show success message to user
 */
function showSuccessMessage(message: string): void {
  // You can replace this with your preferred notification system
  // For now using simple alert as placeholder
  alert(message);
}

/**
 * Show blocked message with details (PRODUCT-PHILOSOPHY.md - store owner language)
 */
function showBlockedMessage(result: SmartDeleteResult): void {
  let message = result.message;

  if (result.details) {
    message += '\n\n';
    message += `السبب: ${result.details.reason}`;

    if (result.details.used_items && result.details.used_items.length > 0) {
      message += '\n\nالمنتجات المتأثرة:\n';
      result.details.used_items.forEach((item, index) => {
        message += `${index + 1}. ${item.product_name}: `;
        message += `الكمية الأصلية ${item.original_quantity}, `;
        message += `تم بيع ${item.sold_quantity}\n`;
      });
    }

    if (result.details.suggested_action) {
      message += `\n\n${result.details.suggested_action}`;
    }
  }

  alert(message);
}

/**
 * Show error message
 */
function showErrorMessage(error: Error): void {
  alert(`حدث خطأ: ${error.message}`);
}

/**
 * Create a smart delete handler for React components
 */
export function useSmartDeleteHandler() {
  const handleDelete = async (
    deleteFunction: () => Promise<SmartDeleteResult>,
    options: SmartDeleteHandlerOptions = {}
  ): Promise<boolean> => {
    return handleSmartDelete(deleteFunction, options);
  };

  return { handleDelete };
}

/**
 * Check if a delete result was successful
 */
export function isDeleteSuccessful(result: SmartDeleteResult): boolean {
  return result.can_proceed && (result.action === 'deleted' || result.action === 'reversed');
}

/**
 * Check if a delete result was blocked
 */
export function isDeleteBlocked(result: SmartDeleteResult): boolean {
  return result.action === 'blocked';
}

/**
 * Get user-friendly action description (PRODUCT-PHILOSOPHY.md)
 */
export function getActionDescription(result: SmartDeleteResult): string {
  switch (result.action) {
    case 'deleted':
      return 'تم الحذف';
    case 'reversed':
      return 'تم إلغاء العملية';
    case 'blocked':
      return 'لم يتم الحذف';
    default:
      return 'حالة غير معروفة';
  }
}
