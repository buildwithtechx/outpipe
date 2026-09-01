import { useState } from 'react';
import { Button } from '#/components/ui/button';
import { Input } from '#/components/ui/input';
import { Label } from '#/components/ui/label';
import { Textarea } from '#/components/ui/textarea';
import type { Tunnel } from '#/interfaces/tunnel';
import { useTunnelConfigurationMutation } from '../hooks/use-tunnel-configuration-mutation';

export function TunnelConfigurationForm({ tunnel }: { tunnel: Tunnel }) {
  const mutation = useTunnelConfigurationMutation(tunnel.id);
  const [expiresAt, setExpiresAt] = useState(toInputDate(tunnel.expiresAt));
  const [accessPolicy, setAccessPolicy] = useState(tunnel.accessPolicy || '{}');
  const save = () => {
    try {
      const parsed = JSON.parse(accessPolicy);
      if (!parsed || Array.isArray(parsed) || typeof parsed !== 'object')
        throw new Error();
      mutation.mutate({
        accessPolicy: JSON.stringify(parsed),
        expiresAt: expiresAt ? new Date(expiresAt).toISOString() : undefined,
      });
    } catch {
      window.alert('Access policy must be a JSON object.');
    }
  };
  return (
    <section className="mt-6 rounded-2xl border border-white/10 bg-white/[0.025] p-5 sm:p-6">
      <div>
        <h2 className="text-lg font-medium">Tunnel controls</h2>
        <p className="mt-1 text-sm text-white/45">
          Set a lifetime and the policy applied to incoming connections.
        </p>
      </div>
      <div className="mt-5 grid gap-5 lg:grid-cols-2">
        <div className="grid gap-2">
          <Label htmlFor="tunnel-expires" className="text-white/70">
            Expires at
          </Label>
          <Input
            id="tunnel-expires"
            type="datetime-local"
            value={expiresAt}
            onChange={(event) => setExpiresAt(event.target.value)}
            className="border-white/10 bg-black text-white"
          />
          <p className="text-xs text-white/40">
            Leave empty to keep this tunnel running without an expiry.
          </p>
        </div>
        <div className="grid gap-2">
          <Label htmlFor="tunnel-policy" className="text-white/70">
            Access policy (JSON)
          </Label>
          <Textarea
            id="tunnel-policy"
            value={accessPolicy}
            onChange={(event) => setAccessPolicy(event.target.value)}
            rows={4}
            className="border-white/10 bg-black font-mono text-sm text-white"
          />
        </div>
      </div>
      <Button
        type="button"
        className="mt-5"
        onClick={save}
        disabled={mutation.isPending}
      >
        {mutation.isPending ? 'Saving…' : 'Save tunnel controls'}
      </Button>
      {mutation.isError && (
        <p className="mt-3 text-sm text-rose-200">
          Could not update this tunnel.
        </p>
      )}
      {mutation.isSuccess && (
        <p className="mt-3 text-sm text-emerald-200">
          Tunnel controls updated.
        </p>
      )}
    </section>
  );
}

function toInputDate(value?: string) {
  if (!value) return '';
  const date = new Date(value);
  const offset = date.getTimezoneOffset();
  return new Date(date.getTime() - offset * 60_000).toISOString().slice(0, 16);
}
