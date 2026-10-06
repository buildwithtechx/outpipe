import { CopyCommand } from '#/components/ui/copy-command';

export function TunnelDetailItem({
  label,
  value,
  copyValue,
}: {
  label: string;
  value: string;
  copyValue?: string;
}) {
  return (
    <div className="min-w-0 space-y-2">
      <p className="text-xs font-medium uppercase tracking-wider text-muted-foreground">
        {label}
      </p>
      {copyValue ? (
        <CopyCommand text={copyValue} label={`Copy ${label.toLowerCase()}`} />
      ) : (
        <p className="break-words font-mono text-sm">{value}</p>
      )}
    </div>
  );
}
