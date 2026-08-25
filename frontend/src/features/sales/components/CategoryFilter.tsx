import { cn } from '../../../utils';
import { Tag } from 'lucide-react';

interface CategoryFilterProps {
  categories: any[];
  selectedCategory: string | null;
  onCategorySelect: (categoryId: string | null) => void;
}

export function CategoryFilter({
  categories,
  selectedCategory,
  onCategorySelect,
}: CategoryFilterProps) {
  const iconMap: Record<string, any> = {
    package: Tag,
    tag: Tag,
    monitor: Tag,
    cpu: Tag,
    'hard-drive': Tag,
    zap: Tag,
    box: Tag,
    thermometer: Tag,
    keyboard: Tag,
    mouse: Tag,
    headphones: Tag,
    camera: Tag,
    printer: Tag,
    wifi: Tag,
    shield: Tag,
    wrench: Tag,
  };

  return (
    <div className="flex flex-wrap items-center gap-2 rounded-xl border border-border-default bg-bg-surface p-3">
      <button
        onClick={() => onCategorySelect(null)}
        className={cn(
          'inline-flex items-center gap-1.5 rounded-full border px-4 py-2 text-sm font-medium transition-all duration-200',
          'cursor-pointer',
          selectedCategory === null
            ? 'border-2 border-primary bg-primary/10 font-semibold text-primary'
            : 'border-border-default bg-bg-surface text-text-primary hover:border-primary'
        )}
      >
        <Tag className="h-4 w-4" />
        <span>الكل</span>
      </button>

      {categories.map((category: any) => {
        const IconComponent = iconMap[category.icon] || Tag;
        const isSelected = selectedCategory === category.id;
        const color = category.color || '#2563eb';

        return (
          <button
            key={category.id}
            onClick={() => onCategorySelect(category.id)}
            className={cn(
              'inline-flex items-center gap-1.5 rounded-full border px-4 py-2 text-sm font-medium transition-all duration-200',
              'cursor-pointer',
              isSelected
                ? 'border-2 border-primary bg-primary/10 font-semibold text-primary'
                : 'border-border-default bg-bg-surface text-text-primary hover:border-primary'
            )}
          >
            <IconComponent
              className={cn(
                'h-4 w-4',
                isSelected ? 'text-primary' : 'text-text-secondary'
              )}
              style={{ color: isSelected ? color : undefined }}
            />
            <span>{category.name}</span>
          </button>
        );
      })}
    </div>
  );
}
