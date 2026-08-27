import { cn } from '../../../utils';
import { Tag, Smartphone, Laptop, Monitor, Cpu, HardDrive, Camera, Printer, Wifi, Headphones, Speaker, Cable } from 'lucide-react';
import type { ComponentType, CSSProperties } from 'react';

interface CategoryOption {
  id: string;
  name: string;
  icon?: string;
  color?: string;
}
interface CategoryFilterProps {
  categories: CategoryOption[];
  selectedCategory: string | null;
  onCategorySelect: (categoryId: string | null) => void;
}

export function CategoryFilter({
  categories,
  selectedCategory,
  onCategorySelect,
}: CategoryFilterProps) {
  const iconMap: Record<string, ComponentType<{ className?: string; style?: CSSProperties }>> = {
    smartphone: Smartphone,
    laptop: Laptop,
    monitor: Monitor,
    cpu: Cpu,
    'hard-drive': HardDrive,
    camera: Camera,
    printer: Printer,
    wifi: Wifi,
    headphones: Headphones,
    speaker: Speaker,
    cable: Cable,
    // Fallback for any other icons
    package: Tag,
    tag: Tag,
    zap: Tag,
    box: Tag,
    thermometer: Tag,
    keyboard: Tag,
    mouse: Tag,
    shield: Tag,
    wrench: Tag,
  };

  return (
    <div className="flex flex-wrap items-center gap-2 rounded-xl border border-border-default p-3 backdrop-blur-xl" style={{ 
      background: 'var(--bg-surface)',
      backdropFilter: 'blur(20px)' 
    }}>
      <button
        onClick={() => onCategorySelect(null)}
        className={cn(
          'inline-flex items-center gap-1.5 rounded-full border px-4 py-2 text-sm font-medium transition-all duration-200',
          'cursor-pointer backdrop-blur-md',
          selectedCategory === null
            ? 'border-2 border-primary font-semibold shadow-lg'
            : 'border-border-default backdrop-blur-sm hover:border-primary'
        )}
        style={{
          background: selectedCategory === null ? 'rgba(37, 99, 235, 0.15)' : 'rgba(128, 128, 128, 0.08)',
          color: selectedCategory === null ? 'var(--primary)' : 'var(--text-primary)',
          boxShadow: selectedCategory === null ? '0 10px 15px -3px rgba(37, 99, 235, 0.15)' : undefined
        }}
      >
        <Tag 
          className="h-4 w-4"
          style={{ 
            color: selectedCategory === null ? 'var(--primary)' : 'var(--text-secondary)',
            opacity: selectedCategory === null ? 1 : 0.6
          }}
        />
        <span>الكل</span>
      </button>

      {categories.map((category) => {
        const IconComponent = iconMap[category.icon] || Tag;
        const isSelected = selectedCategory === category.id;
        const color = category.color || '#2563eb';

        return (
          <button
            key={category.id}
            onClick={() => onCategorySelect(category.id)}
            className={cn(
              'inline-flex items-center gap-1.5 rounded-full border px-4 py-2 text-sm font-medium transition-all duration-200',
              'cursor-pointer backdrop-blur-md',
              isSelected
                ? 'border-2 border-primary font-semibold shadow-lg'
                : 'border-border-default backdrop-blur-sm hover:border-primary'
            )}
            style={{
              background: isSelected ? 'rgba(37, 99, 235, 0.15)' : 'rgba(128, 128, 128, 0.08)',
              color: isSelected ? 'var(--primary)' : 'var(--text-primary)',
              boxShadow: isSelected ? `0 10px 15px -3px ${color}26` : undefined
            }}
          >
            <IconComponent
              className="h-4 w-4"
              style={{ 
                color: color,
                opacity: isSelected ? 1 : 0.5
              }}
            />
            <span>{category.name}</span>
          </button>
        );
      })}
    </div>
  );
}
