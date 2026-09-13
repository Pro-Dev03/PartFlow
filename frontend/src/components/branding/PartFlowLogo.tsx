import logoUrl from '../../assets/partflow-logo.png';

interface PartFlowLogoProps {
  size?: number;
  className?: string;
  priority?: boolean;
}

export function PartFlowLogo({ size = 44, className = '', priority = false }: PartFlowLogoProps) {
  return (
    <img
      src={logoUrl}
      alt="PartFlow"
      width={size}
      height={size}
      loading={priority ? 'eager' : 'lazy'}
      decoding="async"
      className={`block object-contain ${className}`}
      style={{ width: size, height: size }}
    />
  );
}
