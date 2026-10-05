import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
  createEnvironment,
  createProject,
  deleteProject,
  deleteSecret,
  getEnvironments,
  getProjects,
  getSecrets,
  getShares,
  restoreSecret,
  revokeShare,
  setSecret,
} from '../services/secret-service';

export function useProjects(organizationId?: string) {
  return useQuery({
    queryKey: ['secret-projects', organizationId],
    queryFn: () => getProjects(organizationId as string),
    enabled: Boolean(organizationId),
  });
}

export function useEnvironments(organizationId?: string, projectId?: string) {
  return useQuery({
    queryKey: ['secret-environments', organizationId, projectId],
    queryFn: () =>
      getEnvironments(organizationId as string, projectId as string),
    enabled: Boolean(organizationId && projectId),
  });
}

export function useSecretsList(
  organizationId?: string,
  projectId?: string,
  environmentId?: string,
  reveal = false,
) {
  return useQuery({
    queryKey: ['secrets', organizationId, projectId, environmentId, reveal],
    queryFn: () =>
      getSecrets(
        organizationId as string,
        projectId as string,
        environmentId as string,
        reveal,
      ),
    enabled: Boolean(organizationId && projectId && environmentId),
  });
}

export function useSecretShares(organizationId?: string) {
  return useQuery({
    queryKey: ['secret-shares', organizationId],
    queryFn: () => getShares(organizationId as string),
    enabled: Boolean(organizationId),
  });
}

export function useSecretMutations(organizationId: string) {
  const queryClient = useQueryClient();

  const addProject = useMutation({
    mutationFn: (input: { slug: string; name: string; description?: string }) =>
      createProject(organizationId, input),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: ['secret-projects', organizationId],
      });
    },
  });

  const removeProject = useMutation({
    mutationFn: (projectId: string) => deleteProject(organizationId, projectId),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: ['secret-projects', organizationId],
      });
    },
  });

  const addEnvironment = useMutation({
    mutationFn: ({
      projectId,
      input,
    }: {
      projectId: string;
      input: { slug: string; name: string };
    }) => createEnvironment(organizationId, projectId, input),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: ['secret-environments', organizationId],
      });
    },
  });

  const saveSecret = useMutation({
    mutationFn: ({
      projectId,
      environmentId,
      input,
    }: {
      projectId: string;
      environmentId: string;
      input: { key: string; value: string; comment?: string };
    }) => setSecret(organizationId, projectId, environmentId, input),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: ['secrets', organizationId],
      });
    },
  });

  const removeSecret = useMutation({
    mutationFn: (secretId: string) => deleteSecret(organizationId, secretId),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: ['secrets', organizationId],
      });
    },
  });

  const recoverSecret = useMutation({
    mutationFn: (secretId: string) => restoreSecret(organizationId, secretId),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: ['secrets', organizationId],
      });
    },
  });

  const removeShare = useMutation({
    mutationFn: (shareId: string) => revokeShare(organizationId, shareId),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: ['secret-shares', organizationId],
      });
    },
  });

  return {
    addProject,
    removeProject,
    addEnvironment,
    saveSecret,
    removeSecret,
    recoverSecret,
    removeShare,
  };
}
