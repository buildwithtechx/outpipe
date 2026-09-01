import { useMutation, useQueryClient } from '@tanstack/react-query';
import { useAuthStore } from '#/stores/auth-store';
import { logout } from '../services/auth-service';

export function useLogout() {
  const clear = useAuthStore((state) => state.clear);
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: logout,
    onSuccess: () => {
      clear();
      queryClient.clear();
      window.location.assign('/');
    },
  });
}
