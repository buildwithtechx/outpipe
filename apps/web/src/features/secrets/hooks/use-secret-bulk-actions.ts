import { completeShareUrl, encryptShare } from '@outpipe/shared/crypto';
import { useState } from 'react';
import type { SecretItem } from '#/interfaces/secret';
import {
  batchDeleteSecrets,
  batchMoveSecrets,
  createSecretShare,
  rollbackSecretVersion,
} from '../services/secret-service';

export function useSecretBulkActions({
  organizationId,
  projectId,
  environmentId,
  secrets,
  onRefresh,
}: {
  organizationId?: string;
  projectId: string;
  environmentId: string;
  secrets: SecretItem[];
  onRefresh: () => void;
}) {
  const [selectedSecretIds, setSelectedSecretIds] = useState<string[]>([]);
  const [historySecret, setHistorySecret] = useState<SecretItem | null>(null);
  const [bulkAction, setBulkAction] = useState<
    'move' | 'delete' | 'share' | null
  >(null);
  const [isBulkOperating, setIsBulkOperating] = useState(false);

  const handleToggleSelect = (id: string, checked: boolean) => {
    setSelectedSecretIds((prev) =>
      checked ? [...prev, id] : prev.filter((item) => item !== id),
    );
  };

  const handleToggleSelectAll = (checked: boolean) => {
    setSelectedSecretIds(checked ? secrets.map((s) => s.id) : []);
  };

  const handleConfirmBulkMove = async (
    targetEnvId: string,
    conflictMode: 'skip' | 'overwrite',
  ) => {
    if (!organizationId || !projectId || !environmentId) return;
    setIsBulkOperating(true);
    try {
      await batchMoveSecrets(
        organizationId,
        projectId,
        environmentId,
        targetEnvId,
        selectedSecretIds,
        conflictMode,
      );
      setSelectedSecretIds([]);
      setBulkAction(null);
      onRefresh();
    } finally {
      setIsBulkOperating(false);
    }
  };

  const handleConfirmBulkDelete = async () => {
    if (!organizationId) return;
    setIsBulkOperating(true);
    try {
      await batchDeleteSecrets(organizationId, selectedSecretIds);
      setSelectedSecretIds([]);
      setBulkAction(null);
      onRefresh();
    } finally {
      setIsBulkOperating(false);
    }
  };

  const handleConfirmBulkShare = async (
    _expiry: string,
    maxViews: number,
  ): Promise<string | null> => {
    if (!organizationId) return null;
    setIsBulkOperating(true);
    try {
      const selectedItems = secrets.filter((s) =>
        selectedSecretIds.includes(s.id),
      );
      const payload = JSON.stringify(
        selectedItems.map((s) => ({ key: s.key, value: s.value || '' })),
      );
      const encrypted = await encryptShare({ type: 'text', text: payload });
      const share = await createSecretShare(organizationId, {
        encryptedPayload: encrypted.ciphertext,
        expiresAt: new Date(Date.now() + 7 * 86400000).toISOString(),
        maxViews,
      });
      return completeShareUrl(window.location.origin, share.id, encrypted.key);
    } finally {
      setIsBulkOperating(false);
    }
  };

  const handleRollback = async (version: number) => {
    if (!organizationId || !historySecret) return;
    await rollbackSecretVersion(organizationId, historySecret.id, version);
    onRefresh();
    setHistorySecret(null);
  };

  return {
    selectedSecretIds,
    historySecret,
    bulkAction,
    isBulkOperating,
    setHistorySecret,
    setBulkAction,
    handleToggleSelect,
    handleToggleSelectAll,
    handleConfirmBulkMove,
    handleConfirmBulkDelete,
    handleConfirmBulkShare,
    handleRollback,
  };
}
