import { LoaderCircle } from 'lucide-react';
import type { FormEvent } from 'react';
import { Button } from '#/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '#/components/ui/card';
import { Input } from '#/components/ui/input';
import { Label } from '#/components/ui/label';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '#/components/ui/select';
import type { CreateTunnelRequest, TunnelProtocol } from '#/interfaces/tunnel';

const protocols: TunnelProtocol[] = ['http', 'https', 'tcp', 'udp'];

type TunnelCreateFormProps = {
  request: CreateTunnelRequest;
  isPending: boolean;
  error: string | null;
  onChange: (value: CreateTunnelRequest) => void;
  onSubmit: () => void;
};

export function TunnelCreateForm({
  request,
  isPending,
  error,
  onChange,
  onSubmit,
}: TunnelCreateFormProps) {
  function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    onSubmit();
  }

  return (
    <Card className="mt-8 gap-0 rounded-2xl border-indigo-300/25 bg-indigo-300/5 py-0 shadow-none">
      <CardHeader className="px-5 pt-5 pb-0 sm:px-6 sm:pt-6">
        <CardTitle>Create a tunnel</CardTitle>
        <p className="text-sm text-white/50">
          The public hostname is assigned automatically.
        </p>
      </CardHeader>
      <CardContent className="px-5 pt-5 pb-5 sm:px-6 sm:pt-6 sm:pb-6">
        <form onSubmit={submit} className="grid gap-5">
          <div className="grid gap-4 sm:grid-cols-2">
            <div className="grid gap-2">
              <Label htmlFor="tunnel-name" className="text-white/70">
                Tunnel name
              </Label>
              <Input
                id="tunnel-name"
                required
                value={request.name}
                onChange={(event) =>
                  onChange({ ...request, name: event.target.value })
                }
                placeholder="checkout-api"
                className="h-11 rounded-xl border-white/10 bg-black/40 text-white placeholder:text-white/25"
              />
            </div>
            <div className="grid gap-2">
              <Label className="text-white/70">Protocol</Label>
              <Select
                value={request.protocol}
                onValueChange={(protocol: TunnelProtocol) =>
                  onChange({ ...request, protocol })
                }
              >
                <SelectTrigger className="h-11 w-full rounded-xl border-white/10 bg-black/40 text-white">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {protocols.map((protocol) => (
                    <SelectItem key={protocol} value={protocol}>
                      {protocol.toUpperCase()}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <div className="grid gap-2">
              <Label htmlFor="target-host" className="text-white/70">
                Target host
              </Label>
              <Input
                id="target-host"
                required
                value={request.targetHost}
                onChange={(event) =>
                  onChange({ ...request, targetHost: event.target.value })
                }
                placeholder="localhost"
                className="h-11 rounded-xl border-white/10 bg-black/40 text-white placeholder:text-white/25"
              />
            </div>
            <div className="grid gap-2">
              <Label htmlFor="target-port" className="text-white/70">
                Target port
              </Label>
              <Input
                id="target-port"
                required
                type="number"
                min="1"
                max="65535"
                value={request.targetPort}
                onChange={(event) =>
                  onChange({
                    ...request,
                    targetPort: Number(event.target.value),
                  })
                }
                className="h-11 rounded-xl border-white/10 bg-black/40 text-white"
              />
            </div>
          </div>
          {error && <p className="text-sm text-rose-200">{error}</p>}
          <Button
            type="submit"
            size="lg"
            disabled={isPending}
            className="w-fit rounded-full bg-white px-5 text-black hover:bg-indigo-100"
          >
            {isPending && <LoaderCircle className="animate-spin" />}
            {isPending ? 'Creating tunnel…' : 'Create tunnel'}
          </Button>
        </form>
      </CardContent>
    </Card>
  );
}
