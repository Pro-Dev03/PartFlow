import { useState, useMemo } from 'react';
import { cn } from '../../lib/utils';
import { Button } from '../ui/button';
import { Badge } from '../ui/badge';
import {
  Table,
  TableHeader,
  TableBody,
  TableFooter,
  TableHead,
  TableRow,
  TableCell,
} from '../ui/table';
import { 
  ChevronDown, 
  ChevronUp, 
  MoreHorizontal, 
  Download, 
  Filter,
  Eye,
  EyeOff,
  RefreshCw
} from 'lucide-react';

export interface Column<T> {
  key: string;
  title: string;
  width?: string;
  sortable?: boolean;
  filterable?: boolean;
  render?: (row: T) => React.ReactNode;
}

export interface DataTableProps<T> {
  data: T[];
  columns: Column<T>[];
  selectable?: boolean;
  onSelectionChange?: (selected: T[]) => void;
  onRowClick?: (row: T) => void;
  loading?: boolean;
  empty?: React.ReactNode;
  pagination?: {
    page: number;
    pageSize: number;
    total: number;
    onPageChange: (page: number) => void;
  };
  bulkActions?: {
    label: string;
    icon?: any;
    onClick: (selected: T[]) => void;
    variant?: 'primary' | 'secondary' | 'danger';
  }[];
  onExport?: (data: T[]) => void;
  refreshable?: boolean;
  onRefresh?: () => void;
  expandable?: boolean;
  renderExpanded?: (row: T) => React.ReactNode;
}

export function DataTable<T extends Record<string, any>>({
  data,
  columns,
  selectable = false,
  onSelectionChange,
  onRowClick,
  loading = false,
  empty,
  pagination,
  bulkActions = [],
  onExport,
  refreshable = false,
  onRefresh,
  expandable = false,
  renderExpanded,
}: DataTableProps<T>) {
  const [selectedRows, setSelectedRows] = useState<Set<string>>(new Set());
  const [sortColumn, setSortColumn] = useState<string | null>(null);
  const [sortDirection, setSortDirection] = useState<'asc' | 'desc'>('asc');
  const [expandedRows, setExpandedRows] = useState<Set<string>>(new Set());
  const [columnVisibility, setColumnVisibility] = useState<Set<string>>(
    new Set(columns.map(col => col.key))
  );
  const [showColumnMenu, setShowColumnMenu] = useState(false);
  const [showFilterMenu, setShowFilterMenu] = useState(false);

  const visibleColumns = useMemo(() => 
    columns.filter(col => columnVisibility.has(col.key)),
    [columns, columnVisibility]
  );

  const handleRowSelect = (rowId: string) => {
    const newSelected = new Set(selectedRows);
    if (newSelected.has(rowId)) {
      newSelected.delete(rowId);
    } else {
      newSelected.add(rowId);
    }
    setSelectedRows(newSelected);
    
    if (onSelectionChange) {
      const selectedData = data.filter(row => {
        const id = row.id || row._id || JSON.stringify(row);
        return newSelected.has(id);
      });
      onSelectionChange(selectedData);
    }
  };

  const handleSelectAll = () => {
    if (selectedRows.size === data.length) {
      setSelectedRows(new Set());
    } else {
      const allIds = new Set(
        data.map(row => row.id || row._id || JSON.stringify(row))
      );
      setSelectedRows(allIds);
    }
    
    if (onSelectionChange) {
      if (selectedRows.size === data.length) {
        onSelectionChange([]);
      } else {
        onSelectionChange(data);
      }
    }
  };

  const handleSort = (columnKey: string) => {
    if (sortColumn === columnKey) {
      setSortDirection(sortDirection === 'asc' ? 'desc' : 'asc');
    } else {
      setSortColumn(columnKey);
      setSortDirection('asc');
    }
  };

  const handleRowExpand = (rowId: string) => {
    const newExpanded = new Set(expandedRows);
    if (newExpanded.has(rowId)) {
      newExpanded.delete(rowId);
    } else {
      newExpanded.add(rowId);
    }
    setExpandedRows(newExpanded);
  };

  const handleColumnToggle = (columnKey: string) => {
    const newVisibility = new Set(columnVisibility);
    if (newVisibility.has(columnKey)) {
      // Don't allow hiding the last column
      if (newVisibility.size > 1) {
        newVisibility.delete(columnKey);
      }
    } else {
      newVisibility.add(columnKey);
    }
    setColumnVisibility(newVisibility);
  };

  const handleExport = () => {
    if (onExport) {
      const selectedData = data.filter(row => {
        const id = row.id || row._id || JSON.stringify(row);
        return selectedRows.size > 0 ? selectedRows.has(id) : true;
      });
      onExport(selectedData);
    }
  };

  const getSortedData = () => {
    if (!sortColumn) return data;
    
    return [...data].sort((a, b) => {
      const aVal = a[sortColumn];
      const bVal = b[sortColumn];
      
      if (typeof aVal === 'string' && typeof bVal === 'string') {
        return sortDirection === 'asc' 
          ? aVal.localeCompare(bVal)
          : bVal.localeCompare(aVal);
      }
      
      if (typeof aVal === 'number' && typeof bVal === 'number') {
        return sortDirection === 'asc' ? aVal - bVal : bVal - aVal;
      }
      
      return 0;
    });
  };

  const sortedData = getSortedData();
  const selectedData = data.filter(row => {
    const id = row.id || row._id || JSON.stringify(row);
    return selectedRows.has(id);
  });

  if (loading) {
    return (
      <div className="flex items-center justify-center p-8">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
      </div>
    );
  }

  if (data.length === 0 && empty) {
    return <div className="p-8">{empty}</div>;
  }

  return (
    <div className="space-y-4">
      {/* Toolbar */}
      <div className="flex items-center justify-between gap-4">
        <div className="flex items-center gap-2">
          {selectable && selectedRows.size > 0 && (
            <Badge variant="info" className="gap-1">
              {selectedRows.size} selected
            </Badge>
          )}
          
          {selectedRows.size > 0 && bulkActions.length > 0 && (
            <div className="flex items-center gap-2">
              {bulkActions.map((action, index) => (
                <Button
                  key={index}
                  variant={action.variant || 'secondary'}
                  size="sm"
                  onClick={() => action.onClick(selectedData)}
                >
                  {action.icon && <action.icon className="w-4 h-4" />}
                  {action.label}
                </Button>
              ))}
            </div>
          )}
        </div>

        <div className="flex items-center gap-2">
          {refreshable && onRefresh && (
            <Button
              variant="ghost"
              size="sm"
              onClick={onRefresh}
              className="text-text-secondary hover:text-text-primary"
            >
              <RefreshCw className="w-4 h-4" />
            </Button>
          )}
          
          {onExport && (
            <Button
              variant="secondary"
              size="sm"
              onClick={handleExport}
              className="gap-2"
            >
              <Download className="w-4 h-4" />
              Export
            </Button>
          )}

          <div className="relative">
            <Button
              variant="ghost"
              size="sm"
              onClick={() => setShowColumnMenu(!showColumnMenu)}
              className="text-text-secondary hover:text-text-primary"
            >
              <MoreHorizontal className="w-4 h-4" />
            </Button>
            
            {showColumnMenu && (
              <div className="absolute top-full right-0 mt-2 bg-surface border border-border rounded-lg shadow-lg p-2 min-w-[150px] z-10">
                {columns.map((column) => (
                  <div
                    key={column.key}
                    className="flex items-center gap-2 px-3 py-2 hover:bg-surface-elevated rounded cursor-pointer text-sm"
                    onClick={() => handleColumnToggle(column.key)}
                  >
                    {columnVisibility.has(column.key) ? (
                      <Eye className="w-4 h-4 text-text-tertiary" />
                    ) : (
                      <EyeOff className="w-4 h-4 text-text-tertiary" />
                    )}
                    <span className="text-text-primary">{column.title}</span>
                  </div>
                ))}
              </div>
            )}
          </div>
        </div>
      </div>

      {/* Table */}
      <div className="rounded-lg border border-border bg-surface overflow-hidden">
        <Table>
          <TableHeader>
            <TableRow>
              {selectable && (
                <TableHead className="w-12">
                  <input
                    type="checkbox"
                    checked={selectedRows.size === data.length && data.length > 0}
                    onChange={handleSelectAll}
                    className="rounded border-border"
                  />
                </TableHead>
              )}
              {expandable && <TableHead className="w-12"></TableHead>}
              {visibleColumns.map((column) => (
                <TableHead
                  key={column.key}
                  className={cn(
                    'cursor-pointer hover:bg-surface-elevated transition-colors',
                    column.sortable && 'select-none'
                  )}
                  style={{ width: column.width }}
                  onClick={() => column.sortable && handleSort(column.key)}
                >
                  <div className="flex items-center gap-2">
                    {column.title}
                    {column.sortable && sortColumn === column.key && (
                      <span className="text-text-tertiary">
                        {sortDirection === 'asc' ? (
                          <ChevronUp className="w-4 h-4" />
                        ) : (
                          <ChevronDown className="w-4 h-4" />
                        )}
                      </span>
                    )}
                  </div>
                </TableHead>
              ))}
            </TableRow>
          </TableHeader>
          <TableBody>
            {sortedData.map((row, index) => {
              const rowId = row.id || row._id || JSON.stringify(row);
              const isSelected = selectedRows.has(rowId);
              const isExpanded = expandedRows.has(rowId);
              
              return (
                <>
                  <TableRow
                    key={rowId}
                    className={cn(
                      'cursor-pointer transition-colors',
                      isSelected && 'bg-primary/5',
                      onRowClick && 'hover:bg-surface-elevated'
                    )}
                    onClick={() => onRowClick && onRowClick(row)}
                  >
                    {selectable && (
                      <TableCell>
                        <input
                          type="checkbox"
                          checked={isSelected}
                          onChange={() => handleRowSelect(rowId)}
                          className="rounded border-border"
                          onClick={(e) => e.stopPropagation()}
                        />
                      </TableCell>
                    )}
                    {expandable && (
                      <TableCell>
                        <button
                          onClick={(e) => {
                            e.stopPropagation();
                            handleRowExpand(rowId);
                          }}
                          className="text-text-tertiary hover:text-text-primary transition-colors"
                        >
                          {isExpanded ? (
                            <ChevronDown className="w-4 h-4" />
                          ) : (
                            <ChevronUp className="w-4 h-4" />
                          )}
                        </button>
                      </TableCell>
                    )}
                    {visibleColumns.map((column) => (
                      <TableCell key={column.key}>
                        {column.render ? column.render(row) : String(row[column.key] || '')}
                      </TableCell>
                    ))}
                  </TableRow>
                  {expandable && isExpanded && renderExpanded && (
                    <TableRow key={`${rowId}-expanded`}>
                      <TableCell colSpan={visibleColumns.length + (selectable ? 1 : 0) + (expandable ? 1 : 0)} className="bg-surface-elevated">
                        {renderExpanded(row)}
                      </TableCell>
                    </TableRow>
                  )}
                </>
              );
            })}
          </TableBody>
        </Table>
      </div>

      {/* Pagination */}
      {pagination && (
        <div className="flex items-center justify-between">
          <div className="text-sm text-text-secondary">
            Showing {((pagination.page - 1) * pagination.pageSize) + 1} to{' '}
            {Math.min(pagination.page * pagination.pageSize, pagination.total)} of {pagination.total} results
          </div>
          <div className="flex items-center gap-2">
            <button
              onClick={() => pagination.onPageChange(pagination.page - 1)}
              disabled={pagination.page === 1}
              className="px-3 py-1 rounded border border-border hover:bg-surface-elevated disabled:opacity-50 disabled:cursor-not-allowed text-sm text-text-primary"
            >
              Previous
            </button>
            <span className="text-sm text-text-primary">
              Page {pagination.page}
            </span>
            <button
              onClick={() => pagination.onPageChange(pagination.page + 1)}
              disabled={pagination.page * pagination.pageSize >= pagination.total}
              className="px-3 py-1 rounded border border-border hover:bg-surface-elevated disabled:opacity-50 disabled:cursor-not-allowed text-sm text-text-primary"
            >
              Next
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
