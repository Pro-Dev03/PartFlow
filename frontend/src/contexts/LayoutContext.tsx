import { createContext, useContext, useState, ReactNode, useCallback, useRef } from 'react';

interface LayoutContextType {
  fullWidth: boolean;
  setFullWidth: (fullWidth: boolean) => void;
}

const LayoutContext = createContext<LayoutContextType | undefined>(undefined);

export function LayoutProvider({ children }: { children: ReactNode }) {
  const [fullWidth, setFullWidthState] = useState(false);
  const isSettingWidth = useRef(false);

  // Use useCallback to prevent recreation of the function
  const setFullWidth = useCallback((value: boolean) => {
    if (isSettingWidth.current) return; // Prevent rapid re-setting
    isSettingWidth.current = true;
    setFullWidthState(value);
    setTimeout(() => {
      isSettingWidth.current = false;
    }, 100);
  }, []);

  return (
    <LayoutContext.Provider value={{ fullWidth, setFullWidth }}>
      {children}
    </LayoutContext.Provider>
  );
}

export function useLayout() {
  const context = useContext(LayoutContext);
  if (context === undefined) {
    throw new Error('useLayout must be used within a LayoutProvider');
  }
  return context;
}
