import type { ReactNode } from 'react';
import { cn } from '#/lib/utils';

type MarketingContainerProps = {
  children: ReactNode;
  className?: string;
};

export function MarketingContainer({
  children,
  className,
}: MarketingContainerProps) {
  return (
    <div className={cn('mx-auto w-full max-w-7xl px-6 lg:px-8', className)}>
      {children}
    </div>
  );
}
