import { Modal } from '../../../components/ui/modal';
import { Button } from '../../../components/ui/button';
import { Input } from '../../../components/ui/input';
import { getButtonSize } from '../../../config/button-sizes';
import { Product } from '../types/inventory.types';

interface InventoryModalsProps {
  isViewModalOpen: boolean;
  setIsViewModalOpen: (open: boolean) => void;
  isEditModalOpen: boolean;
  setIsEditModalOpen: (open: boolean) => void;
  selectedProduct: Product | null;
  setSelectedProduct: (product: Product | null) => void;
  onSaveProduct: (productData: Product) => void;
}

export function InventoryModals({
  isViewModalOpen,
  setIsViewModalOpen,
  isEditModalOpen,
  setIsEditModalOpen,
  selectedProduct,
  setSelectedProduct,
  onSaveProduct,
}: InventoryModalsProps) {
  return (
    <>
      {/* View Product Modal */}
      <Modal
        isOpen={isViewModalOpen}
        onClose={() => setIsViewModalOpen(false)}
        title="تفاصيل المنتج"
      >
        {selectedProduct && (
          <div className="space-y-md">
            <div>
              <label className="text-small font-medium text-text mb-sm block">الاسم</label>
              <Input value={selectedProduct.name} disabled />
            </div>
            <div>
              <label className="text-small font-medium text-text mb-sm block">SKU</label>
              <Input value={selectedProduct.sku} disabled />
            </div>
            <div>
              <label className="text-small font-medium text-text mb-sm block">السعر</label>
              <Input value={`₪${selectedProduct.sellingPrice}`} disabled />
            </div>
            <div>
              <label className="text-small font-medium text-text mb-sm block">المخزون</label>
              <Input value={selectedProduct.stock} disabled />
            </div>
            <div className="flex gap-sm justify-end">
              <Button variant="secondary" size={getButtonSize('inventory', 'modalAction')} onClick={() => setIsViewModalOpen(false)}>
                إغلاق
              </Button>
            </div>
          </div>
        )}
      </Modal>

      {/* Edit Product Modal */}
      <Modal
        isOpen={isEditModalOpen}
        onClose={() => {
          setIsEditModalOpen(false);
          setSelectedProduct(null);
        }}
        title={selectedProduct ? "تعديل المنتج" : "إضافة منتج جديد"}
      >
        {selectedProduct ? (
          <div className="space-y-md">
            <div>
              <label className="text-small font-medium text-text mb-sm block">الاسم</label>
              <Input 
                value={selectedProduct.name}
                onChange={(e) => setSelectedProduct({ ...selectedProduct, name: e.target.value })}
              />
            </div>
            <div>
              <label className="text-small font-medium text-text mb-sm block">SKU</label>
              <Input 
                value={selectedProduct.sku}
                onChange={(e) => setSelectedProduct({ ...selectedProduct, sku: e.target.value })}
              />
            </div>
            <div>
              <label className="text-small font-medium text-text mb-sm block">السعر</label>
              <Input 
                type="number"
                value={selectedProduct.sellingPrice}
                onChange={(e) => setSelectedProduct({ ...selectedProduct, sellingPrice: Number(e.target.value) })}
              />
            </div>
            <div>
              <label className="text-small font-medium text-text mb-sm block">المخزون</label>
              <Input 
                type="number"
                value={selectedProduct.stock}
                onChange={(e) => setSelectedProduct({ ...selectedProduct, stock: Number(e.target.value) })}
              />
            </div>
            <div className="flex gap-sm justify-end">
              <Button variant="secondary" size={getButtonSize('inventory', 'modalAction')} onClick={() => {
                setIsEditModalOpen(false);
                setSelectedProduct(null);
              }}>
                إلغاء
              </Button>
              <Button variant="primary" size={getButtonSize('inventory', 'modalAction')} onClick={() => onSaveProduct(selectedProduct)}>
                حفظ
              </Button>
            </div>
          </div>
        ) : (
          <div className="space-y-md">
            <div>
              <label className="text-small font-medium text-text mb-sm block">الاسم</label>
              <Input 
                placeholder="أدخل اسم المنتج"
                onChange={(e) => setSelectedProduct({ name: e.target.value } as Product)}
              />
            </div>
            <div>
              <label className="text-small font-medium text-text mb-sm block">SKU</label>
              <Input 
                placeholder="أدخل SKU"
                onChange={(e) => setSelectedProduct((prev: Product | null) => ({ ...prev, sku: e.target.value } as Product))}
              />
            </div>
            <div>
              <label className="text-small font-medium text-text mb-sm block">السعر</label>
              <Input 
                type="number"
                placeholder="أدخل السعر"
                onChange={(e) => setSelectedProduct((prev: Product | null) => ({ ...prev, sellingPrice: Number(e.target.value) } as Product))}
              />
            </div>
            <div>
              <label className="text-small font-medium text-text mb-sm block">المخزون</label>
              <Input 
                type="number"
                placeholder="أدخل الكمية"
                onChange={(e) => setSelectedProduct((prev: Product | null) => ({ ...prev, stock: Number(e.target.value) } as Product))}
              />
            </div>
            <div className="flex gap-sm justify-end">
              <Button variant="secondary" size={getButtonSize('inventory', 'modalAction')} onClick={() => {
                setIsEditModalOpen(false);
                setSelectedProduct(null);
              }}>
                إلغاء
              </Button>
              <Button variant="primary" size={getButtonSize('inventory', 'modalAction')} onClick={() => onSaveProduct(selectedProduct as Product)}>
                حفظ
              </Button>
            </div>
          </div>
        )}
      </Modal>
    </>
  );
}