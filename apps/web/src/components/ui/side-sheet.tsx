'use client';

import { XIcon } from 'lucide-react';
import { Dialog as DialogPrimitive } from 'radix-ui';
import type * as React from 'react';
import { Button } from '#/components/ui/button';
import { cn } from '#/lib/utils';

export interface SideSheetProps {
  open: boolean;
  onClose: () => void;
  title: string;
  description?: string;
  children: React.ReactNode;
  footer?: React.ReactNode;
  closeDisabled?: boolean;
  className?: string;
}

export function SideSheet({
  open,
  onClose,
  title,
  description,
  children,
  footer,
  closeDisabled = false,
  className,
}: SideSheetProps) {
  return (
    <DialogPrimitive.Root
      open={open}
      onOpenChange={(next) => {
        if (!next && !closeDisabled) onClose();
      }}
    >
      <DialogPrimitive.Portal>
        <DialogPrimitive.Overlay className="fixed inset-0 z-50 bg-black/60 backdrop-blur-xs data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0" />
        <DialogPrimitive.Content
          aria-describedby={description ? undefined : undefined}
          className={cn(
            'fixed inset-y-0 right-0 z-50 flex h-full w-full max-w-md flex-col border-l border-white/10 bg-card p-0 text-white shadow-2xl transition ease-in-out data-[state=closed]:duration-200 data-[state=open]:duration-300 data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:slide-out-to-right data-[state=open]:slide-in-from-right',
            className,
          )}
        >
          <div className="flex items-center justify-between border-b border-white/10 px-5 py-4">
            <div>
              <DialogPrimitive.Title className="text-sm font-semibold text-white">
                {title}
              </DialogPrimitive.Title>
              {description && (
                <DialogPrimitive.Description className="mt-1 text-xs text-white/60">
                  {description}
                </DialogPrimitive.Description>
              )}
            </div>
            <Button
              type="button"
              variant="ghost"
              size="icon-sm"
              disabled={closeDisabled}
              onClick={onClose}
              className="text-white/60 hover:text-white"
            >
              <XIcon className="size-4" />
              <span className="sr-only">Close</span>
            </Button>
          </div>

          <div className="flex-1 overflow-y-auto px-5 py-4">{children}</div>

          {footer && (
            <div className="border-t border-white/10 bg-white/2 px-5 py-3">
              {footer}
            </div>
          )}
        </DialogPrimitive.Content>
      </DialogPrimitive.Portal>
    </DialogPrimitive.Root>
  );
}
