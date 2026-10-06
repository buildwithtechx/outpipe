import { useState } from 'react';
import { Button } from '#/components/ui/button';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '#/components/ui/dialog';
import { Label } from '#/components/ui/label';
import { Textarea } from '#/components/ui/textarea';
import { useSubmitFeedback } from '#/hooks/use-submit-feedback';
import type { UptimeIncident } from '#/interfaces/uptime';
import { useUptimeMutations } from '../hooks/use-uptime';

export function IncidentUpdateForm({ incident }: { incident: UptimeIncident }) {
  const [open, setOpen] = useState(false);
  const [status, setStatus] = useState<UptimeIncident['status']>(
    incident.status,
  );
  const [message, setMessage] = useState('');
  const { addIncidentUpdate } = useUptimeMutations(incident.organizationId);
  const { error, submit } = useSubmitFeedback(open);
  return (
    <>
      <Button
        variant="outline"
        size="sm"
        onClick={() => {
          setStatus(incident.status);
          setOpen(true);
        }}
      >
        Update incident
      </Button>
      <Dialog
        open={open}
        onOpenChange={(value) => {
          if (!addIncidentUpdate.isPending) setOpen(value);
        }}
      >
        <DialogContent showCloseButton={!addIncidentUpdate.isPending}>
          <DialogHeader>
            <DialogTitle>Update incident</DialogTitle>
            <DialogDescription>{incident.title}</DialogDescription>
          </DialogHeader>
          <form
            onSubmit={async (event) => {
              event.preventDefault();
              if (!message.trim()) return;
              const saved = await submit(() =>
                addIncidentUpdate.mutateAsync({
                  incidentId: incident.id,
                  input: { status, message: message.trim() },
                }),
              );
              if (saved) {
                setMessage('');
                setOpen(false);
              }
            }}
          >
            <fieldset
              disabled={addIncidentUpdate.isPending}
              className="space-y-4"
            >
              {error && (
                <p role="alert" className="text-sm text-rose-300">
                  Could not save this update. Your message has been kept.
                </p>
              )}
              <div className="space-y-2">
                <Label htmlFor="incident-update-status">Status</Label>
                <select
                  id="incident-update-status"
                  value={status}
                  onChange={(event) =>
                    setStatus(event.target.value as UptimeIncident['status'])
                  }
                  className="min-h-11 w-full rounded-lg border border-border bg-background px-3"
                >
                  <option value="investigating">Investigating</option>
                  <option value="identified">Identified</option>
                  <option value="monitoring">Monitoring</option>
                  <option value="resolved">Resolved</option>
                </select>
              </div>
              <div className="space-y-2">
                <Label htmlFor="incident-update-message">Public update</Label>
                <Textarea
                  id="incident-update-message"
                  required
                  maxLength={4096}
                  rows={4}
                  value={message}
                  onChange={(event) => setMessage(event.target.value)}
                />
                <p className="text-xs text-muted-foreground">
                  Updates appear on your published status page.
                </p>
              </div>
              <Button type="submit">
                {addIncidentUpdate.isPending ? 'Saving...' : 'Publish update'}
              </Button>
            </fieldset>
          </form>
        </DialogContent>
      </Dialog>
    </>
  );
}
