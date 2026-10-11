import { Trash2 } from 'lucide-react';
import {
  type KeyboardEvent,
  type PointerEvent,
  useLayoutEffect,
  useRef,
  useState,
} from 'react';
import { Button } from '#/components/ui/button';
import { cn } from '#/lib/utils';
import { createHoldToDeleteController } from './hold-to-delete-controller';

export interface HoldToDeleteProps {
  onConfirm: () => void;
  disabled?: boolean;
  loading?: boolean;
  label?: string;
  className?: string;
}

export function HoldToDelete({
  onConfirm,
  disabled = false,
  loading = false,
  label = 'Hold to delete',
  className,
}: HoldToDeleteProps) {
  const [progress, setProgress] = useState(0);
  const latest = useRef({ onConfirm, disabled, loading });
  const mounted = useRef(true);
  const controller = useRef<ReturnType<
    typeof createHoldToDeleteController
  > | null>(null);

  useLayoutEffect(() => {
    latest.current = { onConfirm, disabled, loading };
  }, [onConfirm, disabled, loading]);

  useLayoutEffect(() => {
    mounted.current = true;
    const hold = createHoldToDeleteController({
      onProgress: (val) => {
        if (mounted.current) setProgress(val);
      },
      onConfirm: () => {
        const current = latest.current;
        if (mounted.current && !current.disabled && !current.loading) {
          current.onConfirm();
        }
      },
    });
    controller.current = hold;

    const cancelPointer = (e: globalThis.PointerEvent) => {
      hold.cancel(`pointer:${e.pointerId}`);
    };
    const cancelOnHide = () => {
      if (document.visibilityState !== 'visible') hold.cancel();
    };
    const cancelOnEscape = (e: globalThis.KeyboardEvent) => {
      if (e.key === 'Escape') hold.cancel();
    };
    const cancelOnBlur = () => hold.cancel();

    window.addEventListener('pointerup', cancelPointer);
    window.addEventListener('pointercancel', cancelPointer);
    window.addEventListener('blur', cancelOnBlur);
    window.addEventListener('keydown', cancelOnEscape);
    document.addEventListener('visibilitychange', cancelOnHide);

    return () => {
      mounted.current = false;
      hold.cancel();
      controller.current = null;
      window.removeEventListener('pointerup', cancelPointer);
      window.removeEventListener('pointercancel', cancelPointer);
      window.removeEventListener('blur', cancelOnBlur);
      window.removeEventListener('keydown', cancelOnEscape);
      document.removeEventListener('visibilitychange', cancelOnHide);
    };
  }, []);

  const handlePointerDown = (e: PointerEvent<HTMLButtonElement>) => {
    if (disabled || loading || e.button !== 0) return;
    controller.current?.start(`pointer:${e.pointerId}`);
  };

  const handlePointerUp = (e: PointerEvent<HTMLButtonElement>) => {
    controller.current?.cancel(`pointer:${e.pointerId}`);
  };

  const handleKeyDown = (e: KeyboardEvent<HTMLButtonElement>) => {
    if (disabled || loading) return;
    if (e.key === ' ' || e.key === 'Enter') {
      if (!e.repeat) {
        controller.current?.start('keyboard');
      }
    }
  };

  const handleKeyUp = (e: KeyboardEvent<HTMLButtonElement>) => {
    if (e.key === ' ' || e.key === 'Enter') {
      controller.current?.cancel('keyboard');
    }
  };

  const isHolding = progress > 0;

  return (
    <Button
      type="button"
      variant="outline"
      size="sm"
      disabled={disabled || loading}
      onPointerDown={handlePointerDown}
      onPointerUp={handlePointerUp}
      onKeyDown={handleKeyDown}
      onKeyUp={handleKeyUp}
      aria-label={`${label}. Press and hold to confirm.`}
      className={cn(
        'relative overflow-hidden text-xs border-rose-500/20 text-rose-300 hover:border-rose-500/40 hover:bg-rose-500/10 select-none transition-colors',
        isHolding && 'border-rose-500/60 bg-rose-500/15',
        className,
      )}
    >
      <div
        className="absolute inset-y-0 left-0 bg-rose-500/30 transition-[width] ease-linear pointer-events-none"
        style={{ width: `${Math.round(progress * 100)}%` }}
        aria-hidden="true"
      />
      <Trash2 className="size-3.5 shrink-0 z-10" />
      <span className="z-10 font-sans">
        {loading
          ? 'Deleting...'
          : isHolding
            ? `${Math.round(progress * 100)}%`
            : label}
      </span>
      <span className="sr-only" role="status" aria-live="polite">
        {isHolding
          ? `Holding to delete: ${Math.round(progress * 100)} percent`
          : ''}
      </span>
    </Button>
  );
}
