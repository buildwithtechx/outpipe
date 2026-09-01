import { Radio } from 'lucide-react';

export function TunnelEmptyState() {
  return (
    <div className="rounded-2xl border border-dashed border-white/15 bg-white/[0.02] px-6 py-14 text-center">
      <Radio className="mx-auto size-5 text-indigo-200" />
      <h3 className="mt-4 text-lg font-semibold">No tunnels yet</h3>
      <p className="mx-auto mt-2 max-w-md text-sm leading-6 text-white/50">
        Create your first endpoint to connect a local app, webhook, or service
        to the public internet.
      </p>
    </div>
  );
}
