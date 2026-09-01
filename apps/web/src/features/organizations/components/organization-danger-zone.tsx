import { useState } from 'react';
import { Button } from '#/components/ui/button';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '#/components/ui/select';
import { useAccountActions } from '#/features/account/hooks/use-account-actions';
import type { OrganizationMember } from '#/interfaces/organization';

export function OrganizationDangerZone({
  organizationId,
  members,
}: {
  organizationId: string;
  members: OrganizationMember[];
}) {
  const actions = useAccountActions(organizationId);
  const [newOwnerId, setNewOwnerId] = useState('');
  return (
    <section className="mt-8 rounded-2xl border border-rose-300/20 bg-rose-300/[0.04] p-6">
      <h2 className="text-lg font-medium text-rose-100">
        Ownership and account safety
      </h2>
      <p className="mt-2 text-sm text-white/50">
        Transfer this workspace before deleting your account.
      </p>
      <div className="mt-5 flex flex-wrap items-center gap-3">
        <Select value={newOwnerId} onValueChange={setNewOwnerId}>
          <SelectTrigger className="w-64 border-white/10 bg-black text-white">
            <SelectValue placeholder="Choose a new owner" />
          </SelectTrigger>
          <SelectContent>
            {members
              .filter((member) => member.role !== 'owner')
              .map((member) => (
                <SelectItem key={member.userId} value={member.userId}>
                  {member.userId}
                </SelectItem>
              ))}
          </SelectContent>
        </Select>
        <Button
          type="button"
          variant="outline"
          disabled={!newOwnerId || actions.transfer.isPending}
          onClick={() => {
            if (window.confirm('Transfer ownership of this workspace?'))
              actions.transfer.mutate(newOwnerId);
          }}
        >
          Transfer ownership
        </Button>
      </div>
      <Button
        type="button"
        variant="destructive"
        className="mt-8"
        disabled={actions.remove.isPending}
        onClick={() => {
          if (window.confirm('Delete your account? This cannot be undone.'))
            actions.remove.mutate();
        }}
      >
        Delete account
      </Button>
    </section>
  );
}
