import { useState } from 'react';
import { Button } from '../../components/ui/button';
import { Input } from '../../components/ui/input';
import { Select } from '../../components/ui/select';
import { Card, CardContent, CardHeader, CardTitle } from '../../components/ui/card';
import { Badge } from '../../components/ui/badge';
import { EmptyState } from '../../components/ui/empty-state';
import { ErrorState } from '../../components/ui/error-state';
import { LoadingSpinner } from '../../components/ui/loading-spinner';
import { PageHeader } from '../../components/ui/page-header';
import { 
  Search, 
  Plus, 
  Check, 
  X, 
  AlertCircle, 
  Inbox,
  Package,
  Filter,
  ArrowUpDown
} from 'lucide-react';

export function DesignSystemShowcase() {
  const [buttonVariant, setButtonVariant] = useState<'primary' | 'secondary' | 'ghost' | 'danger' | 'success' | 'outline'>('primary');
  const [buttonSize, setButtonSize] = useState<'sm' | 'md' | 'lg'>('md');
  const [buttonLoading, setButtonLoading] = useState(false);
  const [inputError, setInputError] = useState('');
  const [cardVariant, setCardVariant] = useState<'default' | 'interactive' | 'featured' | 'warning' | 'ai'>('default');
  const [badgeVariant, setBadgeVariant] = useState<'default' | 'success' | 'warning' | 'danger' | 'info' | 'secondary' | 'outline'>('default');

  return (
    <div className="space-y-md">
      <PageHeader
        eyebrow="Development"
        title="Design System Showcase"
        description="Internal page for developers to view all components and their states"
      />

      {/* BUTTONS */}
      <Card>
        <CardHeader>
          <CardTitle>Buttons</CardTitle>
        </CardHeader>
        <CardContent className="space-y-md">
          <div className="flex flex-wrap gap-sm items-center">
            <span className="text-small text-text-muted">Variant:</span>
            {(['primary', 'secondary', 'ghost', 'danger', 'success', 'outline'] as const).map((variant) => (
              <Button
                key={variant}
                variant={buttonVariant === variant ? 'primary' : 'secondary'}
                size="sm"
                onClick={() => setButtonVariant(variant)}
              >
                {variant}
              </Button>
            ))}
          </div>
          
          <div className="flex flex-wrap gap-sm items-center">
            <span className="text-small text-text-muted">Size:</span>
            {(['sm', 'md', 'lg'] as const).map((size) => (
              <Button
                key={size}
                variant={buttonSize === size ? 'primary' : 'secondary'}
                size="sm"
                onClick={() => setButtonSize(size)}
              >
                {size}
              </Button>
            ))}
          </div>

          <div className="flex flex-wrap gap-sm items-center">
            <span className="text-small text-text-muted">States:</span>
            <Button variant={buttonVariant} size={buttonSize}>
              Default
            </Button>
            <Button variant={buttonVariant} size={buttonSize} disabled>
              Disabled
            </Button>
            <Button 
              variant={buttonVariant} 
              size={buttonSize} 
              isLoading={buttonLoading}
              onClick={() => {
                setButtonLoading(true);
                setTimeout(() => setButtonLoading(false), 2000);
              }}
            >
              Loading
            </Button>
          </div>
        </CardContent>
      </Card>

      {/* INPUTS */}
      <Card>
        <CardHeader>
          <CardTitle>Inputs</CardTitle>
        </CardHeader>
        <CardContent className="space-y-md">
          <Input label="Default Input" placeholder="Enter text..." />
          <Input label="With Error" error="This field is required" />
          <Input label="With Success" success="Value is valid" />
          <Input label="With Helper Text" helperText="This is additional information" />
          <Input label="Disabled" disabled placeholder="Cannot edit" />
          <Input label="Read Only" readOnly value="Cannot be modified" />
        </CardContent>
      </Card>

      {/* SELECTS */}
      <Card>
        <CardHeader>
          <CardTitle>Selects</CardTitle>
        </CardHeader>
        <CardContent className="space-y-md">
          <Select label="Default Select">
            <option value="">Select an option</option>
            <option value="1">Option 1</option>
            <option value="2">Option 2</option>
          </Select>
          <Select label="With Error" error="Please select an option">
            <option value="">Select an option</option>
            <option value="1">Option 1</option>
            <option value="2">Option 2</option>
          </Select>
        </CardContent>
      </Card>

      {/* CARDS */}
      <Card>
        <CardHeader>
          <CardTitle>Cards</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-md">
            {(['default', 'interactive', 'featured', 'warning', 'ai'] as const).map((variant) => (
              <Card key={variant} variant={variant} className="cursor-pointer">
                <CardHeader>
                  <CardTitle className="capitalize">{variant}</CardTitle>
                </CardHeader>
                <CardContent>
                  <p className="text-small text-text-muted">
                    Card variant demonstration
                  </p>
                </CardContent>
              </Card>
            ))}
          </div>
        </CardContent>
      </Card>

      {/* BADGES */}
      <Card>
        <CardHeader>
          <CardTitle>Badges</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex flex-wrap gap-sm">
            {(['default', 'success', 'warning', 'danger', 'info', 'secondary', 'outline'] as const).map((variant) => (
              <Badge key={variant} variant={variant}>
                {variant}
              </Badge>
            ))}
          </div>
          <div className="flex flex-wrap gap-sm mt-md">
            {(['default', 'success', 'warning', 'danger', 'info'] as const).map((variant) => (
              <Badge key={variant} variant={variant} dot>
                {variant} with dot
              </Badge>
            ))}
          </div>
        </CardContent>
      </Card>

      {/* STATES */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-md">
        <Card>
          <CardHeader>
            <CardTitle>Empty State</CardTitle>
          </CardHeader>
          <CardContent>
            <EmptyState
              icon={<Inbox className="w-8 h-8" />}
              title="No items found"
              description="There are no items to display at this time"
              action={<Button variant="primary">Add Item</Button>}
            />
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Error State</CardTitle>
          </CardHeader>
          <CardContent>
            <ErrorState
              icon={<AlertCircle className="w-8 h-8" />}
              title="Something went wrong"
              description="An error occurred while loading the data"
              error="Error: Connection failed"
              action={<Button variant="secondary">Retry</Button>}
            />
          </CardContent>
        </Card>
      </div>

      {/* LOADING */}
      <Card>
        <CardHeader>
          <CardTitle>Loading States</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex items-center gap-md">
            <div className="flex items-center gap-sm">
              <LoadingSpinner size="sm" color="cyan" />
              <span className="text-small text-text-muted">Small</span>
            </div>
            <div className="flex items-center gap-sm">
              <LoadingSpinner size="md" color="cyan" />
              <span className="text-small text-text-muted">Medium</span>
            </div>
            <div className="flex items-center gap-sm">
              <LoadingSpinner size="lg" color="cyan" />
              <span className="text-small text-text-muted">Large</span>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* REAL WORLD EXAMPLE */}
      <Card variant="featured">
        <CardHeader>
          <CardTitle>Real World Example</CardTitle>
        </CardHeader>
        <CardContent className="space-y-md">
          <PageHeader
            eyebrow="Example"
            title="Inventory Management"
            description="This demonstrates how components work together"
            actions={
              <Button variant="primary" className="gap-2">
                <Plus className="w-4 h-4" />
                Add Item
              </Button>
            }
          />
          
          <div className="flex flex-col md:flex-row gap-md">
            <div className="flex-1">
              <Input
                placeholder="Search products..."
                leftIcon={<Search className="w-4 h-4" />}
              />
            </div>
            <div className="flex gap-sm">
              <Button variant="secondary" className="gap-2">
                <Filter className="w-4 h-4" />
                Filter
              </Button>
              <Button variant="secondary" className="gap-2">
                <ArrowUpDown className="w-4 h-4" />
                Sort
              </Button>
            </div>
          </div>

          <div className="flex gap-sm">
            <Button variant="primary" className="gap-2">
              <Package className="w-4 h-4" />
              Products
            </Button>
            <Button variant="secondary" className="gap-2">
              <Package className="w-4 h-4" />
              Items
            </Button>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-4 gap-md">
            <Card>
              <CardContent className="p-lg">
                <div className="flex items-start justify-between">
                  <div>
                    <p className="text-small text-text-muted">Total Items</p>
                    <p className="text-metric font-bold text-text mt-1">1,284</p>
                  </div>
                  <div className="w-10 h-10 rounded-sm bg-cyan/10 flex items-center justify-center">
                    <Package className="w-5 h-5 text-cyan" />
                  </div>
                </div>
              </CardContent>
            </Card>
            <Card variant="warning">
              <CardContent className="p-lg">
                <div className="flex items-start justify-between">
                  <div>
                    <p className="text-small text-text-muted">Low Stock</p>
                    <p className="text-metric font-bold text-text mt-1">8</p>
                  </div>
                  <div className="w-10 h-10 rounded-sm bg-yellow/10 flex items-center justify-center">
                    <AlertCircle className="w-5 h-5 text-yellow" />
                  </div>
                </div>
              </CardContent>
            </Card>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}