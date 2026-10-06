import { Check, Copy } from 'lucide-react';
import { useCopyText } from '#/hooks/use-copy-text';
import { Button } from './button';

export function CopyCommand({
  text,
  label = 'Copy command',
}: {
  text: string;
  label?: string;
}) {
  const { state, copy } = useCopyText();
  return (
    <div className="space-y-2">
      <div className="flex items-center gap-3 rounded-xl border border-border bg-background p-3">
        <code className="min-w-0 flex-1 overflow-x-auto whitespace-nowrap text-xs leading-6 text-foreground select-all">
          {text}
        </code>
        <Button
          type="button"
          variant="ghost"
          size="sm"
          aria-label={label}
          onClick={() => void copy(text)}
        >
          {state === 'copied' ? (
            <Check className="text-emerald-300" />
          ) : (
            <Copy />
          )}
          {state === 'copied' ? 'Copied' : 'Copy'}
        </Button>
      </div>
      <p aria-live="polite" className="text-xs text-muted-foreground">
        {state === 'error'
          ? 'Could not copy. Select the command and copy it manually.'
          : state === 'copied'
            ? 'Copied to clipboard.'
            : ''}
      </p>
    </div>
  );
}
